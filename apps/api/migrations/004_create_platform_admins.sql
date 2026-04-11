CREATE TABLE IF NOT EXISTS platform_admins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(32) NOT NULL CHECK (role IN ('admin', 'support_admin', 'finance_admin')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
