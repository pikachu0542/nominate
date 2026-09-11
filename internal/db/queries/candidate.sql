-- name: CreateCandidate :one
INSERT INTO candidate (period_id, position_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetCandidate :one
SELECT *
FROM candidate
WHERE id = $1;

-- name: ListCandidatesForPeriod :many
SELECT *
FROM candidate
WHERE period_id = $1
ORDER BY position_id ASC;

-- name: ListCandidatesWithMembers :many
SELECT c.id AS candidate_id, c.period_id, c.position_id, array_agg(cm.username) AS members
FROM candidate c
LEFT JOIN candidate_member cm ON c.id = cm.candidate_id
WHERE c.period_id = $1
GROUP BY c.id, c.period_id, c.position_id
ORDER BY c.position_id ASC;