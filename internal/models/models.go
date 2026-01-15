package models

import (
	"time"
)

type Pod struct {
	ID        int64     `json:"id"`
	Namespace string    `json:"namespace"`
	PodName   string    `json:"pod_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Container struct {
	ID            int64     `json:"id"`
	PodID         int64     `json:"pod_id"`
	ContainerName string    `json:"container_name"`
	Image         string    `json:"image"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type MetricsSnapshot struct {
	ID          int64     `json:"id"`
	ContainerID int64     `json:"container_id"`
	Timestamp   time.Time `json:"timestamp"`
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage int64     `json:"memory_usage"`
}

type ResourceRequest struct {
	ID          int64     `json:"id"`
	ContainerID int64     `json:"container_id"`
	CPURequest  float64   `json:"cpu_request"`
	CPULimit    float64   `json:"cpu_limit"`
	MemRequest  int64     `json:"mem_request"`
	MemLimit    int64     `json:"mem_limit"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Analysis struct {
	ID                 int64     `json:"id"`
	ContainerID        int64     `json:"container_id"`
	AnalyzedAt         time.Time `json:"analyzed_at"`
	WindowStart        time.Time `json:"window_start"`
	WindowEnd          time.Time `json:"window_end"`
	AvgCPU             float64   `json:"avg_cpu"`
	MaxCPU             float64   `json:"max_cpu"`
	P95CPU             float64   `json:"p95_cpu"`
	P99CPU             float64   `json:"p99_cpu"`
	AvgMemory          int64     `json:"avg_memory"`
	MaxMemory          int64     `json:"max_memory"`
	P95Memory          int64     `json:"p95_memory"`
	P99Memory          int64     `json:"p99_memory"`
	CurrentCPURequest  float64   `json:"current_cpu_request"`
	CurrentMemRequest  int64     `json:"current_mem_request"`
	RecommendedCPU     float64   `json:"recommended_cpu"`
	RecommendedMemory  int64     `json:"recommended_memory"`
	CPUWastePercent    float64   `json:"cpu_waste_percent"`
	MemoryWastePercent float64   `json:"memory_waste_percent"`
	MonthlySavings     float64   `json:"monthly_savings"`
	Status             string    `json:"status"`
	Confidence         string    `json:"confidence"`
}

type Recommendation struct {
	ID                int64     `json:"id"`
	AnalysisID        int64     `json:"analysis_id"`
	Namespace         string    `json:"namespace"`
	PodName           string    `json:"pod_name"`
	ContainerName     string    `json:"container_name"`
	CurrentCPU        float64   `json:"current_cpu"`
	CurrentMemory     int64     `json:"current_memory"`
	RecommendedCPU    float64   `json:"recommended_cpu"`
	RecommendedMemory int64     `json:"recommended_memory"`
	MonthlySavings    float64   `json:"monthly_savings"`
	Confidence        string    `json:"confidence"`
	Status            string    `json:"status"`
	Reason            string    `json:"reason"`
	Applied           bool      `json:"applied"`
	CreatedAt         time.Time `json:"created_at"`
}

type PodDetail struct {
	Namespace          string  `json:"namespace"`
	PodName            string  `json:"pod_name"`
	ContainerName      string  `json:"container_name"`
	Status             string  `json:"status"`
	CPUWastePercent    float64 `json:"cpu_waste_percent"`
	MemoryWastePercent float64 `json:"memory_waste_percent"`
	MonthlySavings     float64 `json:"monthly_savings"`
	CurrentCPU         float64 `json:"current_cpu"`
	CurrentMemory      int64   `json:"current_memory"`
	RecommendedCPU     float64 `json:"recommended_cpu"`
	RecommendedMemory  int64   `json:"recommended_memory"`
	Confidence         string  `json:"confidence"`
}

type Statistics struct {
	TotalPods           int       `json:"total_pods"`
	OverProvisioned     int       `json:"over_provisioned"`
	UnderProvisioned    int       `json:"under_provisioned"`
	Optimal             int       `json:"optimal"`
	TotalMonthlySavings float64   `json:"total_monthly_savings"`
	TotalCPUWasteCores  float64   `json:"total_cpu_waste_cores"`
	TotalMemoryWasteGB  float64   `json:"total_memory_waste_gb"`
	LastAnalysis        time.Time `json:"last_analysis"`
	LastCollection      time.Time `json:"last_collection"`
}

type UsageHistory struct {
	Timestamp time.Time `json:"timestamp"`
	CPU       float64   `json:"cpu"`
	Memory    int64     `json:"memory"`
}
