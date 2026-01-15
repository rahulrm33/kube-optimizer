package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/scaleops/k8s-optimizer/internal/config"
	"github.com/scaleops/k8s-optimizer/internal/database"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned"
)

func main() {
	// Parse command line flags
	kubeconfig := flag.String("kubeconfig", "", "Path to kubeconfig file")
	kubecontext := flag.String("context", "", "Kubernetes context to use")
	namespace := flag.String("namespace", "", "Namespace to collect metrics from (empty for all)")
	once := flag.Bool("once", false, "Run once and exit (default: continuous collection)")
	interval := flag.Duration("interval", 5*time.Minute, "Collection interval")
	flag.Parse()

	// Load application config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	db, err := database.NewDB(cfg.Database.ConnectionString())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize schema
	if err := db.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	// Build kubeconfig
	kubeconfigPath := *kubeconfig
	if kubeconfigPath == "" {
		if home := os.Getenv("HOME"); home != "" {
			kubeconfigPath = filepath.Join(home, ".kube", "config")
		}
	}

	// Build config from kubeconfig file
	configOverrides := &clientcmd.ConfigOverrides{}
	if *kubecontext != "" {
		configOverrides.CurrentContext = *kubecontext
	}

	kubeConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		configOverrides,
	).ClientConfig()
	if err != nil {
		log.Fatalf("Failed to build kubeconfig: %v", err)
	}

	// Create Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Create Metrics clientset
	metricsClient, err := metricsv1beta1.NewForConfig(kubeConfig)
	if err != nil {
		log.Fatalf("Failed to create metrics client: %v", err)
	}

	log.Printf("Connected to Kubernetes cluster")
	if *kubecontext != "" {
		log.Printf("Using context: %s", *kubecontext)
	}

	collector := &Collector{
		db:            db,
		clientset:     clientset,
		metricsClient: metricsClient,
		config:        cfg,
		namespace:     *namespace,
	}

	if *once {
		log.Println("Running single collection...")
		if err := collector.Collect(context.Background()); err != nil {
			log.Fatalf("Collection failed: %v", err)
		}
		log.Println("Collection complete!")
		return
	}

	// Continuous collection
	log.Printf("Starting continuous collection every %v", *interval)
	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	// Run immediately
	if err := collector.Collect(context.Background()); err != nil {
		log.Printf("Collection error: %v", err)
	}

	for range ticker.C {
		if err := collector.Collect(context.Background()); err != nil {
			log.Printf("Collection error: %v", err)
		}
	}
}

type Collector struct {
	db            *database.DB
	clientset     *kubernetes.Clientset
	metricsClient *metricsv1beta1.Clientset
	config        *config.Config
	namespace     string
}

func (c *Collector) Collect(ctx context.Context) error {
	log.Println("Starting metrics collection...")

	// Get all pods
	listOptions := metav1.ListOptions{}
	var pods *corev1.PodList
	var err error

	if c.namespace != "" {
		pods, err = c.clientset.CoreV1().Pods(c.namespace).List(ctx, listOptions)
	} else {
		pods, err = c.clientset.CoreV1().Pods("").List(ctx, listOptions)
	}

	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}

	log.Printf("Found %d pods", len(pods.Items))

	// Get pod metrics
	var podMetricsList interface{ Items() interface{} }
	
	if c.namespace != "" {
		metrics, err := c.metricsClient.MetricsV1beta1().PodMetricses(c.namespace).List(ctx, listOptions)
		if err != nil {
			log.Printf("Warning: Could not get metrics (metrics-server might not be installed): %v", err)
		} else {
			log.Printf("Got metrics for %d pods", len(metrics.Items))
			_ = podMetricsList
		}
	} else {
		metrics, err := c.metricsClient.MetricsV1beta1().PodMetricses("").List(ctx, listOptions)
		if err != nil {
			log.Printf("Warning: Could not get metrics (metrics-server might not be installed): %v", err)
		} else {
			log.Printf("Got metrics for %d pods", len(metrics.Items))
		}
	}

	// Store each pod and its containers
	for _, pod := range pods.Items {
		// Skip pods that are not running
		if pod.Status.Phase != corev1.PodRunning {
			continue
		}

		// Skip system pods (optional)
		if pod.Namespace == "kube-system" && !isImportantSystemPod(pod.Name) {
			continue
		}

		if err := c.storePod(ctx, &pod); err != nil {
			log.Printf("Error storing pod %s/%s: %v", pod.Namespace, pod.Name, err)
			continue
		}
	}

	// Run analysis
	if err := c.runAnalysis(ctx); err != nil {
		log.Printf("Error running analysis: %v", err)
	}

	log.Println("Collection complete!")
	return nil
}

func (c *Collector) storePod(ctx context.Context, pod *corev1.Pod) error {
	// Insert or update pod
	var podID int64
	err := c.db.QueryRow(`
		INSERT INTO pods (namespace, pod_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (namespace, pod_name) DO UPDATE SET updated_at = $4
		RETURNING id
	`, pod.Namespace, pod.Name, time.Now(), time.Now()).Scan(&podID)

	if err != nil {
		return fmt.Errorf("failed to insert pod: %w", err)
	}

	// Process each container
	for _, container := range pod.Spec.Containers {
		if err := c.storeContainer(ctx, podID, pod.Namespace, pod.Name, &container); err != nil {
			log.Printf("Error storing container %s: %v", container.Name, err)
		}
	}

	return nil
}

func (c *Collector) storeContainer(ctx context.Context, podID int64, namespace, podName string, container *corev1.Container) error {
	// Insert or update container
	var containerID int64
	err := c.db.QueryRow(`
		INSERT INTO containers (pod_id, container_name, image, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (pod_id, container_name) DO UPDATE SET image = $3, updated_at = $5
		RETURNING id
	`, podID, container.Name, container.Image, time.Now(), time.Now()).Scan(&containerID)

	if err != nil {
		return fmt.Errorf("failed to insert container: %w", err)
	}

	// Store resource requests
	cpuRequest := float64(0)
	cpuLimit := float64(0)
	memRequest := int64(0)
	memLimit := int64(0)

	if container.Resources.Requests != nil {
		if cpu := container.Resources.Requests.Cpu(); cpu != nil {
			cpuRequest = float64(cpu.MilliValue()) / 1000.0
		}
		if mem := container.Resources.Requests.Memory(); mem != nil {
			memRequest = mem.Value()
		}
	}

	if container.Resources.Limits != nil {
		if cpu := container.Resources.Limits.Cpu(); cpu != nil {
			cpuLimit = float64(cpu.MilliValue()) / 1000.0
		}
		if mem := container.Resources.Limits.Memory(); mem != nil {
			memLimit = mem.Value()
		}
	}

	// Insert or update resource requests
	_, err = c.db.Exec(`
		INSERT INTO resource_requests (container_id, cpu_request, cpu_limit, mem_request, mem_limit, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT DO NOTHING
	`, containerID, cpuRequest, cpuLimit, memRequest, memLimit, time.Now())

	if err != nil {
		log.Printf("Warning: failed to insert resource request: %v", err)
	}

	// Try to get metrics for this container
	if err := c.storeMetrics(ctx, containerID, namespace, podName, container.Name); err != nil {
		log.Printf("Warning: failed to store metrics for %s/%s/%s: %v", namespace, podName, container.Name, err)
	}

	return nil
}

func (c *Collector) storeMetrics(ctx context.Context, containerID int64, namespace, podName, containerName string) error {
	// Get pod metrics
	podMetrics, err := c.metricsClient.MetricsV1beta1().PodMetricses(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// Find container metrics
	for _, cm := range podMetrics.Containers {
		if cm.Name == containerName {
			cpuUsage := float64(cm.Usage.Cpu().MilliValue()) / 1000.0
			memUsage := cm.Usage.Memory().Value()

			_, err := c.db.Exec(`
				INSERT INTO metrics_snapshots (container_id, timestamp, cpu_usage, memory_usage)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT (container_id, timestamp) DO NOTHING
			`, containerID, time.Now().Truncate(time.Minute), cpuUsage, memUsage)

			if err != nil {
				return err
			}

			log.Printf("  Stored metrics for %s/%s/%s: CPU=%.3f cores, Memory=%d MB",
				namespace, podName, containerName, cpuUsage, memUsage/(1024*1024))
			return nil
		}
	}

	return fmt.Errorf("container not found in metrics")
}

func (c *Collector) runAnalysis(ctx context.Context) error {
	log.Println("Running analysis...")

	// Get containers with enough metrics data
	rows, err := c.db.Query(`
		SELECT DISTINCT c.id, c.pod_id, c.container_name, p.namespace, p.pod_name
		FROM containers c
		JOIN pods p ON p.id = c.pod_id
		WHERE EXISTS (
			SELECT 1 FROM metrics_snapshots m 
			WHERE m.container_id = c.id
		)
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type containerInfo struct {
		id            int64
		podID         int64
		containerName string
		namespace     string
		podName       string
	}

	var containers []containerInfo
	for rows.Next() {
		var ci containerInfo
		if err := rows.Scan(&ci.id, &ci.podID, &ci.containerName, &ci.namespace, &ci.podName); err != nil {
			continue
		}
		containers = append(containers, ci)
	}

	for _, ci := range containers {
		if err := c.analyzeContainer(ctx, ci.id, ci.namespace, ci.podName, ci.containerName); err != nil {
			log.Printf("Error analyzing %s/%s/%s: %v", ci.namespace, ci.podName, ci.containerName, err)
		}
	}

	return nil
}

func (c *Collector) analyzeContainer(ctx context.Context, containerID int64, namespace, podName, containerName string) error {
	// Get metrics for last 7 days
	windowStart := time.Now().Add(-7 * 24 * time.Hour)
	windowEnd := time.Now()

	rows, err := c.db.Query(`
		SELECT cpu_usage, memory_usage
		FROM metrics_snapshots
		WHERE container_id = $1 AND timestamp >= $2
		ORDER BY timestamp
	`, containerID, windowStart)
	if err != nil {
		return err
	}
	defer rows.Close()

	var cpuValues []float64
	var memValues []int64

	for rows.Next() {
		var cpu float64
		var mem int64
		if err := rows.Scan(&cpu, &mem); err != nil {
			continue
		}
		cpuValues = append(cpuValues, cpu)
		memValues = append(memValues, mem)
	}

	if len(cpuValues) == 0 {
		return fmt.Errorf("no metrics data")
	}

	// Calculate statistics
	avgCPU, maxCPU, p95CPU, p99CPU := calculateStats(cpuValues)
	avgMem, maxMem, p95Mem, p99Mem := calculateStatsInt(memValues)

	// Get current resource requests
	var currentCPU float64
	var currentMem int64
	err = c.db.QueryRow(`
		SELECT COALESCE(cpu_request, 0), COALESCE(mem_request, 0)
		FROM resource_requests
		WHERE container_id = $1
		ORDER BY updated_at DESC
		LIMIT 1
	`, containerID).Scan(&currentCPU, &currentMem)
	if err != nil {
		// No resource requests, use defaults
		currentCPU = 0.1
		currentMem = 128 * 1024 * 1024
	}

	// Calculate recommended resources (P95 + 20% buffer)
	recommendedCPU := p95CPU * 1.2
	recommendedMem := int64(float64(p95Mem) * 1.2)

	// Ensure minimums
	if recommendedCPU < 0.01 {
		recommendedCPU = 0.01
	}
	if recommendedMem < 32*1024*1024 {
		recommendedMem = 32 * 1024 * 1024
	}

	// Calculate waste percentages
	cpuWaste := float64(0)
	memWaste := float64(0)
	if currentCPU > 0 {
		cpuWaste = ((currentCPU - recommendedCPU) / currentCPU) * 100
	}
	if currentMem > 0 {
		memWaste = float64(currentMem-recommendedMem) / float64(currentMem) * 100
	}

	// Clamp waste to reasonable bounds
	if cpuWaste < -100 {
		cpuWaste = -100
	}
	if memWaste < -100 {
		memWaste = -100
	}

	// Calculate monthly savings
	cpuSavings := float64(0)
	memSavings := float64(0)
	if cpuWaste > 0 {
		cpuSavings = (currentCPU - recommendedCPU) * c.config.Analysis.CPUCostPerCore
	}
	if memWaste > 0 {
		memSavings = float64(currentMem-recommendedMem) / (1024 * 1024 * 1024) * c.config.Analysis.MemoryCostPerGB
	}
	monthlySavings := cpuSavings + memSavings
	if monthlySavings < 0 {
		monthlySavings = 0
	}

	// Determine status
	status := "optimal"
	if cpuWaste > 30 || memWaste > 30 {
		status = "over-provisioned"
	} else if cpuWaste < -20 || memWaste < -20 {
		status = "under-provisioned"
	}

	// Determine confidence based on data points
	confidence := "low"
	if len(cpuValues) >= 100 {
		confidence = "high"
	} else if len(cpuValues) >= 20 {
		confidence = "medium"
	}

	// Insert analysis
	var analysisID int64
	err = c.db.QueryRow(`
		INSERT INTO analyses (
			container_id, analyzed_at, window_start, window_end,
			avg_cpu, max_cpu, p95_cpu, p99_cpu,
			avg_memory, max_memory, p95_memory, p99_memory,
			current_cpu_request, current_mem_request,
			recommended_cpu, recommended_memory,
			cpu_waste_percent, memory_waste_percent,
			monthly_savings, status, confidence
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		RETURNING id
	`, containerID, time.Now(), windowStart, windowEnd,
		avgCPU, maxCPU, p95CPU, p99CPU,
		avgMem, maxMem, p95Mem, p99Mem,
		currentCPU, currentMem, recommendedCPU, recommendedMem,
		cpuWaste, memWaste, monthlySavings, status, confidence).Scan(&analysisID)

	if err != nil {
		return err
	}

	// Generate recommendation
	reason := fmt.Sprintf("Based on %d data points over 7 days. CPU waste: %.1f%%, Memory waste: %.1f%%",
		len(cpuValues), cpuWaste, memWaste)

	_, err = c.db.Exec(`
		INSERT INTO recommendations (
			analysis_id, namespace, pod_name, container_name,
			current_cpu, current_memory,
			recommended_cpu, recommended_memory,
			monthly_savings, confidence, status, reason, applied
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, analysisID, namespace, podName, containerName,
		currentCPU, currentMem, recommendedCPU, recommendedMem,
		monthlySavings, confidence, status, reason, false)

	log.Printf("  Analyzed %s/%s/%s: status=%s, savings=$%.2f/month",
		namespace, podName, containerName, status, monthlySavings)

	return err
}

func calculateStats(values []float64) (avg, max, p95, p99 float64) {
	if len(values) == 0 {
		return
	}

	// Sort for percentiles
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	// Calculate average
	var sum float64
	for _, v := range values {
		sum += v
	}
	avg = sum / float64(len(values))

	// Max
	max = sorted[len(sorted)-1]

	// P95
	p95Index := int(float64(len(sorted)) * 0.95)
	if p95Index >= len(sorted) {
		p95Index = len(sorted) - 1
	}
	p95 = sorted[p95Index]

	// P99
	p99Index := int(float64(len(sorted)) * 0.99)
	if p99Index >= len(sorted) {
		p99Index = len(sorted) - 1
	}
	p99 = sorted[p99Index]

	return
}

func calculateStatsInt(values []int64) (avg, max, p95, p99 int64) {
	if len(values) == 0 {
		return
	}

	// Sort for percentiles
	sorted := make([]int64, len(values))
	copy(sorted, values)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	// Calculate average
	var sum int64
	for _, v := range values {
		sum += v
	}
	avg = sum / int64(len(values))

	// Max
	max = sorted[len(sorted)-1]

	// P95
	p95Index := int(float64(len(sorted)) * 0.95)
	if p95Index >= len(sorted) {
		p95Index = len(sorted) - 1
	}
	p95 = sorted[p95Index]

	// P99
	p99Index := int(float64(len(sorted)) * 0.99)
	if p99Index >= len(sorted) {
		p99Index = len(sorted) - 1
	}
	p99 = sorted[p99Index]

	return
}

func isImportantSystemPod(name string) bool {
	// Include some important system pods
	importantPrefixes := []string{
		"coredns",
		"metrics-server",
		"aws-node",
		"kube-proxy",
	}
	for _, prefix := range importantPrefixes {
		if len(name) >= len(prefix) && name[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

