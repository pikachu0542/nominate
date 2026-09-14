package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Opens a pool of connections to the database. Pings the database to ensure connections
// were successfully established. Returns the pool if successful, otherwise returns an error.
func NewPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("Error while creating pgx pool: %w", err)
	}

	// Ensure that we are able to actually connect to the database
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("Error while pinging database: %w", err)
	}
	return pool, nil
}
