-- B2B Dashboard: Organizations, Members, and Service Requests

-- ============================================================
-- ENUMS
-- ============================================================

CREATE TYPE org_type AS ENUM ('housing_society', 'company', 'institution');
CREATE TYPE org_status AS ENUM ('active', 'suspended', 'pending');
CREATE TYPE org_role AS ENUM ('admin', 'manager', 'member');
CREATE TYPE member_status AS ENUM ('active', 'inactive', 'invited');
CREATE TYPE request_priority AS ENUM ('low', 'medium', 'high', 'urgent');
CREATE TYPE request_status AS ENUM ('pending', 'assigned', 'in_progress', 'completed', 'cancelled');

-- ============================================================
-- ORGANIZATIONS
-- ============================================================

CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    type org_type NOT NULL,
    address TEXT,
    postcode VARCHAR(20),
    city VARCHAR(100),
    state VARCHAR(100),
    country VARCHAR(100) NOT NULL DEFAULT 'India',
    contact_phone VARCHAR(20),
    contact_email VARCHAR(255),
    logo_url TEXT,
    settings JSONB NOT NULL DEFAULT '{}',
    status org_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_organizations_status ON organizations(status);
CREATE INDEX idx_organizations_type ON organizations(type);
CREATE INDEX idx_organizations_city ON organizations(city);

-- ============================================================
-- ORGANIZATION MEMBERS
-- ============================================================

CREATE TABLE organization_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role org_role NOT NULL DEFAULT 'member',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status member_status NOT NULL DEFAULT 'invited'
);

CREATE INDEX idx_org_members_org ON organization_members(org_id);
CREATE INDEX idx_org_members_user ON organization_members(user_id);
CREATE INDEX idx_org_members_status ON organization_members(status);
CREATE UNIQUE INDEX idx_org_members_unique ON organization_members(org_id, user_id);

-- ============================================================
-- ORGANIZATION SERVICE REQUESTS
-- ============================================================

CREATE TABLE organization_service_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    requested_by UUID NOT NULL REFERENCES users(id),
    category_id UUID NOT NULL REFERENCES categories(id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    priority request_priority NOT NULL DEFAULT 'medium',
    status request_status NOT NULL DEFAULT 'pending',
    assigned_provider_id UUID REFERENCES users(id),
    scheduled_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_org_requests_org ON organization_service_requests(org_id);
CREATE INDEX idx_org_requests_status ON organization_service_requests(status);
CREATE INDEX idx_org_requests_priority ON organization_service_requests(priority);
CREATE INDEX idx_org_requests_requested_by ON organization_service_requests(requested_by);
CREATE INDEX idx_org_requests_assigned ON organization_service_requests(assigned_provider_id);

-- ============================================================
-- UPDATED_AT TRIGGERS
-- ============================================================

CREATE TRIGGER set_updated_at BEFORE UPDATE ON organizations FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER set_updated_at BEFORE UPDATE ON organization_service_requests FOR EACH ROW EXECUTE FUNCTION update_updated_at();
