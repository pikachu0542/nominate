-- name: CreateNominationPeriod :one
INSERT INTO nomination_period (position_id, opens_at, closes_at)
VALUES ($1, $2, $3)
RETURNING id, position_id, opens_at, closes_at;

-- name: GetNominationPeriod :one
SELECT *
FROM nomination_period
WHERE id = $1;

-- name: ListOpenNominationPeriods :many
SELECT *
FROM nomination_period
WHERE consolidated_at IS NULL
ORDER BY opens_at ASC;

-- name: ListAllNominationPeriods :many
SELECT *
FROM nomination_period
ORDER BY opens_at ASC;

-- name: ListPeriodsPendingConsolidation :many
SELECT *
FROM nomination_period
WHERE consolidated_at IS NULL
ORDER BY opens_at ASC;

-- name: SetPeriodConsolidated :exec
UPDATE nomination_period
SET consolidated_at = CURRENT_TIMESTAMP
WHERE id = $1;