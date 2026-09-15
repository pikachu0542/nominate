package worker

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/computersciencehouse/nominate/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Runs the consolidation process on nominations submitted for a particular period
func RunConsolidation(ctx context.Context, pool *pgxpool.Pool, q *db.Queries) error {
	periods, err := q.ListPeriodsPendingConsolidation(ctx)
	if err != nil {
		return fmt.Errorf("Error while listing periods pending consolidation: %w", err)
	}

	for _, period := range periods {
		if err := consolidatePeriod(ctx, pool, q, period.ID, period.PositionID); err != nil {
			fmt.Printf("Consolidation failed for period %d: %v\n", period.ID, err)
			continue
		}
	}

	return nil
}

// Checks all submitted nominations for a given period, and stores all unique submissions in the database as candidates
// This operation is performed in a transaction, which ensures that if an error occurs before all members of a candidate
// are processed, any already completed operations will be rolled back. This means there should never be partially populated candidates
func consolidatePeriod(ctx context.Context, pool *pgxpool.Pool, q *db.Queries, periodID, positionID int32) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("Error while beginning transaction: %w", err)
	}
	// If an error occurred before adding all members to a candidate, rollback the transaction
	defer tx.Rollback(ctx)

	qtx := q.WithTx(tx)

	rows, err := qtx.GetNominatedUsersForPeriod(ctx, periodID)
	if err != nil {
		return fmt.Errorf("Error while fetching nominated users for period %d: %w", periodID, err)
	}

	if len(rows) == 0 {
		if err := qtx.SetPeriodConsolidated(ctx, periodID); err != nil {
			return fmt.Errorf("Error marking period %d as consolidated: %w", periodID, err)
		}
		// Assuming all operations in the transaction were successful, commit to the database to make them permanent
		return tx.Commit(ctx)
	}

	membersByNomination := map[int32][]string{}

	for _, r := range rows {
		membersByNomination[r.NominationID] = append(membersByNomination[r.NominationID], r.Username)
	}

	type groupKey string
	groups := map[groupKey][]string{}

	for _, members := range membersByNomination {
		sorted := append([]string(nil), members...)
		sort.Strings(sorted)
		key := groupKey(strings.Join(sorted, ","))
		if _, exists := groups[key]; !exists {
			groups[key] = sorted
		}
	}

	// Create candidates in the database as necessary
	for _, members := range groups {
		candidate, err := q.CreateCandidate(ctx, db.CreateCandidateParams{
			PeriodID:   periodID,
			PositionID: positionID,
		})
		if err != nil {
			return fmt.Errorf("Error trying to create a candidate for period %d: %w", periodID, err)
		}

		// Add the users to candidates as necessary
		for _, username := range members {
			if err := q.AddMemberToCandidate(ctx, db.AddMemberToCandidateParams{
				CandidateID: candidate.ID,
				Username:    username,
			}); err != nil {
				return fmt.Errorf("Error while adding user %s to candidate %d: %w", username, candidate.ID, err)
			}
		}
	}

	if err := qtx.SetPeriodConsolidated(ctx, periodID); err != nil {
		return fmt.Errorf("Error while marking period %d as consolidated: %w", periodID, err)
	}

	// Commit the transaction to the database, making the changes permanent
	// This only runs if all operations within the transaction were successful
	return tx.Commit(ctx)
}
