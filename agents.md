# VoxMeet — Agent Development Guide

## Project Overview

VoxMeet is a fully self-hosted communication suite — group calls, messaging, screen sharing.  
**Stack:** Go (Pion WebRTC), RabbitMQ, PostgreSQL, React + TailwindCSS  
**Deployment:** Docker Compose (single host, all services)

## Architecture (6 Services)

```
React SPA → api-gateway (8080) ─── RabbitMQ ─── sfu (Pion SFU)
                                              ├── chat-service
                                              ├── room-service
                                              └── presence-service
```

Infrastructure: PostgreSQL, RabbitMQ, Coturn (STUN/TURN)

## Development Workflow — TDD (Red-Green-Refactor)

Every feature follows strict TDD:

1. **Red** — Write a failing test that defines expected behaviour
2. **Green** — Write the minimum code to make it pass
3. **Refactor** — Clean up while keeping tests green

### Testing rules

- `*_test.go` goes next to the file it tests
- Use `testing` standard library + `github.com/stretchr/testify` for assertions
- Tests MUST NOT depend on external services (RabbitMQ, PostgreSQL) unless explicitly marked as `_integration_test.go`
- Use interfaces + mocks/test doubles for external dependencies
- Run `go test ./...` before every commit
- Coverage target: 80%+ for core logic (auth, signaling, room management)

### Test naming

- `Test_<Package>_<Function>` — e.g. `Test_Auth_RegisterUser`
- Clear subtest names: `t.Run("invalid email returns error", ...)`

## Implementation Phases

| Phase | What | Testing Focus | Status |
|---|---|---|---|
| **1. Foundation** | Go modules, shared RabbitMQ lib, api-gateway skeleton, docker-compose, DB migrations | Connection manager, config loading, health endpoint | ✅ Done |
| **2. Auth + Rooms** | Registration, JWT auth, room CRUD, WebSocket hub | Auth middleware, room handlers, WS hub lifecycle | ✅ Done |
| **3. Calls (SFU)** | Pion WebRTC SFU, signaling flow, audio/video, screen share | PeerConnection setup, track routing, ICE handling | ✅ Done (signaling + room mgmt, no Pion yet) |
| **3. Calls (SFU)** | Pion WebRTC SFU, signaling flow, audio/video, screen share | PeerConnection setup, track routing, ICE handling, RTP forwarding | ✅ Done (full Pion + media relay) |
| **4. Chat** | Message persistence, broadcast, history retrieval | Message store, broadcast fan-out | ✅ Done |
| **5. Frontend** | React SPA (Vite + Tailwind) | Component unit tests (Vitest + testing-library) | 🔲 Planned |

## Services

| Service | Port | What it does | Status |
|---|---|---|---|
| **api-gateway** | 8080 | Auth (JWT), REST (users, rooms), WS signaling hub, serves SPA | ✅ Built |
| **sfu** | internal | WebRTC SFU — room/peer/track management, RabbitMQ signaling relay | ✅ Built (signaling) |
| **chat-service** | internal | Message persistence (PostgreSQL), broadcast (RabbitMQ), history retrieval | ✅ Built |
| **presence-service** | internal | Online/offline tracking, heartbeat timeout, typing indicators | ✅ Built |
| **room-service** | internal | Room lifecycle — extracted from api-gateway (planned) | 🔲 Stub |
| **web-client** | — | React SPA | 🔲 Empty |

## DB Schema Management

### Schema source of truth

`schema.sql` at the repo root is the single source of truth for the database schema.

### Atlas (migrations)

```bash
# Generate a new migration after editing schema.sql
docker run --rm --net=host -v $(pwd):/work arigaio/atlas migrate diff my_migration_name \
  --to file:///work/schema.sql \
  --dev-url "docker://postgres/17/dev?search_path=public" \
  --dir file:///work/migrations

# Lint the latest migration
docker run --rm --net=host -v $(pwd):/work arigaio/atlas migrate lint \
  --dev-url "docker://postgres/17/dev?search_path=public" \
  --latest 1 \
  --dir file:///work/migrations

# Apply all pending migrations
docker run --rm --net=host -v $(pwd):/work arigaio/atlas migrate apply \
  --url "$DATABASE_URL" \
  --dir file:///work/migrations
```

### sqlc (Go codegen)

After modifying `schema.sql`, regenerate sqlc:

```bash
cd services/<service>
sqlc generate
```

This produces typed Go code in `internal/db/` — models, queries, and the `Querier` interface.

- **Error handling:** Always check errors. No silent swallows. Wrap with context: `fmt.Errorf("fetch room %s: %w", id, err)`
- **Logging:** Use `zerolog`. Structured fields only, no formatted strings. `log.Info().Str("user", id).Msg("user joined room")`
- **Config:** All config from environment variables via `os.Getenv` + defaults. No config files.
- **Context:** Pass `context.Context` as first param on all public functions. Use it for cancellation and deadlines.
- **Interfaces:** Define interfaces where you consume (not produce). Keep them small (1-3 methods).
- **Concurrency:** Use `errgroup` for goroutine groups. Always handle graceful shutdown via `signal.NotifyContext`.
- **Dependency injection:** Pass dependencies explicitly through constructors. No global state.

## Project Structure

```
voxmeet/
├── agents.md                  # This file
├── docker-compose.yml         # All services + infra
├── docker-compose.dev.yml     # Dev overrides (hot reload)
├── .env.example
├── Makefile                   # Top-level build/test/run
├── pkgs/                      # Shared Go libraries
│   ├── go.mod
│   └── rabbitmq/              # Connection manager, RPC
├── services/
│   ├── api-gateway/           # HTTP + WS, auth, routing
│   ├── sfu/                   # Pion WebRTC SFU
│   ├── chat-service/          # Message persistence
│   ├── room-service/          # Room lifecycle
│   ├── presence-service/      # Online status
│   └── web-client/            # React + Vite + Tailwind
├── migrations/                # SQL migration files
└── deploy/                    # Configs (coturn, nginx)
```

## RabbitMQ Topology

| Exchange | Type | Use |
|---|---|---|
| `voxmeet.signal` | topic | WebRTC signaling (offer/answer/ICE/join/leave) |
| `voxmeet.chat` | topic | Chat messages + typing |
| `voxmeet.room` | topic | Room lifecycle events |
| `voxmeet.presence` | topic | Presence (online/offline) |
| `voxmeet.rpc` | direct | RPC request/reply |

Routing keys: `{signal,chat,presence}.room.{roomID}`

## Database (PostgreSQL)

4 tables: `users`, `rooms`, `room_members`, `messages`. UUID PKs, bcrypt passwords, JWT auth.

## WebSocket Protocol

All signaling happens over a single WS connection per client to api-gateway.

**Client → Server:**
```json
{"type":"join_room","room_id":"...","offer":{...}}
{"type":"answer","room_id":"...","target_peer":"...","sdp":{...}}
{"type":"ice_candidate","room_id":"...","target_peer":"...","candidate":{...}}
{"type":"leave_room","room_id":"..."}
{"type":"chat.send","room_id":"...","content":"..."}
{"type":"chat.typing","room_id":"...","is_typing":true}
```

**Server → Client:**
```json
{"type":"room_joined","room_id":"...","peers":[...]}
{"type":"peer_joined","room_id":"...","peer":{...}}
{"type":"peer_left","room_id":"...","peer_id":"..."}
{"type":"offer","room_id":"...","from_peer":"...","sdp":{...}}
{"type":"ice_candidate","room_id":"...","from_peer":"...","candidate":{...}}
{"type":"chat.message","room_id":"...","user":{...},"content":"...","timestamp":"..."}
{"type":"chat.typing","room_id":"...","user_id":"...","is_typing":true}
{"type":"error","code":"...","message":"..."}
```

## Key Go Dependencies

- `github.com/pion/webrtc/v4` — WebRTC
- `github.com/gorilla/websocket` — WebSocket server
- `github.com/rabbitmq/amqp091-go` — RabbitMQ client
- `github.com/golang-jwt/jwt/v5` — JWT
- `github.com/rs/zerolog` — Structured logging
- `github.com/jackc/pgx/v5` — PostgreSQL driver
- `github.com/golang-migrate/migrate/v4` — DB migrations
- `github.com/stretchr/testify` — Test assertions
- `golang.org/x/crypto` — bcrypt
- `golang.org/x/sync` — errgroup, semaphore
