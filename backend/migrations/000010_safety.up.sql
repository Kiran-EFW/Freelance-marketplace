-- Safety Features: SOS Alerts, Live Location Tracking, Emergency Contacts

-- ============================================================
-- ENUMS
-- ============================================================

CREATE TYPE sos_status AS ENUM ('active', 'responded', 'resolved', 'false_alarm');

-- ============================================================
-- SOS ALERTS
-- ============================================================

CREATE TABLE sos_alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    job_id UUID REFERENCES jobs(id),
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    status sos_status NOT NULL DEFAULT 'active',
    emergency_contacts_notified BOOLEAN NOT NULL DEFAULT false,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

CREATE INDEX idx_sos_alerts_user ON sos_alerts(user_id);
CREATE INDEX idx_sos_alerts_status ON sos_alerts(status);
CREATE INDEX idx_sos_alerts_job ON sos_alerts(job_id);
CREATE INDEX idx_sos_alerts_created ON sos_alerts(created_at DESC);

-- ============================================================
-- LOCATION SHARES (live tracking during jobs)
-- ============================================================

CREATE TABLE location_shares (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_id UUID NOT NULL REFERENCES jobs(id),
    user_id UUID NOT NULL REFERENCES users(id),
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    accuracy FLOAT,
    shared_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_location_shares_job ON location_shares(job_id);
CREATE INDEX idx_location_shares_user ON location_shares(user_id);
CREATE INDEX idx_location_shares_shared ON location_shares(shared_at DESC);

-- ============================================================
-- EMERGENCY CONTACTS
-- ============================================================

CREATE TABLE emergency_contacts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    relationship VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_emergency_contacts_user ON emergency_contacts(user_id);
