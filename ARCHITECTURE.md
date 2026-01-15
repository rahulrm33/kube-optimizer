# Architecture Documentation

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Web Browser                              │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              Dashboard UI (HTML/JS/CSS)                   │  │
│  │  • Bootstrap 5 • Chart.js • Vanilla JavaScript            │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              ▲ │
                              │ │ HTTP/JSON
                              │ ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Gin Web Server (Go)                         │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    Middleware Layer                       │  │
│  │  • Logger • Recovery • CORS • (BasicAuth)                 │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    HTTP Handlers                          │  │
│  │  • Dashboard • API Endpoints • YAML Generation            │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                  Repository Layer                         │  │
│  │  • GetPods • GetStats • GetRecommendations                │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              ▲ │
                              │ │ SQL
                              │ ▼
┌─────────────────────────────────────────────────────────────────┐
│                      PostgreSQL Database                         │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Tables:                                                  │  │
│  │  • pods • containers • metrics_snapshots                  │  │
│  │  • resource_requests • analyses • recommendations         │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## Request Flow

### 1. Dashboard Page Load (GET /)

```
Browser
  │
  ├─→ GET /
  │
  ▼
Gin Router
  │
  ├─→ handlers.GetDashboard()
  │     │
  │     ├─→ repo.GetStatistics()
  │     │     │
  │     │     └─→ PostgreSQL: Aggregate queries
  │     │
  │     ├─→ repo.GetPods(status="over-provisioned", limit=10)
  │     │     │
  │     │     └─→ PostgreSQL: SELECT with JOIN
  │     │
  │     └─→ Render dashboard.html template
  │
  ▼
Browser: Display dashboard with data
```

### 2. API Request (GET /api/pods)

```
Browser (AJAX)
  │
  ├─→ GET /api/pods?namespace=production&status=over-provisioned
  │
  ▼
Gin Router
  │
  ├─→ Middleware: Logger, CORS
  │
  ├─→ handlers.GetPods()
  │     │
  │     ├─→ Parse query parameters
  │     │
  │     ├─→ repo.GetPods(namespace, status, sortBy, limit)
  │     │     │
  │     │     └─→ PostgreSQL: Filtered SELECT
  │     │
  │     └─→ Return JSON response
  │
  ▼
Browser: Update table with filtered data
```

### 3. Download YAML (GET /api/recommendations/:id/yaml)

```
Browser
  │
  ├─→ GET /api/recommendations/1/yaml
  │
  ▼
Gin Router
  │
  ├─→ handlers.GetRecommendationYAML()
  │     │
  │     ├─→ repo.GetRecommendationByID(1)
  │     │     │
  │     │     └─→ PostgreSQL: SELECT by ID
  │     │
  │     ├─→ generateResourcePatch(recommendation)
  │     │
  │     ├─→ yaml.Marshal(patch)
  │     │
  │     └─→ Return YAML with Content-Disposition header
  │
  ▼
Browser: Download patch-namespace-pod-container.yaml
```

## Component Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                          cmd/web/                                │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  main.go                                                  │  │
│  │  • Load config                                            │  │
│  │  • Connect to database                                    │  │
│  │  • Initialize Gin router                                  │  │
│  │  • Setup routes and middleware                            │  │
│  │  • Start server                                           │  │
│  │  • Handle graceful shutdown                               │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ uses
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                        internal/                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  config/config.go                                         │  │
│  │  • Environment variable parsing                           │  │
│  │  • Configuration structs                                  │  │
│  │  • Default values                                         │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  database/db.go                                           │  │
│  │  • PostgreSQL connection                                  │  │
│  │  • Schema initialization                                  │  │
│  │  • Table creation                                         │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  models/models.go                                         │  │
│  │  • Data structures                                        │  │
│  │  • JSON tags                                              │  │
│  │  • View models                                            │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  repository/repository.go                                 │  │
│  │  • Database queries                                       │  │
│  │  • Data access methods                                    │  │
│  │  • SQL with parameterization                              │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ used by
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                           web/                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  handlers/handlers.go                                     │  │
│  │  • HTTP request handlers                                  │  │
│  │  • Business logic                                         │  │
│  │  • Response formatting                                    │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  middleware/middleware.go                                 │  │
│  │  • Request logging                                        │  │
│  │  • Error recovery                                         │  │
│  │  • CORS headers                                           │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  templates/                                               │  │
│  │  • dashboard.html - Main UI                               │  │
│  │  • error.html - Error page                                │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## Database Schema

```
┌──────────────────┐
│      pods        │
├──────────────────┤
│ id (PK)          │
│ namespace        │
│ pod_name         │◄─────┐
│ created_at       │      │
│ updated_at       │      │
└──────────────────┘      │
         │                │
         │ 1:N            │
         ▼                │
┌──────────────────┐      │
│   containers     │      │
├──────────────────┤      │
│ id (PK)          │      │
│ pod_id (FK)      │──────┘
│ container_name   │◄─────┐
│ image            │      │
│ created_at       │      │
│ updated_at       │      │
└──────────────────┘      │
         │                │
         │ 1:N            │
         ├────────────────┼────────────┐
         ▼                │            │
┌──────────────────┐      │            │
│metrics_snapshots │      │            │
├──────────────────┤      │            │
│ id (PK)          │      │            │
│ container_id(FK) │──────┘            │
│ timestamp        │                   │
│ cpu_usage        │                   │
│ memory_usage     │                   │
└──────────────────┘                   │
                                       │
         ┌─────────────────────────────┤
         ▼                             │
┌──────────────────┐                   │
│resource_requests │                   │
├──────────────────┤                   │
│ id (PK)          │                   │
│ container_id(FK) │───────────────────┘
│ cpu_request      │
│ cpu_limit        │
│ mem_request      │
│ mem_limit        │
│ updated_at       │
└──────────────────┘
         │
         │ 1:1
         ▼
┌──────────────────┐
│    analyses      │
├──────────────────┤
│ id (PK)          │
│ container_id(FK) │
│ analyzed_at      │
│ window_start     │
│ window_end       │
│ avg_cpu          │
│ max_cpu          │
│ p95_cpu          │
│ p99_cpu          │
│ avg_memory       │
│ max_memory       │
│ p95_memory       │
│ p99_memory       │
│ current_cpu_req  │
│ current_mem_req  │
│ recommended_cpu  │
│ recommended_mem  │
│ cpu_waste_%      │
│ memory_waste_%   │
│ monthly_savings  │
│ status           │
│ confidence       │
└──────────────────┘
         │
         │ 1:1
         ▼
┌──────────────────┐
│recommendations   │
├──────────────────┤
│ id (PK)          │
│ analysis_id (FK) │
│ namespace        │
│ pod_name         │
│ container_name   │
│ current_cpu      │
│ current_memory   │
│ recommended_cpu  │
│ recommended_mem  │
│ monthly_savings  │
│ confidence       │
│ status           │
│ reason           │
│ applied          │
│ created_at       │
└──────────────────┘
```

## Data Flow

### Metrics Collection (Future Implementation)

```
Kubernetes Cluster
  │
  ├─→ Metrics Server API
  │     │
  │     └─→ Pod CPU/Memory metrics
  │
  ▼
Collector Service (Not Implemented)
  │
  ├─→ Fetch pod list
  ├─→ Fetch resource requests/limits
  ├─→ Fetch metrics for each container
  │
  ▼
Database
  │
  ├─→ INSERT INTO pods
  ├─→ INSERT INTO containers
  ├─→ INSERT INTO resource_requests
  └─→ INSERT INTO metrics_snapshots
```

### Analysis Process (Future Implementation)

```
Scheduler (Cron/Timer)
  │
  ▼
Analyzer Service (Not Implemented)
  │
  ├─→ Query metrics_snapshots (last 7 days)
  │
  ├─→ Calculate statistics:
  │     • Average CPU/Memory
  │     • Max CPU/Memory
  │     • P95/P99 percentiles
  │
  ├─→ Compare with current requests
  │
  ├─→ Generate recommendations:
  │     • Recommended resources = P95 * 1.2
  │     • Calculate waste percentage
  │     • Calculate monthly savings
  │     • Determine status (over/under/optimal)
  │     • Assign confidence level
  │
  ├─→ INSERT INTO analyses
  │
  └─→ INSERT INTO recommendations
```

### Current Implementation (Sample Data)

```
seed.go Script
  │
  ├─→ Generate random pods
  ├─→ Generate random containers
  ├─→ Generate random resource requests
  ├─→ Generate 100 metrics snapshots per container
  ├─→ Calculate statistics
  ├─→ Generate analyses
  └─→ Generate recommendations
  │
  ▼
Database: Populated with sample data
```

## API Architecture

### RESTful Endpoints

```
/
├── /                           [GET]  Dashboard HTML
├── /health                     [GET]  Health check
│
└── /api/
    ├── /pods                   [GET]  List pods (filtered)
    ├── /pod/:ns/:name          [GET]  Pod detail
    ├── /recommendations        [GET]  List recommendations
    ├── /recommendations/:id
    │   ├── /yaml               [GET]  Download YAML
    │   └── /apply              [POST] Mark applied
    ├── /stats                  [GET]  Statistics
    └── /namespaces             [GET]  Namespace list
```

### Response Format

**Success**:
```json
{
  "pods": [...],
  "total": 50,
  "page": 1
}
```

**Error**:
```json
{
  "error": "Pod not found"
}
```

## Frontend Architecture

### Component Structure

```
dashboard.html
├── Navbar
│   ├── Title
│   ├── Last Updated
│   ├── Refresh Button
│   └── Theme Toggle
│
├── Summary Cards
│   ├── Total Savings
│   ├── Over-provisioned
│   ├── Under-provisioned
│   └── Optimal
│
├── Charts Section
│   ├── Pie Chart (Status Distribution)
│   └── Bar Chart (Top 10 Wasteful)
│
├── Filters Section
│   ├── Namespace Dropdown
│   ├── Status Dropdown
│   ├── Sort By Dropdown
│   └── Search Box
│
└── Recommendations Table
    ├── Table Headers (sortable)
    ├── Table Body (dynamic)
    └── Action Buttons per Row
```

### JavaScript Modules

```javascript
// Global State
- allPods: []
- currentSort: {column, direction}
- autoRefreshInterval

// Initialization
- initializeCharts()
- loadNamespaces()
- loadRecommendations()
- startAutoRefresh()
- loadTheme()

// Data Loading
- loadRecommendations()
- applyFilters()
- handleSearch()
- refreshData()

// Rendering
- renderTable(pods)
- updateCharts(stats)

// Actions
- viewDetails(namespace, podName)
- downloadYAML(namespace, podName)
- toggleTheme()

// Utilities
- showLoading(show)
- showToast(message, type)
- formatCPU(millicores)
- formatMemory(bytes)
```

## Deployment Architecture

### Docker Compose

```
┌─────────────────────────────────────┐
│         Docker Network              │
│                                     │
│  ┌──────────────────────────────┐  │
│  │   PostgreSQL Container       │  │
│  │   • postgres:15-alpine       │  │
│  │   • Port: 5432               │  │
│  │   • Volume: postgres_data    │  │
│  └──────────────────────────────┘  │
│              ▲                      │
│              │ TCP                  │
│              ▼                      │
│  ┌──────────────────────────────┐  │
│  │   Web App Container          │  │
│  │   • Built from Dockerfile    │  │
│  │   • Port: 8080               │  │
│  │   • Depends on postgres      │  │
│  └──────────────────────────────┘  │
│              │                      │
└──────────────┼──────────────────────┘
               │
               │ HTTP
               ▼
         Host: localhost:8080
```

### Kubernetes (Future)

```
┌─────────────────────────────────────┐
│         Kubernetes Cluster          │
│                                     │
│  ┌──────────────────────────────┐  │
│  │   PostgreSQL StatefulSet     │  │
│  │   • Persistent Volume        │  │
│  │   • Service (ClusterIP)      │  │
│  └──────────────────────────────┘  │
│              ▲                      │
│              │                      │
│  ┌──────────────────────────────┐  │
│  │   Web App Deployment         │  │
│  │   • 3 replicas               │  │
│  │   • ConfigMap for config     │  │
│  │   • Secret for DB password   │  │
│  │   • Service (ClusterIP)      │  │
│  └──────────────────────────────┘  │
│              │                      │
│  ┌──────────────────────────────┐  │
│  │   Ingress                    │  │
│  │   • HTTPS/TLS                │  │
│  │   • Domain: optimizer.com    │  │
│  └──────────────────────────────┘  │
└─────────────────────────────────────┘
```

## Security Architecture

### Current Implementation

```
Browser
  │
  ├─→ HTTP Request
  │
  ▼
Gin Middleware Stack
  │
  ├─→ Logger (log all requests)
  │
  ├─→ Recovery (catch panics)
  │
  ├─→ CORS (allow cross-origin)
  │
  ├─→ [BasicAuth] (optional, commented out)
  │
  ▼
Handlers
  │
  ├─→ Input validation
  │
  ├─→ Parameterized SQL queries
  │
  └─→ Error handling
```

### Production Recommendations

```
Internet
  │
  ▼
Reverse Proxy (nginx/Traefik)
  │
  ├─→ HTTPS/TLS termination
  ├─→ Rate limiting
  ├─→ IP filtering
  │
  ▼
Load Balancer
  │
  ├─→ Multiple app instances
  │
  ▼
Web App
  │
  ├─→ Authentication (OAuth/OIDC)
  ├─→ Authorization (RBAC)
  ├─→ Audit logging
  │
  ▼
Database
  │
  └─→ SSL connection
      └─→ Encrypted at rest
```

## Monitoring Architecture (Future)

```
Application
  │
  ├─→ Prometheus Metrics
  │     • Request count
  │     • Response time
  │     • Error rate
  │
  ├─→ Structured Logging
  │     • JSON format
  │     • Log levels
  │     • Correlation IDs
  │
  └─→ Health Checks
        • /health endpoint
        • Database connectivity
        • Dependency checks
  │
  ▼
Monitoring Stack
  │
  ├─→ Prometheus (metrics)
  ├─→ Grafana (dashboards)
  ├─→ Loki (logs)
  └─→ AlertManager (alerts)
```

## Scalability Considerations

### Horizontal Scaling

```
Load Balancer
  │
  ├─→ Web App Instance 1
  ├─→ Web App Instance 2
  ├─→ Web App Instance 3
  │
  └─→ Shared PostgreSQL
        │
        └─→ Read Replicas (for queries)
```

### Caching Layer (Future)

```
Web App
  │
  ├─→ Redis Cache
  │     • Statistics (TTL: 5 min)
  │     • Namespace list (TTL: 1 hour)
  │     • Pod list (TTL: 1 min)
  │
  └─→ PostgreSQL (cache miss)
```

### Database Optimization

```
PostgreSQL
  │
  ├─→ Indexes on:
  │     • pods(namespace)
  │     • analyses(status)
  │     • metrics_snapshots(timestamp)
  │     • recommendations(applied)
  │
  ├─→ Partitioning:
  │     • metrics_snapshots by date
  │
  └─→ Archiving:
        • Old metrics to cold storage
```

---

## Summary

This architecture provides:

✅ **Separation of Concerns**: Clear layers (handlers, repository, database)  
✅ **Scalability**: Stateless web app, can scale horizontally  
✅ **Maintainability**: Well-organized code structure  
✅ **Extensibility**: Easy to add new features  
✅ **Performance**: Indexed queries, efficient data access  
✅ **Security**: Parameterized queries, middleware protection  

The implementation follows Go best practices and web development standards, making it production-ready with minor enhancements (authentication, HTTPS, monitoring).

