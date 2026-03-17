-- name: GetCropCatalogByJurisdiction :many
SELECT * FROM crop_catalog WHERE jurisdiction_id = $1 AND is_active = true ORDER BY crop_slug;

-- name: GetCropBySlug :one
SELECT * FROM crop_catalog WHERE jurisdiction_id = $1 AND crop_slug = $2;
