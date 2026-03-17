-- name: ListCategories :many
SELECT * FROM categories
WHERE is_active = true
ORDER BY sort_order, slug;

-- name: ListTopLevelCategories :many
SELECT * FROM categories
WHERE parent_id IS NULL AND is_active = true
ORDER BY sort_order, slug;

-- name: ListSubcategories :many
SELECT * FROM categories
WHERE parent_id = $1 AND is_active = true
ORDER BY sort_order, slug;

-- name: GetCategoryBySlug :one
SELECT * FROM categories WHERE slug = $1;

-- name: GetCategoryByID :one
SELECT * FROM categories WHERE id = $1;

-- name: CreateCategory :one
INSERT INTO categories (slug, name, parent_id, icon, sort_order, is_active, requires_license, pricing_model, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: UpdateCategory :one
UPDATE categories
SET name = COALESCE(sqlc.narg('name'), name),
    icon = COALESCE(sqlc.narg('icon'), icon),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    sort_order = COALESCE(sqlc.narg('sort_order'), sort_order),
    metadata = COALESCE(sqlc.narg('metadata'), metadata)
WHERE id = $1
RETURNING *;

-- name: CountProvidersByCategory :many
SELECT c.id, c.slug, c.name, COUNT(pc.provider_id) as provider_count
FROM categories c
LEFT JOIN provider_categories pc ON pc.category_id = c.id
WHERE c.is_active = true
GROUP BY c.id, c.slug, c.name
ORDER BY provider_count DESC;
