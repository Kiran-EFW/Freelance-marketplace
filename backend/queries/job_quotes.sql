-- name: CreateJobQuote :one
INSERT INTO job_quotes (job_id, provider_id, amount, currency, message, estimated_duration_hours, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetJobQuoteByID :one
SELECT * FROM job_quotes WHERE id = $1;

-- name: ListQuotesByJob :many
SELECT * FROM job_quotes WHERE job_id = $1 ORDER BY created_at;

-- name: ListQuotesByProvider :many
SELECT * FROM job_quotes WHERE provider_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: AcceptJobQuote :one
UPDATE job_quotes SET is_accepted = true WHERE id = $1 RETURNING *;
