package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/scaleops/k8s-optimizer/internal/database"
	"github.com/scaleops/k8s-optimizer/internal/models"
)

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetPods(namespace, status, sortBy string, limit int) ([]models.PodDetail, error) {
	query := `
		SELECT 
			p.namespace,
			p.pod_name,
			c.container_name,
			a.status,
			a.cpu_waste_percent,
			a.memory_waste_percent,
			a.monthly_savings,
			a.current_cpu_request,
			a.current_mem_request,
			a.recommended_cpu,
			a.recommended_memory,
			a.confidence
		FROM pods p
		JOIN containers c ON c.pod_id = p.id
		JOIN analyses a ON a.container_id = c.id
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	if namespace != "" {
		query += fmt.Sprintf(" AND p.namespace = $%d", argCount)
		args = append(args, namespace)
		argCount++
	}

	if status != "" {
		query += fmt.Sprintf(" AND a.status = $%d", argCount)
		args = append(args, status)
		argCount++
	}

	// Default sort by savings descending
	orderBy := "a.monthly_savings DESC"
	if sortBy == "waste" {
		orderBy = "a.cpu_waste_percent DESC"
	} else if sortBy == "name" {
		orderBy = "p.pod_name ASC"
	}
	query += " ORDER BY " + orderBy

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, limit)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pods []models.PodDetail
	for rows.Next() {
		var p models.PodDetail
		err := rows.Scan(
			&p.Namespace,
			&p.PodName,
			&p.ContainerName,
			&p.Status,
			&p.CPUWastePercent,
			&p.MemoryWastePercent,
			&p.MonthlySavings,
			&p.CurrentCPU,
			&p.CurrentMemory,
			&p.RecommendedCPU,
			&p.RecommendedMemory,
			&p.Confidence,
		)
		if err != nil {
			return nil, err
		}
		pods = append(pods, p)
	}

	return pods, nil
}

func (r *Repository) GetPodDetail(namespace, podName string) (*models.PodDetail, *models.Analysis, []models.UsageHistory, error) {
	// Get pod detail
	query := `
		SELECT 
			p.namespace,
			p.pod_name,
			c.container_name,
			a.status,
			a.cpu_waste_percent,
			a.memory_waste_percent,
			a.monthly_savings,
			a.current_cpu_request,
			a.current_mem_request,
			a.recommended_cpu,
			a.recommended_memory,
			a.confidence
		FROM pods p
		JOIN containers c ON c.pod_id = p.id
		JOIN analyses a ON a.container_id = c.id
		WHERE p.namespace = $1 AND p.pod_name = $2
		LIMIT 1
	`

	var pod models.PodDetail
	err := r.db.QueryRow(query, namespace, podName).Scan(
		&pod.Namespace,
		&pod.PodName,
		&pod.ContainerName,
		&pod.Status,
		&pod.CPUWastePercent,
		&pod.MemoryWastePercent,
		&pod.MonthlySavings,
		&pod.CurrentCPU,
		&pod.CurrentMemory,
		&pod.RecommendedCPU,
		&pod.RecommendedMemory,
		&pod.Confidence,
	)
	if err != nil {
		return nil, nil, nil, err
	}

	// Get full analysis
	analysisQuery := `
		SELECT 
			a.id, a.container_id, a.analyzed_at, a.window_start, a.window_end,
			a.avg_cpu, a.max_cpu, a.p95_cpu, a.p99_cpu,
			a.avg_memory, a.max_memory, a.p95_memory, a.p99_memory,
			a.current_cpu_request, a.current_mem_request,
			a.recommended_cpu, a.recommended_memory,
			a.cpu_waste_percent, a.memory_waste_percent, a.monthly_savings,
			a.status, a.confidence
		FROM pods p
		JOIN containers c ON c.pod_id = p.id
		JOIN analyses a ON a.container_id = c.id
		WHERE p.namespace = $1 AND p.pod_name = $2
		LIMIT 1
	`

	var analysis models.Analysis
	err = r.db.QueryRow(analysisQuery, namespace, podName).Scan(
		&analysis.ID, &analysis.ContainerID, &analysis.AnalyzedAt,
		&analysis.WindowStart, &analysis.WindowEnd,
		&analysis.AvgCPU, &analysis.MaxCPU, &analysis.P95CPU, &analysis.P99CPU,
		&analysis.AvgMemory, &analysis.MaxMemory, &analysis.P95Memory, &analysis.P99Memory,
		&analysis.CurrentCPURequest, &analysis.CurrentMemRequest,
		&analysis.RecommendedCPU, &analysis.RecommendedMemory,
		&analysis.CPUWastePercent, &analysis.MemoryWastePercent, &analysis.MonthlySavings,
		&analysis.Status, &analysis.Confidence,
	)
	if err != nil {
		return &pod, nil, nil, err
	}

	// Get usage history
	historyQuery := `
		SELECT timestamp, cpu_usage, memory_usage
		FROM metrics_snapshots
		WHERE container_id = $1
		ORDER BY timestamp DESC
		LIMIT 100
	`

	rows, err := r.db.Query(historyQuery, analysis.ContainerID)
	if err != nil {
		return &pod, &analysis, nil, err
	}
	defer rows.Close()

	var history []models.UsageHistory
	for rows.Next() {
		var h models.UsageHistory
		if err := rows.Scan(&h.Timestamp, &h.CPU, &h.Memory); err != nil {
			return &pod, &analysis, nil, err
		}
		history = append(history, h)
	}

	return &pod, &analysis, history, nil
}

func (r *Repository) GetRecommendations(confidence string, minSavings float64, limit int) ([]models.Recommendation, error) {
	query := `
		SELECT 
			id, analysis_id, namespace, pod_name, container_name,
			current_cpu, current_memory, recommended_cpu, recommended_memory,
			monthly_savings, confidence, status, reason, applied, created_at
		FROM recommendations
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	if confidence != "" {
		query += fmt.Sprintf(" AND confidence = $%d", argCount)
		args = append(args, confidence)
		argCount++
	}

	if minSavings > 0 {
		query += fmt.Sprintf(" AND monthly_savings >= $%d", argCount)
		args = append(args, minSavings)
		argCount++
	}

	query += " ORDER BY monthly_savings DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, limit)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recommendations []models.Recommendation
	for rows.Next() {
		var r models.Recommendation
		err := rows.Scan(
			&r.ID, &r.AnalysisID, &r.Namespace, &r.PodName, &r.ContainerName,
			&r.CurrentCPU, &r.CurrentMemory, &r.RecommendedCPU, &r.RecommendedMemory,
			&r.MonthlySavings, &r.Confidence, &r.Status, &r.Reason, &r.Applied, &r.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		recommendations = append(recommendations, r)
	}

	return recommendations, nil
}

func (r *Repository) GetRecommendationByID(id int64) (*models.Recommendation, error) {
	query := `
		SELECT 
			id, analysis_id, namespace, pod_name, container_name,
			current_cpu, current_memory, recommended_cpu, recommended_memory,
			monthly_savings, confidence, status, reason, applied, created_at
		FROM recommendations
		WHERE id = $1
	`

	var rec models.Recommendation
	err := r.db.QueryRow(query, id).Scan(
		&rec.ID, &rec.AnalysisID, &rec.Namespace, &rec.PodName, &rec.ContainerName,
		&rec.CurrentCPU, &rec.CurrentMemory, &rec.RecommendedCPU, &rec.RecommendedMemory,
		&rec.MonthlySavings, &rec.Confidence, &rec.Status, &rec.Reason, &rec.Applied, &rec.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &rec, nil
}

func (r *Repository) MarkRecommendationApplied(id int64, applied bool) error {
	query := `UPDATE recommendations SET applied = $1 WHERE id = $2`
	_, err := r.db.Exec(query, applied, id)
	return err
}

func (r *Repository) GetStatistics() (*models.Statistics, error) {
	var stats models.Statistics

	// Get pod counts by status
	statusQuery := `
		SELECT 
			COUNT(DISTINCT p.id) as total,
			COUNT(DISTINCT CASE WHEN a.status = 'over-provisioned' THEN p.id END) as over_prov,
			COUNT(DISTINCT CASE WHEN a.status = 'under-provisioned' THEN p.id END) as under_prov,
			COUNT(DISTINCT CASE WHEN a.status = 'optimal' THEN p.id END) as optimal,
			COALESCE(SUM(a.monthly_savings), 0) as total_savings,
			COALESCE(SUM(CASE WHEN a.status = 'over-provisioned' 
				THEN a.current_cpu_request - a.recommended_cpu 
				ELSE 0 END), 0) as cpu_waste,
			COALESCE(SUM(CASE WHEN a.status = 'over-provisioned' 
				THEN (a.current_mem_request - a.recommended_memory)::DOUBLE PRECISION / 1073741824.0
				ELSE 0 END), 0) as mem_waste
		FROM pods p
		JOIN containers c ON c.pod_id = p.id
		JOIN analyses a ON a.container_id = c.id
	`

	err := r.db.QueryRow(statusQuery).Scan(
		&stats.TotalPods,
		&stats.OverProvisioned,
		&stats.UnderProvisioned,
		&stats.Optimal,
		&stats.TotalMonthlySavings,
		&stats.TotalCPUWasteCores,
		&stats.TotalMemoryWasteGB,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Get last analysis time
	err = r.db.QueryRow("SELECT MAX(analyzed_at) FROM analyses").Scan(&stats.LastAnalysis)
	if err != nil && err != sql.ErrNoRows {
		stats.LastAnalysis = time.Time{}
	}

	// Get last collection time
	err = r.db.QueryRow("SELECT MAX(timestamp) FROM metrics_snapshots").Scan(&stats.LastCollection)
	if err != nil && err != sql.ErrNoRows {
		stats.LastCollection = time.Time{}
	}

	return &stats, nil
}

func (r *Repository) GetNamespaces() ([]string, error) {
	rows, err := r.db.Query("SELECT DISTINCT namespace FROM pods ORDER BY namespace")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var namespaces []string
	for rows.Next() {
		var ns string
		if err := rows.Scan(&ns); err != nil {
			return nil, err
		}
		namespaces = append(namespaces, ns)
	}

	return namespaces, nil
}

func (r *Repository) SearchPods(searchTerm string) ([]models.PodDetail, error) {
	searchTerm = "%" + strings.ToLower(searchTerm) + "%"
	query := `
		SELECT 
			p.namespace,
			p.pod_name,
			c.container_name,
			a.status,
			a.cpu_waste_percent,
			a.memory_waste_percent,
			a.monthly_savings,
			a.current_cpu_request,
			a.current_mem_request,
			a.recommended_cpu,
			a.recommended_memory,
			a.confidence
		FROM pods p
		JOIN containers c ON c.pod_id = p.id
		JOIN analyses a ON a.container_id = c.id
		WHERE LOWER(p.pod_name) LIKE $1 OR LOWER(p.namespace) LIKE $1
		ORDER BY a.monthly_savings DESC
		LIMIT 50
	`

	rows, err := r.db.Query(query, searchTerm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pods []models.PodDetail
	for rows.Next() {
		var p models.PodDetail
		err := rows.Scan(
			&p.Namespace,
			&p.PodName,
			&p.ContainerName,
			&p.Status,
			&p.CPUWastePercent,
			&p.MemoryWastePercent,
			&p.MonthlySavings,
			&p.CurrentCPU,
			&p.CurrentMemory,
			&p.RecommendedCPU,
			&p.RecommendedMemory,
			&p.Confidence,
		)
		if err != nil {
			return nil, err
		}
		pods = append(pods, p)
	}

	return pods, nil
}

