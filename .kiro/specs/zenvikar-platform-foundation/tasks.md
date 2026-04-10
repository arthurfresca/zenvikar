# Implementation Plan: Zenvikar Platform Foundation

## Overview

Infrastructure-first foundation for the Zenvikar SaaS multi-tenant booking platform. Tasks are ordered by dependency: monorepo scaffolding first, then backend core, multi-tenancy primitives, booking domain, frontend apps, local dev infrastructure, observability, Terraform, CI/CD, and developer experience. Go for the backend API, TypeScript/Next.js for frontend apps.

## Tasks

- [x] 1. Monorepo root scaffolding and shared packages
  - [x] 1.1 Create root monorepo structure with workspace configuration
    - Create root `package.json` with npm/pnpm workspaces pointing to `apps/*` and `packages/*`
    - Create root `tsconfig.json` with path aliases for `@zenvikar/*` packages
    - Create root `.gitignore` covering Node, Go, Terraform, Docker, IDE files
    - Create root `.editorconfig` and `.prettierrc` for consistent formatting
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

  - [x] 1.2 Create `packages/types` shared TypeScript package
    - Create `packages/types/package.json` with name `@zenvikar/types`
    - Create `packages/types/tsconfig.json`
    - Create `packages/types/src/index.ts` barrel export
    - Create `packages/types/src/tenant.ts` with `Tenant`, `TenantColors` interfaces
    - Create `packages/types/src/user.ts` with `User`, `TenantRole`, `PlatformRole`, `TenantMembership` types
    - Create `packages/types/src/booking.ts` with `Booking`, `Service`, `OpeningHours`, `BlockedDate` interfaces
    - _Requirements: 1.3, 6.1, 15.1_

  - [x] 1.3 Create `packages/config` shared configuration package
    - Create `packages/config/package.json` with name `@zenvikar/config`
    - Create `packages/config/src/i18n.ts` with locales array (`en`, `pt`), `Locale` type, `defaultLocale`
    - Create `packages/config/src/api.ts` with API client base configuration and environment helpers
    - Create `packages/config/src/index.ts` barrel export
    - _Requirements: 1.3, 15.1_

  - [x] 1.4 Create `packages/ui` shared UI component package skeleton
    - Create `packages/ui/package.json` with name `@zenvikar/ui`
    - Create `packages/ui/tsconfig.json`
    - Create `packages/ui/src/index.ts` barrel export
    - Create placeholder component `packages/ui/src/Button.tsx`
    - _Requirements: 1.3_

- [x] 2. Backend Go API scaffolding and platform module
  - [x] 2.1 Initialize Go module and project structure
    - Create `apps/api/go.mod` with module `github.com/zenvikar/api`
    - Create `apps/api/cmd/api/main.go` with module registry pattern, dependency injection, health endpoints, metrics endpoint, and migration runner as shown in design
    - Create `apps/api/internal/platform/` directory structure
    - _Requirements: 1.2, 9.5, 11.1_

  - [x] 2.2 Implement configuration loading
    - Create `apps/api/internal/platform/config/config.go` with `Config` struct and `Load()` function
    - Support env vars: `DATABASE_URL`, `REDIS_URL`, `PORT`, `OTEL_EXPORTER_OTLP_ENDPOINT`, `BASE_DOMAIN`, `ALLOWED_ORIGINS`
    - _Requirements: 9.2, 9.3_

  - [x] 2.3 Implement structured logging setup
    - Create `apps/api/internal/platform/logger/logger.go` with slog JSON handler setup
    - Configure log level from environment
    - _Requirements: 10.5_

  - [x] 2.4 Implement database and Redis connection helpers
    - Create `apps/api/internal/platform/db/db.go` with PostgreSQL connection setup, connection pool config (`MaxOpenConns`, `MaxIdleConns`, `ConnMaxLifetime`)
    - Create `apps/api/internal/platform/cache/redis.go` with Redis client setup
    - _Requirements: 9.2, 9.3, 9.4_

  - [x] 2.5 Implement health and readiness endpoints
    - Create `apps/api/internal/platform/handlers/health.go` with `/healthz` (liveness) and `/readyz` (readiness checking DB + Redis connectivity)
    - _Requirements: 9.5, 16.6_

  - [x] 2.6 Implement core middleware stack
    - Create `apps/api/internal/platform/middleware/logging.go` — request logging middleware using slog
    - Create `apps/api/internal/platform/middleware/recovery.go` — panic recovery middleware
    - Create `apps/api/internal/platform/middleware/cors.go` — CORS middleware with configurable allowed origins
    - _Requirements: 10.5_

  - [x] 2.7 Implement Module interface and registration
    - Create `apps/api/internal/platform/module.go` with `Module` interface (`Name()`, `RegisterRoutes()`, `Migrate()`) and `Dependencies` struct
    - _Requirements: 1.2, 11.1_

- [x] 3. Checkpoint — Backend API boots and serves health endpoints
  - Ensure `go build` succeeds for `cmd/api/main.go`, health endpoints are wired, and module registration pattern works. Ask the user if questions arise.

- [x] 4. Multi-tenancy primitives — Tenant model, validation, and resolution
  - [x] 4.1 Implement tenant slug validation
    - Create `apps/api/internal/tenants/validation.go` with `ValidateTenantSlug(slug string) error`
    - Enforce: 3–63 chars, `^[a-z0-9][a-z0-9-]*[a-z0-9]$`, no consecutive hyphens, not in reserved slugs list
    - Define reserved slugs constant list: `www`, `api`, `admin`, `manage`, `app`, `mail`, `smtp`, `ftp`, `ssh`, `git`, `cdn`, `static`, `assets`, `media`, `blog`, `docs`, `help`, `support`, `status`, `billing`, `zenvikar`
    - Return descriptive errors for each failure case
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_

  - [x] 4.2 Write property test for slug validation (Property 1)
    - **Property 1: Slug Validation Correctness**
    - Use `pgregory.net/rapid` to generate arbitrary strings and verify: accepted iff length 3–63, matches pattern, no consecutive hyphens, not reserved
    - **Validates: Requirements 2.1, 2.2, 2.3, 2.4**

  - [x] 4.3 Implement subdomain extraction
    - Create `apps/api/internal/platform/middleware/tenant.go` with `ExtractTenantSlugFromHost(host, baseDomain string) (string, error)`
    - Handle: port stripping, lowercase normalization, base domain match, nested subdomain rejection
    - Define errors: `ErrNoSubdomain`, `ErrInvalidHost`
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

  - [x] 4.4 Write property test for subdomain extraction round-trip (Property 2)
    - **Property 2: Subdomain Extraction Round-Trip**
    - Generate valid slugs and base domains, construct `{slug}.{baseDomain}` (with/without port), verify extraction returns original slug
    - **Validates: Requirements 3.1, 3.4**

  - [x] 4.5 Write property test for invalid host rejection (Property 3)
    - **Property 3: Invalid Host Rejection**
    - Generate hostnames that do not end with `.{baseDomain}` and are not equal to `baseDomain`, verify `ErrInvalidHost` is returned
    - **Validates: Requirement 3.3**

  - [x] 4.6 Implement tenant data model and migration
    - Create `apps/api/internal/tenants/model.go` with `Tenant` struct as defined in design
    - Create `apps/api/migrations/001_create_tenants.sql` with table, partial index on `slug WHERE enabled = true`
    - Create `apps/api/internal/tenants/repository.go` with `FindBySlug(ctx, slug)` database query
    - _Requirements: 4.1, 11.2, 11.3, 14.2_

  - [x] 4.7 Implement tenant resolution service with Redis caching
    - Create `apps/api/internal/tenants/service.go` with `ResolveTenantBySlug(ctx, slug) (*Tenant, error)`
    - Implement: validate slug → check Redis cache → query DB → check enabled → cache in Redis (5min TTL) → return
    - Define errors: `ErrTenantNotFound`, `ErrTenantDisabled`
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6_

  - [x] 4.8 Write property test for tenant JSON serialization round-trip (Property 4)
    - **Property 4: Tenant JSON Serialization Round-Trip**
    - Generate valid `Tenant` objects, serialize to JSON, deserialize back, verify equality
    - **Validates: Requirement 4.6**

  - [x] 4.9 Implement tenant resolution middleware
    - Create tenant resolution middleware that extracts slug from `X-Tenant-ID` header and injects `TenantContext` into request context
    - _Requirements: 4.1, 14.1_

  - [x] 4.10 Register tenants module with route and migration
    - Create `apps/api/internal/tenants/module.go` implementing `Module` interface
    - Register `GET /api/v1/tenants/resolve?slug={slug}` endpoint
    - _Requirements: 4.1, 11.1_

- [x] 5. Multi-tenancy primitives — Users, memberships, and RBAC
  - [x] 5.1 Implement user data model and migration
    - Create `apps/api/internal/users/model.go` with `User` struct
    - Create `apps/api/migrations/002_create_users.sql` with table, unique email constraint, locale check
    - _Requirements: 11.2, 15.2_

  - [x] 5.2 Implement platform admin model and migration
    - Create `apps/api/internal/users/platform_admin.go` with `PlatformAdmin` struct and role constants
    - Create `apps/api/migrations/004_create_platform_admins.sql`
    - _Requirements: 5.4, 11.2_

  - [x] 5.3 Implement tenant membership model, repository, and migration
    - Create `apps/api/internal/tenant_memberships/model.go` with `TenantMembership` struct and role constants
    - Create `apps/api/migrations/003_create_tenant_memberships.sql` with unique(tenant_id, user_id) constraint, foreign keys, indexes
    - Create `apps/api/internal/tenant_memberships/repository.go` with `FindByUserAndTenant(ctx, userID, tenantID)`
    - _Requirements: 5.1, 5.2, 11.2, 11.3, 14.2, 14.4_

  - [x] 5.4 Implement membership check service
    - Create `apps/api/internal/tenant_memberships/service.go` with `CheckMembership(ctx, userID, tenantID) (*TenantMembership, error)`
    - Return `ErrNotAMember` if no membership found
    - _Requirements: 5.1, 5.2, 14.3_

  - [x] 5.5 Implement RBAC authorization engine
    - Create `apps/api/internal/platform/authz/authz.go` with `Authorize(ctx, userID, tenantID, permission) error`
    - Implement static permission matrix: `tenant_owner` → all, `tenant_manager` → bookings/services/staff/availability/branding:read, `tenant_staff` → limited, `tenant_finance_viewer` → read-only
    - Platform admin bypass: check `platform_admins` table first
    - Define errors: `ErrNotAMember`, `ErrInsufficientPermissions`
    - _Requirements: 5.3, 5.4, 5.5, 5.6, 5.7_

  - [x] 5.6 Write property test for RBAC permission matrix (Property 5)
    - **Property 5: RBAC Permission Matrix Correctness**
    - Generate all (role, permission) combinations, verify authorization result matches static permission matrix; platform admins always authorized; tenant_owner has all permissions
    - **Validates: Requirements 5.5, 5.6, 5.7**

  - [x] 5.7 Register users and tenant_memberships modules
    - Create module files implementing `Module` interface for users and tenant_memberships
    - _Requirements: 11.1_

- [x] 6. Checkpoint — Multi-tenancy primitives compile and unit tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 7. Booking domain data models and migrations
  - [x] 7.1 Implement service model and migration
    - Create `apps/api/internal/services/model.go` with `Service` struct
    - Create migration in `apps/api/migrations/005_create_booking_domain.sql` for `services` table with `duration_minutes > 0` check, buffer constraints
    - _Requirements: 6.4, 11.2_

  - [x] 7.2 Implement availability models and migration
    - Create `apps/api/internal/availability/model.go` with `OpeningHours` and `BlockedDate` structs
    - Add `opening_hours` table with `UNIQUE(tenant_id, day_of_week)` and `day_of_week BETWEEN 0 AND 6` check
    - Add `blocked_dates` table with `UNIQUE(tenant_id, date)`
    - _Requirements: 6.5, 6.6, 6.7, 11.2_

  - [x] 7.3 Implement booking model and migration
    - Create `apps/api/internal/bookings/model.go` with `Booking` struct and status constants
    - Add `bookings` table with `CHECK (end_time > start_time)`, status check constraint, unique index for double-booking prevention on `(tenant_id, service_id, start_time) WHERE status != 'cancelled'`
    - Add composite index on `(tenant_id, start_time, end_time)`
    - _Requirements: 6.1, 6.2, 6.3, 6.8, 6.9, 11.2, 11.3_

  - [x] 7.4 Implement booking availability checking logic
    - Create `apps/api/internal/bookings/availability.go` with `CheckAvailability(ctx, tenantID, serviceID, startTime)` following the algorithm in design
    - Check blocked dates → opening hours → overlapping bookings (with buffers) → calculate end_time
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5_

  - [x] 7.5 Write property test for blocked date rejection (Property 6)
    - **Property 6: Blocked Date Booking Rejection**
    - Generate tenants with blocked dates and booking requests targeting those dates, verify rejection with "date_blocked"
    - **Validates: Requirement 7.1**

  - [x] 7.6 Write property test for outside hours rejection (Property 7)
    - **Property 7: Outside Hours Booking Rejection**
    - Generate tenants with opening hours and booking requests outside those hours, verify rejection with "outside_hours"
    - **Validates: Requirement 7.2**

  - [x] 7.7 Write property test for booking overlap rejection with buffers (Property 8)
    - **Property 8: Booking Overlap Rejection with Buffers**
    - Generate existing bookings and new requests with overlapping effective windows (including buffers), verify rejection with "slot_taken"
    - **Validates: Requirements 7.3, 7.5**

  - [x] 7.8 Write property test for booking end time calculation (Property 9)
    - **Property 9: Booking End Time Calculation**
    - Generate valid booking requests with varying start_time and service duration_minutes, verify `end_time == start_time + duration_minutes`
    - **Validates: Requirement 7.4**

  - [x] 7.9 Implement booking service and create booking handler
    - Create `apps/api/internal/bookings/service.go` with `CreateBooking(ctx, tenantID, req)` using availability check, transaction, status set to "pending"
    - _Requirements: 6.8, 7.1, 7.2, 7.3, 7.4, 7.5_

  - [x] 7.10 Register booking domain modules (bookings, services, availability)
    - Create module files implementing `Module` interface for bookings, services, and availability
    - Register placeholder routes
    - _Requirements: 11.1_

- [x] 8. Checkpoint — Booking domain models compile and property tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 9. API error handling and response formatting
  - [x] 9.1 Implement standardized error response types
    - Create `apps/api/internal/platform/errors/errors.go` with error codes and HTTP status mapping
    - Map: `tenant_not_found` → 404, `tenant_disabled` → 403, `slot_unavailable` → 409, `slug_reserved` → 422, membership errors → 403, infra errors → 503
    - Create error response middleware that converts domain errors to JSON responses
    - _Requirements: 16.1, 16.2, 16.3, 16.4, 16.5, 16.6_

- [x] 10. OpenTelemetry integration in API
  - [x] 10.1 Implement OpenTelemetry tracer and metrics setup
    - Create `apps/api/internal/platform/otel/otel.go` with tracer provider initialization, OTLP exporter config
    - Create `apps/api/internal/platform/middleware/tracing.go` — OTel tracing middleware for HTTP requests
    - Wire Prometheus metrics handler at `/metrics`
    - _Requirements: 10.2, 10.3_

- [x] 11. Frontend app scaffolding
  - [x] 11.1 Scaffold marketing-web Next.js app
    - Create `apps/marketing-web/` with Next.js App Router, TypeScript, Tailwind CSS
    - Create `apps/marketing-web/package.json`, `tsconfig.json`, `next.config.js`
    - Create minimal `app/page.tsx` landing page with i18n setup using `next-intl`
    - Add `@zenvikar/types`, `@zenvikar/config`, `@zenvikar/ui` as workspace dependencies
    - _Requirements: 1.1, 15.1_

  - [x] 11.2 Scaffold booking-web Next.js app with tenant resolution
    - Create `apps/booking-web/` with Next.js App Router, TypeScript, Tailwind CSS
    - Create `src/lib/tenant.ts` with `resolveTenant(host)` function as shown in design
    - Create `src/middleware.ts` extracting `x-forwarded-host` and setting `x-tenant-host` header
    - Create `src/messages/en.json` and `src/messages/pt.json` with booking translations
    - Create minimal `app/page.tsx` that resolves tenant and applies branding + locale
    - _Requirements: 1.1, 3.1, 15.1, 15.3_

  - [x] 11.3 Scaffold tenant-web Next.js app with URL-path tenant switching
    - Create `apps/tenant-web/` with Next.js App Router, TypeScript, Tailwind CSS
    - Create `app/t/[tenantSlug]/layout.tsx` with tenant resolution from URL path as shown in design
    - Create minimal dashboard page placeholder
    - _Requirements: 1.1, 15.1_

  - [x] 11.4 Scaffold admin-web Next.js app
    - Create `apps/admin-web/` with Next.js App Router, TypeScript, Tailwind CSS
    - Create minimal `app/page.tsx` admin dashboard placeholder
    - _Requirements: 1.1, 15.1_

- [x] 12. Checkpoint — All frontend apps build successfully
  - Ensure all tests pass, ask the user if questions arise.

- [x] 13. Local development infrastructure — Docker and Traefik
  - [x] 13.1 Create Dockerfiles for API and frontend apps
    - Create `apps/api/Dockerfile` — multi-stage Go build
    - Create `apps/booking-web/Dockerfile` — Next.js production build
    - Create `apps/tenant-web/Dockerfile` — Next.js production build
    - Create `apps/admin-web/Dockerfile` — Next.js production build
    - Create `apps/marketing-web/Dockerfile` — Next.js production build
    - _Requirements: 9.1, 13.2_

  - [x] 13.2 Create Docker Compose configuration
    - Create `docker-compose.yml` with services: traefik, api, marketing-web, booking-web, tenant-web, admin-web, postgres, redis
    - Configure PostgreSQL with `POSTGRES_DB=zenvikar`, `POSTGRES_USER=zenvikar`, `POSTGRES_PASSWORD=zenvikar`, port 5432
    - Configure Redis on port 6379
    - Configure `depends_on` so API starts after postgres and redis
    - Configure environment variables for API (DATABASE_URL, REDIS_URL, OTEL endpoint)
    - _Requirements: 9.1, 9.2, 9.3, 9.4_

  - [x] 13.3 Configure Traefik routing rules
    - Configure Traefik with Docker provider and entrypoint on port 80
    - Add routing labels: `zenvikar.localhost` → marketing-web, `api.zenvikar.localhost` → api, `manage.zenvikar.localhost` → tenant-web, `admin.zenvikar.localhost` → admin-web
    - Add wildcard routing: `{subdomain}.zenvikar.localhost` → booking-web with lower priority
    - Configure `X-Forwarded-Host` header forwarding for booking-web
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5, 8.6, 8.7_

- [x] 14. Observability stack configuration
  - [x] 14.1 Create OpenTelemetry Collector configuration
    - Create `infra/observability/otel-collector-config.yaml` with OTLP receiver, Loki exporter for logs, Tempo exporter for traces
    - _Requirements: 10.1, 10.3_

  - [x] 14.2 Create Prometheus configuration
    - Create `infra/observability/prometheus.yml` with scrape config targeting API `/metrics` endpoint
    - _Requirements: 10.1, 10.2_

  - [x] 14.3 Create Loki and Tempo configurations
    - Create `infra/observability/loki-config.yaml` with local storage config
    - Create `infra/observability/tempo-config.yaml` with local storage config
    - _Requirements: 10.1_

  - [x] 14.4 Create Grafana provisioning with data sources
    - Create `infra/observability/grafana/provisioning/datasources/datasources.yaml` pre-configuring Prometheus, Loki, and Tempo data sources
    - _Requirements: 10.4_

  - [x] 14.5 Add observability services to Docker Compose
    - Add Prometheus, Grafana, Loki, Tempo, and OTel Collector services to `docker-compose.yml`
    - Mount configuration files as volumes
    - _Requirements: 10.1_

- [x] 15. Checkpoint — Docker Compose starts all services
  - Ensure all tests pass, ask the user if questions arise.

- [x] 16. Terraform module structure
  - [x] 16.1 Create Terraform module skeletons
    - Create `infra/terraform/modules/networking/main.tf`, `variables.tf`, `outputs.tf` — VPC, subnets, security groups placeholders
    - Create `infra/terraform/modules/dns/main.tf`, `variables.tf`, `outputs.tf` — DNS zones, wildcard subdomains
    - Create `infra/terraform/modules/tls/main.tf`, `variables.tf`, `outputs.tf` — TLS certificates
    - Create `infra/terraform/modules/app-hosting/main.tf`, `variables.tf`, `outputs.tf` — Container hosting
    - Create `infra/terraform/modules/database/main.tf`, `variables.tf`, `outputs.tf` — Managed PostgreSQL
    - Create `infra/terraform/modules/cache/main.tf`, `variables.tf`, `outputs.tf` — Managed Redis
    - Create `infra/terraform/modules/secrets/main.tf`, `variables.tf`, `outputs.tf` — Secret management
    - Create `infra/terraform/modules/observability/main.tf`, `variables.tf`, `outputs.tf` — Cloud monitoring
    - _Requirements: 12.1_

  - [x] 16.2 Create Terraform environment configurations
    - Create `infra/terraform/environments/dev/main.tf`, `variables.tf`, `terraform.tfvars` composing modules
    - Create `infra/terraform/environments/staging/main.tf`, `variables.tf`, `terraform.tfvars`
    - Create `infra/terraform/environments/prod/main.tf`, `variables.tf`, `terraform.tfvars`
    - _Requirements: 12.2, 12.3_

- [x] 17. CI/CD GitHub Actions workflows
  - [x] 17.1 Create CI workflow for pull requests
    - Create `.github/workflows/ci.yml` with jobs: lint + type-check Go code, lint + type-check TypeScript code, run Go unit tests, run TypeScript tests
    - _Requirements: 13.1_

  - [x] 17.2 Create Docker build workflow
    - Create `.github/workflows/docker-build.yml` to build Docker images for API and all frontend apps
    - _Requirements: 13.2_

  - [x] 17.3 Create migration test workflow
    - Create `.github/workflows/migration-test.yml` to spin up PostgreSQL, apply all migrations, verify they apply cleanly
    - _Requirements: 13.3_

- [x] 18. Developer experience — Makefile, env, seed scripts, README
  - [x] 18.1 Create Makefile with common commands
    - Create `Makefile` with targets: `up` (docker-compose up), `down`, `build`, `migrate`, `seed`, `test`, `lint`, `logs`
    - _Requirements: 9.1_

  - [x] 18.2 Create environment configuration template
    - Create `.env.example` with all required environment variables documented
    - _Requirements: 9.2_

  - [x] 18.3 Create database seed script
    - Create `scripts/seed.sql` or `scripts/seed.go` that inserts sample tenants (e.g., "acme", "beta-salon"), sample users, memberships, services, opening hours, and bookings for local development
    - _Requirements: 11.2_

  - [x] 18.4 Create project README
    - Create `README.md` with: project overview, prerequisites, quick start guide (`make up`), architecture diagram reference, directory structure, available Make targets, links to design/requirements docs
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

- [x] 19. Placeholder modules for deferred features
  - [x] 19.1 Create placeholder module stubs
    - Create minimal module files for: `billing`, `reports`, `notifications`, `audit`, `staff`, `branding` — each implementing `Module` interface with empty `RegisterRoutes` and no-op `Migrate`
    - _Requirements: 1.2_

- [x] 20. Final checkpoint — Full build, all tests pass, docker-compose up works
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests (Properties 1–9) validate universal correctness properties from the design document using `pgregory.net/rapid` for Go
- Go is used for all backend code; TypeScript/Next.js for all frontend apps
- Migrations are numbered sequentially and run on API startup
