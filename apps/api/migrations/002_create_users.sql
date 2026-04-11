CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    password_hash TEXT NOT NULL,
    phone VARCHAR(20),
    preferred_contact VARCHAR(10) NOT NULL DEFAULT 'email' CHECK (preferred_contact IN ('email', 'phone', 'whatsapp')),
    locale VARCHAR(2) NOT NULL DEFAULT 'en' CHECK (locale IN ('en', 'pt')),
    email_verified BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
