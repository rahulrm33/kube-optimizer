package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/scaleops/k8s-optimizer/internal/config"
	"github.com/scaleops/k8s-optimizer/internal/models"
	"github.com/scaleops/k8s-optimizer/internal/repository"
	"gopkg.in/yaml.v3"
)

type Handler struct {
	repo   *repository.Repository
	config *config.Config
}

func NewHandler(repo *repository.Repository, cfg *config.Config) *Handler {
	return &Handler{
		repo:   repo,
		config: cfg,
	}
}

// Dashboard home page
func (h *Handler) GetDashboard(c *gin.Context) {
	stats, err := h.repo.GetStatistics()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to load statistics",
		})
		return
	}

	// Get top 10 wasteful pods
	topPods, err := h.repo.GetPods("", "over-provisioned", "savings", 10)
	if err != nil {
		topPods = []models.PodDetail{}
	}

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"stats":   stats,
		"topPods": topPods,
	})
}

// GET /api/pods - List analyzed pods
func (h *Handler) GetPods(c *gin.Context) {
	namespace := c.Query("namespace")
	status := c.Query("status")
	sortBy := c.Query("sort_by")
	limitStr := c.DefaultQuery("limit", "50")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	// Handle search
	searchTerm := c.Query("search")
	var pods []models.PodDetail

	if searchTerm != "" {
		pods, err = h.repo.SearchPods(searchTerm)
	} else {
		pods, err = h.repo.GetPods(namespace, status, sortBy, limit)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch pods",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pods":  pods,
		"total": len(pods),
		"page":  1,
	})
}

// GET /api/pod/:namespace/:name - Pod detail
func (h *Handler) GetPodDetail(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	pod, analysis, history, err := h.repo.GetPodDetail(namespace, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Pod not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pod":           pod,
		"analysis":      analysis,
		"usage_history": history,
	})
}

// GET /api/recommendations - All recommendations
func (h *Handler) GetRecommendations(c *gin.Context) {
	confidence := c.Query("confidence")
	minSavingsStr := c.DefaultQuery("min_savings", "0")
	limitStr := c.DefaultQuery("limit", "100")

	minSavings, _ := strconv.ParseFloat(minSavingsStr, 64)
	limit, _ := strconv.Atoi(limitStr)

	recommendations, err := h.repo.GetRecommendations(confidence, minSavings, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch recommendations",
		})
		return
	}

	// Calculate total savings
	var totalSavings float64
	for _, rec := range recommendations {
		totalSavings += rec.MonthlySavings
	}

	c.JSON(http.StatusOK, gin.H{
		"recommendations": recommendations,
		"total_savings":   totalSavings,
		"total_count":     len(recommendations),
	})
}

// GET /api/recommendations/:id/yaml - Download YAML patch
func (h *Handler) GetRecommendationYAML(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid recommendation ID",
		})
		return
	}

	rec, err := h.repo.GetRecommendationByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Recommendation not found",
		})
		return
	}

	// Generate YAML patch
	patch := generateResourcePatch(rec)

	yamlData, err := yaml.Marshal(patch)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate YAML",
		})
		return
	}

	// Set headers for download
	filename := fmt.Sprintf("patch-%s-%s-%s.yaml", rec.Namespace, rec.PodName, rec.ContainerName)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "text/yaml", yamlData)
}

// POST /api/recommendations/:id/apply - Mark as applied
func (h *Handler) ApplyRecommendation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid recommendation ID",
		})
		return
	}

	var body struct {
		Applied bool `json:"applied"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	if err := h.repo.MarkRecommendationApplied(id, body.Applied); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update recommendation",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Recommendation updated successfully",
	})
}

// GET /api/stats - Overall statistics
func (h *Handler) GetStats(c *gin.Context) {
	stats, err := h.repo.GetStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch statistics",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GET /api/namespaces - Get all namespaces
func (h *Handler) GetNamespaces(c *gin.Context) {
	namespaces, err := h.repo.GetNamespaces()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch namespaces",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"namespaces": namespaces,
	})
}

// Helper function to generate Kubernetes resource patch
func generateResourcePatch(rec *models.Recommendation) map[string]interface{} {
	cpuRequest := fmt.Sprintf("%.0fm", rec.RecommendedCPU*1000)
	memRequest := fmt.Sprintf("%dMi", rec.RecommendedMemory/(1024*1024))

	// Calculate limits (20% headroom)
	cpuLimit := fmt.Sprintf("%.0fm", rec.RecommendedCPU*1000*1.2)
	memLimit := fmt.Sprintf("%dMi", rec.RecommendedMemory/(1024*1024)*120/100)

	patch := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata": map[string]interface{}{
			"name":      rec.PodName,
			"namespace": rec.Namespace,
		},
		"spec": map[string]interface{}{
			"containers": []map[string]interface{}{
				{
					"name": rec.ContainerName,
					"resources": map[string]interface{}{
						"requests": map[string]interface{}{
							"cpu":    cpuRequest,
							"memory": memRequest,
						},
						"limits": map[string]interface{}{
							"cpu":    cpuLimit,
							"memory": memLimit,
						},
					},
				},
			},
		},
	}

	return patch
}

// Health check endpoint
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

