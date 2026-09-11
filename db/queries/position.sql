-- name: ListPositions :many
SELECT position_name 
FROM position
ORDER BY id ASC;

-- name: GetPosition :one
SELECT position_name
FROM position
WHERE id = $1;

-- name: CreatePosition :one
INSERT INTO position (position_name)
VALUES ($1)
RETURNING id, position_name;

-- name: UpdatePosition :one
UPDATE position
SET position_name = $2
WHERE id = $1
RETURNING id, position_name;

-- name: DeletePosition :one
DELETE FROM position
WHERE id = $1
RETURNING id, position_name;