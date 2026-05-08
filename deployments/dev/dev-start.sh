#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_FILE="$SCRIPT_DIR/docker-compose.dev.yml"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

parse_args() {
    BUILD=false
    INFRA_ONLY=false

    while [[ "$#" -gt 0 ]]; do
        case $1 in
            --build) BUILD=true ;;
            --infra-only) INFRA_ONLY=true ;;
            --help) 
                echo -e "${BLUE}SingerOS Dev Environment Startup${NC}"
                echo ""
                echo "Usage: $0 [OPTIONS]"
                echo ""
                echo "Options:"
                echo "  --build       Build Docker image from source before starting"
                echo "  --infra-only  Start only infrastructure services (Postgres, NATS, Redis)"
                echo "  --help        Show this help message"
                echo ""
                echo "Examples:"
                echo "  $0                    # Start with pre-built image"
                echo "  $0 --build            # Rebuild image then start"
                echo "  $0 --infra-only       # Start only infrastructure for local debugging"
                exit 0
                ;;
            *) 
                echo -e "${RED}Unknown parameter: $1${NC}"
                exit 1
                ;;
        esac
        shift
    done
}

check_config_files() {
    if [ ! -f "$SCRIPT_DIR/.env" ]; then
        echo -e "${YELLOW}Warning: .env file not found. Creating from .env.example...${NC}"
        cp "$SCRIPT_DIR/.env.example" "$SCRIPT_DIR/.env"
        echo -e "${YELLOW}Please edit .env file and set your environment variables.${NC}"
    fi

    if [ ! -f "$SCRIPT_DIR/server.config.yaml" ]; then
        echo -e "${YELLOW}Warning: server.config.yaml not found. Creating from template...${NC}"
        cp "$SCRIPT_DIR/server.config.example.yaml" "$SCRIPT_DIR/server.config.yaml"
        echo -e "${YELLOW}Please edit server.config.yaml file and set your configuration.${NC}"
    fi

    if [ ! -f "$SCRIPT_DIR/worker.config.yaml" ]; then
        echo -e "${YELLOW}Warning: worker.config.yaml not found. Creating from template...${NC}"
        cp "$SCRIPT_DIR/worker.config.example.yaml" "$SCRIPT_DIR/worker.config.yaml"
        echo -e "${YELLOW}Please edit worker.config.yaml file and set your configuration.${NC}"
    fi
}

build_image() {
    if [ "$BUILD" = true ]; then
        echo -e "${BLUE}Building Docker image...${NC}"
        cd "$SCRIPT_DIR/../.."
        make docker-build
        if [ $? -ne 0 ]; then
            echo -e "${RED}Build failed!${NC}"
            exit 1
        fi
        docker tag registry.yygu.cn/insmtx/SingerOS:latest localhost/dev_singer:latest
        echo -e "${GREEN}Docker image built successfully${NC}"
        echo ""
        cd "$SCRIPT_DIR"
    fi
}

start_infrastructure() {
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

    if [ "$INFRA_ONLY" = true ]; then
        echo -e "${BLUE}Infrastructure-only mode. Start application components manually as needed.${NC}"
        echo ""
        echo -e "${YELLOW}Available services:${NC}"
        echo "  - PostgreSQL: localhost:5433"
        echo "  - NATS:       localhost:4223 (monitoring: 8223)"
        echo "  - Redis:      localhost:6380"
        exit 0
    fi
}

show_info() {
    echo ""
    echo -e "${GREEN}Dev environment is ready!${NC}"
    echo ""
    echo -e "${BLUE}Quick Commands:${NC}"
    echo "  View logs:     ./dev-logs.sh"
    echo "  Check status:  ./dev-status.sh"
    echo "  Stop:          ./dev-stop.sh"
    echo "  Restart:       ./dev-restart.sh"
    echo "  Start server:  ./dev-server.sh"
    echo "  Start worker:  ./dev-worker.sh"
    echo "  Start frontend: ./dev-frontend.sh"
    echo ""
    echo -e "${BLUE}Service Ports:${NC}"
    echo "  - API Server:  http://localhost:8081"
    echo "  - PostgreSQL:  localhost:5433"
    echo "  - NATS:        localhost:4223 (monitoring: 8223)"
    echo "  - Redis:       localhost:6380"
}

main() {
    parse_args "$@"

    echo -e "${BLUE}╔══════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║  SingerOS Development Environment        ║${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════════╝${NC}"
    echo ""

    check_config_files
    build_image
    start_infrastructure
    show_info
}

main "$@"
