# Zenvikar Platform

Multi-tenant SaaS booking platform with bilingual support (English / Portuguese). Built as a monorepo with a Go modular-monolith API and four Next.js frontend apps, connected through Traefik reverse proxy.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) & Docker Compose
- [Go 1.22+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/) & [pnpm](https://pnpm.io/)

## Quick Start

```bash
cp .env.example .env
make build
make up
```

Services will be available at:

| URL | App |
|---|---|
| `http://zenvikar.localhost` | Marketing website |
| `http://acme.zenvikar.localhost` | Booking page (tenant: acme) |
| `http://manage.zenvikar.localhost` | Tenant management portal |
| `http://admin.zenvikar.localhost` | Platform admin |
| `http://api.zenvikar.localhost` | Backend API |
| `http://api.zenvikar.localhost/swagger` | API docs (non-production only) |
| `http://localhost:3000` | Grafana (admin/admin) |
| `http://localhost:9090` | Prometheus |

Seed the database with sample data:

```bash
make seed
```

## Google OAuth Setup

The API owns the OAuth callback. In Google Cloud Console, create an OAuth 2.0
Web application client and add this authorized redirect URI for local Docker:

```text
http://localhost:8081/api/v1/auth/oauth/google/callback
```

Then set these values in `.env` before rebuilding/restarting the API:

```env
GOOGLE_CLIENT_ID=your-web-client-id
GOOGLE_CLIENT_SECRET=your-web-client-secret
API_PUBLIC_URL=http://localhost:8081
BASE_DOMAIN=zenvikar.localhost
```

If Google rejects the request with `redirect_uri_mismatch`, the value above
does not exactly match the callback generated from `API_PUBLIC_URL`. For local
development, use `localhost` for the Google callback; `*.zenvikar.localhost`
is still used by the frontend apps, but Google's OAuth redirect URI rules only
exempt `localhost` from the HTTPS requirement.

## Architecture

See [design document](.kiro/specs/zenvikar-platform-foundation/design.md) for the full architecture diagram and sequence flows.

```
Browser → Traefik → Frontend Apps (Next.js) → API (Go) → PostgreSQL / Redis
                                                       → OTel Collector → Loki / Tempo
```

## Directory Structure

```
zenvikar/
├── apps/
│   ├── api/              # Go modular-monolith backend
│   ├── admin-web/        # Platform admin (Next.js)
│   ├── booking-web/      # Customer booking by subdomain (Next.js)
│   ├── marketing-web/    # Public landing page (Next.js)
│   └── tenant-web/       # Tenant management portal (Next.js)
├── packages/
│   ├── config/           # Shared configuration (i18n, API client)
│   ├── types/            # Shared TypeScript interfaces
│   └── ui/               # Shared UI components
├── infra/
│   ├── observability/    # Prometheus, Grafana, Loki, Tempo configs
│   └── terraform/        # IaC modules and environment configs
├── scripts/
│   └── seed.sql          # Sample data for local development
├── .github/workflows/    # CI/CD pipelines
├── docker-compose.yml
└── Makefile
```

## Make Targets

| Target | Description |
|---|---|
| `make up` | Start all services (`docker compose up -d`) |
| `make down` | Stop all services |
| `make build` | Build all Docker images |
| `make migrate` | Run database migrations |
| `make seed` | Insert sample data into PostgreSQL |
| `make test` | Run Go tests |
| `make lint` | Run Go linters |
| `make logs` | Tail logs from all services |

## Documentation

- [Requirements](.kiro/specs/zenvikar-platform-foundation/requirements.md)
- [Design](.kiro/specs/zenvikar-platform-foundation/design.md)
- [Tasks](.kiro/specs/zenvikar-platform-foundation/tasks.md)
