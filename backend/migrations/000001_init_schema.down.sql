-- Rollback initial schema

DROP TRIGGER IF EXISTS set_updated_at ON routes;
DROP TRIGGER IF EXISTS set_updated_at ON crop_catalog;
DROP TRIGGER IF EXISTS set_updated_at ON disputes;
DROP TRIGGER IF EXISTS set_updated_at ON reviews;
DROP TRIGGER IF EXISTS set_updated_at ON transactions;
DROP TRIGGER IF EXISTS set_updated_at ON jobs;
DROP TRIGGER IF EXISTS set_updated_at ON categories;
DROP TRIGGER IF EXISTS set_updated_at ON provider_profiles;
DROP TRIGGER IF EXISTS set_updated_at ON users;
DROP TRIGGER IF EXISTS set_updated_at ON jurisdictions;

DROP FUNCTION IF EXISTS update_updated_at();

DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS route_stops;
DROP TABLE IF EXISTS routes;
DROP TABLE IF EXISTS crop_catalog;
DROP TABLE IF EXISTS provider_availability;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS otp_codes;
DROP TABLE IF EXISTS points_ledger;
DROP TABLE IF EXISTS disputes;
DROP TABLE IF EXISTS reviews;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS job_quotes;
DROP TABLE IF EXISTS jobs;
DROP TABLE IF EXISTS provider_categories;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS provider_profiles;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS jurisdictions;

DROP TYPE IF EXISTS pricing_model;
DROP TYPE IF EXISTS moderation_status;
DROP TYPE IF EXISTS dispute_status;
DROP TYPE IF EXISTS dispute_severity;
DROP TYPE IF EXISTS dispute_type;
DROP TYPE IF EXISTS escrow_status;
DROP TYPE IF EXISTS payment_status;
DROP TYPE IF EXISTS payment_method;
DROP TYPE IF EXISTS job_status;
DROP TYPE IF EXISTS subscription_tier;
DROP TYPE IF EXISTS provider_level;
DROP TYPE IF EXISTS verification_status;
DROP TYPE IF EXISTS device_type;
DROP TYPE IF EXISTS user_type;

DROP EXTENSION IF EXISTS "postgis";
DROP EXTENSION IF EXISTS "uuid-ossp";
