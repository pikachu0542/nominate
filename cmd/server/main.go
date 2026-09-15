package main

import (
	"context"
	"log"

	"github.com/computersciencehouse/nominate/internal/config"
	"github.com/computersciencehouse/nominate/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	cshAuth "github.com/computersciencehouse/csh-auth"
)

func main() {

	godotenv.Load()

	ctx := context.Background()

	// Load env vars
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

	log.Println("Value of NOMINATE_HOST:" + cfg.Auth.Host)

	// Initialize pool of database connections
	pool, err := db.NewPool(ctx, cfg.DBString)
	if err != nil {
		log.Fatalf("Error attempting to initialize connection pool: %v", err)
	}
	defer pool.Close()

	csh := cshAuth.CSHAuth{}

	// Initialize CSH auth
	csh.Init(
		cfg.Auth.ClientID,
		cfg.Auth.ClientSecret,
		cfg.Auth.JWTSecret,
		cfg.Auth.State,
		cfg.Auth.Host,
		cfg.Auth.Host+"/auth/callback",
		cfg.Auth.Host+"/auth/login",
		[]string{"profile", "email", "groups"},
	)

	if err != nil {
		log.Panicf("Error initializing CSH auth: %v", err)
	}

	queries := db.New(pool)
	_ = queries

	router := gin.Default()

	// Routes to handle CSH auth
	router.GET("/auth/login", csh.AuthRequest)
	router.GET("/auth/callback", csh.AuthCallback)
	router.GET("/auth/logout", csh.AuthLogout)

	// Health check route - basically just for ensuring the server is still up
	router.GET("/health", csh.AuthWrapper(func(c *gin.Context) {
		c.String(200, "Ok")
	}))

	// Start the server
	router.Run(":" + cfg.Port)
	log.Printf("Listening on port %s", cfg.Port)
}
