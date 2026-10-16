-- name: CreateCandidateDecision :one
INSERT INTO candidate_decision (candidate_id, username, status, response_deadline)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetCandidateDecision :one
SELECT *
FROM candidate_decision
WHERE candidate_id = $1 AND username = $2;

-- name: CreatePendingDecisionsForPeriod :exec
INSERT INTO candidate_decision (candidate_id, username, status, response_deadline)
SELECT cm.candidate_id, cm.username, 'pending'::decision_status, sqlc.arg(response_deadline)::timestamptz
FROM candidate_member cm
JOIN candidate c ON c.id = cm.candidate_id
WHERE c.period_id = sqlc.arg(period_id)
ON CONFLICT (candidate_id, username) DO NOTHING;

-- name: MarkNotified :one
UPDATE candidate_decision
SET notified_at = CURRENT_TIMESTAMP
WHERE candidate_id = $1 AND username = $2
RETURNING *;

-- name: ListPendingDecisionsPastDeadline :many
SELECT * 
FROM candidate_decision
WHERE status = 'pending' AND response_deadline < now();

-- name: DeclineExpiredDecision :exec
UPDATE candidate_decision
SET status = 'declined', responded_at = now()
WHERE id = $1 AND status = 'pending';

-- name: RespondToDecision :one
UPDATE candidate_decision
SET status = $2, responded_at = now()
WHERE id = $1 
    AND username = $3 
    AND status = 'pending'
    AND response_deadline > now()
RETURNING *;

-- name: ListPendingDecisionsForUser :many
SELECT cd.id, cd.status, cd.notified_at, cd.responded_at, cd.response_deadline,
       p.position_name,
       (cd.status = 'pending' AND cd.response_deadline > now())::boolean AS can_respond,
       COALESCE((
           SELECT string_agg(cm.username, ', ' ORDER BY cm.username)
           FROM candidate_member cm
           WHERE cm.candidate_id = cd.candidate_id
             AND cm.username <> cd.username
       ), '')::text AS co_nominees
FROM candidate_decision cd
JOIN candidate c ON c.id = cd.candidate_id
JOIN position p ON p.id = c.position_id
WHERE cd.username = $1
ORDER BY cd.response_deadline DESC;