# SingerOS Development Environment

Docker Compose-based development environment for SingerOS with offset ports and independent component management.

## Quick Start

### Initial Setup (First Time Only)

```bash
cd deployments/dev
./dev-setup.sh
```

Edit `.env` file and set your `LLM_API_KEY` and other configuration.

### Start Infrastructure

Start infrastructure (PostgreSQL, NATS, Redis):

```bash
docker-compose -f docker-compose.dev.yml up -d
```

### Start Individual Components

After starting infrastructure, start components independently:

```bash
# Start server
./dev-server.sh

# Start worker
./dev-worker.sh

# Start frontend (requires Node.js)
./dev-frontend.sh
```

Each component supports `--build` flag to rebuild before starting.

### View Logs

```bash
# All services
docker-compose -f docker-compose.dev.yml logs -f

# Specific service
docker-compose -f docker-compose.dev.yml logs -f postgresql
```

### Stop Environment

```bash
docker-compose -f docker-compose.dev.yml down
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
- `server.config.example.yaml` - Server config template (copy to `server.config.yaml`)
- `worker.config.example.yaml` - Worker config template (copy to `worker.config.yaml`)

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
make dev-server    # Start server
make dev-worker    # Start worker
make dev-frontend  # Start frontend
```
