-- name: AddNominatedUser :exec
INSERT INTO nominated_user (username, nomination_id)
VALUES ($1, $2);

-- name: ListUsersInNomination :many
SELECT *
FROM nominated_user
WHERE nomination_id = $1
ORDER BY username ASC;