#!/bin/bash

# K8s Resource Optimizer - Startup Script
# Usage: ./start.sh [context]

set -e

PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$PROJECT_DIR"

# Default context (change this to your cluster)
KUBE_CONTEXT="${1:-}"
COLLECTION_INTERVAL="${COLLECTION_INTERVAL:-5m}"
WEB_PORT="${WEB_PORT:-8080}"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  K8s Resource Optimizer${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# Check if PostgreSQL is running
echo -e "${YELLOW}[1/4] Checking PostgreSQL...${NC}"
if ! docker ps | grep -q k8s-optimizer-db; then
    echo "Starting PostgreSQL..."
    docker-compose up -d postgres
    sleep 5
fi
echo -e "${GREEN}✓ PostgreSQL is running${NC}"

# Build the applications
echo -e "${YELLOW}[2/4] Building applications...${NC}"
go build -o bin/web-server cmd/web/main.go
go build -o bin/collector cmd/collector/main.go
echo -e "${GREEN}✓ Build complete${NC}"

# Start web server
echo -e "${YELLOW}[3/4] Starting web server on port ${WEB_PORT}...${NC}"
./bin/web-server > logs/web-server.log 2>&1 &
WEB_PID=$!
echo $WEB_PID > .web-server.pid
sleep 2

if ps -p $WEB_PID > /dev/null; then
    echo -e "${GREEN}✓ Web server started (PID: $WEB_PID)${NC}"
else
    echo -e "${RED}✗ Failed to start web server${NC}"
    exit 1
fi

# Start collector
echo -e "${YELLOW}[4/4] Starting metrics collector (interval: ${COLLECTION_INTERVAL})...${NC}"
./bin/collector --context "$KUBE_CONTEXT" --interval "$COLLECTION_INTERVAL" > logs/collector.log 2>&1 &
COLLECTOR_PID=$!
echo $COLLECTOR_PID > .collector.pid
sleep 2

if ps -p $COLLECTOR_PID > /dev/null; then
    echo -e "${GREEN}✓ Collector started (PID: $COLLECTOR_PID)${NC}"
else
    echo -e "${RED}✗ Failed to start collector${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  All services started successfully!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "Dashboard: ${GREEN}http://localhost:${WEB_PORT}${NC}"
echo -e "Cluster:   ${GREEN}${KUBE_CONTEXT}${NC}"
echo -e "Interval:  ${GREEN}${COLLECTION_INTERVAL}${NC}"
echo ""
echo "Logs:"
echo "  - Web server: logs/web-server.log"
echo "  - Collector:  logs/collector.log"
echo ""
echo "To stop all services: ./stop.sh"
echo ""

