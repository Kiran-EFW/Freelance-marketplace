-- name: CreateNotification :one
INSERT INTO notifications (user_id, type, title, body, data, channel)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListUnreadNotifications :many
SELECT * FROM notifications
WHERE user_id = $1 AND is_read = false
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: MarkNotificationRead :exec
UPDATE notifications SET is_read = true, read_at = NOW()
WHERE id = $1 AND user_id = $2;

-- name: MarkAllRead :exec
UPDATE notifications SET is_read = true, read_at = NOW()
WHERE user_id = $1 AND is_read = false;

-- name: CountUnread :one
SELECT COUNT(*) FROM notifications
WHERE user_id = $1 AND is_read = false;
