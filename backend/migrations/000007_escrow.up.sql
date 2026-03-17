-- Escrow transactions for holding payments until job completion.
-- The escrow_status enum already exists in 000001_init_schema.up.sql.

CREATE TABLE escrow_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_id UUID NOT NULL REFERENCES jobs(id),
    customer_id UUID NOT NULL REFERENCES users(id),
    provider_id UUID NOT NULL REFERENCES users(id),
    amount NUMERIC(12,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'INR',
    status escrow_status NOT NULL DEFAULT 'held',
    gateway_payment_id VARCHAR(255),
    held_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    released_at TIMESTAMPTZ,
    refunded_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_escrow_job ON escrow_transactions(job_id);
CREATE INDEX idx_escrow_customer ON escrow_transactions(customer_id);
CREATE INDEX idx_escrow_provider ON escrow_transactions(provider_id);
CREATE INDEX idx_escrow_status ON escrow_transactions(status);

-- Apply updated_at trigger
CREATE TRIGGER set_updated_at BEFORE UPDATE ON escrow_transactions FOR EACH ROW EXECUTE FUNCTION update_updated_at();
