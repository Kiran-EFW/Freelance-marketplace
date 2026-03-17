-- name: GetJurisdiction :one
SELECT * FROM jurisdictions WHERE id = $1;

-- name: ListActiveJurisdictions :many
SELECT * FROM jurisdictions WHERE is_active = true ORDER BY name;

-- name: UpdateJurisdictionConfig :exec
UPDATE jurisdictions SET config = $2 WHERE id = $1;
