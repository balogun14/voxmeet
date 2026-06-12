# Development Guide

## Prerequisites

- Go 1.25+
- Docker + Docker Compose
- sqlc 1.27+

## Getting Started

```bash
# Clone
git clone https://github.com/your-org/voxmeet.git
cd voxmeet

# Set up environment
cp .env.example .env
# Edit JWT_SECRET

# Start infrastructure (PostgreSQL + RabbitMQ)
docker compose up -d postgres rabbitmq

# Run tests
make test

# Start development
make dev
```

## Running Tests

```bash
# All unit tests (no external services required)
make test

# Specific service
cd services/api-gateway
go test ./... -short -count=1

# Integration tests (requires RabbitMQ + PostgreSQL running)
make test-integration
```

## Working with the Database

### Schema

Edit `schema.sql` at the repo root, then:

```bash
# Generate migration
docker run --rm --net=host -v $(pwd):/work arigaio/atlas migrate diff my_migration_name \
  --to file:///work/schema.sql \
  --dev-url "docker://postgres/17/dev?search_path=public" \
  --dir file:///work/migrations

# Regenerate sqlc code
cd services/api-gateway
sqlc generate
```

### Adding a New Query

1. Add the query to `services/<service>/queries/<table>.sql`
2. Run `sqlc generate` in that service directory
3. The `Querier` interface and implementation are regenerated automatically

## Project Structure

```
voxmeet/
├── schema.sql              # Database schema source of truth
├── atlas.hcl               # Atlas migration config
├── docker-compose.yml      # All services + infrastructure
├── Makefile                # Top-level commands
├── pkgs/                   # Shared Go libraries
│   └── rabbitmq/           # Connection manager, RPC helpers
├── services/
│   ├── api-gateway/        # HTTP + WebSocket + auth
│   ├── sfu/                # WebRTC SFU (Pion)
│   ├── chat-service/       # Message persistence
│   ├── room-service/       # Room lifecycle
│   ├── presence-service/   # Online status
│   └── web-client/         # React SPA
├── migrations/             # Atlas-generated SQL migrations
└── docs/                   # Documentation
```

## Code Generation

```bash
# sqlc — generates typed Go models and queries
cd services/api-gateway && sqlc generate

# Atlas — generates DB migrations from schema.sql diff
docker run --rm --net=host -v $(pwd):/work arigaio/atlas migrate diff \
  --to file:///work/schema.sql \
  --dev-url "docker://postgres/17/dev?search_path=public" \
  --dir file:///work/migrations
```

## Conventional Commits

We follow a simplified conventional commits format:

- `feat(scope):` — New feature
- `fix(scope):` — Bug fix
- `refactor(scope):` — Code refactoring
- `test(scope):` — Adding or updating tests
- `docs(scope):` — Documentation changes
- `chore(scope):` — Maintenance, tooling, CI

Examples:
```
feat(auth): add refresh token endpoint
fix(sfu): handle nil ICE candidate on gathering complete
docs: update signaling protocol for publish flow
```
