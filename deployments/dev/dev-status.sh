#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}╔══════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  SingerOS Dev Environment Status         ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════╝${NC}"
echo ""

cd "$SCRIPT_DIR"

echo -e "${BLUE}Infrastructure Services:${NC}"
for service in postgresql nats redis; do
    container="singer-dev-$service"
    if docker ps -q -f name=$container 2>/dev/null | grep -q .; then
        health=$(docker inspect --format='{{.State.Health.Status}}' $container 2>/dev/null || echo "no healthcheck")
        if [ "$health" = "healthy" ]; then
            echo -e "  $service: ${GREEN}healthy${NC}"
        else
            echo -e "  $service: ${YELLOW}$health${NC}"
        fi
    else
        echo -e "  $service: ${RED}stopped${NC}"
    fi
done

echo ""

echo -e "${BLUE}Application Services:${NC}"
for app in server worker frontend; do
    container="singer-dev-$app"
    if docker ps -q -f name=$container 2>/dev/null | grep -q .; then
        echo -e "  $app: ${GREEN}running${NC}"
    else
        echo -e "  $app: ${RED}stopped${NC}"
    fi
done

echo ""

echo -e "${BLUE}Port Mappings:${NC}"
echo "  API Server:  http://localhost:8081"
echo "  PostgreSQL:  localhost:5433"
echo "  NATS:        localhost:4223 (monitoring: 8223)"
echo "  Redis:       localhost:6380"
echo ""

echo -e "${BLUE}Configuration:${NC}"
if [ -f "$SCRIPT_DIR/.env" ]; then
    echo -e "  .env:       ${GREEN}exists${NC}"
else
    echo -e "  .env:       ${RED}missing (copy from .env.example)${NC}"
fi

if [ -f "$SCRIPT_DIR/server.config.yaml" ]; then
    echo -e "  server.config.yaml: ${GREEN}exists${NC}"
else
    echo -e "  server.config.yaml: ${RED}missing (copy from server.config.example.yaml)${NC}"
fi

if [ -f "$SCRIPT_DIR/worker.config.yaml" ]; then
    echo -e "  worker.config.yaml: ${GREEN}exists${NC}"
else
    echo -e "  worker.config.yaml: ${RED}missing (copy from worker.config.example.yaml)${NC}"
fi

echo ""
