# Design Document: Zenvikar Platform Foundation

## Overview

Zenvikar is a SaaS multi-tenant booking platform with bilingual support (English and Portuguese). The platform consists of five applications: a marketing website, a customer-facing booking app (multi-tenant by subdomain), a tenant management portal (tenant context via URL path), a platform administration portal, and a Go modular monolith backend API.

This design document covers the infrastructure-first foundation: monorepo scaffolding, multi-tenancy primitives, booking domain structure, local development environment with Docker Compose and Traefik, observability stack (Prometheus, Grafana, Loki, Tempo, OpenTelemetry, Sentry placeholders), Terraform modules for future cloud deployment, and CI/CD via GitHub Actions. The goal is to establish a clear, maintainable foundation optimized for MVP speed without premature enterprise complexity.

The architecture follows a modular monolith pattern for the backend, avoiding microservices. All frontend apps share UI components, types, and configuration through a monorepo with shared packages. Tenant resolution happens at the edge (subdomain for booking-web, URL path for tenant-web) and is propagated to the backend via headers and middleware.

## Architecture

```mermaid
graph TD
    subgraph "Public Internet"
        Browser[Browser]
    end

    subgraph "Reverse Proxy"
        Traefik[Traefik]
    end

    subgraph "Frontend Apps"
        MW[marketing-web<br/>zenvikar.localhost]
        BW[booking-web<br/>*.zenvikar.localhost]
        TW[tenant-web<br/>manage.zenvikar.localhost]
        AW[admin-web<br/>admin.zenvikar.localhost]
    end

    subgraph "Backend"
        API[api - Go Modular Monolith<br/>api.zenvikar.localhost]
    end

    subgraph "Data Layer"
        PG[(PostgreSQL)]
        RD[(Redis)]
    end

    subgraph "Observability"
        PROM[Prometheus]
        GRAF[Grafana]
        LOKI[Loki]
        TEMPO[Tempo]
        OTEL[OTel Collector]
    end

    Browser --> Traefik
    Traefik --> MW
    Traefik --> BW
    Traefik --> TW
    Traefik --> AW
    Traefik --> API

    API --> PG
    API --> RD
    API --> OTEL

    OTEL --> LOKI
    OTEL --> TEMPO
    PROM --> API
    GRAF --> PROM
    GRAF --> LOKI
    GRAF --> TEMPO
```

## Sequence Diagrams

### Booking-Web Tenant Resolution Flow

```mermaid
sequenceDiagram
    participant B as Browser
    participant T as Traefik
    participant BW as booking-web
    participant API as api (Go)
    participant PG as PostgreSQL
    participant RD as Redis

    B->>T: GET https://acme.zenvikar.localhost
    T->>BW: Forward (X-Forwarded-Host: acme.zenvikar.localhost)
    BW->>BW: Extract subdomain "acme" from Host header
    BW->>API: GET /api/v1/tenants/resolve?slug=acme
    API->>RD: Check tenant cache (tenant:slug:acme)
    alt Cache hit
        RD-->>API: Tenant config JSON
    else Cache miss
        API->>PG: SELECT * FROM tenants WHERE slug = 'acme' AND enabled = true
        PG-->>API: Tenant row
        API->>RD: SET tenant:slug:acme (TTL 5min)
    end
    API-->>BW: { id, slug, displayName, logoUrl, colors, timezone, locale }
    BW->>BW: Apply tenant branding, set locale
    BW-->>B: Rendered booking page with tenant branding
```

### Tenant-Web Multi-Tenant Switching Flow

```mermaid
sequenceDiagram
    participant B as Browser
    participant T as Traefik
    participant TW as tenant-web
    participant API as api (Go)

    B->>T: GET https://manage.zenvikar.localhost/t/acme/dashboard
    T->>TW: Forward request
    TW->>TW: Extract tenantSlug "acme" from URL path
    TW->>API: GET /api/v1/tenants/acme (with auth token)
    API->>API: Verify user has membership in tenant "acme"
    API-->>TW: Tenant config + user role
    TW-->>B: Rendered dashboard for tenant "acme"

    Note over B: User switches tenant
    B->>T: GET https://manage.zenvikar.localhost/t/beta-salon/dashboard
    T->>TW: Forward request
    TW->>TW: Extract tenantSlug "beta-salon" from URL path
    TW->>API: GET /api/v1/tenants/beta-salon (with auth token)
    API->>API: Verify user has membership in tenant "beta-salon"
    API-->>TW: Tenant config + user role
    TW-->>B: Rendered dashboard for tenant "beta-salon"
```

### API Request Authentication & Authorization Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant API as api (Go)
    participant Auth as auth module
    participant Tenants as tenants module
    participant PG as PostgreSQL

    C->>API: POST /api/v1/bookings (Authorization: Bearer <token>)
    API->>Auth: Middleware: ValidateToken(token)
    Auth->>Auth: Decode JWT, extract userID + claims
    Auth-->>API: UserContext { userID, email }

    API->>Tenants: Middleware: ResolveTenant(X-Tenant-ID header)
    Tenants->>PG: Lookup tenant + membership for userID
    Tenants-->>API: TenantContext { tenantID, role }

    API->>API: RBAC check: role has permission for endpoint
    alt Authorized
        API->>API: Execute handler with UserContext + TenantContext
        API-->>C: 200 OK
    else Unauthorized
        API-->>C: 403 Forbidden
    end
```


## Components and Interfaces

### Component 1: Traefik Reverse Proxy

**Purpose**: Routes all incoming requests to the correct frontend or backend service based on hostname. Passes original `Host` header to booking-web for tenant resolution.

**Routing Rules**:
| Hostname Pattern | Target Service |
|---|---|
| `zenvikar.localhost` | marketing-web:3000 |
| `*.zenvikar.localhost` (wildcard) | booking-web:3001 |
| `manage.zenvikar.localhost` | tenant-web:3002 |
| `admin.zenvikar.localhost` | admin-web:3003 |
| `api.zenvikar.localhost` | api:8080 |

**Responsibilities**:
- TLS termination (self-signed for local dev)
- Wildcard subdomain routing for booking-web
- Forward `X-Forwarded-Host` header to booking-web
- Priority rules: explicit hostnames match before wildcard
- Health check routing for all services

### Component 2: API - Go Modular Monolith

**Purpose**: Single Go binary serving all backend logic, organized as internal modules with clear boundaries.

**Interface**:
```go
// cmd/api/main.go - Application entry point
package main

// Module registry - each module registers its routes
type Module interface {
    Name() string
    RegisterRoutes(router Router, deps Dependencies)
    Migrate(db *sql.DB) error
}

// Dependencies injected into each module
type Dependencies struct {
    DB          *sql.DB
    Redis       *redis.Client
    Logger      *slog.Logger
    Tracer      trace.Tracer
    Config      *config.Config
}

// Health endpoints (not behind auth)
// GET  /healthz          - liveness probe
// GET  /readyz           - readiness probe (checks DB + Redis)
// GET  /metrics          - Prometheus metrics
```

**Module Structure**:
```
apps/api/
├── cmd/api/main.go
├── internal/
│   ├── platform/          # Shared utilities (logger, config, db, redis, otel, middleware)
│   ├── auth/              # Authentication (JWT, sessions)
│   ├── users/             # User management
│   ├── tenants/           # Tenant CRUD, resolution, branding
│   ├── tenant_memberships/ # User-tenant relationships, RBAC
│   ├── branding/          # Tenant branding config
│   ├── bookings/          # Booking domain
│   ├── availability/      # Opening hours, blocked dates, slots
│   ├── services/          # Bookable services
│   ├── staff/             # Staff management
│   ├── billing/           # Billing (placeholder)
│   ├── reports/           # Reporting (placeholder)
│   ├── notifications/     # Notification dispatch (placeholder)
│   └── audit/             # Audit logging (placeholder)
├── migrations/
├── go.mod
└── go.sum
```

**Responsibilities**:
- Serve REST API for all frontend apps
- Tenant resolution and caching
- Authentication and RBAC enforcement
- Database migrations
- Structured logging with slog
- OpenTelemetry tracing and metrics
- Health/readiness/liveness endpoints

### Component 3: Frontend Apps (Next.js + TypeScript)

**Purpose**: Four Next.js applications sharing UI components, types, and configuration through monorepo packages.

**Interface**:
```typescript
// packages/types/src/tenant.ts
export interface Tenant {
  id: string;
  slug: string;
  displayName: string;
  logoUrl: string | null;
  colors: TenantColors;
  timezone: string;
  defaultLocale: "en" | "pt";
  enabled: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface TenantColors {
  primary: string;
  secondary: string;
  accent: string;
}

// packages/types/src/user.ts
export interface User {
  id: string;
  email: string;
  name: string;
  locale: "en" | "pt";
}

export type TenantRole =
  | "tenant_owner"
  | "tenant_manager"
  | "tenant_staff"
  | "tenant_finance_viewer";

export type PlatformRole =
  | "admin"
  | "support_admin"
  | "finance_admin";

export interface TenantMembership {
  tenantId: string;
  userId: string;
  role: TenantRole;
}
```

**Shared Packages**:
- `packages/ui` - Shared React components (buttons, forms, layouts, tenant-branded wrappers)
- `packages/types` - TypeScript interfaces shared between all apps and API client
- `packages/config` - Shared configuration (i18n setup, API client config, environment helpers)

**Responsibilities**:
- `marketing-web`: Public landing pages, SEO, bilingual content
- `booking-web`: Tenant-branded booking experience, subdomain-based tenant resolution
- `tenant-web`: Tenant management dashboard, URL-path-based tenant switching
- `admin-web`: Platform administration, user/tenant management

### Component 4: Observability Stack

**Purpose**: Full observability for local development with production-ready patterns.

**Components**:
| Tool | Role | Data Source |
|---|---|---|
| Prometheus | Metrics collection | Scrapes /metrics from api |
| Grafana | Dashboards | Queries Prometheus, Loki, Tempo |
| Loki | Log aggregation | Receives logs from OTel Collector |
| Tempo | Distributed tracing | Receives traces from OTel Collector |
| OTel Collector | Telemetry pipeline | Receives from api via OTLP |
| Sentry | Error tracking | Placeholder DSN in frontend/backend |

### Component 5: Terraform Infrastructure

**Purpose**: Infrastructure as Code for cloud deployment with reusable modules.

**Module Structure**:
```
infra/terraform/
├── modules/
│   ├── networking/        # VPC, subnets, security groups
│   ├── dns/               # DNS zones, wildcard subdomains
│   ├── tls/               # TLS certificates (ACM/Let's Encrypt)
│   ├── app-hosting/       # Container hosting (ECS/Cloud Run/etc.)
│   ├── database/          # Managed PostgreSQL
│   ├── cache/             # Managed Redis
│   ├── secrets/           # Secret management
│   └── observability/     # Cloud monitoring setup
└── environments/
    ├── dev/
    │   ├── main.tf
    │   ├── variables.tf
    │   └── terraform.tfvars
    ├── staging/
    └── prod/
```


## Data Models

### Model 1: Tenant

```go
// internal/tenants/model.go
type Tenant struct {
    ID            uuid.UUID       `json:"id" db:"id"`
    Slug          string          `json:"slug" db:"slug"`
    DisplayName   string          `json:"displayName" db:"display_name"`
    LogoURL       *string         `json:"logoUrl" db:"logo_url"`
    ColorPrimary  string          `json:"colorPrimary" db:"color_primary"`
    ColorSecondary string         `json:"colorSecondary" db:"color_secondary"`
    ColorAccent   string          `json:"colorAccent" db:"color_accent"`
    Timezone      string          `json:"timezone" db:"timezone"`
    DefaultLocale string          `json:"defaultLocale" db:"default_locale"`
    Enabled       bool            `json:"enabled" db:"enabled"`
    CreatedAt     time.Time       `json:"createdAt" db:"created_at"`
    UpdatedAt     time.Time       `json:"updatedAt" db:"updated_at"`
}
```

**Validation Rules**:
- `slug`: lowercase alphanumeric + hyphens, 3-63 chars, unique, not in reserved slugs list
- `displayName`: 1-255 chars, non-empty
- `timezone`: valid IANA timezone string
- `defaultLocale`: must be "en" or "pt"
- `colorPrimary/Secondary/Accent`: valid hex color codes (#RRGGBB)

**Reserved Slugs**: `www`, `api`, `admin`, `manage`, `app`, `mail`, `smtp`, `ftp`, `ssh`, `git`, `cdn`, `static`, `assets`, `media`, `blog`, `docs`, `help`, `support`, `status`, `billing`, `zenvikar`

```sql
-- migrations/001_create_tenants.sql
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(63) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    logo_url TEXT,
    color_primary VARCHAR(7) NOT NULL DEFAULT '#000000',
    color_secondary VARCHAR(7) NOT NULL DEFAULT '#666666',
    color_accent VARCHAR(7) NOT NULL DEFAULT '#0066FF',
    timezone VARCHAR(64) NOT NULL DEFAULT 'UTC',
    default_locale VARCHAR(2) NOT NULL DEFAULT 'en' CHECK (default_locale IN ('en', 'pt')),
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tenants_slug ON tenants(slug) WHERE enabled = true;
```

### Model 2: User

```go
// internal/users/model.go
type User struct {
    ID            uuid.UUID  `json:"id" db:"id"`
    Email         string     `json:"email" db:"email"`
    Name          string     `json:"name" db:"name"`
    PasswordHash  string     `json:"-" db:"password_hash"`
    Locale        string     `json:"locale" db:"locale"`
    EmailVerified bool       `json:"emailVerified" db:"email_verified"`
    CreatedAt     time.Time  `json:"createdAt" db:"created_at"`
    UpdatedAt     time.Time  `json:"updatedAt" db:"updated_at"`
}
```

```sql
-- migrations/002_create_users.sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    password_hash TEXT NOT NULL,
    locale VARCHAR(2) NOT NULL DEFAULT 'en' CHECK (locale IN ('en', 'pt')),
    email_verified BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Model 3: Tenant Membership

```go
// internal/tenant_memberships/model.go
type TenantMembership struct {
    ID        uuid.UUID  `json:"id" db:"id"`
    TenantID  uuid.UUID  `json:"tenantId" db:"tenant_id"`
    UserID    uuid.UUID  `json:"userId" db:"user_id"`
    Role      string     `json:"role" db:"role"`
    CreatedAt time.Time  `json:"createdAt" db:"created_at"`
    UpdatedAt time.Time  `json:"updatedAt" db:"updated_at"`
}

// Valid tenant roles
const (
    RoleTenantOwner         = "tenant_owner"
    RoleTenantManager       = "tenant_manager"
    RoleTenantStaff         = "tenant_staff"
    RoleTenantFinanceViewer = "tenant_finance_viewer"
)
```

```sql
-- migrations/003_create_tenant_memberships.sql
CREATE TABLE tenant_memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(32) NOT NULL CHECK (role IN (
        'tenant_owner', 'tenant_manager', 'tenant_staff', 'tenant_finance_viewer'
    )),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, user_id)
);

CREATE INDEX idx_tenant_memberships_user ON tenant_memberships(user_id);
CREATE INDEX idx_tenant_memberships_tenant ON tenant_memberships(tenant_id);
```

### Model 4: Platform Admin

```go
// internal/users/platform_admin.go
type PlatformAdmin struct {
    ID        uuid.UUID  `json:"id" db:"id"`
    UserID    uuid.UUID  `json:"userId" db:"user_id"`
    Role      string     `json:"role" db:"role"`
    CreatedAt time.Time  `json:"createdAt" db:"created_at"`
}

const (
    RoleAdmin        = "admin"
    RoleSupportAdmin = "support_admin"
    RoleFinanceAdmin = "finance_admin"
)
```

```sql
-- migrations/004_create_platform_admins.sql
CREATE TABLE platform_admins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(32) NOT NULL CHECK (role IN ('admin', 'support_admin', 'finance_admin')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Model 5: Booking Domain (Foundation)

```go
// internal/services/model.go
type Service struct {
    ID              uuid.UUID  `json:"id" db:"id"`
    TenantID        uuid.UUID  `json:"tenantId" db:"tenant_id"`
    Name            string     `json:"name" db:"name"`
    DurationMinutes int        `json:"durationMinutes" db:"duration_minutes"`
    BufferBefore    int        `json:"bufferBefore" db:"buffer_before_minutes"`
    BufferAfter     int        `json:"bufferAfter" db:"buffer_after_minutes"`
    Enabled         bool       `json:"enabled" db:"enabled"`
    CreatedAt       time.Time  `json:"createdAt" db:"created_at"`
    UpdatedAt       time.Time  `json:"updatedAt" db:"updated_at"`
}

// internal/availability/model.go
type OpeningHours struct {
    ID        uuid.UUID  `json:"id" db:"id"`
    TenantID  uuid.UUID  `json:"tenantId" db:"tenant_id"`
    DayOfWeek int        `json:"dayOfWeek" db:"day_of_week"` // 0=Sunday, 6=Saturday
    OpenTime  string     `json:"openTime" db:"open_time"`     // "09:00"
    CloseTime string     `json:"closeTime" db:"close_time"`   // "18:00"
    Enabled   bool       `json:"enabled" db:"enabled"`
}

type BlockedDate struct {
    ID        uuid.UUID  `json:"id" db:"id"`
    TenantID  uuid.UUID  `json:"tenantId" db:"tenant_id"`
    Date      time.Time  `json:"date" db:"date"`
    Reason    *string    `json:"reason" db:"reason"`
}

// internal/bookings/model.go
type Booking struct {
    ID         uuid.UUID  `json:"id" db:"id"`
    TenantID   uuid.UUID  `json:"tenantId" db:"tenant_id"`
    ServiceID  uuid.UUID  `json:"serviceId" db:"service_id"`
    CustomerID uuid.UUID  `json:"customerId" db:"customer_id"`
    StartTime  time.Time  `json:"startTime" db:"start_time"`
    EndTime    time.Time  `json:"endTime" db:"end_time"`
    Status     string     `json:"status" db:"status"` // pending, confirmed, cancelled
    Timezone   string     `json:"timezone" db:"timezone"`
    CreatedAt  time.Time  `json:"createdAt" db:"created_at"`
    UpdatedAt  time.Time  `json:"updatedAt" db:"updated_at"`
}
```

```sql
-- migrations/005_create_booking_domain.sql
CREATE TABLE services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    duration_minutes INT NOT NULL CHECK (duration_minutes > 0),
    buffer_before_minutes INT NOT NULL DEFAULT 0 CHECK (buffer_before_minutes >= 0),
    buffer_after_minutes INT NOT NULL DEFAULT 0 CHECK (buffer_after_minutes >= 0),
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE opening_hours (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    day_of_week INT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    open_time TIME NOT NULL,
    close_time TIME NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    UNIQUE(tenant_id, day_of_week)
);

CREATE TABLE blocked_dates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    reason TEXT,
    UNIQUE(tenant_id, date)
);

CREATE TABLE bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    service_id UUID NOT NULL REFERENCES services(id),
    customer_id UUID NOT NULL REFERENCES users(id),
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'cancelled')),
    timezone VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (end_time > start_time)
);

-- Prevent double-booking per tenant+time window
CREATE UNIQUE INDEX idx_bookings_no_overlap
    ON bookings(tenant_id, service_id, start_time)
    WHERE status != 'cancelled';

CREATE INDEX idx_bookings_tenant_time ON bookings(tenant_id, start_time, end_time);
```


## Key Functions with Formal Specifications

### Function 1: ResolveTenantBySlug

```go
// internal/tenants/service.go
func (s *TenantService) ResolveTenantBySlug(ctx context.Context, slug string) (*Tenant, error)
```

**Preconditions:**
- `slug` is a non-empty, lowercase string matching `^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$`
- `slug` is not in the reserved slugs list
- Database connection is available

**Postconditions:**
- If tenant exists and is enabled: returns `*Tenant` with all fields populated, `error` is nil
- If tenant exists but is disabled: returns nil, `ErrTenantDisabled`
- If tenant does not exist: returns nil, `ErrTenantNotFound`
- Result is cached in Redis with TTL of 5 minutes on success
- No mutations to any data

**Loop Invariants:** N/A

### Function 2: ExtractTenantSlugFromHost

```go
// internal/platform/middleware/tenant.go
func ExtractTenantSlugFromHost(host string, baseDomain string) (string, error)
```

**Preconditions:**
- `host` is a non-empty string (may include port)
- `baseDomain` is a non-empty string representing the platform domain (e.g., "zenvikar.localhost")

**Postconditions:**
- If host is `{slug}.{baseDomain}`: returns slug, nil
- If host is exactly `baseDomain`: returns "", `ErrNoSubdomain`
- If host does not end with `.{baseDomain}`: returns "", `ErrInvalidHost`
- Returned slug is always lowercase
- No side effects

**Loop Invariants:** N/A

### Function 3: CheckMembership

```go
// internal/tenant_memberships/service.go
func (s *MembershipService) CheckMembership(
    ctx context.Context,
    userID uuid.UUID,
    tenantID uuid.UUID,
) (*TenantMembership, error)
```

**Preconditions:**
- `userID` and `tenantID` are valid non-zero UUIDs
- Database connection is available
- User exists in users table

**Postconditions:**
- If membership exists: returns `*TenantMembership` with role, nil
- If no membership: returns nil, `ErrNotAMember`
- Returned membership always has a valid role from the allowed set
- No mutations to any data

**Loop Invariants:** N/A

### Function 4: ValidateTenantSlug

```go
// internal/tenants/validation.go
func ValidateTenantSlug(slug string) error
```

**Preconditions:**
- `slug` is a string (may be empty)

**Postconditions:**
- Returns nil if and only if all of:
  - `slug` length is between 3 and 63 characters
  - `slug` matches `^[a-z0-9][a-z0-9-]*[a-z0-9]$`
  - `slug` does not contain consecutive hyphens
  - `slug` is not in the reserved slugs list
- Returns descriptive error otherwise
- No side effects

**Loop Invariants:** N/A

### Function 5: CreateBooking

```go
// internal/bookings/service.go
func (s *BookingService) CreateBooking(
    ctx context.Context,
    tenantID uuid.UUID,
    req CreateBookingRequest,
) (*Booking, error)
```

**Preconditions:**
- `tenantID` is a valid UUID for an enabled tenant
- `req.ServiceID` references an enabled service belonging to `tenantID`
- `req.CustomerID` is a valid user UUID
- `req.StartTime` is in the future
- Database connection is available

**Postconditions:**
- If slot is available and valid: returns `*Booking` with status "pending", nil
- If time slot overlaps existing non-cancelled booking: returns nil, `ErrSlotUnavailable`
- If date is blocked: returns nil, `ErrDateBlocked`
- If outside opening hours: returns nil, `ErrOutsideOpeningHours`
- `booking.EndTime = booking.StartTime + service.DurationMinutes`
- Booking is persisted in database within a transaction
- No partial writes on error

**Loop Invariants:** N/A

## Algorithmic Pseudocode

### Tenant Resolution Algorithm (with caching)

```go
// ResolveTenantBySlug - full algorithm
func (s *TenantService) ResolveTenantBySlug(ctx context.Context, slug string) (*Tenant, error) {
    // Step 1: Validate slug format
    if err := ValidateTenantSlug(slug); err != nil {
        return nil, fmt.Errorf("invalid slug: %w", err)
    }

    // Step 2: Check Redis cache
    cacheKey := fmt.Sprintf("tenant:slug:%s", slug)
    cached, err := s.redis.Get(ctx, cacheKey).Result()
    if err == nil {
        var tenant Tenant
        if err := json.Unmarshal([]byte(cached), &tenant); err == nil {
            return &tenant, nil
        }
    }

    // Step 3: Query database
    tenant, err := s.repo.FindBySlug(ctx, slug)
    if err != nil {
        return nil, ErrTenantNotFound
    }

    // Step 4: Check enabled status
    if !tenant.Enabled {
        return nil, ErrTenantDisabled
    }

    // Step 5: Cache result
    data, _ := json.Marshal(tenant)
    s.redis.Set(ctx, cacheKey, data, 5*time.Minute)

    return tenant, nil
}
```

### Subdomain Extraction Algorithm

```go
// ExtractTenantSlugFromHost - extracts tenant slug from hostname
func ExtractTenantSlugFromHost(host string, baseDomain string) (string, error) {
    // Step 1: Strip port if present
    hostname := host
    if idx := strings.LastIndex(host, ":"); idx != -1 {
        hostname = host[:idx]
    }
    hostname = strings.ToLower(hostname)

    // Step 2: Check if host ends with base domain
    if !strings.HasSuffix(hostname, "."+baseDomain) {
        if hostname == baseDomain {
            return "", ErrNoSubdomain
        }
        return "", ErrInvalidHost
    }

    // Step 3: Extract subdomain
    slug := strings.TrimSuffix(hostname, "."+baseDomain)

    // Step 4: Validate no nested subdomains
    if strings.Contains(slug, ".") {
        return "", ErrInvalidHost
    }

    return slug, nil
}
```

### Booking Availability Check Algorithm

```go
// CheckAvailability - determines if a time slot is bookable
func (s *BookingService) CheckAvailability(
    ctx context.Context,
    tenantID uuid.UUID,
    serviceID uuid.UUID,
    startTime time.Time,
) (*AvailabilityResult, error) {
    // Step 1: Load tenant config (for timezone)
    tenant, err := s.tenants.ResolveTenantByID(ctx, tenantID)
    if err != nil {
        return nil, err
    }

    // Step 2: Load service (for duration + buffers)
    service, err := s.services.GetByID(ctx, serviceID)
    if err != nil {
        return nil, err
    }

    loc, _ := time.LoadLocation(tenant.Timezone)
    localStart := startTime.In(loc)
    endTime := startTime.Add(time.Duration(service.DurationMinutes) * time.Minute)

    // Step 3: Check if date is blocked
    blocked, err := s.availability.IsDateBlocked(ctx, tenantID, localStart)
    if blocked {
        return &AvailabilityResult{Available: false, Reason: "date_blocked"}, nil
    }

    // Step 4: Check opening hours for the day
    dayOfWeek := int(localStart.Weekday())
    hours, err := s.availability.GetOpeningHours(ctx, tenantID, dayOfWeek)
    if err != nil || !hours.Enabled {
        return &AvailabilityResult{Available: false, Reason: "outside_hours"}, nil
    }

    // Step 5: Verify time falls within opening hours
    openTime := parseTimeInLocation(hours.OpenTime, localStart, loc)
    closeTime := parseTimeInLocation(hours.CloseTime, localStart, loc)
    if localStart.Before(openTime) || endTime.After(closeTime) {
        return &AvailabilityResult{Available: false, Reason: "outside_hours"}, nil
    }

    // Step 6: Check for overlapping bookings (including buffers)
    bufferStart := startTime.Add(-time.Duration(service.BufferBefore) * time.Minute)
    bufferEnd := endTime.Add(time.Duration(service.BufferAfter) * time.Minute)

    hasConflict, err := s.repo.HasOverlappingBooking(ctx, tenantID, serviceID, bufferStart, bufferEnd)
    if hasConflict {
        return &AvailabilityResult{Available: false, Reason: "slot_taken"}, nil
    }

    return &AvailabilityResult{Available: true, EndTime: endTime}, nil
}
```

### RBAC Authorization Algorithm

```go
// Authorize - checks if user has required permission for a tenant action
func (s *AuthzService) Authorize(
    ctx context.Context,
    userID uuid.UUID,
    tenantID uuid.UUID,
    permission string,
) error {
    // Step 1: Check platform admin (bypasses tenant RBAC)
    platformAdmin, err := s.platformAdmins.GetByUserID(ctx, userID)
    if err == nil && platformAdmin != nil {
        return nil // Platform admins have full access
    }

    // Step 2: Check tenant membership
    membership, err := s.memberships.CheckMembership(ctx, userID, tenantID)
    if err != nil {
        return ErrNotAMember
    }

    // Step 3: Check role has permission
    if !roleHasPermission(membership.Role, permission) {
        return ErrInsufficientPermissions
    }

    return nil
}

// roleHasPermission - static permission matrix
var rolePermissions = map[string][]string{
    "tenant_owner":          {"*"}, // all permissions
    "tenant_manager":        {"bookings:*", "services:*", "staff:*", "availability:*", "branding:read"},
    "tenant_staff":          {"bookings:read", "bookings:create", "bookings:update", "services:read", "availability:read"},
    "tenant_finance_viewer": {"bookings:read", "billing:read", "reports:read"},
}
```


## Example Usage

### Tenant Resolution in booking-web (Next.js)

```typescript
// apps/booking-web/src/lib/tenant.ts
import { Tenant } from "@zenvikar/types";

export async function resolveTenant(host: string): Promise<Tenant> {
  const baseDomain = process.env.NEXT_PUBLIC_BASE_DOMAIN || "zenvikar.localhost";
  const slug = host.replace(`.${baseDomain}`, "").split(":")[0];

  const res = await fetch(
    `${process.env.API_INTERNAL_URL}/api/v1/tenants/resolve?slug=${slug}`,
    { next: { revalidate: 300 } }
  );

  if (!res.ok) throw new Error(`Tenant not found: ${slug}`);
  return res.json();
}

// apps/booking-web/src/middleware.ts
import { NextRequest, NextResponse } from "next/server";

export function middleware(request: NextRequest) {
  const host = request.headers.get("x-forwarded-host") || request.headers.get("host") || "";
  // Pass host to server components via header
  const response = NextResponse.next();
  response.headers.set("x-tenant-host", host);
  return response;
}
```

### Tenant Switching in tenant-web (Next.js)

```typescript
// apps/tenant-web/src/app/t/[tenantSlug]/layout.tsx
import { notFound } from "next/navigation";
import { Tenant } from "@zenvikar/types";

interface Props {
  params: { tenantSlug: string };
  children: React.ReactNode;
}

export default async function TenantLayout({ params, children }: Props) {
  const res = await fetch(
    `${process.env.API_INTERNAL_URL}/api/v1/tenants/resolve?slug=${params.tenantSlug}`,
    { next: { revalidate: 300 } }
  );

  if (!res.ok) return notFound();
  const tenant: Tenant = await res.json();

  return (
    <TenantProvider tenant={tenant}>
      <TenantSidebar currentTenant={tenant} />
      <main>{children}</main>
    </TenantProvider>
  );
}
```

### i18n Setup (Bilingual EN/PT)

```typescript
// packages/config/src/i18n.ts
export const locales = ["en", "pt"] as const;
export type Locale = (typeof locales)[number];
export const defaultLocale: Locale = "en";

// apps/booking-web/src/messages/en.json
{
  "booking": {
    "title": "Book an Appointment",
    "selectService": "Select a service",
    "selectDate": "Select a date",
    "confirm": "Confirm Booking"
  }
}

// apps/booking-web/src/messages/pt.json
{
  "booking": {
    "title": "Agendar um Compromisso",
    "selectService": "Selecione um serviço",
    "selectDate": "Selecione uma data",
    "confirm": "Confirmar Agendamento"
  }
}
```

### Go API Module Registration

```go
// cmd/api/main.go
func main() {
    cfg := config.Load()
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

    db, err := sql.Open("postgres", cfg.DatabaseURL)
    if err != nil {
        logger.Error("failed to connect to database", "error", err)
        os.Exit(1)
    }

    rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisURL})

    tp := otel.InitTracer(cfg.OTelEndpoint)
    defer tp.Shutdown(context.Background())

    deps := Dependencies{
        DB:     db,
        Redis:  rdb,
        Logger: logger,
        Tracer: tp.Tracer("zenvikar-api"),
        Config: cfg,
    }

    // Register modules
    modules := []Module{
        auth.New(),
        users.New(),
        tenants.New(),
        tenant_memberships.New(),
        branding.New(),
        bookings.New(),
        availability.New(),
        services.New(),
        staff.New(),
        billing.New(),
        reports.New(),
        notifications.New(),
        audit.New(),
    }

    router := chi.NewRouter()
    router.Use(middleware.Logger(logger))
    router.Use(middleware.Recoverer)
    router.Use(middleware.OTelTracing(tp))
    router.Use(middleware.CORS(cfg.AllowedOrigins))

    // Health endpoints
    router.Get("/healthz", handlers.Liveness)
    router.Get("/readyz", handlers.Readiness(db, rdb))
    router.Handle("/metrics", promhttp.Handler())

    // Register module routes
    for _, m := range modules {
        m.RegisterRoutes(router, deps)
        logger.Info("registered module", "name", m.Name())
    }

    // Run migrations
    for _, m := range modules {
        if err := m.Migrate(db); err != nil {
            logger.Error("migration failed", "module", m.Name(), "error", err)
            os.Exit(1)
        }
    }

    server := &http.Server{Addr: ":" + cfg.Port, Handler: router}
    logger.Info("starting server", "port", cfg.Port)
    if err := server.ListenAndServe(); err != nil {
        logger.Error("server error", "error", err)
    }
}
```

### Docker Compose (Key Services)

```yaml
# docker-compose.yml (excerpt)
services:
  traefik:
    image: traefik:v3.0
    command:
      - "--api.insecure=true"
      - "--providers.docker=true"
      - "--entrypoints.web.address=:80"
    ports:
      - "80:80"
      - "8080:8080"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock

  api:
    build: ./apps/api
    labels:
      - "traefik.http.routers.api.rule=Host(`api.zenvikar.localhost`)"
    environment:
      - DATABASE_URL=postgres://zenvikar:zenvikar@postgres:5432/zenvikar?sslmode=disable
      - REDIS_URL=redis:6379
      - OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317
    depends_on:
      - postgres
      - redis

  booking-web:
    build: ./apps/booking-web
    labels:
      - "traefik.http.routers.booking.rule=HostRegexp(`{subdomain:[a-z0-9-]+}.zenvikar.localhost`)"
      - "traefik.http.routers.booking.priority=1"
      - "traefik.http.middlewares.booking-headers.headers.customrequestheaders.X-Forwarded-Host={host}"
    environment:
      - NEXT_PUBLIC_BASE_DOMAIN=zenvikar.localhost
      - API_INTERNAL_URL=http://api:8080

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: zenvikar
      POSTGRES_USER: zenvikar
      POSTGRES_PASSWORD: zenvikar
    volumes:
      - pgdata:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system — essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Slug Validation Correctness

*For any* string input, the slug validator SHALL accept the string if and only if: its length is between 3 and 63 characters, it matches `^[a-z0-9][a-z0-9-]*[a-z0-9]$`, it contains no consecutive hyphens, and it is not in the reserved slugs list. All other inputs SHALL be rejected with a descriptive error.

**Validates: Requirements 2.1, 2.2, 2.3, 2.4**

### Property 2: Subdomain Extraction Round-Trip

*For any* valid tenant slug and any base domain, constructing a host as `{slug}.{baseDomain}` (with or without a port suffix) and then calling ExtractTenantSlugFromHost SHALL return the original slug.

**Validates: Requirements 3.1, 3.4**

### Property 3: Invalid Host Rejection

*For any* hostname that does not end with `.{baseDomain}` and is not equal to `baseDomain`, the Slug_Extractor SHALL return ErrInvalidHost.

**Validates: Requirement 3.3**

### Property 4: Tenant JSON Serialization Round-Trip

*For any* valid Tenant object, serializing it to JSON and deserializing it back SHALL produce an object equal to the original. This ensures cache consistency between Redis and database results.

**Validates: Requirement 4.6**

### Property 5: RBAC Permission Matrix Correctness

*For any* combination of (role, permission), the RBAC_Engine authorization result SHALL match the static permission matrix — including that platform admins are always authorized and tenant_owner has all permissions.

**Validates: Requirements 5.5, 5.6, 5.7**

### Property 6: Blocked Date Booking Rejection

*For any* tenant with blocked dates and any booking request targeting a blocked date, the Booking_Service SHALL reject the booking with reason "date_blocked".

**Validates: Requirement 7.1**

### Property 7: Outside Hours Booking Rejection

*For any* tenant with defined opening hours and any booking request with a start_time or computed end_time outside those hours, the Booking_Service SHALL reject the booking with reason "outside_hours".

**Validates: Requirement 7.2**

### Property 8: Booking Overlap Rejection with Buffers

*For any* existing non-cancelled booking and any new booking request whose effective time window (including service buffer_before and buffer_after) overlaps the existing booking's effective window, the Booking_Service SHALL reject the new booking with reason "slot_taken".

**Validates: Requirements 7.3, 7.5**

### Property 9: Booking End Time Calculation

*For any* valid booking request with a given start_time and a service with duration_minutes, the resulting booking's end_time SHALL equal start_time + duration_minutes.

**Validates: Requirement 7.4**


## Error Handling

### Error Scenario 1: Tenant Not Found

**Condition**: Subdomain slug does not match any enabled tenant in the database
**Response**: booking-web shows a generic "Tenant not found" page (not a 500 error). API returns 404 with `{"error": "tenant_not_found", "message": "No tenant found for this address"}`
**Recovery**: No recovery needed. User should check the URL.

### Error Scenario 2: Tenant Disabled

**Condition**: Tenant exists but `enabled = false`
**Response**: booking-web shows "This booking page is currently unavailable". API returns 403 with `{"error": "tenant_disabled"}`
**Recovery**: Tenant owner or platform admin re-enables the tenant.

### Error Scenario 3: Booking Slot Conflict

**Condition**: Two concurrent requests attempt to book the same time slot
**Response**: Database unique constraint prevents double-booking. Second request receives 409 Conflict with `{"error": "slot_unavailable"}`
**Recovery**: Client retries with a different time slot. Frontend refreshes available slots.

### Error Scenario 4: Reserved Slug Registration

**Condition**: Tenant creation attempted with a reserved slug (e.g., "api", "admin", "www")
**Response**: API returns 422 with `{"error": "slug_reserved", "message": "This name is not available"}`
**Recovery**: User chooses a different slug.

### Error Scenario 5: Unauthorized Tenant Access

**Condition**: User attempts to access a tenant they don't have membership in
**Response**: API returns 403 Forbidden. tenant-web redirects to tenant selection page.
**Recovery**: User selects a tenant they have access to, or requests access from tenant owner.

### Error Scenario 6: Database Connection Failure

**Condition**: PostgreSQL or Redis is unreachable
**Response**: `/readyz` endpoint returns 503. Traefik stops routing to unhealthy instances. API returns 503 for all data-dependent requests.
**Recovery**: Automatic reconnection with exponential backoff. Kubernetes/Docker restarts unhealthy containers.

## Testing Strategy

### Unit Testing Approach

- Go backend: standard `testing` package with table-driven tests
- Key test areas: slug validation, subdomain extraction, RBAC permission checks, booking time calculations
- Frontend: Jest + React Testing Library for component tests
- Coverage target: 80% for core business logic modules (tenants, bookings, auth, RBAC)

### Property-Based Testing Approach

**Property Test Library**: Go: `pgregory.net/rapid`, TypeScript: `fast-check`

Key properties to test:
- Slug validation accepts all valid slugs and rejects all invalid ones
- Subdomain extraction is the inverse of subdomain construction
- RBAC permission checks are deterministic for the same inputs
- Booking time calculations respect timezone boundaries
- Tenant resolution returns consistent results regardless of cache state

### Integration Testing Approach

- Docker Compose test environment with real PostgreSQL and Redis
- API integration tests: HTTP requests against running server
- Database migration tests: verify all migrations apply cleanly and are reversible
- Tenant resolution end-to-end: subdomain → API → database → response
- Multi-tenant isolation: verify data from tenant A is never visible to tenant B

## Performance Considerations

- **Tenant Resolution Caching**: Redis cache with 5-minute TTL for tenant lookups by slug. This is the hottest path (every booking-web request).
- **Database Indexing**: Partial index on `tenants(slug) WHERE enabled = true` for fast lookups. Composite indexes on bookings for time-range queries.
- **Connection Pooling**: Go's `sql.DB` with configured `MaxOpenConns`, `MaxIdleConns`, `ConnMaxLifetime`.
- **Static Asset Caching**: Next.js static assets served with long cache headers. Tenant branding loaded once and cached client-side.
- **Deferred Optimizations**: Read replicas, query result caching, CDN for static assets, connection pooling via PgBouncer. These are not needed for MVP.

## Security Considerations

- **Tenant Isolation**: All database queries include `tenant_id` filter. No cross-tenant data leakage.
- **Authentication**: JWT-based with short-lived access tokens and refresh tokens. Passwords hashed with bcrypt.
- **RBAC**: Role-based access control enforced at middleware level before handlers execute.
- **Input Validation**: All user inputs validated at API boundary. SQL injection prevented by parameterized queries.
- **CORS**: Strict origin allowlist per environment. Wildcard subdomains handled carefully.
- **Secrets Management**: Environment variables for local dev. Terraform-managed secrets for production (AWS Secrets Manager / Vault placeholder).
- **Rate Limiting**: Placeholder middleware for rate limiting per tenant and per IP. Not implemented in MVP but structure is in place.
- **Deferred**: CSRF protection, Content Security Policy headers, audit logging implementation, IP allowlisting for admin portal.

## Dependencies

### Backend (Go)
- `net/http` + `go-chi/chi` - HTTP router
- `lib/pq` or `jackc/pgx` - PostgreSQL driver
- `go-redis/redis` - Redis client
- `golang-migrate/migrate` - Database migrations
- `google/uuid` - UUID generation
- `go.opentelemetry.io/otel` - OpenTelemetry SDK
- `prometheus/client_golang` - Prometheus metrics

### Frontend (TypeScript/Next.js)
- `next` 14+ - React framework with App Router
- `next-intl` - i18n for bilingual support
- `@sentry/nextjs` - Error tracking (placeholder)
- `tailwindcss` - Styling

### Infrastructure
- Docker + Docker Compose - Containerization
- Traefik v3 - Reverse proxy
- PostgreSQL 16 - Primary database
- Redis 7 - Caching
- Prometheus + Grafana + Loki + Tempo - Observability
- OpenTelemetry Collector - Telemetry pipeline
- Terraform - Infrastructure as Code
- GitHub Actions - CI/CD

## Deferred Items & Future Evolution

### Deferred (Not in MVP)
- Full booking CRUD implementation (only structure/schema in MVP)
- Payment processing / billing integration
- Email/SMS notification dispatch
- Audit log querying UI
- Report generation
- Staff scheduling and resource-based booking
- Customer self-service portal features
- Advanced branding (custom CSS, custom domains per tenant)
- Rate limiting implementation
- CSRF protection
- Read replicas and PgBouncer

### Recommended Evolution Path
1. **Phase 1 (MVP)**: This foundation → basic auth → tenant CRUD → service CRUD → simple booking flow
2. **Phase 2**: Notifications, billing integration, staff management, availability rules
3. **Phase 3**: Advanced booking (resource-based, recurring), reporting, audit UI
4. **Phase 4**: Custom domains per tenant, advanced branding, CDN, read replicas
5. **Phase 5**: Mobile apps, webhooks, public API for integrations
