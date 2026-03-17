-- Initial schema for Seva — Service Marketplace Platform
-- Requires PostgreSQL 15+ with PostGIS extension

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";

-- ============================================================
-- ENUMS
-- ============================================================

CREATE TYPE user_type AS ENUM ('customer', 'provider', 'company');
CREATE TYPE device_type AS ENUM ('smartphone', 'basic_phone', 'web');
CREATE TYPE verification_status AS ENUM ('unverified', 'pending', 'verified', 'rejected');
CREATE TYPE provider_level AS ENUM ('new', 'active', 'trusted', 'expert', 'local_champion');
CREATE TYPE subscription_tier AS ENUM ('free', 'basic', 'professional', 'enterprise');
CREATE TYPE job_status AS ENUM ('draft', 'posted', 'matched', 'quoted', 'accepted', 'in_progress', 'completed', 'cancelled', 'disputed');
CREATE TYPE payment_method AS ENUM ('online', 'cash', 'wallet');
CREATE TYPE payment_status AS ENUM ('pending', 'authorized', 'captured', 'failed', 'refunded', 'partially_refunded');
CREATE TYPE escrow_status AS ENUM ('held', 'released', 'refunded', 'disputed');
CREATE TYPE dispute_type AS ENUM ('quality', 'no_show', 'late_arrival', 'overcharge', 'non_payment', 'damage', 'harassment', 'other');
CREATE TYPE dispute_severity AS ENUM ('low', 'medium', 'high', 'critical');
CREATE TYPE dispute_status AS ENUM ('open', 'under_review', 'mediation', 'escalated', 'resolved', 'closed');
CREATE TYPE moderation_status AS ENUM ('pending', 'approved', 'rejected', 'flagged');
CREATE TYPE pricing_model AS ENUM ('hourly', 'fixed', 'per_sqft', 'per_unit', 'per_tree', 'per_day', 'negotiable');

-- ============================================================
-- JURISDICTIONS
-- ============================================================

CREATE TABLE jurisdictions (
    id VARCHAR(10) PRIMARY KEY,              -- e.g. "in", "us", "uk"
    name VARCHAR(100) NOT NULL,
    default_language VARCHAR(10) NOT NULL DEFAULT 'en',
    currency VARCHAR(3) NOT NULL,             -- ISO 4217
    currency_symbol VARCHAR(5) NOT NULL,
    phone_prefix VARCHAR(5) NOT NULL,
    timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',
    is_active BOOLEAN NOT NULL DEFAULT false,
    config JSONB NOT NULL DEFAULT '{}',       -- full jurisdiction config
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- USERS
-- ============================================================

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type user_type NOT NULL,
    phone VARCHAR(20) NOT NULL,
    email VARCHAR(255),
    name VARCHAR(255),
    avatar_url TEXT,
    jurisdiction_id VARCHAR(10) NOT NULL REFERENCES jurisdictions(id),
    preferred_language VARCHAR(10) NOT NULL DEFAULT 'en',
    device_type device_type NOT NULL DEFAULT 'smartphone',
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_seen_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_users_type ON users(type);
CREATE INDEX idx_users_jurisdiction ON users(jurisdiction_id);
CREATE INDEX idx_users_created_at ON users(created_at);

-- ============================================================
-- PROVIDER PROFILES
-- ============================================================

CREATE TABLE provider_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    business_name VARCHAR(255),
    description TEXT,
    skills TEXT[] NOT NULL DEFAULT '{}',
    service_radius_km DECIMAL(5,1) NOT NULL DEFAULT 10.0,
    postcode VARCHAR(20),
    location GEOGRAPHY(POINT, 4326),          -- PostGIS point
    trust_score DECIMAL(3,2) NOT NULL DEFAULT 0.00,  -- 0.00 to 5.00
    level provider_level NOT NULL DEFAULT 'new',
    verification_status verification_status NOT NULL DEFAULT 'unverified',
    subscription_tier subscription_tier NOT NULL DEFAULT 'free',
    total_jobs_completed INTEGER NOT NULL DEFAULT 0,
    total_reviews INTEGER NOT NULL DEFAULT 0,
    avg_response_time_minutes INTEGER,
    availability_schedule JSONB DEFAULT '{}',  -- structured weekly schedule
    documents JSONB DEFAULT '[]',              -- uploaded verification docs
    bank_account_id VARCHAR(255),
    wallet_balance DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    is_available BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_provider_location ON provider_profiles USING GIST(location);
CREATE INDEX idx_provider_skills ON provider_profiles USING GIN(skills);
CREATE INDEX idx_provider_postcode ON provider_profiles(postcode);
CREATE INDEX idx_provider_level ON provider_profiles(level);
CREATE INDEX idx_provider_trust ON provider_profiles(trust_score DESC);
CREATE INDEX idx_provider_verification ON provider_profiles(verification_status);

-- ============================================================
-- CATEGORIES
-- ============================================================

CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug VARCHAR(100) NOT NULL UNIQUE,
    name JSONB NOT NULL DEFAULT '{}',         -- {"en": "Plumbing", "hi": "प्लंबिंग"}
    parent_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    icon VARCHAR(50),
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    requires_license BOOLEAN NOT NULL DEFAULT false,
    pricing_model pricing_model NOT NULL DEFAULT 'negotiable',
    metadata JSONB DEFAULT '{}',              -- category-specific config
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_categories_parent ON categories(parent_id);
CREATE INDEX idx_categories_slug ON categories(slug);
CREATE INDEX idx_categories_active ON categories(is_active);

-- ============================================================
-- PROVIDER ↔ CATEGORY (many-to-many)
-- ============================================================

CREATE TABLE provider_categories (
    provider_id UUID NOT NULL REFERENCES provider_profiles(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    years_experience INTEGER,
    is_primary BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (provider_id, category_id)
);

-- ============================================================
-- JOBS
-- ============================================================

CREATE TABLE jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES users(id),
    provider_id UUID REFERENCES users(id),
    category_id UUID NOT NULL REFERENCES categories(id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    postcode VARCHAR(20),
    location GEOGRAPHY(POINT, 4326),
    address TEXT,
    status job_status NOT NULL DEFAULT 'draft',
    urgency VARCHAR(20) DEFAULT 'normal',      -- normal, urgent, flexible
    scheduled_at TIMESTAMPTZ,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    quoted_price DECIMAL(12,2),
    final_price DECIMAL(12,2),
    currency VARCHAR(3) NOT NULL DEFAULT 'INR',
    payment_method payment_method NOT NULL DEFAULT 'online',
    is_recurring BOOLEAN NOT NULL DEFAULT false,
    recurrence_rule JSONB,                     -- iCal-like recurrence
    photos TEXT[] DEFAULT '{}',                -- before photos
    completion_photos TEXT[] DEFAULT '{}',     -- after photos
    jurisdiction_id VARCHAR(10) NOT NULL REFERENCES jurisdictions(id),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_jobs_customer ON jobs(customer_id);
CREATE INDEX idx_jobs_provider ON jobs(provider_id);
CREATE INDEX idx_jobs_category ON jobs(category_id);
CREATE INDEX idx_jobs_status ON jobs(status);
CREATE INDEX idx_jobs_location ON jobs USING GIST(location);
CREATE INDEX idx_jobs_postcode ON jobs(postcode);
CREATE INDEX idx_jobs_scheduled ON jobs(scheduled_at);
CREATE INDEX idx_jobs_created ON jobs(created_at DESC);
CREATE INDEX idx_jobs_jurisdiction ON jobs(jurisdiction_id);

-- ============================================================
-- JOB QUOTES (providers can submit quotes)
-- ============================================================

CREATE TABLE job_quotes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    provider_id UUID NOT NULL REFERENCES users(id),
    amount DECIMAL(12,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'INR',
    message TEXT,
    estimated_duration_hours DECIMAL(5,1),
    is_accepted BOOLEAN NOT NULL DEFAULT false,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_quotes_job ON job_quotes(job_id);
CREATE INDEX idx_quotes_provider ON job_quotes(provider_id);
CREATE UNIQUE INDEX idx_quotes_job_provider ON job_quotes(job_id, provider_id);

-- ============================================================
-- TRANSACTIONS
-- ============================================================

CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_id UUID NOT NULL REFERENCES jobs(id),
    amount DECIMAL(12,2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    commission_rate DECIMAL(4,2) NOT NULL,     -- percentage
    commission_amount DECIMAL(12,2) NOT NULL,
    tax_amount DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    provider_payout_amount DECIMAL(12,2) NOT NULL,
    payment_status payment_status NOT NULL DEFAULT 'pending',
    escrow_status escrow_status,
    payment_gateway VARCHAR(50),
    gateway_order_id VARCHAR(255),
    gateway_payment_id VARCHAR(255),
    gateway_signature VARCHAR(255),
    paid_at TIMESTAMPTZ,
    settled_at TIMESTAMPTZ,
    refund_amount DECIMAL(12,2),
    refunded_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_job ON transactions(job_id);
CREATE INDEX idx_transactions_status ON transactions(payment_status);
CREATE INDEX idx_transactions_gateway ON transactions(gateway_order_id);

-- ============================================================
-- REVIEWS
-- ============================================================

CREATE TABLE reviews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_id UUID NOT NULL REFERENCES jobs(id),
    reviewer_id UUID NOT NULL REFERENCES users(id),
    reviewee_id UUID NOT NULL REFERENCES users(id),
    rating SMALLINT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    language VARCHAR(10) DEFAULT 'en',
    photos TEXT[] DEFAULT '{}',
    moderation_status moderation_status NOT NULL DEFAULT 'pending',
    is_public BOOLEAN NOT NULL DEFAULT true,
    response TEXT,                             -- reviewee can respond
    responded_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reviews_job ON reviews(job_id);
CREATE INDEX idx_reviews_reviewer ON reviews(reviewer_id);
CREATE INDEX idx_reviews_reviewee ON reviews(reviewee_id);
CREATE INDEX idx_reviews_rating ON reviews(rating);
CREATE UNIQUE INDEX idx_reviews_unique ON reviews(job_id, reviewer_id);

-- ============================================================
-- DISPUTES
-- ============================================================

CREATE TABLE disputes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_id UUID NOT NULL REFERENCES jobs(id),
    raised_by UUID NOT NULL REFERENCES users(id),
    against UUID NOT NULL REFERENCES users(id),
    type dispute_type NOT NULL,
    severity dispute_severity NOT NULL DEFAULT 'low',
    status dispute_status NOT NULL DEFAULT 'open',
    description TEXT NOT NULL,
    evidence JSONB DEFAULT '[]',               -- [{type, url, description}]
    resolution TEXT,
    resolved_by UUID REFERENCES users(id),
    resolution_amount DECIMAL(12,2),
    escalated_at TIMESTAMPTZ,
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_disputes_job ON disputes(job_id);
CREATE INDEX idx_disputes_raised_by ON disputes(raised_by);
CREATE INDEX idx_disputes_status ON disputes(status);
CREATE INDEX idx_disputes_severity ON disputes(severity);

-- ============================================================
-- POINTS LEDGER (Gamification)
-- ============================================================

CREATE TABLE points_ledger (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    points INTEGER NOT NULL,
    reason VARCHAR(100) NOT NULL,              -- e.g. "job_completed", "review_given"
    reference_id UUID,                         -- job_id, review_id, etc.
    reference_type VARCHAR(50),
    balance_after INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_points_user ON points_ledger(user_id);
CREATE INDEX idx_points_created ON points_ledger(created_at DESC);

-- ============================================================
-- OTP MANAGEMENT
-- ============================================================

CREATE TABLE otp_codes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone VARCHAR(20) NOT NULL,
    code VARCHAR(6) NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 3,
    expires_at TIMESTAMPTZ NOT NULL,
    verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_otp_phone ON otp_codes(phone);
CREATE INDEX idx_otp_expires ON otp_codes(expires_at);

-- ============================================================
-- NOTIFICATIONS
-- ============================================================

CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    type VARCHAR(50) NOT NULL,                 -- job_update, payment, promotion, system
    title VARCHAR(255) NOT NULL,
    body TEXT,
    data JSONB DEFAULT '{}',
    channel VARCHAR(20) NOT NULL DEFAULT 'push', -- push, sms, email, in_app
    is_read BOOLEAN NOT NULL DEFAULT false,
    sent_at TIMESTAMPTZ,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user ON notifications(user_id);
CREATE INDEX idx_notifications_unread ON notifications(user_id, is_read) WHERE is_read = false;
CREATE INDEX idx_notifications_created ON notifications(created_at DESC);

-- ============================================================
-- PROVIDER SCHEDULE / AVAILABILITY
-- ============================================================

CREATE TABLE provider_availability (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider_id UUID NOT NULL REFERENCES provider_profiles(id) ON DELETE CASCADE,
    day_of_week SMALLINT NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6), -- 0=Sun
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    is_available BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_availability_provider ON provider_availability(provider_id);
CREATE UNIQUE INDEX idx_availability_slot ON provider_availability(provider_id, day_of_week, start_time);

-- ============================================================
-- CROP WORK CATALOG (for crop & land services)
-- ============================================================

CREATE TABLE crop_catalog (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    jurisdiction_id VARCHAR(10) NOT NULL REFERENCES jurisdictions(id),
    crop_slug VARCHAR(100) NOT NULL,
    name JSONB NOT NULL DEFAULT '{}',          -- {"en": "Coconut", "ml": "തെങ്ങ്"}
    work_types JSONB NOT NULL DEFAULT '[]',    -- [{slug, name, pricing_model, typical_price_range}]
    seasonal_calendar JSONB DEFAULT '{}',      -- {month: [available_work_types]}
    metadata JSONB DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_crop_jurisdiction ON crop_catalog(jurisdiction_id, crop_slug);

-- ============================================================
-- ROUTE MANAGEMENT (for circuit-based workers)
-- ============================================================

CREATE TABLE routes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider_id UUID NOT NULL REFERENCES provider_profiles(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    postcodes TEXT[] NOT NULL DEFAULT '{}',
    frequency VARCHAR(50) NOT NULL,            -- weekly, biweekly, monthly
    day_of_week SMALLINT,
    max_stops INTEGER NOT NULL DEFAULT 20,
    current_stops INTEGER NOT NULL DEFAULT 0,
    price_per_stop DECIMAL(10,2),
    currency VARCHAR(3) NOT NULL DEFAULT 'INR',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_routes_provider ON routes(provider_id);
CREATE INDEX idx_routes_postcodes ON routes USING GIN(postcodes);

CREATE TABLE route_stops (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    route_id UUID NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
    customer_id UUID NOT NULL REFERENCES users(id),
    address TEXT NOT NULL,
    location GEOGRAPHY(POINT, 4326),
    stop_order INTEGER NOT NULL,
    notes TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_route_stops_route ON route_stops(route_id);

-- ============================================================
-- AUDIT LOG
-- ============================================================

CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID,
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_user ON audit_log(user_id);
CREATE INDEX idx_audit_entity ON audit_log(entity_type, entity_id);
CREATE INDEX idx_audit_created ON audit_log(created_at DESC);

-- ============================================================
-- UPDATED_AT TRIGGER
-- ============================================================

CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply trigger to all tables with updated_at
CREATE TRIGGER set_updated_at BEFORE UPDATE ON jurisdictions FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER set_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER set_updated_at BEFORE UPDATE ON provider_profiles FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER set_updated_at BEFORE UPDATE ON categories FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER set_updated_at BEFORE UPDATE ON jobs FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER set_updated_at BEFORE UPDATE ON transactions FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER set_updated_at BEFORE UPDATE ON reviews FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER set_updated_at BEFORE UPDATE ON disputes FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER set_updated_at BEFORE UPDATE ON crop_catalog FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER set_updated_at BEFORE UPDATE ON routes FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================
-- SEED: Default jurisdiction
-- ============================================================

INSERT INTO jurisdictions (id, name, default_language, currency, currency_symbol, phone_prefix, timezone, is_active) VALUES
('in', 'India', 'en', 'INR', '₹', '+91', 'Asia/Kolkata', true);
