# Contributing to VoxMeet

First off, thank you for considering contributing! We welcome contributions from everyone.

## Code of Conduct

All contributors must abide by our [Code of Conduct](CODE_OF_CONDUCT.md).

## How to Contribute

### 1. Find an Issue

- Look for issues labeled `good-first-issue` or `help-wanted`
- Or open a new issue describing what you'd like to work on

### 2. Set Up

```bash
git clone https://github.com/your-org/voxmeet.git
cd voxmeet
cp .env.example .env
make dev
```

### 3. Make Changes

- **TDD is required** — write tests before or alongside implementation
- One feature per PR
- Keep changes focused and minimal

### 4. Test

```bash
# Run all tests (no external services required)
make test

# Run integration tests (requires Docker)
make test-integration
```

### 5. Commit

```
type(scope): brief description

- Bullet points for details
- References issue number if applicable
```

Types: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`

### 6. Submit a Pull Request

- PRs need at least one review before merging
- CI must pass (tests + lint)
- Squash merge preferred

## Development Workflow

We follow **TDD (Red-Green-Refactor)**:

1. **Red** — Write a failing test that defines expected behaviour
2. **Green** — Write the minimum code to make it pass
3. **Refactor** — Clean up while keeping tests green

### Testing Rules

- Tests go next to the file they test (`file_test.go`)
- Use `testing` standard library + `github.com/stretchr/testify`
- Tests MUST NOT depend on external services (RabbitMQ, PostgreSQL) unless marked `_integration_test.go`
- Use interfaces + mocks/test doubles for external dependencies
- Run `go test ./...` before every commit
- Coverage target: 80%+ for core logic

### Test Naming

```
Test_<Package>_<Function>
t.Run("specific behaviour", ...)
```

Example: `Test_Auth_RegisterUser` with subtests like `"invalid email returns error"`.

## Service Directory Structure

```
services/<service>/
├── cmd/server/main.go       # Entry point
├── internal/
│   ├── config/              # Env-based config
│   ├── handler/             # RPC or HTTP handlers
│   ├── store/               # Database access (if any)
│   └── model/               # Domain types
├── go.mod
└── Dockerfile
```

Shared code lives in `pkgs/` — currently `pkgs/rabbitmq/` for connection management.

## Code Standards

- **Error handling:** Always check errors. Wrap with context: `fmt.Errorf("fetch room %s: %w", id, err)`
- **Logging:** Use `zerolog`. Structured fields only: `log.Info().Str("user", id).Msg("user joined room")`
- **Config:** All config from environment variables via `os.Getenv` + defaults. No config files.
- **Context:** Pass `context.Context` as first param on all public functions
- **Interfaces:** Define interfaces where you consume (not produce). Keep them small (1-3 methods)
- **Concurrency:** Use `errgroup` for goroutine groups. Handle graceful shutdown via `signal.NotifyContext`
- **Dependency injection:** Pass dependencies explicitly through constructors. No global state

## Database Migrations

`schema.sql` at the repo root is the single source of truth. Generate migrations with Atlas:

```bash
docker run --rm --net=host -v $(pwd):/work arigaio/atlas migrate diff my_migration \
  --to file:///work/schema.sql \
  --dev-url "docker://postgres/17/dev?search_path=public" \
  --dir file:///work/migrations
```

After modifying the schema, regenerate sqlc:

```bash
cd services/<service>
sqlc generate
```
