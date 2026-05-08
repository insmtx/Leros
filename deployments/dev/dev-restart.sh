#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

BUILD=false

while [[ "$#" -gt 0 ]]; do
    case $1 in
        --build) BUILD=true ;;
        --help) 
            echo "Usage: $0 [--build]"
            echo "  --build    Rebuild Docker image before restart"
            exit 0
            ;;
        *) 
            echo "Unknown parameter: $1"
            exit 1
            ;;
    esac
    shift
done

echo -e "${BLUE}Restarting SingerOS Dev Environment...${NC}"
echo ""

"$SCRIPT_DIR/dev-stop.sh"
echo ""

if [ "$BUILD" = true ]; then
    "$SCRIPT_DIR/dev-start.sh" --build
else
    "$SCRIPT_DIR/dev-start.sh"
fi
