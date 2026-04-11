-- =============================================================================
-- Zenvikar Platform — Local Development Seed Data
-- Usage: make seed
-- =============================================================================

BEGIN;

-- ---------------------------------------------------------------------------
-- Tenants
-- ---------------------------------------------------------------------------
INSERT INTO tenants (id, slug, display_name, logo_url, color_primary, color_secondary, color_accent, timezone, default_locale, enabled)
VALUES
  ('a1b2c3d4-0001-4000-8000-000000000001', 'acme', 'Acme Salon', NULL, '#E63946', '#457B9D', '#1D3557', 'America/New_York', 'en', true),
  ('a1b2c3d4-0002-4000-8000-000000000002', 'beta-salon', 'Beta Salon', NULL, '#2A9D8F', '#264653', '#E9C46A', 'America/Sao_Paulo', 'pt', true)
ON CONFLICT (slug) DO NOTHING;

-- ---------------------------------------------------------------------------
-- Users
-- ---------------------------------------------------------------------------
INSERT INTO users (id, email, name, password_hash, locale, email_verified)
VALUES
  ('b1b2c3d4-0001-4000-8000-000000000001', 'owner@acme.test', 'Alice Owner', '$2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ012', 'en', true),
  ('b1b2c3d4-0002-4000-8000-000000000002', 'manager@acme.test', 'Bob Manager', '$2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ012', 'en', true),
  ('b1b2c3d4-0003-4000-8000-000000000003', 'staff@acme.test', 'Carol Staff', '$2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ012', 'en', true),
  ('b1b2c3d4-0004-4000-8000-000000000004', 'owner@beta.test', 'Diana Owner', '$2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ012', 'pt', true),
  ('b1b2c3d4-0005-4000-8000-000000000005', 'customer@example.test', 'Eve Customer', '$2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ012', 'en', true),
  ('b1b2c3d4-0006-4000-8000-000000000006', 'admin@zenvikar.test', 'Frank Admin', '$2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ012', 'en', true)
ON CONFLICT (email) DO NOTHING;

-- ---------------------------------------------------------------------------
-- Platform Admins
-- ---------------------------------------------------------------------------
INSERT INTO platform_admins (id, user_id, role)
VALUES
  ('c1b2c3d4-0001-4000-8000-000000000001', 'b1b2c3d4-0006-4000-8000-000000000006', 'admin')
ON CONFLICT (user_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- Tenant Memberships
-- ---------------------------------------------------------------------------
INSERT INTO tenant_memberships (id, tenant_id, user_id, role)
VALUES
  -- Acme Salon memberships
  ('d1b2c3d4-0001-4000-8000-000000000001', 'a1b2c3d4-0001-4000-8000-000000000001', 'b1b2c3d4-0001-4000-8000-000000000001', 'tenant_owner'),
  ('d1b2c3d4-0002-4000-8000-000000000002', 'a1b2c3d4-0001-4000-8000-000000000001', 'b1b2c3d4-0002-4000-8000-000000000002', 'tenant_manager'),
  ('d1b2c3d4-0003-4000-8000-000000000003', 'a1b2c3d4-0001-4000-8000-000000000001', 'b1b2c3d4-0003-4000-8000-000000000003', 'tenant_staff'),
  -- Beta Salon memberships
  ('d1b2c3d4-0004-4000-8000-000000000004', 'a1b2c3d4-0002-4000-8000-000000000002', 'b1b2c3d4-0004-4000-8000-000000000004', 'tenant_owner')
ON CONFLICT (tenant_id, user_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- Services
-- ---------------------------------------------------------------------------
INSERT INTO services (id, tenant_id, name, duration_minutes, buffer_before_minutes, buffer_after_minutes, enabled)
VALUES
  -- Acme Salon services
  ('e1b2c3d4-0001-4000-8000-000000000001', 'a1b2c3d4-0001-4000-8000-000000000001', 'Haircut', 30, 5, 5, true),
  ('e1b2c3d4-0002-4000-8000-000000000002', 'a1b2c3d4-0001-4000-8000-000000000001', 'Hair Coloring', 90, 10, 10, true),
  ('e1b2c3d4-0003-4000-8000-000000000003', 'a1b2c3d4-0001-4000-8000-000000000001', 'Beard Trim', 15, 0, 5, true),
  -- Beta Salon services
  ('e1b2c3d4-0004-4000-8000-000000000004', 'a1b2c3d4-0002-4000-8000-000000000002', 'Corte de Cabelo', 30, 5, 5, true),
  ('e1b2c3d4-0005-4000-8000-000000000005', 'a1b2c3d4-0002-4000-8000-000000000002', 'Manicure', 45, 5, 5, true)
ON CONFLICT DO NOTHING;

-- ---------------------------------------------------------------------------
-- Opening Hours (Mon–Sat for both tenants)
-- ---------------------------------------------------------------------------
INSERT INTO opening_hours (id, tenant_id, day_of_week, open_time, close_time, enabled)
VALUES
  -- Acme Salon: Mon(1)–Sat(6), 09:00–18:00
  ('f1b2c3d4-0001-4000-8000-000000000001', 'a1b2c3d4-0001-4000-8000-000000000001', 1, '09:00', '18:00', true),
  ('f1b2c3d4-0002-4000-8000-000000000002', 'a1b2c3d4-0001-4000-8000-000000000001', 2, '09:00', '18:00', true),
  ('f1b2c3d4-0003-4000-8000-000000000003', 'a1b2c3d4-0001-4000-8000-000000000001', 3, '09:00', '18:00', true),
  ('f1b2c3d4-0004-4000-8000-000000000004', 'a1b2c3d4-0001-4000-8000-000000000001', 4, '09:00', '18:00', true),
  ('f1b2c3d4-0005-4000-8000-000000000005', 'a1b2c3d4-0001-4000-8000-000000000001', 5, '09:00', '18:00', true),
  ('f1b2c3d4-0006-4000-8000-000000000006', 'a1b2c3d4-0001-4000-8000-000000000001', 6, '10:00', '16:00', true),
  -- Beta Salon: Mon(1)–Fri(5), 08:00–20:00
  ('f1b2c3d4-0007-4000-8000-000000000007', 'a1b2c3d4-0002-4000-8000-000000000002', 1, '08:00', '20:00', true),
  ('f1b2c3d4-0008-4000-8000-000000000008', 'a1b2c3d4-0002-4000-8000-000000000002', 2, '08:00', '20:00', true),
  ('f1b2c3d4-0009-4000-8000-000000000009', 'a1b2c3d4-0002-4000-8000-000000000002', 3, '08:00', '20:00', true),
  ('f1b2c3d4-000a-4000-8000-00000000000a', 'a1b2c3d4-0002-4000-8000-000000000002', 4, '08:00', '20:00', true),
  ('f1b2c3d4-000b-4000-8000-00000000000b', 'a1b2c3d4-0002-4000-8000-000000000002', 5, '08:00', '20:00', true)
ON CONFLICT (tenant_id, day_of_week) DO NOTHING;

-- ---------------------------------------------------------------------------
-- Blocked Dates
-- ---------------------------------------------------------------------------
INSERT INTO blocked_dates (id, tenant_id, date, reason)
VALUES
  ('01b2c3d4-0001-4000-8000-000000000001', 'a1b2c3d4-0001-4000-8000-000000000001', '2025-12-25', 'Christmas Day'),
  ('01b2c3d4-0002-4000-8000-000000000002', 'a1b2c3d4-0001-4000-8000-000000000001', '2025-01-01', 'New Year''s Day'),
  ('01b2c3d4-0003-4000-8000-000000000003', 'a1b2c3d4-0002-4000-8000-000000000002', '2025-12-25', 'Natal')
ON CONFLICT (tenant_id, date) DO NOTHING;

-- ---------------------------------------------------------------------------
-- Bookings
-- ---------------------------------------------------------------------------
INSERT INTO bookings (id, tenant_id, service_id, customer_id, start_time, end_time, status, timezone)
VALUES
  -- Acme: confirmed haircut
  ('a0b2c3d4-0001-4000-8000-000000000001', 'a1b2c3d4-0001-4000-8000-000000000001', 'e1b2c3d4-0001-4000-8000-000000000001', 'b1b2c3d4-0005-4000-8000-000000000005',
   '2025-07-14 10:00:00+00', '2025-07-14 10:30:00+00', 'confirmed', 'America/New_York'),
  -- Acme: pending hair coloring
  ('a0b2c3d4-0002-4000-8000-000000000002', 'a1b2c3d4-0001-4000-8000-000000000001', 'e1b2c3d4-0002-4000-8000-000000000002', 'b1b2c3d4-0005-4000-8000-000000000005',
   '2025-07-14 14:00:00+00', '2025-07-14 15:30:00+00', 'pending', 'America/New_York'),
  -- Beta: pending corte de cabelo
  ('a0b2c3d4-0003-4000-8000-000000000003', 'a1b2c3d4-0002-4000-8000-000000000002', 'e1b2c3d4-0004-4000-8000-000000000004', 'b1b2c3d4-0005-4000-8000-000000000005',
   '2025-07-15 09:00:00+00', '2025-07-15 09:30:00+00', 'pending', 'America/Sao_Paulo')
ON CONFLICT DO NOTHING;

COMMIT;
