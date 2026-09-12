-- name: CreateNomination :one
INSERT INTO nomination (period_id, position_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetNominationsForPeriod :many
SELECT *
FROM nomination
WHERE period_id = $1
ORDER BY position_id ASC;

-- name: GetNominatedUsersForPeriod :many
SELECT n.id AS nomination_id, n.position_id, nu.username
FROM nomination n
JOIN nominated_user nu ON nu.nomination_id = n.id
WHERE n.period_id = $1
ORDER BY n.id;