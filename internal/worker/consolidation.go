package worker

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/computersciencehouse/nominate/internal/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

// How long a nominated member has to accept or decline
const responseWindow = 5 * 24 * time.Hour

// RunConsolidation turns the raw nominations of every closed period into
// candidates, then creates a pending decision for each candidate member.
func RunConsolidation(ctx context.Context, pool *pgxpool.Pool, q *db.Queries) error {
	periods, err := q.ListPeriodsPendingConsolidation(ctx)
	if err != nil {
		return fmt.Errorf("listing periods pending consolidation: %w", err)
	}

	for _, period := range periods {
		if err := consolidatePeriod(ctx, pool, q, period.ID, period.PositionID); err != nil {
			log.Printf("consolidation failed for period %d: %v", period.ID, err)
			continue
		}
		log.Printf("consolidated period %d", period.ID)
	}

	return nil
}

// consolidatePeriod does all of its writes in one transaction, so a failure
// part-way through leaves the period untouched and it is retried next run.
func consolidatePeriod(ctx context.Context, pool *pgxpool.Pool, q *db.Queries, periodID, positionID int32) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := q.WithTx(tx)

	rows, err := qtx.ListNominatedUsersForPeriod(ctx, periodID)
	if err != nil {
		return fmt.Errorf("fetching nominated users: %w", err)
	}

	// Nobody was nominated: mark the period done so it isn't re-checked forever
	if len(rows) == 0 {
		if err := qtx.SetPeriodConsolidated(ctx, periodID); err != nil {
			return fmt.Errorf("marking empty period consolidated: %w", err)
		}
		return tx.Commit(ctx)
	}

	// Collect the usernames named in each individual nomination
	membersByNomination := map[int32][]string{}
	for _, r := range rows {
		membersByNomination[r.NominationID] = append(membersByNomination[r.NominationID], r.Username)
	}

	// Collapse nominations naming the same set of people into one group
	groups := map[string][]string{}
	for _, members := range membersByNomination {
		sorted := append([]string(nil), members...)
		sort.Strings(sorted)
		key := strings.Join(sorted, ",")
		if _, exists := groups[key]; !exists {
			groups[key] = sorted
		}
	}

	// One candidate per distinct group, with a row per member
	for _, members := range groups {
		candidate, err := qtx.CreateCandidate(ctx, db.CreateCandidateParams{
			PeriodID:   periodID,
			PositionID: positionID,
		})
		if err != nil {
			return fmt.Errorf("creating candidate: %w", err)
		}

		for _, username := range members {
			if err := qtx.AddMemberToCandidate(ctx, db.AddMemberToCandidateParams{
				CandidateID: candidate.ID,
				Username:    username,
			}); err != nil {
				return fmt.Errorf("adding candidate member %s: %w", username, err)
			}
		}
	}

	// A pending decision for every member of every candidate in this period
	if err := qtx.CreatePendingDecisionsForPeriod(ctx, db.CreatePendingDecisionsForPeriodParams{
		PeriodID:         periodID,
		ResponseDeadline: time.Now().Add(responseWindow),
	}); err != nil {
		return fmt.Errorf("creating pending decisions: %w", err)
	}

	if err := qtx.SetPeriodConsolidated(ctx, periodID); err != nil {
		return fmt.Errorf("marking period consolidated: %w", err)
	}

	return tx.Commit(ctx)
}
