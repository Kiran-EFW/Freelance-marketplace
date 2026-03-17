-- name: CreateEscrowTransaction :one
INSERT INTO escrow_transactions (job_id, customer_id, provider_id, amount, currency, status, gateway_payment_id)
VALUES ($1, $2, $3, $4, $5, 'held', $6)
RETURNING *;

-- name: GetEscrowByID :one
SELECT * FROM escrow_transactions WHERE id = $1;

-- name: GetEscrowByJobID :one
SELECT * FROM escrow_transactions WHERE job_id = $1;

-- name: ReleaseEscrowTransaction :one
UPDATE escrow_transactions
SET status = 'released', released_at = NOW()
WHERE id = $1 AND status = 'held'
RETURNING *;

-- name: RefundEscrowTransaction :one
UPDATE escrow_transactions
SET status = 'refunded', refunded_at = NOW()
WHERE id = $1 AND status IN ('held', 'disputed')
RETURNING *;

-- name: DisputeEscrowTransaction :one
UPDATE escrow_transactions
SET status = 'disputed'
WHERE id = $1 AND status = 'held'
RETURNING *;

-- name: ListEscrowByUser :many
SELECT * FROM escrow_transactions
WHERE customer_id = $1 OR provider_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
