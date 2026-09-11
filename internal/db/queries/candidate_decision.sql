-- name: CreateCandidateDecision :one
INSERT INTO candidate_decision (candidate_id, username, status, response_deadline)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetCandidateDecision :one
SELECT *
FROM candidate_decision
WHERE candidate_id = $1 AND username = $2;

-- name: MarkNotified :one
UPDATE candidate_decision
SET notified_at = CURRENT_TIMESTAMP
WHERE candidate_id = $1 AND username = $2
RETURNING *;