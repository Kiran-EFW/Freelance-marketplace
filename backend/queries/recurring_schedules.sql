-- name: CreateSchedule :one
INSERT INTO recurring_schedules (customer_id, provider_id, category_id, title, description, frequency, day_of_week, day_of_month, preferred_time, amount, currency, status, next_occurrence)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'active', $12)
RETURNING *;

-- name: GetScheduleByID :one
SELECT * FROM recurring_schedules WHERE id = $1;

-- name: ListSchedulesByCustomer :many
SELECT * FROM recurring_schedules
WHERE customer_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListSchedulesByProvider :many
SELECT * FROM recurring_schedules
WHERE provider_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateScheduleStatus :one
UPDATE recurring_schedules
SET status = $2
WHERE id = $1
RETURNING *;

-- name: UpdateNextOccurrence :exec
UPDATE recurring_schedules
SET next_occurrence = $2,
    last_occurrence = $3,
    total_occurrences = total_occurrences + 1
WHERE id = $1;

-- name: ListDueSchedules :many
SELECT * FROM recurring_schedules
WHERE next_occurrence <= NOW()
  AND status = 'active'
ORDER BY next_occurrence ASC
LIMIT $1;
