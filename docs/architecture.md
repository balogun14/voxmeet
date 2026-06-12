# Architecture

## Overview

VoxMeet is a microservice-oriented monolith — Go services that communicate over RabbitMQ, deployable via Docker Compose on a single host.

```
Client (React SPA)
    │  HTTP REST + WebSocket
    ▼
api-gateway ── RabbitMQ ──► sfu (Pion WebRTC SFU)
                │               │  Direct RTP/RTCP
                ├── chat-service  ◄──► Client
                ├── room-service
                └── presence-service
```

## Services

### api-gateway (port 8080)

The only service clients connect to directly. Responsibilities:
- **Auth** — User registration, login, JWT issuance and validation
- **REST API** — Room CRUD, user profiles, chat history
- **WebSocket hub** — Signaling gateway. All WS messages from clients are published to RabbitMQ and responses are routed back to the correct client
- **Static files** — Serves the React SPA bundle

### SFU (internal, no public port)

The Selective Forwarding Unit built on Pion WebRTC. No HTTP interface — communicates entirely over RabbitMQ. Responsibilities:
- Manage Pion `PeerConnection` per client
- Handle SDP negotiation (SFU creates offers, clients answer)
- Route ICE candidates bidirectionally
- Forward RTP/RTCP media between peers (publish/subscribe model)
- Active speaker detection and last-N forwarding (planned)

### chat-service (internal)

Message persistence and broadcast. Writes to PostgreSQL, publishes to RabbitMQ for real-time delivery.

### room-service (internal)

Room lifecycle — create, update, delete rooms, manage memberships and roles.

### presence-service (internal)

Tracks who's online, which room they're in, typing indicators. Heartbeat-based with automatic timeout.

### web-client

React + TypeScript + TailwindCSS SPA built with Vite. Served by api-gateway or a dedicated Nginx container.

## Communication

### RabbitMQ Topology

| Exchange | Type | Purpose |
|---|---|---|
| `voxmeet.signal` | topic | WebRTC signaling (join, SDP, ICE, leave) |
| `voxmeet.chat` | topic | Chat messages and typing indicators |
| `voxmeet.room` | topic | Room lifecycle events |
| `voxmeet.presence` | topic | Presence events (online/offline) |
| `voxmeet.rpc` | direct | RPC-style request/reply |

**Routing keys:** `{exchange}.room.{roomID}` — e.g. `signal.room.abc-123`

### WebSocket Protocol

All real-time communication happens over a single WebSocket connection per client. See [signaling-protocol.md](./signaling-protocol.md).

## Database

PostgreSQL 16 with 5 tables: `users`, `rooms`, `room_members`, `messages`, `sessions`.

Schema management:
- **Source of truth:** `schema.sql` at repo root
- **Migrations:** Atlas generates migration files from `schema.sql` diffs
- **Codegen:** sqlc generates typed Go code from queries

## Deployment

Single-host deployment via Docker Compose:

```
docker compose up --build -d
```

All services are containerized. Coturn provides STUN/TURN for WebRTC connectivity behind NATs.
