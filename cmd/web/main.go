package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/scaleops/k8s-optimizer/internal/config"
	"github.com/scaleops/k8s-optimizer/internal/database"
	"github.com/scaleops/k8s-optimizer/internal/repository"
	"github.com/scaleops/k8s-optimizer/web/handlers"
	"github.com/scaleops/k8s-optimizer/web/middleware"
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

	// Initialize database schema
	if err := db.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	log.Println("Database connected and schema initialized")

	// Create repository and handlers
	repo := repository.NewRepository(db)
	handler := handlers.NewHandler(repo, cfg)

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Create Gin router
	router := gin.New()

	// Apply middleware
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())

	// Load HTML templates
	router.LoadHTMLGlob(cfg.Web.TemplatesDir + "/*.html")

	// Serve static files (if you have any)
	if _, err := os.Stat(cfg.Web.StaticDir); err == nil {
		router.Static("/static", cfg.Web.StaticDir)
	}

	// Setup routes
	setupRoutes(router, handler)

	// Create HTTP server
	addr := fmt.Sprintf(":%d", cfg.Web.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting web server on http://localhost%s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func setupRoutes(router *gin.Engine, h *handlers.Handler) {
	// Health check
	router.GET("/health", h.HealthCheck)

	// Dashboard home
	router.GET("/", h.GetDashboard)

	// API routes
	api := router.Group("/api")
	{
		// Pods
		api.GET("/pods", h.GetPods)
		api.GET("/pod/:namespace/:name", h.GetPodDetail)

		// Recommendations
		api.GET("/recommendations", h.GetRecommendations)
		api.GET("/recommendations/:id/yaml", h.GetRecommendationYAML)
		api.POST("/recommendations/:id/apply", h.ApplyRecommendation)

		// Statistics
		api.GET("/stats", h.GetStats)

		// Namespaces
		api.GET("/namespaces", h.GetNamespaces)
	}

	// 404 handler
	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "Page not found",
		})
	})
}
