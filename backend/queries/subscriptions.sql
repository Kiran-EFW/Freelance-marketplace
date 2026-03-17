-- name: CreateSubscription :one
INSERT INTO subscriptions (provider_id, tier, expires_at, amount, currency, payment_method)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: GetActiveSubscription :one
SELECT * FROM subscriptions
WHERE provider_id = $1 AND status = 'active' AND (expires_at IS NULL OR expires_at > NOW())
ORDER BY created_at DESC LIMIT 1;

-- name: GetSubscriptionByID :one
SELECT * FROM subscriptions WHERE id = $1;

-- name: UpdateSubscriptionStatus :exec
UPDATE subscriptions SET status = $2 WHERE id = $1;

-- name: ListExpiredSubscriptions :many
SELECT * FROM subscriptions WHERE status = 'active' AND expires_at < NOW();

-- name: CancelSubscription :exec
UPDATE subscriptions SET status = 'cancelled', auto_renew = false WHERE id = $1;

-- name: ListSubscriptionsByProvider :many
SELECT * FROM subscriptions
WHERE provider_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateGatewaySubscriptionID :exec
UPDATE subscriptions SET gateway_subscription_id = $2 WHERE id = $1;
