#!/bin/bash

# K8s Resource Optimizer - Stop Script

PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$PROJECT_DIR"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}Stopping K8s Resource Optimizer...${NC}"

# Stop web server
if [ -f .web-server.pid ]; then
    PID=$(cat .web-server.pid)
    if ps -p $PID > /dev/null 2>&1; then
        kill $PID 2>/dev/null
        echo -e "${GREEN}✓ Web server stopped (PID: $PID)${NC}"
    fi
    rm -f .web-server.pid
fi

# Stop collector
if [ -f .collector.pid ]; then
    PID=$(cat .collector.pid)
    if ps -p $PID > /dev/null 2>&1; then
        kill $PID 2>/dev/null
        echo -e "${GREEN}✓ Collector stopped (PID: $PID)${NC}"
    fi
    rm -f .collector.pid
fi

# Kill any remaining processes
pkill -f "bin/web-server" 2>/dev/null || true
pkill -f "bin/collector" 2>/dev/null || true

echo -e "${GREEN}All services stopped.${NC}"
echo ""
echo "To stop PostgreSQL: docker-compose down"

