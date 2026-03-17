-- name: AddProviderCategory :exec
INSERT INTO provider_categories (provider_id, category_id, years_experience, is_primary)
VALUES ($1, $2, $3, $4)
ON CONFLICT (provider_id, category_id) DO UPDATE SET years_experience = $3, is_primary = $4;

-- name: ListProviderCategories :many
SELECT * FROM provider_categories WHERE provider_id = $1;

-- name: RemoveProviderCategory :exec
DELETE FROM provider_categories WHERE provider_id = $1 AND category_id = $2;
