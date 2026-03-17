-- Recurring job scheduling tables.

CREATE TYPE recurrence_frequency AS ENUM ('daily', 'weekly', 'biweekly', 'monthly', 'quarterly');
CREATE TYPE schedule_status AS ENUM ('active', 'paused', 'cancelled');

CREATE TABLE recurring_schedules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES users(id),
    provider_id UUID NOT NULL REFERENCES users(id),
    category_id UUID NOT NULL REFERENCES categories(id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    frequency recurrence_frequency NOT NULL,
    day_of_week INT,        -- 0=Sunday, 1=Monday, ... 6=Saturday (nullable)
    day_of_month INT,       -- 1-31 (nullable)
    preferred_time TIME NOT NULL DEFAULT '09:00:00',
    amount NUMERIC(12,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'INR',
    status schedule_status NOT NULL DEFAULT 'active',
    next_occurrence TIMESTAMPTZ,
    last_occurrence TIMESTAMPTZ,
    total_occurrences INT NOT NULL DEFAULT 0,
    max_occurrences INT,    -- null means unlimited
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_recurring_customer ON recurring_schedules(customer_id);
CREATE INDEX idx_recurring_provider ON recurring_schedules(provider_id);
CREATE INDEX idx_recurring_status ON recurring_schedules(status);
CREATE INDEX idx_recurring_next_occurrence ON recurring_schedules(next_occurrence);

-- Apply updated_at trigger
CREATE TRIGGER set_updated_at BEFORE UPDATE ON recurring_schedules FOR EACH ROW EXECUTE FUNCTION update_updated_at();
