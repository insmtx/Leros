#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_FILE="$SCRIPT_DIR/docker-compose.dev.yml"

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}Stopping SingerOS Dev Environment...${NC}"
echo ""

cd "$SCRIPT_DIR"

echo -e "${BLUE}Stopping application containers...${NC}"
docker stop singer-dev-server singer-dev-worker singer-dev-frontend 2>/dev/null || true
docker rm singer-dev-server singer-dev-worker singer-dev-frontend 2>/dev/null || true

echo -e "${BLUE}Stopping infrastructure services...${NC}"
docker-compose -f docker-compose.dev.yml down

echo ""
echo -e "${GREEN}Dev environment stopped.${NC}"
echo ""
echo -e "${BLUE}Note: Volumes are preserved. To remove volumes, run:${NC}"
echo "  docker-compose -f $COMPOSE_FILE down -v"
