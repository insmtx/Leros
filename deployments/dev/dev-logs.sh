#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

FOLLOW=false
SERVICE=""
TAIL=100

while [[ "$#" -gt 0 ]]; do
    case $1 in
        -f|--follow) FOLLOW=true ;;
        -n|--lines) TAIL="$2"; shift ;;
        --help)
            echo "Usage: $0 [OPTIONS] [SERVICE]"
            echo ""
            echo "Options:"
            echo "  -f, --follow    Follow log output"
            echo "  -n, --lines     Number of lines to show"
            echo ""
            echo "Services:"
            echo "  postgresql, nats, redis"
            echo ""
            echo "Examples:"
            echo "  $0                    # Show recent logs from all services"
            echo "  $0 -f                 # Follow logs"
            echo "  $0 -f postgresql      # Follow PostgreSQL logs"
            echo "  $0 -n 50              # Show last 50 lines"
            exit 0
            ;;
        *) SERVICE="$1" ;;
    esac
    shift
done

cd "$SCRIPT_DIR"

CMD="docker-compose -f docker-compose.dev.yml logs"

if [ "$FOLLOW" = true ]; then
    CMD="$CMD -f"
fi

CMD="$CMD --tail=$TAIL"

if [ -n "$SERVICE" ]; then
    CMD="$CMD $SERVICE"
fi

eval "$CMD"
