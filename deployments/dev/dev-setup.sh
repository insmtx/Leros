#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}SingerOS Dev Environment Setup${NC}"
echo ""

echo -e "${BLUE}Checking prerequisites...${NC}"

if ! command -v docker &> /dev/null; then
    echo -e "${RED}Docker is not installed${NC}"
    exit 1
fi
echo -e "${GREEN}Docker: $(docker --version)${NC}"

if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}Docker Compose is not installed${NC}"
    exit 1
fi
echo -e "${GREEN}Docker Compose: $(docker-compose --version)${NC}"

echo ""

if [ ! -f "$SCRIPT_DIR/.env" ]; then
    echo -e "${YELLOW}Creating .env file from template...${NC}"
    cp "$SCRIPT_DIR/.env.example" "$SCRIPT_DIR/.env"
    echo -e "${GREEN}.env file created${NC}"
    echo -e "${YELLOW}Please edit .env and set your LLM_API_KEY${NC}"
else
    echo -e "${GREEN}.env file already exists${NC}"
fi

if [ ! -f "$SCRIPT_DIR/server.config.yaml" ]; then
    echo -e "${YELLOW}Creating server.config.yaml from template...${NC}"
    cp "$SCRIPT_DIR/server.config.example.yaml" "$SCRIPT_DIR/server.config.yaml"
    echo -e "${GREEN}server.config.yaml created${NC}"
else
    echo -e "${GREEN}server.config.yaml already exists${NC}"
fi

if [ ! -f "$SCRIPT_DIR/worker.config.yaml" ]; then
    echo -e "${YELLOW}Creating worker.config.yaml from template...${NC}"
    cp "$SCRIPT_DIR/worker.config.example.yaml" "$SCRIPT_DIR/worker.config.yaml"
    echo -e "${GREEN}worker.config.yaml created${NC}"
else
    echo -e "${GREEN}worker.config.yaml already exists${NC}"
fi

echo ""

# Build Docker image if requested
BUILD=false
START_SERVICES=false

while [[ "$#" -gt 0 ]]; do
    case $1 in
        --build) BUILD=true ;;
        --start) START_SERVICES=true ;;
        --help) 
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --build    Build Docker image from source"
            echo "  --start    Start services after setup"
            exit 0
            ;;
        *) 
            echo "Unknown parameter: $1"
            exit 1
            ;;
    esac
    shift
done

if [ "$BUILD" = true ]; then
    echo -e "${BLUE}Building Docker image...${NC}"
    cd "$SCRIPT_DIR/../.."
    make docker-build
    docker tag registry.yygu.cn/insmtx/SingerOS:latest localhost/dev_singer:latest
    echo -e "${GREEN}Docker image built successfully${NC}"
    cd "$SCRIPT_DIR"
else
    echo -e "${BLUE}Pulling Docker images...${NC}"
    cd "$SCRIPT_DIR"
    docker-compose -f docker-compose.dev.yml pull
fi

echo ""
echo -e "${GREEN}Setup complete!${NC}"
echo ""

if [ "$START_SERVICES" = true ]; then
    echo -e "${BLUE}Starting infrastructure services...${NC}"
    docker-compose -f docker-compose.dev.yml up -d postgresql nats redis

    echo -e "${YELLOW}Waiting for services to be healthy...${NC}"
    sleep 5

    for service in postgresql nats redis; do
        echo -n "Checking $service... "
        for i in {1..30}; do
            status=$(docker inspect --format='{{.State.Health.Status}}' singer-dev-$service 2>/dev/null || echo "starting")
            if [ "$status" = "healthy" ]; then
                echo -e "${GREEN}healthy${NC}"
                break
            fi
            if [ $i -eq 30 ]; then
                echo -e "${RED}timeout${NC}"
            fi
            sleep 2
        done
    done

    echo ""
    echo -e "${GREEN}Infrastructure services started!${NC}"
    echo ""
else
    echo -e "${BLUE}Next steps:${NC}"
    echo "  1. Edit .env file and set your LLM_API_KEY"
    echo "  2. Edit server.config.yaml and worker.config.yaml with your GitHub app credentials (if needed)"
    echo "  3. Start infrastructure: docker-compose -f docker-compose.dev.yml up -d"
    echo "  4. Start server: ./dev-server.sh"
    echo "  5. Start worker: ./dev-worker.sh"
fi
