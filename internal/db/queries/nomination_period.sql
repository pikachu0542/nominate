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

-- name: ListOpenPeriodsWithPosition :many
SELECT np.id AS period_id, np.position_id, p.position_name, np.opens_at, np.closes_at
FROM nomination_period np
JOIN position p ON p.id = np.position_id
WHERE np.opens_at <= now() AND np.closes_at > now()
ORDER BY np.opens_at;
    
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