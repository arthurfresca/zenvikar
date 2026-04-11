CREATE TABLE IF NOT EXISTS services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    duration_minutes INT NOT NULL CHECK (duration_minutes > 0),
    buffer_before_minutes INT NOT NULL DEFAULT 0 CHECK (buffer_before_minutes >= 0),
    buffer_after_minutes INT NOT NULL DEFAULT 0 CHECK (buffer_after_minutes >= 0),
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Links services to tenant members who can perform them.
-- Each member can set their own price for the service.
-- A service must have at least one member to be bookable.
CREATE TABLE IF NOT EXISTS service_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    membership_id UUID NOT NULL REFERENCES tenant_memberships(id) ON DELETE CASCADE,
    price_cents INT NOT NULL DEFAULT 0 CHECK (price_cents >= 0),
    description TEXT,
    UNIQUE(service_id, membership_id)
);

CREATE INDEX IF NOT EXISTS idx_service_members_service ON service_members(service_id);
CREATE INDEX IF NOT EXISTS idx_service_members_membership ON service_members(membership_id);
