# Quick Start Guide

Get the K8s Resource Optimizer dashboard up and running in 5 minutes!

## Prerequisites

- Go 1.21+ installed
- PostgreSQL 12+ installed and running
- (Optional) Docker and Docker Compose

## Option 1: Local Development

### 1. Clone and Setup

```bash
git clone <repository-url>
cd scaleops-own
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Setup Database

```bash
# Create the database
createdb k8s_optimizer

# Or use the Makefile
make db-setup
```

### 4. Configure Environment

```bash
# Copy example env file
cp .env.example .env

# Edit .env if needed (optional, defaults are fine for local dev)
```

### 5. Seed Sample Data

```bash
# Generate sample data for testing
go run scripts/seed.go
```

### 6. Start the Server

```bash
# Using go run
go run cmd/web/main.go

# Or using Makefile
make run

# Or build and run
make build
./bin/k8s-optimizer
```

### 7. Open Dashboard

Visit http://localhost:8080 in your browser.

You should see:
- Summary cards with statistics
- Charts showing pod distribution
- Table with recommendations

## Option 2: Docker Compose (Easiest!)

### 1. Start Everything

```bash
docker-compose up -d
```

This will:
- Start PostgreSQL database
- Build and start the web application
- Initialize the database schema
- Expose the dashboard on port 8080

### 2. Seed Sample Data

```bash
# Wait a few seconds for services to be ready, then:
docker-compose exec web sh -c "go run scripts/seed.go"
```

### 3. Open Dashboard

Visit http://localhost:8080

### Stop Services

```bash
docker-compose down
```

## Testing the Dashboard

### 1. Explore Summary Cards

At the top, you'll see:
- **Total Monthly Savings**: Potential cost savings
- **Over-provisioned**: Pods with excess resources
- **Under-provisioned**: Pods needing more resources
- **Optimal**: Well-configured pods

### 2. View Charts

- **Pie Chart**: Distribution of pod statuses
- **Bar Chart**: Top 10 pods by potential savings

### 3. Filter and Search

- Filter by namespace (dropdown)
- Filter by status (all/over/under/optimal)
- Sort by savings, waste %, or name
- Search for specific pods

### 4. Take Actions

For each recommendation:
- **👁 View**: See detailed metrics (opens detail view)
- **⬇ Download**: Get YAML patch file
- **✓ Apply**: Mark as applied (checkbox)

### 5. Test Auto-Refresh

- Dashboard auto-refreshes every 5 minutes
- Click "Refresh" button for manual update
- Watch for toast notifications

### 6. Try Dark Mode

Click the moon/sun icon in the top-right to toggle theme.

## API Testing

Test API endpoints directly:

```bash
# Get statistics
curl http://localhost:8080/api/stats

# Get all pods
curl http://localhost:8080/api/pods

# Get pods in specific namespace
curl "http://localhost:8080/api/pods?namespace=production"

# Get over-provisioned pods only
curl "http://localhost:8080/api/pods?status=over-provisioned"

# Search for pods
curl "http://localhost:8080/api/pods?search=api"

# Get recommendations
curl http://localhost:8080/api/recommendations

# Get namespaces
curl http://localhost:8080/api/namespaces

# Health check
curl http://localhost:8080/health
```

## Troubleshooting

### Database Connection Error

```bash
# Check PostgreSQL is running
pg_isready

# Check database exists
psql -l | grep k8s_optimizer

# Recreate database
make db-reset
```

### Port Already in Use

```bash
# Change port in .env
echo "WEB_PORT=8081" >> .env

# Or set environment variable
export WEB_PORT=8081
go run cmd/web/main.go
```

### No Data Showing

```bash
# Run the seed script
go run scripts/seed.go

# Or check database
psql k8s_optimizer -c "SELECT COUNT(*) FROM pods;"
```

### Template Not Found

Make sure you're running from the project root:

```bash
cd /path/to/scaleops-own
go run cmd/web/main.go
```

## Next Steps

### Connect to Real Kubernetes Cluster

To collect real metrics (not implemented in this MVP, but structure is ready):

1. Set up Kubernetes client
2. Implement metrics collection service
3. Schedule periodic collection
4. Run analysis on collected data

### Customize Cost Calculations

Edit `.env`:

```bash
# Set your actual cloud costs
CPU_COST_PER_CORE=35.0      # $/month per core
MEMORY_COST_PER_GB=12.0     # $/month per GB
```

### Enable Authentication

In `cmd/web/main.go`, uncomment the BasicAuth middleware:

```go
// Protect all routes except health check
api := router.Group("/api")
api.Use(middleware.BasicAuth("admin", "password"))
```

### Production Deployment

1. Build optimized binary:
   ```bash
   make build
   ```

2. Use Docker:
   ```bash
   make docker-build
   docker run -p 8080:8080 --env-file .env k8s-optimizer:latest
   ```

3. Deploy to Kubernetes (create deployment manifests)

## Need Help?

- Check `README.md` for detailed documentation
- Open an issue on GitHub
- Review the code - it's well-commented!

## Clean Up

```bash
# Stop services
docker-compose down

# Drop database
make db-drop

# Clean build artifacts
make clean
```

Happy optimizing! 🚀

