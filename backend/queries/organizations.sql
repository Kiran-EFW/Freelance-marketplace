-- name: CreateOrganization :one
INSERT INTO organizations (name, type, address, postcode, city, state, country, contact_phone, contact_email, logo_url, settings, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: GetOrganizationByID :one
SELECT * FROM organizations WHERE id = $1;

-- name: ListOrganizations :many
SELECT * FROM organizations
WHERE status = COALESCE(sqlc.narg('status'), status)
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: AddOrganizationMember :one
INSERT INTO organization_members (org_id, user_id, role, status)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: RemoveOrganizationMember :exec
DELETE FROM organization_members WHERE org_id = $1 AND user_id = $2;

-- name: ListOrganizationMembers :many
SELECT om.*, u.name as user_name, u.phone as user_phone, u.email as user_email
FROM organization_members om
JOIN users u ON u.id = om.user_id
WHERE om.org_id = $1
ORDER BY om.joined_at DESC
LIMIT $2 OFFSET $3;

-- name: GetMemberRole :one
SELECT role, status FROM organization_members WHERE org_id = $1 AND user_id = $2;

-- name: CreateOrganizationServiceRequest :one
INSERT INTO organization_service_requests (org_id, requested_by, category_id, title, description, priority, status, scheduled_at, notes)
VALUES ($1, $2, $3, $4, $5, $6, 'pending', $7, $8)
RETURNING *;

-- name: ListOrganizationServiceRequests :many
SELECT osr.*, u.name as requester_name, c.slug as category_slug, c.name as category_name,
       pu.name as provider_name
FROM organization_service_requests osr
JOIN users u ON u.id = osr.requested_by
JOIN categories c ON c.id = osr.category_id
LEFT JOIN users pu ON pu.id = osr.assigned_provider_id
WHERE osr.org_id = $1
  AND osr.status = COALESCE(sqlc.narg('status'), osr.status)
  AND osr.priority = COALESCE(sqlc.narg('priority'), osr.priority)
ORDER BY osr.created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateOrganizationServiceRequestStatus :one
UPDATE organization_service_requests
SET status = $2
WHERE id = $1
RETURNING *;

-- name: AssignOrganizationServiceRequestProvider :one
UPDATE organization_service_requests
SET assigned_provider_id = $2, status = 'assigned'
WHERE id = $1
RETURNING *;

-- name: CountOrganizationServiceRequestsByStatus :one
SELECT COUNT(*) FROM organization_service_requests WHERE org_id = $1 AND status = $2;

-- name: CountOrganizationMembers :one
SELECT COUNT(*) FROM organization_members WHERE org_id = $1 AND status = 'active';
