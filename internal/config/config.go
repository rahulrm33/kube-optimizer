package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Database   DatabaseConfig
	Kubernetes KubernetesConfig
	Analysis   AnalysisConfig
	Web        WebConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type KubernetesConfig struct {
	InCluster  bool
	ConfigPath string
}

type AnalysisConfig struct {
	WindowDays         int
	CollectionInterval time.Duration
	CPUCostPerCore     float64
	MemoryCostPerGB    float64
}

type WebConfig struct {
	Port         int
	TemplatesDir string
	StaticDir    string
}

func Load() (*Config, error) {
	cfg := &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "k8s_optimizer"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Kubernetes: KubernetesConfig{
			InCluster:  getEnvBool("K8S_IN_CLUSTER", false),
			ConfigPath: getEnv("KUBECONFIG", ""),
		},
		Analysis: AnalysisConfig{
			WindowDays:         getEnvInt("ANALYSIS_WINDOW_DAYS", 7),
			CollectionInterval: time.Duration(getEnvInt("COLLECTION_INTERVAL_MINUTES", 5)) * time.Minute,
			CPUCostPerCore:     getEnvFloat("CPU_COST_PER_CORE", 30.0),
			MemoryCostPerGB:    getEnvFloat("MEMORY_COST_PER_GB", 10.0),
		},
		Web: WebConfig{
			Port:         getEnvInt("WEB_PORT", 8080),
			TemplatesDir: getEnv("TEMPLATES_DIR", "web/templates"),
			StaticDir:    getEnv("STATIC_DIR", "web/static"),
		},
	}

	return cfg, nil
}

func (c *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

