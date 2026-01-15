package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func NewDB(connectionString string) (*DB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db}, nil
}

func (db *DB) InitSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS pods (
		id SERIAL PRIMARY KEY,
		namespace VARCHAR(255) NOT NULL,
		pod_name VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(namespace, pod_name)
	);

	CREATE TABLE IF NOT EXISTS containers (
		id SERIAL PRIMARY KEY,
		pod_id INTEGER REFERENCES pods(id) ON DELETE CASCADE,
		container_name VARCHAR(255) NOT NULL,
		image VARCHAR(512),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(pod_id, container_name)
	);

	CREATE TABLE IF NOT EXISTS metrics_snapshots (
		id SERIAL PRIMARY KEY,
		container_id INTEGER REFERENCES containers(id) ON DELETE CASCADE,
		timestamp TIMESTAMP NOT NULL,
		cpu_usage DOUBLE PRECISION NOT NULL,
		memory_usage BIGINT NOT NULL,
		UNIQUE(container_id, timestamp)
	);

	CREATE TABLE IF NOT EXISTS resource_requests (
		id SERIAL PRIMARY KEY,
		container_id INTEGER REFERENCES containers(id) ON DELETE CASCADE,
		cpu_request DOUBLE PRECISION,
		cpu_limit DOUBLE PRECISION,
		mem_request BIGINT,
		mem_limit BIGINT,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS analyses (
		id SERIAL PRIMARY KEY,
		container_id INTEGER REFERENCES containers(id) ON DELETE CASCADE,
		analyzed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		window_start TIMESTAMP NOT NULL,
		window_end TIMESTAMP NOT NULL,
		avg_cpu DOUBLE PRECISION,
		max_cpu DOUBLE PRECISION,
		p95_cpu DOUBLE PRECISION,
		p99_cpu DOUBLE PRECISION,
		avg_memory BIGINT,
		max_memory BIGINT,
		p95_memory BIGINT,
		p99_memory BIGINT,
		current_cpu_request DOUBLE PRECISION,
		current_mem_request BIGINT,
		recommended_cpu DOUBLE PRECISION,
		recommended_memory BIGINT,
		cpu_waste_percent DOUBLE PRECISION,
		memory_waste_percent DOUBLE PRECISION,
		monthly_savings DOUBLE PRECISION,
		status VARCHAR(50),
		confidence VARCHAR(50)
	);

	CREATE TABLE IF NOT EXISTS recommendations (
		id SERIAL PRIMARY KEY,
		analysis_id INTEGER REFERENCES analyses(id) ON DELETE CASCADE,
		namespace VARCHAR(255),
		pod_name VARCHAR(255),
		container_name VARCHAR(255),
		current_cpu DOUBLE PRECISION,
		current_memory BIGINT,
		recommended_cpu DOUBLE PRECISION,
		recommended_memory BIGINT,
		monthly_savings DOUBLE PRECISION,
		confidence VARCHAR(50),
		status VARCHAR(50),
		reason TEXT,
		applied BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_pods_namespace ON pods(namespace);
	CREATE INDEX IF NOT EXISTS idx_metrics_timestamp ON metrics_snapshots(timestamp);
	CREATE INDEX IF NOT EXISTS idx_analyses_status ON analyses(status);
	CREATE INDEX IF NOT EXISTS idx_recommendations_applied ON recommendations(applied);
	`

	_, err := db.Exec(schema)
	return err
}

