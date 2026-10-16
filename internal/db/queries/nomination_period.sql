-- name: CreatePeriod :one
INSERT INTO nomination_period (position_id, opens_at, closes_at)
VALUES ($1, $2, $3)
RETURNING id, position_id, opens_at, closes_at;

-- name: GetPeriodById :one
SELECT np.id AS period_id, np.position_id, p.position_name,
       np.opens_at, np.closes_at, np.consolidated_at
FROM nomination_period np
JOIN position p ON p.id = np.position_id
WHERE np.id = $1;

-- name: ListOpenPeriods :many
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
    
-- name: ListAllPeriods :many
SELECT np.id AS period_id, np.position_id, p.position_name, np.opens_at, np.closes_at,
       (CASE
           WHEN now() < np.opens_at THEN 'upcoming'
           WHEN now() < np.closes_at THEN 'open'
           ELSE 'closed'
       END)::text AS status
FROM nomination_period np
JOIN position p ON p.id = np.position_id
ORDER BY np.opens_at DESC;

-- name: ListPeriodsPendingConsolidation :many
SELECT *
FROM nomination_period
WHERE consolidated_at IS NULL
AND closes_at < CURRENT_TIMESTAMP
ORDER BY opens_at ASC;

-- name: SetPeriodConsolidated :exec
UPDATE nomination_period
SET consolidated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: ListPositionsWithOpenPeriod :many
SELECT p.id, p.position_name, o.open_until
FROM position p
LEFT JOIN (
    SELECT position_id, max(closes_at) AS open_until
    FROM nomination_period
    WHERE opens_at <= now() AND closes_at > now()
    GROUP BY position_id
) o ON o.position_id = p.id
ORDER BY p.position_name;