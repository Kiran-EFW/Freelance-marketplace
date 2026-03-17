-- name: CreateReview :one
INSERT INTO reviews (job_id, reviewer_id, reviewee_id, rating, comment, language)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetReviewByID :one
SELECT r.*, ru.name as reviewer_name, eu.name as reviewee_name
FROM reviews r
JOIN users ru ON ru.id = r.reviewer_id
JOIN users eu ON eu.id = r.reviewee_id
WHERE r.id = $1;

-- name: ListReviewsByProvider :many
SELECT r.*, u.name as reviewer_name, j.title as job_title
FROM reviews r
JOIN users u ON u.id = r.reviewer_id
JOIN jobs j ON j.id = r.job_id
WHERE r.reviewee_id = $1
  AND r.moderation_status = 'approved'
  AND r.is_public = true
ORDER BY r.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetProviderAvgRating :one
SELECT COALESCE(AVG(rating)::DECIMAL(3,2), 0) as avg_rating, COUNT(*) as total_reviews
FROM reviews
WHERE reviewee_id = $1
  AND moderation_status = 'approved';

-- name: RespondToReview :one
UPDATE reviews SET response = $2, responded_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ModerateReview :exec
UPDATE reviews SET moderation_status = $2
WHERE id = $1;
