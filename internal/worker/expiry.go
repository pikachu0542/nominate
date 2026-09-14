package worker

import (
	"context"
	"fmt"

	"github.com/computersciencehouse/nominate/internal/db"
)

// Iterates through all nominations that are past their deadline and where no response was provided.
// Mark each of these nominations as declined.
func RunExpiry(ctx context.Context, q *db.Queries) error {
	decisions, err := q.ListPendingDecisionsPastDeadline(ctx)
	if err != nil {
		return fmt.Errorf("Error attempting to list expired nominations without decisions: %w", err)
	}

	for _, decision := range decisions {
		if err := q.DeclineExpiredDecision(ctx, decision.ID); err != nil {
			fmt.Printf("Failed to auto-decline expired decision %d: %v\n", decision.ID, err)
			continue
		}
	}

	return nil
}
