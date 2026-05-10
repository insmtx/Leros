#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$SCRIPT_DIR/../.."

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}Starting Leros Dev Frontend...${NC}"

if [ ! -d "$ROOT_DIR/frontend" ]; then
    echo -e "${RED}Error: Frontend directory not found at: $ROOT_DIR/frontend${NC}"
    exit 1
fi

echo -e "${YELLOW}Frontend development requires Node.js and npm${NC}"
echo -e "${BLUE}Starting frontend dev server...${NC}"

cd "$ROOT_DIR/frontend"

if [ ! -d "node_modules" ]; then
    echo -e "${YELLOW}Installing dependencies...${NC}"
    npm install
fi

echo -e "${GREEN}Starting frontend dev server...${NC}"
echo -e "${BLUE}Note: Configure frontend to connect to backend at http://localhost:8081${NC}"

npm run dev

echo ""
echo "Frontend should be available at http://localhost:3000 (or as configured)"
