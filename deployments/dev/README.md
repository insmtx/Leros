# SingerOS Development Environment

Docker Compose-based development environment for SingerOS with offset ports and independent component management.

## Quick Start

### Initial Setup (First Time Only)

```bash
cd deployments/dev
./dev-setup.sh
```

Edit `.env` file and set your `LLM_API_KEY` and other configuration.

### Start Environment

Start infrastructure only (PostgreSQL, NATS, Redis):
```bash
./dev-start.sh --infra-only
```

Start with application services:
```bash
./dev-start.sh
```

Rebuild and start:
```bash
./dev-start.sh --build
```

### Start Individual Components

After starting infrastructure (`--infra-only`), start components independently:

```bash
# Start server
./dev-server.sh

# Start worker
./dev-worker.sh

# Start frontend (requires Node.js)
./dev-frontend.sh
```

Each component supports `--build` flag to rebuild before starting.

### View Status

```bash
./dev-status.sh
```

### View Logs

```bash
# All services
./dev-logs.sh

# Follow logs
./dev-logs.sh -f

# Specific service
./dev-logs.sh -f postgresql
```

### Stop Environment

```bash
./dev-stop.sh
```

### Restart Environment

```bash
./dev-restart.sh
./dev-restart.sh --build    # Rebuild before restart
```

## Service Ports

| Service    | Host Port | Container Port |
|------------|-----------|----------------|
| API Server | 8081      | 8080           |
| PostgreSQL | 5433      | 5432           |
| NATS       | 4223      | 4222           |
| NATS Mon.  | 8223      | 8222           |
| Redis      | 6380      | 6379           |

## Configuration Files

- `.env.example` - Environment variables template (copy to `.env`)
- `config.example.yaml` - Application config template (copy to `config.yaml`)

## Architecture

The dev environment separates infrastructure from application services:

1. **Infrastructure** (docker-compose): PostgreSQL, NATS, Redis
2. **Application** (individual scripts): Server, Worker, Frontend

This allows developers to:
- Run infrastructure in containers
- Run application code locally for debugging
- Start/stop components independently

## Makefile Commands

From project root:

```bash
make dev-setup     # Initial setup
make dev-start     # Start all services
make dev-stop      # Stop all services
make dev-status    # Check service status
make dev-logs      # View logs
make dev-server    # Start server only
make dev-worker    # Start worker only
make dev-frontend  # Start frontend only
```
