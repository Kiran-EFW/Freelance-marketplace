-- name: CreateJob :one
INSERT INTO jobs (customer_id, category_id, title, description, postcode, location, address, urgency, scheduled_at, currency, payment_method, is_recurring, recurrence_rule, jurisdiction_id, photos)
VALUES ($1, $2, $3, $4, $5, ST_MakePoint($6, $7)::geography, $8, $9, $10, $11, $12, $13, $14, $15, $16)
RETURNING *;

-- name: GetJobByID :one
SELECT j.*,
       cu.name as customer_name, cu.phone as customer_phone,
       pu.name as provider_name, pu.phone as provider_phone,
       c.slug as category_slug, c.name as category_name
FROM jobs j
JOIN users cu ON cu.id = j.customer_id
LEFT JOIN users pu ON pu.id = j.provider_id
JOIN categories c ON c.id = j.category_id
WHERE j.id = $1;

-- name: ListJobsByCustomer :many
SELECT j.*, c.slug as category_slug, c.name as category_name
FROM jobs j
JOIN categories c ON c.id = j.category_id
WHERE j.customer_id = $1
ORDER BY j.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListJobsByProvider :many
SELECT j.*, c.slug as category_slug, c.name as category_name,
       cu.name as customer_name
FROM jobs j
JOIN categories c ON c.id = j.category_id
JOIN users cu ON cu.id = j.customer_id
WHERE j.provider_id = $1
ORDER BY j.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListJobsByStatus :many
SELECT j.*, c.slug as category_slug, c.name as category_name
FROM jobs j
JOIN categories c ON c.id = j.category_id
WHERE j.status = $1
  AND j.jurisdiction_id = $2
ORDER BY j.created_at DESC
LIMIT $3 OFFSET $4;

-- name: UpdateJobStatus :one
UPDATE jobs SET status = $2 WHERE id = $1 RETURNING *;

-- name: AssignProvider :one
UPDATE jobs SET provider_id = $2, status = 'accepted' WHERE id = $1 RETURNING *;

-- name: CompleteJob :one
UPDATE jobs
SET status = 'completed',
    completed_at = NOW(),
    final_price = $2,
    completion_photos = $3
WHERE id = $1
RETURNING *;

-- name: SearchJobsByLocation :many
SELECT j.*, c.slug as category_slug, c.name as category_name,
       ST_Distance(j.location, ST_MakePoint($1, $2)::geography) as distance_meters
FROM jobs j
JOIN categories c ON c.id = j.category_id
WHERE ST_DWithin(j.location, ST_MakePoint($1, $2)::geography, $3)
  AND j.status = 'posted'
ORDER BY j.created_at DESC
LIMIT $4 OFFSET $5;

-- name: CountJobsByStatus :one
SELECT COUNT(*) FROM jobs WHERE status = $1 AND jurisdiction_id = $2;
