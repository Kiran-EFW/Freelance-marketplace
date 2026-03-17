-- name: CreateUser :one
INSERT INTO users (type, phone, email, name, jurisdiction_id, preferred_language, device_type)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByPhone :one
SELECT * FROM users WHERE phone = $1;

-- name: UpdateUser :one
UPDATE users
SET name = COALESCE(sqlc.narg('name'), name),
    email = COALESCE(sqlc.narg('email'), email),
    avatar_url = COALESCE(sqlc.narg('avatar_url'), avatar_url),
    preferred_language = COALESCE(sqlc.narg('preferred_language'), preferred_language),
    device_type = COALESCE(sqlc.narg('device_type'), device_type)
WHERE id = $1
RETURNING *;

-- name: UpdateLastSeen :exec
UPDATE users SET last_seen_at = NOW() WHERE id = $1;

-- name: DeactivateUser :exec
UPDATE users SET is_active = false WHERE id = $1;

-- name: ListUsersByJurisdiction :many
SELECT * FROM users
WHERE jurisdiction_id = $1 AND is_active = true
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountUsersByType :one
SELECT COUNT(*) FROM users WHERE type = $1 AND is_active = true;
