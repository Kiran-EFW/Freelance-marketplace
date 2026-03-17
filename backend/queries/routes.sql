-- name: CreateRoute :one
INSERT INTO routes (provider_id, name, description, postcodes, frequency, day_of_week, max_stops, price_per_stop, currency)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetRouteByID :one
SELECT * FROM routes WHERE id = $1;

-- name: ListRoutesByProvider :many
SELECT * FROM routes WHERE provider_id = $1 ORDER BY created_at DESC;

-- name: UpdateRoute :one
UPDATE routes SET name = $2, description = $3, is_active = $4 WHERE id = $1 RETURNING *;

-- name: DeleteRoute :exec
DELETE FROM routes WHERE id = $1;

-- name: CreateRouteStop :one
INSERT INTO route_stops (route_id, customer_id, address, location, stop_order, notes)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListStopsByRoute :many
SELECT * FROM route_stops WHERE route_id = $1 ORDER BY stop_order;

-- name: DeleteRouteStop :exec
DELETE FROM route_stops WHERE id = $1 AND route_id = $2;

-- name: UpdateRouteStopOrder :exec
UPDATE route_stops SET stop_order = $2 WHERE id = $1;
