-- =============================================================================
-- Zenvikar Platform — Local Development Seed Data
-- Usage: make seed
-- =============================================================================

BEGIN;

-- ---------------------------------------------------------------------------
-- Tenants
-- ---------------------------------------------------------------------------
INSERT INTO tenants (id, slug, display_name, logo_url, color_primary, color_secondary, color_accent, phone, email, address, currency, slot_interval_minutes, timezone, default_locale, enabled)
VALUES
  ('a1b2c3d4-0001-4000-8000-000000000001', 'acme', 'Acme Salon', NULL, '#E63946', '#457B9D', '#1D3557', '+15551112222', 'contact@acme.test', '123 Main St, New York, NY 10001', 'USD', 15, 'America/New_York', 'en', true),
  ('a1b2c3d4-0002-4000-8000-000000000002', 'beta-salon', 'Beta Salon', NULL, '#2A9D8F', '#264653', '#E9C46A', '+5511988776655', NULL, 'Rua Augusta 100, São Paulo, SP', 'BRL', 30, 'America/Sao_Paulo', 'pt', true)
ON CONFLICT (slug) DO NOTHING;

-- ---------------------------------------------------------------------------
-- Users
-- ---------------------------------------------------------------------------
INSERT INTO users (id, email, name, password_hash, phone, preferred_contact, locale, email_verified)
VALUES
  ('b1b2c3d4-0001-4000-8000-000000000001', 'owner@acme.test', 'Alice Owner', '$2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ012', '+15551234567', 'email', 'en', true),
  ('b1b2c3d4-0002-4000-8000-000000000002', 'manager@acme.test', 'Bob Manager', '$2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ012', '+15559876543', 'phone', 'en', true),
  ('b1b2c3d4-0003-4000-8000-000000000003', 'staff@acme.test', 'Carol Staff', '$2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ012', NULL, 'email', 'en', true),
  ('b1b2c3d4-0004-4000-8000-000000000004', 'owner@beta.test', 'Diana Owner', '$2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ012', '+5511999887766', 'whatsapp', 'pt', true),
  ('b1b2c3d4-0005-4000-8000-000000000005', 'customer@example.test', 'Eve Customer', '$2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ012', '+15555550000', 'whatsapp', 'en', true),
  ('b1b2c3d4-0006-4000-8000-000000000006', 'admin@zenvikar.test', 'Frank Admin', '$2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ012', NULL, 'email', 'en', true)
ON CONFLICT (email) DO NOTHING;

-- ---------------------------------------------------------------------------
-- User Auth Providers
-- ---------------------------------------------------------------------------
INSERT INTO user_auth_providers (id, user_id, provider, provider_id)
VALUES
  -- Most users use email/password
  ('ab12c3d4-0001-4000-8000-000000000001', 'b1b2c3d4-0001-4000-8000-000000000001', 'email', 'b1b2c3d4-0001-4000-8000-000000000001'),
  ('ab12c3d4-0002-4000-8000-000000000002', 'b1b2c3d4-0002-4000-8000-000000000002', 'email', 'b1b2c3d4-0002-4000-8000-000000000002'),
  ('ab12c3d4-0003-4000-8000-000000000003', 'b1b2c3d4-0003-4000-8000-000000000003', 'email', 'b1b2c3d4-0003-4000-8000-000000000003'),
  ('ab12c3d4-0004-4000-8000-000000000004', 'b1b2c3d4-0004-4000-8000-000000000004', 'email', 'b1b2c3d4-0004-4000-8000-000000000004'),
  -- Eve uses Google
  ('ab12c3d4-0005-4000-8000-000000000005', 'b1b2c3d4-0005-4000-8000-000000000005', 'google', '110248495012345678901'),
  -- Frank uses email
  ('ab12c3d4-0006-4000-8000-000000000006', 'b1b2c3d4-0006-4000-8000-000000000006', 'email', 'b1b2c3d4-0006-4000-8000-000000000006')
ON CONFLICT (provider, provider_id) DO NOTHING;

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
INSERT INTO tenant_memberships (id, tenant_id, user_id, role, photo_url, description)
VALUES
  -- Acme Salon memberships
  ('d1b2c3d4-0001-4000-8000-000000000001', 'a1b2c3d4-0001-4000-8000-000000000001', 'b1b2c3d4-0001-4000-8000-000000000001', 'tenant_owner', NULL, 'Salon founder and lead stylist'),
  ('d1b2c3d4-0002-4000-8000-000000000002', 'a1b2c3d4-0001-4000-8000-000000000001', 'b1b2c3d4-0002-4000-8000-000000000002', 'tenant_manager', NULL, 'Front desk and scheduling manager'),
  ('d1b2c3d4-0003-4000-8000-000000000003', 'a1b2c3d4-0001-4000-8000-000000000001', 'b1b2c3d4-0003-4000-8000-000000000003', 'tenant_staff', NULL, NULL),
  -- Beta Salon memberships
  ('d1b2c3d4-0004-4000-8000-000000000004', 'a1b2c3d4-0002-4000-8000-000000000002', 'b1b2c3d4-0004-4000-8000-000000000004', 'tenant_owner', NULL, 'Proprietária do salão')
ON CONFLICT (tenant_id, user_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- Services
-- ---------------------------------------------------------------------------
INSERT INTO services (id, tenant_id, name, description, duration_minutes, buffer_before_minutes, buffer_after_minutes, enabled)
VALUES
  -- Acme Salon services
  ('e1b2c3d4-0001-4000-8000-000000000001', 'a1b2c3d4-0001-4000-8000-000000000001', 'Haircut', 'Includes wash, cut, and style', 30, 5, 5, true),
  ('e1b2c3d4-0002-4000-8000-000000000002', 'a1b2c3d4-0001-4000-8000-000000000001', 'Hair Coloring', 'Full color treatment with premium products', 90, 10, 10, true),
  ('e1b2c3d4-0003-4000-8000-000000000003', 'a1b2c3d4-0001-4000-8000-000000000001', 'Beard Trim', 'Precision beard shaping and trim', 15, 0, 5, true),
  -- Beta Salon services
  ('e1b2c3d4-0004-4000-8000-000000000004', 'a1b2c3d4-0002-4000-8000-000000000002', 'Corte de Cabelo', 'Corte completo com lavagem e finalização', 30, 5, 5, true),
  ('e1b2c3d4-0005-4000-8000-000000000005', 'a1b2c3d4-0002-4000-8000-000000000002', 'Manicure', 'Manicure completa com esmaltação', 45, 5, 5, true)
ON CONFLICT DO NOTHING;

-- ---------------------------------------------------------------------------
-- Service Members (which staff can perform which services)
-- ---------------------------------------------------------------------------
INSERT INTO service_members (id, service_id, membership_id, price_cents, description)
VALUES
  -- Acme: Alice charges $40 for Haircut, Bob charges $35
  ('aa12c3d4-0001-4000-8000-000000000001', 'e1b2c3d4-0001-4000-8000-000000000001', 'd1b2c3d4-0001-4000-8000-000000000001', 4000, '15 years experience with all hair types'),
  ('aa12c3d4-0002-4000-8000-000000000002', 'e1b2c3d4-0001-4000-8000-000000000001', 'd1b2c3d4-0002-4000-8000-000000000002', 3500, NULL),
  -- Acme: only Alice does Hair Coloring at $120
  ('aa12c3d4-0003-4000-8000-000000000003', 'e1b2c3d4-0002-4000-8000-000000000002', 'd1b2c3d4-0001-4000-8000-000000000001', 12000, 'Specializes in balayage and color corrections'),
  -- Acme: Carol does Beard Trims at $15
  ('aa12c3d4-0004-4000-8000-000000000004', 'e1b2c3d4-0003-4000-8000-000000000003', 'd1b2c3d4-0003-4000-8000-000000000003', 1500, NULL),
  -- Beta: Diana does Corte de Cabelo at R$40 and Manicure at R$30
  ('aa12c3d4-0005-4000-8000-000000000005', 'e1b2c3d4-0004-4000-8000-000000000004', 'd1b2c3d4-0004-4000-8000-000000000004', 4000, 'Cortes modernos e clássicos'),
  ('aa12c3d4-0006-4000-8000-000000000006', 'e1b2c3d4-0005-4000-8000-000000000005', 'd1b2c3d4-0004-4000-8000-000000000004', 3000, NULL)
ON CONFLICT DO NOTHING;

-- ---------------------------------------------------------------------------
-- Opening Hours (Mon–Sat for both tenants)
-- ---------------------------------------------------------------------------
INSERT INTO opening_hours (id, service_member_id, day_of_week, open_time, close_time, enabled)
VALUES
  -- Alice does Haircut: Mon-Sat 09:00-18:00
  ('f1b2c3d4-0001-4000-8000-000000000001', 'aa12c3d4-0001-4000-8000-000000000001', 1, '09:00', '18:00', true),
  ('f1b2c3d4-0002-4000-8000-000000000002', 'aa12c3d4-0001-4000-8000-000000000001', 2, '09:00', '18:00', true),
  ('f1b2c3d4-0003-4000-8000-000000000003', 'aa12c3d4-0001-4000-8000-000000000001', 3, '09:00', '18:00', true),
  ('f1b2c3d4-0004-4000-8000-000000000004', 'aa12c3d4-0001-4000-8000-000000000001', 4, '09:00', '18:00', true),
  ('f1b2c3d4-0005-4000-8000-000000000005', 'aa12c3d4-0001-4000-8000-000000000001', 5, '09:00', '18:00', true),
  ('f1b2c3d4-0006-4000-8000-000000000006', 'aa12c3d4-0001-4000-8000-000000000001', 6, '10:00', '16:00', true),
  -- Alice does Hair Coloring: Mon-Fri 09:00-17:00 (shorter day for coloring)
  ('f1b2c3d4-0030-4000-8000-000000000030', 'aa12c3d4-0003-4000-8000-000000000003', 1, '09:00', '17:00', true),
  ('f1b2c3d4-0031-4000-8000-000000000031', 'aa12c3d4-0003-4000-8000-000000000003', 2, '09:00', '17:00', true),
  ('f1b2c3d4-0032-4000-8000-000000000032', 'aa12c3d4-0003-4000-8000-000000000003', 3, '09:00', '17:00', true),
  ('f1b2c3d4-0033-4000-8000-000000000033', 'aa12c3d4-0003-4000-8000-000000000003', 4, '09:00', '17:00', true),
  ('f1b2c3d4-0034-4000-8000-000000000034', 'aa12c3d4-0003-4000-8000-000000000003', 5, '09:00', '17:00', true),
  -- Bob does Haircut: Mon-Fri 10:00-19:00
  ('f1b2c3d4-0010-4000-8000-000000000010', 'aa12c3d4-0002-4000-8000-000000000002', 1, '10:00', '19:00', true),
  ('f1b2c3d4-0011-4000-8000-000000000011', 'aa12c3d4-0002-4000-8000-000000000002', 2, '10:00', '19:00', true),
  ('f1b2c3d4-0012-4000-8000-000000000012', 'aa12c3d4-0002-4000-8000-000000000002', 3, '10:00', '19:00', true),
  ('f1b2c3d4-0013-4000-8000-000000000013', 'aa12c3d4-0002-4000-8000-000000000002', 4, '10:00', '19:00', true),
  ('f1b2c3d4-0014-4000-8000-000000000014', 'aa12c3d4-0002-4000-8000-000000000002', 5, '10:00', '19:00', true),
  -- Carol does Beard Trim: Tue-Sat 08:00-16:00
  ('f1b2c3d4-0020-4000-8000-000000000020', 'aa12c3d4-0004-4000-8000-000000000004', 2, '08:00', '16:00', true),
  ('f1b2c3d4-0021-4000-8000-000000000021', 'aa12c3d4-0004-4000-8000-000000000004', 3, '08:00', '16:00', true),
  ('f1b2c3d4-0022-4000-8000-000000000022', 'aa12c3d4-0004-4000-8000-000000000004', 4, '08:00', '16:00', true),
  ('f1b2c3d4-0023-4000-8000-000000000023', 'aa12c3d4-0004-4000-8000-000000000004', 5, '08:00', '16:00', true),
  ('f1b2c3d4-0024-4000-8000-000000000024', 'aa12c3d4-0004-4000-8000-000000000004', 6, '08:00', '16:00', true),
  -- Diana does Corte de Cabelo: Mon-Fri 08:00-20:00
  ('f1b2c3d4-0007-4000-8000-000000000007', 'aa12c3d4-0005-4000-8000-000000000005', 1, '08:00', '20:00', true),
  ('f1b2c3d4-0008-4000-8000-000000000008', 'aa12c3d4-0005-4000-8000-000000000005', 2, '08:00', '20:00', true),
  ('f1b2c3d4-0009-4000-8000-000000000009', 'aa12c3d4-0005-4000-8000-000000000005', 3, '08:00', '20:00', true),
  ('f1b2c3d4-000a-4000-8000-00000000000a', 'aa12c3d4-0005-4000-8000-000000000005', 4, '08:00', '20:00', true),
  ('f1b2c3d4-000b-4000-8000-00000000000b', 'aa12c3d4-0005-4000-8000-000000000005', 5, '08:00', '20:00', true),
  -- Diana does Manicure: Mon-Fri 08:00-20:00
  ('f1b2c3d4-0040-4000-8000-000000000040', 'aa12c3d4-0006-4000-8000-000000000006', 1, '08:00', '20:00', true),
  ('f1b2c3d4-0041-4000-8000-000000000041', 'aa12c3d4-0006-4000-8000-000000000006', 2, '08:00', '20:00', true),
  ('f1b2c3d4-0042-4000-8000-000000000042', 'aa12c3d4-0006-4000-8000-000000000006', 3, '08:00', '20:00', true),
  ('f1b2c3d4-0043-4000-8000-000000000043', 'aa12c3d4-0006-4000-8000-000000000006', 4, '08:00', '20:00', true),
  ('f1b2c3d4-0044-4000-8000-000000000044', 'aa12c3d4-0006-4000-8000-000000000006', 5, '08:00', '20:00', true)
ON CONFLICT (service_member_id, day_of_week) DO NOTHING;

-- ---------------------------------------------------------------------------
-- Blocked Dates (per member)
-- ---------------------------------------------------------------------------
INSERT INTO blocked_dates (id, membership_id, date, reason)
VALUES
  ('01b2c3d4-0001-4000-8000-000000000001', 'd1b2c3d4-0001-4000-8000-000000000001', '2025-12-25', 'Christmas Day'),
  ('01b2c3d4-0002-4000-8000-000000000002', 'd1b2c3d4-0001-4000-8000-000000000001', '2025-01-01', 'New Year''s Day'),
  ('01b2c3d4-0003-4000-8000-000000000003', 'd1b2c3d4-0004-4000-8000-000000000004', '2025-12-25', 'Natal')
ON CONFLICT (membership_id, date) DO NOTHING;

-- ---------------------------------------------------------------------------
-- Bookings
-- ---------------------------------------------------------------------------
INSERT INTO bookings (id, tenant_id, service_member_id, customer_id, price_cents, start_time, end_time, status)
VALUES
  -- Acme: confirmed haircut with Alice at $40
  ('a0b2c3d4-0001-4000-8000-000000000001', 'a1b2c3d4-0001-4000-8000-000000000001', 'aa12c3d4-0001-4000-8000-000000000001', 'b1b2c3d4-0005-4000-8000-000000000005', 4000,
   '2025-07-14 10:00:00+00', '2025-07-14 10:30:00+00', 'confirmed'),
  -- Acme: pending hair coloring with Alice at $120
  ('a0b2c3d4-0002-4000-8000-000000000002', 'a1b2c3d4-0001-4000-8000-000000000001', 'aa12c3d4-0003-4000-8000-000000000003', 'b1b2c3d4-0005-4000-8000-000000000005', 12000,
   '2025-07-14 14:00:00+00', '2025-07-14 15:30:00+00', 'pending'),
  -- Beta: pending corte de cabelo with Diana at R$40
  ('a0b2c3d4-0003-4000-8000-000000000003', 'a1b2c3d4-0002-4000-8000-000000000002', 'aa12c3d4-0005-4000-8000-000000000005', 'b1b2c3d4-0005-4000-8000-000000000005', 4000,
   '2025-07-15 09:00:00+00', '2025-07-15 09:30:00+00', 'pending')
ON CONFLICT DO NOTHING;

COMMIT;
