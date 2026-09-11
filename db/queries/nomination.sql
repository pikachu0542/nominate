-- name: CreateNomination :one
INSERT INTO nomination (period_id, position_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetNominationsForPeriod :many
SELECT *
FROM nomination
WHERE period_id = $1
ORDER BY position_id ASC;