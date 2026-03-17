DROP TRIGGER IF EXISTS set_updated_at ON subscriptions;
DROP INDEX IF EXISTS idx_subscriptions_expires;
DROP INDEX IF EXISTS idx_subscriptions_status;
DROP INDEX IF EXISTS idx_subscriptions_provider;
DROP TABLE IF EXISTS subscriptions;
