-- name: AddPoints :one
INSERT INTO points_ledger (user_id, points, reason, reference_id, reference_type, balance_after)
VALUES ($1, $2, $3, $4, $5,
    (SELECT COALESCE(SUM(points), 0) + $2 FROM points_ledger WHERE user_id = $1))
RETURNING *;

-- name: GetPointsBalance :one
SELECT COALESCE(SUM(points), 0)::INTEGER as balance
FROM points_ledger
WHERE user_id = $1;

-- name: ListPointsHistory :many
SELECT * FROM points_ledger
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
