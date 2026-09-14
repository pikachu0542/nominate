package main

import (
	"context"
	"log"

	"github.com/computersciencehouse/nominate/internal/config"
	"github.com/computersciencehouse/nominate/internal/db"
	"github.com/gin-gonic/gin"
)

func main() {
	ctx := context.Background()

	// Load env vars
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

	// Initialize pool of database connections
	pool, err := db.NewPool(ctx, cfg.DBString)
	if err != nil {
		log.Fatalf("Error attempting to initialize connection pool: %v", err)
	}
	defer pool.Close()

	queries := db.New(pool)

	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		c.String(200, "Ok")
	})

	// Start the server
	log.Printf("Listening on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Internal server error: %v", err)
	}
}
