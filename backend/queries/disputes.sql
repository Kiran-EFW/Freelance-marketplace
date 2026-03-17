-- name: CreateDispute :one
INSERT INTO disputes (job_id, raised_by, against, type, severity, description, evidence)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetDisputeByID :one
SELECT d.*,
       rb.name as raised_by_name, ag.name as against_name,
       j.title as job_title
FROM disputes d
JOIN users rb ON rb.id = d.raised_by
JOIN users ag ON ag.id = d.against
JOIN jobs j ON j.id = d.job_id
WHERE d.id = $1;

-- name: ListDisputesByStatus :many
SELECT d.*, rb.name as raised_by_name, j.title as job_title
FROM disputes d
JOIN users rb ON rb.id = d.raised_by
JOIN jobs j ON j.id = d.job_id
WHERE d.status = $1
ORDER BY d.severity DESC, d.created_at ASC
LIMIT $2 OFFSET $3;

-- name: UpdateDisputeStatus :one
UPDATE disputes SET status = $2 WHERE id = $1 RETURNING *;

-- name: ResolveDispute :one
UPDATE disputes
SET status = 'resolved',
    resolution = $2,
    resolved_by = $3,
    resolution_amount = $4,
    resolved_at = NOW()
WHERE id = $1
RETURNING *;

-- name: EscalateDispute :one
UPDATE disputes
SET status = 'escalated',
    severity = $2,
    escalated_at = NOW()
WHERE id = $1
RETURNING *;
