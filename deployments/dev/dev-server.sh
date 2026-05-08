#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$SCRIPT_DIR/../.."

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}Starting SingerOS Dev Server...${NC}"

if ! docker ps --format '{{.Names}}' | grep -q "singer-dev-postgresql"; then
    echo -e "${RED}Error: Infrastructure not running. Start it first with:${NC}"
    echo "  docker-compose -f docker-compose.dev.yml up -d"
    exit 1
fi

if [[ "$@" == *"--build"* ]]; then
    echo -e "${BLUE}Building Docker image...${NC}"
    cd "$ROOT_DIR"
    make docker-build
    docker tag registry.yygu.cn/insmtx/SingerOS:latest localhost/dev_singer:latest
fi

docker stop singer-dev-server 2>/dev/null || true
docker rm singer-dev-server 2>/dev/null || true

echo -e "${BLUE}Starting server container...${NC}"
docker run -d \
    --name singer-dev-server \
    --network singer-dev-network \
    -p 8081:8080 \
    -e DATABASE_URL=postgres://singer_dev_user:singer_dev_password@singer-dev-postgresql:5432/singer_dev_db \
    -e NATS_URL=nats://singer-dev-nats:4222 \
    -e REDIS_URL=redis://:redis_dev_password@singer-dev-redis:6379 \
    -e GIN_MODE=debug \
    -v "$SCRIPT_DIR/server.config.yaml:/app/config.yaml" \
    localhost/dev_singer:latest \
    server --config /app/config.yaml

echo ""
echo -e "${GREEN}Server started!${NC}"
echo "  URL: http://localhost:8081"
echo "  View logs: docker logs -f singer-dev-server"
echo "  Stop:      docker stop singer-dev-server"
