-- name: RegisterDeviceToken :one
INSERT INTO device_tokens (user_id, token, platform)
VALUES ($1, $2, $3)
ON CONFLICT (token) DO UPDATE SET user_id = $1, platform = $3, is_active = true, updated_at = NOW()
RETURNING *;

-- name: GetDeviceTokensForUser :many
SELECT * FROM device_tokens WHERE user_id = $1 AND is_active = true;

-- name: DeactivateDeviceToken :exec
UPDATE device_tokens SET is_active = false WHERE token = $1;
