# Requirements Document

## Introduction

This document defines the requirements for the Zenvikar Platform Foundation — the infrastructure-first scaffolding for a SaaS multi-tenant booking platform with bilingual support (English and Portuguese). The foundation covers monorepo structure, multi-tenancy primitives (tenant resolution, RBAC, isolation), booking domain data models, local development environment (Docker Compose, Traefik), observability stack, Terraform module structure, and CI/CD via GitHub Actions. These requirements are derived from the approved design document and focus on the MVP foundation layer.

## Glossary

- **API**: The Go modular monolith backend serving all REST endpoints at `api.zenvikar.localhost`
- **Traefik**: The reverse proxy responsible for routing requests to frontend apps and the API based on hostname
- **Tenant**: An organization using the platform, identified by a unique slug used in subdomains and URL paths
- **Tenant_Slug**: A lowercase alphanumeric string (with hyphens) uniquely identifying a tenant, used in URLs
- **Booking_Web**: The customer-facing Next.js app that resolves tenants via subdomain (`*.zenvikar.localhost`)
- **Tenant_Web**: The tenant management Next.js app that resolves tenants via URL path (`/t/{slug}/...`)
- **Admin_Web**: The platform administration Next.js app at `admin.zenvikar.localhost`
- **Marketing_Web**: The public landing page Next.js app at `zenvikar.localhost`
- **Tenant_Resolver**: The component in the API that looks up tenant configuration by slug, with Redis caching
- **Slug_Extractor**: The function that parses a hostname to extract the tenant slug subdomain
- **Membership_Service**: The component that checks whether a user belongs to a tenant and their role
- **RBAC_Engine**: The authorization component that maps tenant roles to permissions
- **Booking_Service**: The component responsible for creating bookings and checking availability
- **Observability_Stack**: The set of tools (Prometheus, Grafana, Loki, Tempo, OTel Collector) providing metrics, logs, and traces
- **Monorepo**: The single repository containing all apps, packages, infrastructure, and CI/CD configuration
- **Reserved_Slugs**: A predefined list of slugs (e.g., `www`, `api`, `admin`) that cannot be used as tenant slugs

## Requirements

### Requirement 1: Monorepo Structure

**User Story:** As a developer, I want a well-organized monorepo with shared packages, so that all apps and services share types, UI components, and configuration without duplication.

#### Acceptance Criteria

1. THE Monorepo SHALL contain four Next.js frontend applications: Marketing_Web, Booking_Web, Tenant_Web, and Admin_Web
2. THE Monorepo SHALL contain one Go backend application (API) organized as a modular monolith
3. THE Monorepo SHALL include shared TypeScript packages for UI components, types, and configuration
4. THE Monorepo SHALL include an infrastructure directory with Terraform modules and Docker Compose configuration
5. THE Monorepo SHALL include a CI/CD directory with GitHub Actions workflow definitions

### Requirement 2: Tenant Slug Validation

**User Story:** As a platform operator, I want tenant slugs to follow strict validation rules, so that slugs are safe for use in URLs, subdomains, and database lookups.

#### Acceptance Criteria

1. WHEN a tenant slug is submitted, THE Slug_Validator SHALL accept the slug only if its length is between 3 and 63 characters
2. WHEN a tenant slug is submitted, THE Slug_Validator SHALL accept the slug only if it matches the pattern `^[a-z0-9][a-z0-9-]*[a-z0-9]$`
3. WHEN a tenant slug is submitted, THE Slug_Validator SHALL reject the slug if it contains consecutive hyphens
4. WHEN a tenant slug matches an entry in the Reserved_Slugs list, THE Slug_Validator SHALL reject the slug with a descriptive error
5. IF a tenant slug fails validation, THEN THE Slug_Validator SHALL return a specific error describing the validation failure

### Requirement 3: Subdomain Tenant Extraction

**User Story:** As a customer visiting a tenant's booking page, I want the system to identify the tenant from the subdomain, so that I see the correct tenant's branding and services.

#### Acceptance Criteria

1. WHEN a request arrives with host `{slug}.{baseDomain}`, THE Slug_Extractor SHALL return the slug portion as a lowercase string
2. WHEN a request arrives with host equal to the baseDomain (no subdomain), THE Slug_Extractor SHALL return an ErrNoSubdomain error
3. WHEN a request arrives with a host that does not end with `.{baseDomain}`, THE Slug_Extractor SHALL return an ErrInvalidHost error
4. WHEN a host contains a port number, THE Slug_Extractor SHALL strip the port before extracting the slug
5. WHEN a host contains nested subdomains (e.g., `a.b.zenvikar.localhost`), THE Slug_Extractor SHALL return an ErrInvalidHost error

### Requirement 4: Tenant Resolution with Caching

**User Story:** As a frontend application, I want to resolve tenant configuration by slug with caching, so that tenant lookups are fast and do not overload the database.

#### Acceptance Criteria

1. WHEN a valid slug is provided and the tenant exists and is enabled, THE Tenant_Resolver SHALL return the full tenant configuration object
2. WHEN a valid slug is provided and the tenant exists but is disabled, THE Tenant_Resolver SHALL return an ErrTenantDisabled error
3. WHEN a valid slug is provided and no tenant exists with that slug, THE Tenant_Resolver SHALL return an ErrTenantNotFound error
4. WHEN a tenant is resolved successfully, THE Tenant_Resolver SHALL cache the result in Redis with a TTL of 5 minutes
5. WHEN a cached tenant entry exists in Redis, THE Tenant_Resolver SHALL return the cached result without querying the database
6. THE Tenant_Resolver SHALL return identical tenant data regardless of whether the result comes from cache or database

### Requirement 5: Tenant Membership and RBAC

**User Story:** As a tenant owner, I want users to have specific roles within my tenant, so that access to features is controlled based on responsibilities.

#### Acceptance Criteria

1. THE Membership_Service SHALL enforce that each user has at most one membership per tenant
2. WHEN a user's membership is checked, THE Membership_Service SHALL return the membership with role if it exists, or ErrNotAMember if it does not
3. THE RBAC_Engine SHALL support four tenant roles: tenant_owner, tenant_manager, tenant_staff, and tenant_finance_viewer
4. THE RBAC_Engine SHALL support three platform roles: admin, support_admin, and finance_admin
5. WHEN an authorization check is performed, THE RBAC_Engine SHALL grant access if the user is a platform admin, regardless of tenant membership
6. WHEN an authorization check is performed for a tenant action, THE RBAC_Engine SHALL grant access only if the user has a tenant membership with a role that includes the required permission
7. THE RBAC_Engine SHALL grant tenant_owner all permissions within their tenant

### Requirement 6: Booking Domain Data Models

**User Story:** As a developer, I want well-defined booking domain data models with database constraints, so that the booking system has a solid data foundation for the MVP.

#### Acceptance Criteria

1. THE API SHALL store bookings with tenant_id, service_id, customer_id, start_time, end_time, status, and timezone fields
2. THE Database SHALL enforce that booking end_time is greater than booking start_time
3. THE Database SHALL prevent double-booking by enforcing a unique constraint on (tenant_id, service_id, start_time) for non-cancelled bookings
4. THE Database SHALL enforce that service duration_minutes is greater than zero
5. THE Database SHALL enforce that opening_hours day_of_week is between 0 and 6
6. THE Database SHALL enforce that each tenant has at most one opening_hours entry per day_of_week
7. THE Database SHALL enforce that each tenant has at most one blocked_date entry per date
8. WHEN a booking is created, THE Booking_Service SHALL set the status to "pending"
9. THE Database SHALL restrict booking status to one of: pending, confirmed, or cancelled

### Requirement 7: Booking Availability Checking

**User Story:** As a customer, I want the system to check availability before accepting a booking, so that I cannot book a time slot that is unavailable.

#### Acceptance Criteria

1. WHEN a booking is requested for a blocked date, THE Booking_Service SHALL reject the booking with a "date_blocked" reason
2. WHEN a booking is requested outside of the tenant's opening hours, THE Booking_Service SHALL reject the booking with an "outside_hours" reason
3. WHEN a booking is requested for a time slot that overlaps an existing non-cancelled booking (including buffer times), THE Booking_Service SHALL reject the booking with a "slot_taken" reason
4. WHEN a booking is requested, THE Booking_Service SHALL calculate the end_time as start_time plus the service's duration_minutes
5. WHEN checking for overlapping bookings, THE Booking_Service SHALL account for the service's buffer_before_minutes and buffer_after_minutes

### Requirement 8: Traefik Reverse Proxy Routing

**User Story:** As a developer, I want Traefik to route requests to the correct service based on hostname, so that all apps are accessible via their designated subdomains in local development.

#### Acceptance Criteria

1. WHEN a request arrives for `zenvikar.localhost`, THE Traefik SHALL route it to Marketing_Web
2. WHEN a request arrives for `api.zenvikar.localhost`, THE Traefik SHALL route it to the API
3. WHEN a request arrives for `manage.zenvikar.localhost`, THE Traefik SHALL route it to Tenant_Web
4. WHEN a request arrives for `admin.zenvikar.localhost`, THE Traefik SHALL route it to Admin_Web
5. WHEN a request arrives for `{slug}.zenvikar.localhost` (wildcard), THE Traefik SHALL route it to Booking_Web
6. THE Traefik SHALL prioritize explicit hostname matches over wildcard matches
7. WHEN routing to Booking_Web, THE Traefik SHALL forward the original Host header via X-Forwarded-Host

### Requirement 9: Local Development Environment

**User Story:** As a developer, I want a Docker Compose environment that runs all services locally, so that I can develop and test the full platform without cloud dependencies.

#### Acceptance Criteria

1. THE Docker_Compose SHALL define services for Traefik, API, all four frontend apps, PostgreSQL, and Redis
2. THE Docker_Compose SHALL configure PostgreSQL with a default database, user, and password for local development
3. THE Docker_Compose SHALL expose PostgreSQL on port 5432 and Redis on port 6379 for direct access
4. THE Docker_Compose SHALL configure service dependencies so that the API starts after PostgreSQL and Redis
5. THE API SHALL expose a `/healthz` liveness endpoint and a `/readyz` readiness endpoint that checks database and Redis connectivity

### Requirement 10: Observability Stack

**User Story:** As a developer, I want a local observability stack with metrics, logs, and traces, so that I can monitor and debug the platform during development.

#### Acceptance Criteria

1. THE Docker_Compose SHALL include Prometheus, Grafana, Loki, Tempo, and an OpenTelemetry Collector
2. THE API SHALL expose a `/metrics` endpoint in Prometheus format
3. THE API SHALL send traces and logs to the OpenTelemetry Collector via OTLP
4. THE Grafana SHALL be pre-configured with data sources for Prometheus, Loki, and Tempo
5. THE API SHALL use structured JSON logging via Go's slog package

### Requirement 11: Database Migrations

**User Story:** As a developer, I want database migrations managed in code, so that schema changes are versioned, repeatable, and applied automatically.

#### Acceptance Criteria

1. THE API SHALL run database migrations on startup for each registered module
2. THE Migrations SHALL create tables for tenants, users, tenant_memberships, platform_admins, services, opening_hours, blocked_dates, and bookings
3. THE Migrations SHALL create appropriate indexes including a partial index on tenants(slug) where enabled is true
4. IF a migration fails, THEN THE API SHALL log the error and terminate startup

### Requirement 12: Terraform Module Structure

**User Story:** As a DevOps engineer, I want Terraform modules organized by concern with environment separation, so that cloud infrastructure can be provisioned consistently across environments.

#### Acceptance Criteria

1. THE Terraform directory SHALL contain reusable modules for networking, DNS, TLS, app-hosting, database, cache, secrets, and observability
2. THE Terraform directory SHALL contain environment directories for dev, staging, and prod
3. WHEN a new environment is provisioned, THE Terraform modules SHALL be composable and reusable across environments

### Requirement 13: CI/CD Pipeline

**User Story:** As a developer, I want GitHub Actions workflows for continuous integration and deployment, so that code changes are automatically tested and deployable.

#### Acceptance Criteria

1. THE CI pipeline SHALL run linting, type checking, and unit tests for all Go and TypeScript code on pull requests
2. THE CI pipeline SHALL build Docker images for the API and all frontend apps
3. THE CI pipeline SHALL run database migration tests to verify migrations apply cleanly

### Requirement 14: Tenant Data Isolation

**User Story:** As a platform operator, I want strict tenant data isolation, so that no tenant can access another tenant's data.

#### Acceptance Criteria

1. THE API SHALL include tenant_id in all database queries for tenant-scoped resources
2. THE Database SHALL enforce foreign key constraints from tenant-scoped tables to the tenants table
3. WHEN a user requests data, THE API SHALL verify the user has membership in the target tenant before returning results
4. THE Database SHALL cascade-delete tenant-scoped data when a tenant is deleted

### Requirement 15: Bilingual Support

**User Story:** As a user, I want the platform to support English and Portuguese, so that I can use the platform in my preferred language.

#### Acceptance Criteria

1. THE Frontend_Apps SHALL support two locales: "en" (English) and "pt" (Portuguese)
2. THE Database SHALL restrict tenant default_locale and user locale to "en" or "pt"
3. THE Booking_Web SHALL apply the tenant's default locale when rendering the booking page

### Requirement 16: Error Handling

**User Story:** As a user, I want clear error responses when something goes wrong, so that I understand what happened and how to recover.

#### Acceptance Criteria

1. WHEN a tenant is not found, THE API SHALL return HTTP 404 with error code "tenant_not_found"
2. WHEN a tenant is disabled, THE API SHALL return HTTP 403 with error code "tenant_disabled"
3. WHEN a booking slot conflict occurs, THE API SHALL return HTTP 409 with error code "slot_unavailable"
4. WHEN a reserved slug is used for tenant creation, THE API SHALL return HTTP 422 with error code "slug_reserved"
5. WHEN a user lacks tenant membership, THE API SHALL return HTTP 403 Forbidden
6. IF the database or Redis is unreachable, THEN THE API SHALL return HTTP 503 for data-dependent requests and report unhealthy on the readiness endpoint
