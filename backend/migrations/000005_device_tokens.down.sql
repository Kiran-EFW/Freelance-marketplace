-- Revert device tokens table.

DROP TRIGGER IF EXISTS set_updated_at ON device_tokens;
DROP TABLE IF EXISTS device_tokens;
