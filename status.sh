#!/bin/bash

# K8s Resource Optimizer - Status Script

PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$PROJECT_DIR"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}  K8s Resource Optimizer Status${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""

# Check PostgreSQL
echo -n "PostgreSQL:  "
if docker ps | grep -q k8s-optimizer-db; then
    echo -e "${GREEN}Running ✓${NC}"
else
    echo -e "${RED}Stopped ✗${NC}"
fi

# Check Web Server
echo -n "Web Server:  "
if [ -f .web-server.pid ]; then
    PID=$(cat .web-server.pid)
    if ps -p $PID > /dev/null 2>&1; then
        echo -e "${GREEN}Running (PID: $PID) ✓${NC}"
    else
        echo -e "${RED}Stopped ✗${NC}"
    fi
else
    # Check if running without PID file
    if pgrep -f "bin/web-server" > /dev/null 2>&1; then
        echo -e "${GREEN}Running ✓${NC}"
    else
        echo -e "${RED}Stopped ✗${NC}"
    fi
fi

# Check Collector
echo -n "Collector:   "
if [ -f .collector.pid ]; then
    PID=$(cat .collector.pid)
    if ps -p $PID > /dev/null 2>&1; then
        echo -e "${GREEN}Running (PID: $PID) ✓${NC}"
    else
        echo -e "${RED}Stopped ✗${NC}"
    fi
else
    # Check if running without PID file
    if pgrep -f "bin/collector" > /dev/null 2>&1; then
        echo -e "${GREEN}Running ✓${NC}"
    elif pgrep -f "cmd/collector" > /dev/null 2>&1; then
        echo -e "${GREEN}Running (go run) ✓${NC}"
    else
        echo -e "${RED}Stopped ✗${NC}"
    fi
fi

echo ""

# Show recent log entries
echo -e "${YELLOW}Recent Collector Activity:${NC}"
if [ -f logs/collector.log ]; then
    tail -5 logs/collector.log 2>/dev/null || echo "  (no logs yet)"
elif [ -f collector.log ]; then
    tail -5 collector.log 2>/dev/null || echo "  (no logs yet)"
else
    echo "  (no logs yet)"
fi

echo ""
echo -e "Dashboard URL: ${GREEN}http://localhost:8080${NC}"
echo ""

