-- name: CreateProviderProfile :one
INSERT INTO provider_profiles (user_id, business_name, description, skills, service_radius_km, postcode, location)
VALUES ($1, $2, $3, $4, $5, $6, ST_MakePoint($7, $8)::geography)
RETURNING *;

-- name: GetProviderByID :one
SELECT p.*, u.name as user_name, u.phone as user_phone, u.email as user_email, u.avatar_url as user_avatar
FROM provider_profiles p
JOIN users u ON u.id = p.user_id
WHERE p.id = $1;

-- name: GetProviderByUserID :one
SELECT p.*, u.name as user_name, u.phone as user_phone, u.email as user_email, u.avatar_url as user_avatar
FROM provider_profiles p
JOIN users u ON u.id = p.user_id
WHERE p.user_id = $1;

-- name: SearchProvidersByLocation :many
SELECT p.*, u.name as user_name, u.phone as user_phone, u.avatar_url as user_avatar,
       ST_Distance(p.location, ST_MakePoint($1, $2)::geography) as distance_meters
FROM provider_profiles p
JOIN users u ON u.id = p.user_id
WHERE ST_DWithin(p.location, ST_MakePoint($1, $2)::geography, $3)
  AND p.is_available = true
  AND p.verification_status = 'verified'
  AND u.is_active = true
ORDER BY p.trust_score DESC, distance_meters ASC
LIMIT $4 OFFSET $5;

-- name: SearchProvidersByPostcode :many
SELECT p.*, u.name as user_name, u.phone as user_phone, u.avatar_url as user_avatar
FROM provider_profiles p
JOIN users u ON u.id = p.user_id
WHERE p.postcode = $1
  AND p.is_available = true
  AND u.is_active = true
ORDER BY p.trust_score DESC
LIMIT $2 OFFSET $3;

-- name: SearchProvidersBySkill :many
SELECT p.*, u.name as user_name, u.phone as user_phone, u.avatar_url as user_avatar
FROM provider_profiles p
JOIN users u ON u.id = p.user_id
WHERE p.skills @> $1::text[]
  AND p.is_available = true
  AND u.is_active = true
ORDER BY p.trust_score DESC
LIMIT $2 OFFSET $3;

-- name: UpdateProviderProfile :one
UPDATE provider_profiles
SET business_name = COALESCE(sqlc.narg('business_name'), business_name),
    description = COALESCE(sqlc.narg('description'), description),
    skills = COALESCE(sqlc.narg('skills'), skills),
    service_radius_km = COALESCE(sqlc.narg('service_radius_km'), service_radius_km),
    postcode = COALESCE(sqlc.narg('postcode'), postcode),
    is_available = COALESCE(sqlc.narg('is_available'), is_available)
WHERE id = $1
RETURNING *;

-- name: UpdateTrustScore :exec
UPDATE provider_profiles
SET trust_score = $2,
    level = $3
WHERE id = $1;

-- name: IncrementJobsCompleted :exec
UPDATE provider_profiles
SET total_jobs_completed = total_jobs_completed + 1,
    total_reviews = total_reviews + 1
WHERE user_id = $1;

-- name: UpdateVerificationStatus :exec
UPDATE provider_profiles
SET verification_status = $2
WHERE id = $1;
