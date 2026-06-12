# Deployment Guide

## Prerequisites

- Docker 24+ and Docker Compose v2
- A Linux server (or any host with Docker)
- Domain name + TLS certificate (recommended)
- Open ports: 8080 (HTTP/WS), 3478 (STUN/TURN over UDP), 50000-51000 (TURN relay)

## Standard Deployment

```bash
# Clone
git clone https://github.com/your-org/voxmeet.git
cd voxmeet

# Configure
cp .env.example .env
# Edit .env — set JWT_SECRET to a random 64-char string

# Start
docker compose up --build -d

# Verify
curl http://localhost:8080/api/v1/health
```

## Production Deployment

### 1. Use a Reverse Proxy

Run Nginx in front to handle TLS termination:

```nginx
server {
    listen 443 ssl;
    server_name voxmeet.example.com;

    ssl_certificate /etc/letsencrypt/live/voxmeet.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/voxmeet.example.com/privkey.pem;

    location / {
        proxy_pass http://api-gateway:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_read_timeout 86400s;
    }
}
```

### 2. Configure TURN

Edit `deploy/coturn/turnserver.conf`:

```
use-auth-secret
static-auth-secret=<your-turn-secret>
realm=voxmeet.example.com
```

### 3. Scale Services

For larger deployments, you can scale the SFU horizontally. This requires a room registry (Redis) to coordinate which SFU handles which room — planned for a future release.

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `HTTP_PORT` | `:8080` | API gateway listen address |
| `RABBITMQ_URL` | `amqp://guest:guest@rabbitmq:5672/` | RabbitMQ connection string |
| `DATABASE_URL` | `postgres://postgres:postgres@postgres:5432/voxmeet` | PostgreSQL connection string |
| `JWT_SECRET` | — | HMAC key for JWT signing (required) |
| `LOG_LEVEL` | `info` | Log level: debug, info, warn, error |
| `POSTGRES_USER` | `postgres` | PostgreSQL user |
| `POSTGRES_PASSWORD` | `postgres` | PostgreSQL password |
| `RABBITMQ_USER` | `guest` | RabbitMQ user |
| `RABBITMQ_PASSWORD` | `guest` | RabbitMQ password |

## Port Reference

| Port | Service | Protocol |
|---|---|---|
| 8080 | api-gateway | HTTP/WebSocket |
| 5432 | PostgreSQL | Internal only |
| 5672 | RabbitMQ | Internal only |
| 15672 | RabbitMQ admin | Internal only |
| 3478 | Coturn STUN/TURN | UDP |
| 5349 | Coturn TURN/TLS | UDP |
| 50000-51000 | Coturn relay | UDP |
