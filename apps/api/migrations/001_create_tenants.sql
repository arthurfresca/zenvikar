CREATE TABLE IF NOT EXISTS tenants (
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

CREATE INDEX IF NOT EXISTS idx_tenants_slug ON tenants(slug) WHERE enabled = true;
