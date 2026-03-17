-- name: CreateSOSAlert :one
INSERT INTO sos_alerts (user_id, job_id, latitude, longitude, status, emergency_contacts_notified, notes)
VALUES ($1, $2, $3, $4, 'active', $5, $6)
RETURNING *;

-- name: GetSOSAlert :one
SELECT * FROM sos_alerts WHERE id = $1;

-- name: ResolveSOSAlert :one
UPDATE sos_alerts
SET status = $2, resolved_at = NOW(), notes = COALESCE(sqlc.narg('notes'), notes)
WHERE id = $1
RETURNING *;

-- name: ListActiveSOSAlerts :many
SELECT sa.*, u.name as user_name, u.phone as user_phone
FROM sos_alerts sa
JOIN users u ON u.id = sa.user_id
WHERE sa.status = 'active'
ORDER BY sa.created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListSOSAlertsByUser :many
SELECT * FROM sos_alerts
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ShareLocation :one
INSERT INTO location_shares (job_id, user_id, latitude, longitude, accuracy)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetLatestLocation :one
SELECT * FROM location_shares
WHERE job_id = $1 AND user_id = $2
ORDER BY shared_at DESC
LIMIT 1;

-- name: GetLocationHistory :many
SELECT * FROM location_shares
WHERE job_id = $1
ORDER BY shared_at DESC
LIMIT $2 OFFSET $3;

-- name: AddEmergencyContact :one
INSERT INTO emergency_contacts (user_id, name, phone, relationship)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListEmergencyContacts :many
SELECT * FROM emergency_contacts
WHERE user_id = $1
ORDER BY created_at ASC;

-- name: RemoveEmergencyContact :exec
DELETE FROM emergency_contacts WHERE id = $1 AND user_id = $2;

-- name: GetEmergencyContact :one
SELECT * FROM emergency_contacts WHERE id = $1;
