# VoxMeet

**Self-hosted real-time communication suite** — group calls, messaging, and screen sharing.

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

VoxMeet is a fully self-hosted communication platform for teams and communities. It provides real-time group calls, persistent messaging, and screen sharing — no paid third-party services, no data leaving your infrastructure.

## Features

- **Real-time group calls** — Low-latency audio/video using WebRTC with a selective forwarding unit (SFU)
- **Screen sharing** — Share your screen with other participants in a call
- **Persistent chat** — Text messaging with history, typing indicators, and message persistence
- **Presence** — See who's online and what room they're in
- **Self-hosted** — Deploy on your own infrastructure via Docker Compose
- **STUN/TURN support** — Works behind NATs and firewalls using Coturn
- **No paid services** — Fully open source, zero external API dependencies

## Architecture

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

| Service | Role |
|---|---|
| **api-gateway** | Auth, REST API, WebSocket signaling hub, serves SPA |
| **sfu** | Selective Forwarding Unit — Pion WebRTC media routing |
| **chat-service** | Message persistence and broadcast |
| **room-service** | Room lifecycle, membership, permissions |
| **presence-service** | Online/offline tracking, typing indicators |
| **web-client** | React SPA (Vite + TypeScript + TailwindCSS) |

**Infrastructure:** PostgreSQL 16, RabbitMQ 4, Coturn (STUN/TURN)

## Quick Start

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/)
- [Go](https://go.dev/dl/) 1.25+ (for development)
- [sqlc](https://sqlc.dev/) 1.27+ (for code generation)

### Run

```bash
# Clone the repository
git clone https://github.com/your-org/voxmeet.git
cd voxmeet

# Copy environment and customize
cp .env.example .env
# Edit .env — set a strong JWT_SECRET

# Start all services
docker compose up --build -d

# Check health
curl http://localhost:8080/api/v1/health
```

### Development

```bash
# Start dev environment with hot reload
make dev

# Run tests
make test

# Build all services
make build

# View logs
make logs
```

## Documentation

- [Architecture Overview](docs/architecture.md)
- [Signaling Protocol](docs/signaling-protocol.md)
- [Deployment Guide](docs/deployment.md)
- [Development Guide](docs/development.md)

## Project Status

| Phase | Status |
|---|---|
| 1. Foundation | ✅ Complete |
| 2. Auth + Rooms | ✅ Complete |
| 3. Calls (SFU) | 🚧 In Progress |
| 4. Chat | 🔲 Planned |
| 5. Frontend | 🔲 Planned |

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT — see [LICENSE](LICENSE).
