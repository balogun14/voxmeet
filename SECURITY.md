# Security Policy

## Reporting a Vulnerability

We take the security of VoxMeet seriously. If you discover a security vulnerability, please do NOT open a public issue.

Instead, send a private report to the project maintainers via email.

**Please include:**
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

## What to Expect

- **Acknowledgment** within 48 hours
- **Status updates** every 5 business days until resolution
- **Credit** in release notes if you're the reporter (unless you prefer to remain anonymous)

## Scope

VoxMeet is self-hosted software. Security depends partly on how you deploy it. You should:

- Use a strong, randomly generated `JWT_SECRET`
- Run behind a reverse proxy with HTTPS (TLS 1.2+)
- Keep dependencies updated
- Restrict database and RabbitMQ access to internal networks only
- Use firewall rules to limit access to ports 8080 and the TURN ports

## Supported Versions

| Version | Supported |
|---|---|
| Latest | ✅ |
| Older | ❌ |
