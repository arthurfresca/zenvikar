CREATE TABLE IF NOT EXISTS bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    service_member_id UUID NOT NULL REFERENCES service_members(id),
    customer_id UUID NOT NULL REFERENCES users(id),
    price_cents INT NOT NULL DEFAULT 0 CHECK (price_cents >= 0),
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'cancelled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (end_time > start_time)
);

-- Prevent double-booking: same service_member at the same start time
CREATE UNIQUE INDEX IF NOT EXISTS idx_bookings_no_overlap
    ON bookings(service_member_id, start_time)
    WHERE status != 'cancelled';

-- Fast lookups by tenant + time range
CREATE INDEX IF NOT EXISTS idx_bookings_tenant_time
    ON bookings(tenant_id, start_time, end_time);

-- Fast lookups by member (for checking all bookings across services)
CREATE INDEX IF NOT EXISTS idx_bookings_member_time
    ON bookings(service_member_id, start_time, end_time);
