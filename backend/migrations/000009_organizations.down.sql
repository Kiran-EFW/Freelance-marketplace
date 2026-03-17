DROP TRIGGER IF EXISTS set_updated_at ON organization_service_requests;
DROP TRIGGER IF EXISTS set_updated_at ON organizations;

DROP TABLE IF EXISTS organization_service_requests;
DROP TABLE IF EXISTS organization_members;
DROP TABLE IF EXISTS organizations;

DROP TYPE IF EXISTS request_status;
DROP TYPE IF EXISTS request_priority;
DROP TYPE IF EXISTS member_status;
DROP TYPE IF EXISTS org_role;
DROP TYPE IF EXISTS org_status;
DROP TYPE IF EXISTS org_type;
