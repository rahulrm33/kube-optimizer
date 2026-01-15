package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/scaleops/k8s-optimizer/internal/config"
	"github.com/scaleops/k8s-optimizer/internal/database"
)

func main() {
	// Load configuration
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

	log.Println("Seeding database with sample data...")

	// Namespaces
	namespaces := []string{"default", "kube-system", "production", "staging", "development"}
	
	// Pod names
	podPrefixes := []string{"api-server", "web-app", "worker", "database", "cache", "queue", "auth-service"}
	
	// Container names
	containerNames := []string{"app", "sidecar", "init", "proxy"}

	rand.Seed(time.Now().UnixNano())

	// Generate 50 sample pods
	for i := 0; i < 50; i++ {
		namespace := namespaces[rand.Intn(len(namespaces))]
		podName := fmt.Sprintf("%s-%d", podPrefixes[rand.Intn(len(podPrefixes))], rand.Intn(1000))
		
		// Insert pod
		var podID int64
		err := db.QueryRow(`
			INSERT INTO pods (namespace, pod_name, created_at, updated_at)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (namespace, pod_name) DO UPDATE SET updated_at = $4
			RETURNING id
		`, namespace, podName, time.Now(), time.Now()).Scan(&podID)
		
		if err != nil {
			log.Printf("Error inserting pod: %v", err)
			continue
		}

		// Insert container
		containerName := containerNames[rand.Intn(len(containerNames))]
		var containerID int64
		err = db.QueryRow(`
			INSERT INTO containers (pod_id, container_name, image, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (pod_id, container_name) DO UPDATE SET updated_at = $5
			RETURNING id
		`, podID, containerName, "nginx:latest", time.Now(), time.Now()).Scan(&containerID)
		
		if err != nil {
			log.Printf("Error inserting container: %v", err)
			continue
		}

		// Generate random resource requests
		cpuRequest := float64(rand.Intn(4000)+500) / 1000.0        // 0.5 to 4.5 cores
		memRequest := int64(rand.Intn(8000)+512) * 1024 * 1024     // 512MB to 8.5GB
		
		// Insert resource requests
		_, err = db.Exec(`
			INSERT INTO resource_requests (container_id, cpu_request, cpu_limit, mem_request, mem_limit, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, containerID, cpuRequest, cpuRequest*1.5, memRequest, memRequest*2, time.Now())
		
		if err != nil {
			log.Printf("Error inserting resource request: %v", err)
			continue
		}

		// Generate 100 metrics snapshots over the last 7 days
		for j := 0; j < 100; j++ {
			timestamp := time.Now().Add(-time.Duration(j) * time.Hour)
			
			// Generate realistic usage (often less than requested)
			cpuUsage := cpuRequest * (rand.Float64()*0.6 + 0.1) // 10-70% of request
			memUsage := int64(float64(memRequest) * (rand.Float64()*0.7 + 0.1)) // 10-80% of request
			
			_, err = db.Exec(`
				INSERT INTO metrics_snapshots (container_id, timestamp, cpu_usage, memory_usage)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT (container_id, timestamp) DO NOTHING
			`, containerID, timestamp, cpuUsage, memUsage)
			
			if err != nil {
				log.Printf("Error inserting metrics snapshot: %v", err)
			}
		}

		// Calculate statistics for analysis
		avgCPU := cpuRequest * (rand.Float64()*0.5 + 0.15) // 15-65% of request
		p95CPU := cpuRequest * (rand.Float64()*0.6 + 0.25) // 25-85% of request
		
		avgMemory := int64(float64(memRequest) * (rand.Float64()*0.6 + 0.15))
		p95Memory := int64(float64(memRequest) * (rand.Float64()*0.7 + 0.25))
		
		// Recommended resources (p95 + 20% buffer)
		recommendedCPU := p95CPU * 1.2
		recommendedMemory := int64(float64(p95Memory) * 1.2)
		
		// Calculate waste
		cpuWaste := ((cpuRequest - recommendedCPU) / cpuRequest) * 100
		memWaste := float64(memRequest-recommendedMemory) / float64(memRequest) * 100
		
		// Ensure waste is not negative
		if cpuWaste < 0 {
			cpuWaste = 0
		}
		if memWaste < 0 {
			memWaste = 0
		}
		
		// Calculate monthly savings
		cpuSavings := (cpuRequest - recommendedCPU) * cfg.Analysis.CPUCostPerCore * 720 / 24
		memSavings := float64(memRequest-recommendedMemory) / (1024*1024*1024) * cfg.Analysis.MemoryCostPerGB
		monthlySavings := cpuSavings + memSavings
		
		if monthlySavings < 0 {
			monthlySavings = 0
		}

		// Determine status
		status := "optimal"
		if cpuWaste > 30 || memWaste > 30 {
			status = "over-provisioned"
		} else if cpuWaste < -10 || memWaste < -10 {
			status = "under-provisioned"
		}

		// Determine confidence
		confidence := "high"
		if rand.Float64() < 0.3 {
			confidence = "medium"
		} else if rand.Float64() < 0.1 {
			confidence = "low"
		}

		// Insert analysis
		var analysisID int64
		err = db.QueryRow(`
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
		`, containerID, time.Now(), time.Now().Add(-7*24*time.Hour), time.Now(),
			avgCPU, p95CPU*1.1, p95CPU, p95CPU*1.05,
			avgMemory, p95Memory*110/100, p95Memory, p95Memory*105/100,
			cpuRequest, memRequest, recommendedCPU, recommendedMemory,
			cpuWaste, memWaste, monthlySavings, status, confidence).Scan(&analysisID)
		
		if err != nil {
			log.Printf("Error inserting analysis: %v", err)
			continue
		}

		// Insert recommendation
		reason := fmt.Sprintf("Based on 7-day analysis, resources can be optimized. Current CPU waste: %.1f%%, Memory waste: %.1f%%", cpuWaste, memWaste)
		if status == "optimal" {
			reason = "Resources are well-configured based on usage patterns."
		}

		_, err = db.Exec(`
			INSERT INTO recommendations (
				analysis_id, namespace, pod_name, container_name,
				current_cpu, current_memory,
				recommended_cpu, recommended_memory,
				monthly_savings, confidence, status, reason, applied
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		`, analysisID, namespace, podName, containerName,
			cpuRequest, memRequest, recommendedCPU, recommendedMemory,
			monthlySavings, confidence, status, reason, false)
		
		if err != nil {
			log.Printf("Error inserting recommendation: %v", err)
			continue
		}

		log.Printf("Seeded pod %s/%s with status %s (savings: $%.2f/month)", namespace, podName, status, monthlySavings)
	}

	log.Println("Sample data seeding complete!")
	log.Println("You can now access the dashboard at http://localhost:8080")
}

