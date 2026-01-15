.PHONY: help build run test clean docker-build docker-run db-setup

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	@echo "Building k8s-optimizer..."
	@go build -o bin/k8s-optimizer cmd/web/main.go
	@echo "Build complete: bin/k8s-optimizer"

run: ## Run the web server
	@echo "Starting web server..."
	@go run cmd/web/main.go

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@echo "Clean complete"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies updated"

db-setup: ## Create database
	@echo "Creating database..."
	@createdb k8s_optimizer || echo "Database might already exist"
	@echo "Database setup complete"

db-drop: ## Drop database
	@echo "Dropping database..."
	@dropdb k8s_optimizer || echo "Database might not exist"
	@echo "Database dropped"

db-reset: db-drop db-setup ## Reset database

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t k8s-optimizer:latest .
	@echo "Docker image built"

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run -p 8080:8080 --env-file .env k8s-optimizer:latest

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run ./... || echo "Install golangci-lint: https://golangci-lint.run/usage/install/"

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Code formatted"

dev: ## Run in development mode with auto-reload
	@echo "Starting in development mode..."
	@air || echo "Install air: go install github.com/cosmtrek/air@latest"

sample-data: ## Insert sample data for testing
	@echo "Inserting sample data..."
	@go run scripts/seed.go
	@echo "Sample data inserted"

