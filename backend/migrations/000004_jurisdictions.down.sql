-- Revert jurisdiction seed data.

-- Remove non-India jurisdictions added in this migration.
DELETE FROM jurisdictions WHERE id IN ('uk', 'us', 'de', 'fr');

-- Reset India config back to an empty object.
UPDATE jurisdictions SET config = '{}'::jsonb WHERE id = 'in';
