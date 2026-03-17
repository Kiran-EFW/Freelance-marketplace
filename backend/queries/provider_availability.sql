-- name: SetProviderAvailability :exec
INSERT INTO provider_availability (provider_id, day_of_week, start_time, end_time, is_available)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (provider_id, day_of_week, start_time) DO UPDATE SET end_time = $4, is_available = $5;

-- name: GetProviderAvailability :many
SELECT * FROM provider_availability WHERE provider_id = $1 ORDER BY day_of_week;
