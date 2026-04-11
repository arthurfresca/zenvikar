-- Opening hours per service_member: "Bob does Haircut on Monday 09:00-18:00"
CREATE TABLE IF NOT EXISTS opening_hours (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_member_id UUID NOT NULL REFERENCES service_members(id) ON DELETE CASCADE,
    day_of_week INT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    open_time TIME NOT NULL,
    close_time TIME NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    UNIQUE(service_member_id, day_of_week)
);

-- Blocked dates stay at the membership level (a day off blocks all services)
CREATE TABLE IF NOT EXISTS blocked_dates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    membership_id UUID NOT NULL REFERENCES tenant_memberships(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    reason TEXT,
    UNIQUE(membership_id, date)
);
