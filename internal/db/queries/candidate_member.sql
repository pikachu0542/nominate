-- name: AddMemberToCandidate :one
INSERT INTO candidate_member (candidate_id, username)
VALUES ($1, $2)
RETURNING *;

-- name: ListMembersOfCandidate :many
SELECT *
FROM candidate_member
WHERE candidate_id = $1
ORDER BY username ASC;