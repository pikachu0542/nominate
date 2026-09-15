package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/computersciencehouse/nominate/internal/config"
	"github.com/computersciencehouse/nominate/internal/db"
	"github.com/computersciencehouse/nominate/internal/worker"
	"github.com/joho/godotenv"
)

func main() {

	godotenv.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load environment variables: %v", err)
	}

	pool, err := db.NewPool(ctx, cfg.DBString)
	if err != nil {
		log.Fatalf("Database error: %v", err)
	}
	defer pool.Close()

	queries := db.New(pool)

	exitCode := 0

	if err := worker.RunConsolidation(ctx, pool, queries); err != nil {
		log.Printf("Error while running the consolidation process: %v", err)
		exitCode = 1
	}

	os.Exit(exitCode)
}
