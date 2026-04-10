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
| `http://localhost:3000` | Grafana (admin/admin) |
| `http://localhost:9090` | Prometheus |

Seed the database with sample data:

```bash
make seed
```

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
