-- name: CreateTransaction :one
INSERT INTO transactions (job_id, amount, currency, commission_rate, commission_amount, tax_amount, provider_payout_amount, payment_gateway, gateway_order_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetTransactionByID :one
SELECT * FROM transactions WHERE id = $1;

-- name: GetTransactionByJob :one
SELECT * FROM transactions WHERE job_id = $1;

-- name: GetTransactionByGatewayOrder :one
SELECT * FROM transactions WHERE gateway_order_id = $1;

-- name: UpdatePaymentStatus :one
UPDATE transactions
SET payment_status = $2,
    gateway_payment_id = $3,
    gateway_signature = $4,
    paid_at = CASE WHEN $2 = 'captured' THEN NOW() ELSE paid_at END
WHERE id = $1
RETURNING *;

-- name: ReleaseEscrow :one
UPDATE transactions
SET escrow_status = 'released',
    settled_at = NOW()
WHERE id = $1
RETURNING *;

-- name: RefundTransaction :one
UPDATE transactions
SET payment_status = 'refunded',
    refund_amount = $2,
    refunded_at = NOW(),
    escrow_status = 'refunded'
WHERE id = $1
RETURNING *;

-- name: ListUnsettledTransactions :many
SELECT t.*, j.provider_id
FROM transactions t
JOIN jobs j ON j.id = t.job_id
WHERE t.payment_status = 'captured'
  AND t.escrow_status = 'held'
  AND t.paid_at < NOW() - INTERVAL '24 hours'
ORDER BY t.paid_at ASC
LIMIT $1;

-- name: GetProviderEarnings :one
SELECT
    COALESCE(SUM(provider_payout_amount), 0) as total_earnings,
    COUNT(*) as total_transactions
FROM transactions t
JOIN jobs j ON j.id = t.job_id
WHERE j.provider_id = $1
  AND t.payment_status = 'captured'
  AND t.settled_at IS NOT NULL;
