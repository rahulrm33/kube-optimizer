# K8s Resource Optimizer

A comprehensive Kubernetes resource optimization tool that analyzes pod resource usage and provides recommendations to reduce waste and costs.

## Features

- 🔍 **Real-time Analysis**: Monitors CPU and memory usage of all pods in your cluster
- 💰 **Cost Savings**: Calculates potential monthly savings from resource optimization
- 📊 **Beautiful Dashboard**: Modern web UI with charts and visualizations
- 🎯 **Smart Recommendations**: Provides confidence-rated recommendations for resource adjustments
- 📥 **Easy Apply**: Download YAML patches to apply recommendations
- 🔄 **Auto-refresh**: Dashboard updates automatically every 5 minutes
- 🌙 **Dark Mode**: Toggle between light and dark themes

## Architecture

```
k8s-optimizer/
├── cmd/
│   └── web/
│       └── main.go              # Web server entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── database/
│   │   └── db.go                # Database connection and schema
│   ├── models/
│   │   └── models.go            # Data models
│   └── repository/
│       └── repository.go        # Data access layer
├── web/
│   ├── handlers/
│   │   └── handlers.go          # HTTP handlers
│   ├── middleware/
│   │   └── middleware.go        # Middleware (logging, CORS, etc.)
│   └── templates/
│       ├── dashboard.html       # Main dashboard UI
│       └── error.html           # Error page
└── go.mod                       # Go dependencies
```

## Prerequisites

- Go 1.21 or later
- PostgreSQL 12 or later
- Kubernetes cluster access (optional for data collection)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/scaleops/k8s-optimizer.git
cd k8s-optimizer
```

2. Install dependencies:
```bash
go mod download
```

3. Set up PostgreSQL database:
```bash
createdb k8s_optimizer
```

4. Configure environment variables:
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=k8s_optimizer
export WEB_PORT=8080
```

## Running the Application

### Start the Web Server

```bash
go run cmd/web/main.go
```

The web dashboard will be available at http://localhost:8080

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | `postgres` |
| `DB_NAME` | Database name | `k8s_optimizer` |
| `DB_SSLMODE` | SSL mode | `disable` |
| `WEB_PORT` | Web server port | `8080` |
| `TEMPLATES_DIR` | Templates directory | `web/templates` |
| `STATIC_DIR` | Static files directory | `web/static` |
| `CPU_COST_PER_CORE` | Cost per CPU core/month | `30.0` |
| `MEMORY_COST_PER_GB` | Cost per GB memory/month | `10.0` |

## API Endpoints

### Dashboard
- `GET /` - Dashboard home page

### API
- `GET /api/pods` - List all analyzed pods
  - Query params: `namespace`, `status`, `sort_by`, `limit`, `search`
  
- `GET /api/pod/:namespace/:name` - Get pod details
  
- `GET /api/recommendations` - Get all recommendations
  - Query params: `confidence`, `min_savings`, `limit`
  
- `GET /api/recommendations/:id/yaml` - Download YAML patch for recommendation
  
- `POST /api/recommendations/:id/apply` - Mark recommendation as applied
  - Body: `{"applied": true}`
  
- `GET /api/stats` - Get overall statistics
  
- `GET /api/namespaces` - Get all namespaces

### Health
- `GET /health` - Health check endpoint

## Dashboard Features

### Summary Cards
- **Total Monthly Savings**: Aggregate potential savings
- **Over-provisioned Pods**: Count of pods using less than requested
- **Under-provisioned Pods**: Count of pods requesting more resources
- **Optimal Pods**: Count of well-configured pods

### Charts
- **Pie Chart**: Visual breakdown of pods by status
- **Bar Chart**: Top 10 wasteful pods by savings potential

### Filters
- Filter by namespace
- Filter by status (over/under/optimal)
- Sort by savings, waste %, or pod name
- Search by pod name

### Recommendations Table
Shows detailed information for each pod:
- Namespace, Pod, Container names
- Current vs Recommended resources
- Monthly savings potential
- Confidence level
- Status badge
- Actions: View details, Download YAML

### Actions
- **View Details**: Opens detailed view of pod metrics and history
- **Download YAML**: Downloads a patch file with recommended resources
- **Auto-refresh**: Data refreshes every 5 minutes
- **Manual Refresh**: Click refresh button to update immediately
- **Dark Mode**: Toggle theme for comfortable viewing

## Database Schema

The application uses the following tables:

- `pods` - Kubernetes pods
- `containers` - Containers within pods
- `metrics_snapshots` - Historical resource usage data
- `resource_requests` - Current resource requests/limits
- `analyses` - Analysis results with recommendations
- `recommendations` - Generated recommendations

All tables are automatically created on first run.

## Development

### Build
```bash
go build -o k8s-optimizer cmd/web/main.go
```

### Run Tests (when implemented)
```bash
go test ./...
```

### Docker Build (when Dockerfile is added)
```bash
docker build -t k8s-optimizer .
docker run -p 8080:8080 k8s-optimizer
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Support

For issues and questions, please open an issue on GitHub.

