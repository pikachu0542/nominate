package worker

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/computersciencehouse/nominate/internal/db"
)

// Runs the consolidation process on nominations submitted for a particular period
func RunConsolidation(ctx context.Context, q *db.Queries) error {
	periods, err := q.ListPeriodsPendingConsolidation(ctx)
	if err != nil {
		return fmt.Errorf("listing periods pending consolidation: %w", err)
	}

	for _, period := range periods {
		if err := consolidatePeriod(ctx, q, period.ID, period.PositionID); err != nil {
			fmt.Printf("consolidation failed for period %d: %v\n", period.ID, err)
			continue
		}
	}

	return nil
}

// Checks all submitted nominations for a given period, and stores all unique submissions in the database as candidates
func consolidatePeriod(ctx context.Context, q *db.Queries, periodID, positionID int32) error {
	rows, err := q.GetNominatedUsersForPeriod(ctx, periodID)
	if err != nil {
		return fmt.Errorf("fetching nominated users: %w", err)
	}

	if len(rows) == 0 {
		return q.SetPeriodConsolidated(ctx, periodID)
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
			return fmt.Errorf("Error while creating candidate: %w", err)
		}

		// Add the users to candidates as necessary
		for _, username := range members {
			if err := q.AddMemberToCandidate(ctx, db.AddMemberToCandidateParams{
				CandidateID: candidate.ID,
				Username:    username,
			}); err != nil {
				return fmt.Errorf("Error while adding candidate member %s: %w", username, err)
			}
		}
	}

	return q.SetPeriodConsolidated(ctx, periodID)
}
