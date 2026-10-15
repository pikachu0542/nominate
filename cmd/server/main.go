package main

import (
	"context"
	"log"
	"net/http"
	"text/template"

	"github.com/computersciencehouse/nominate/internal/api"
	"github.com/computersciencehouse/nominate/internal/config"
	"github.com/computersciencehouse/nominate/internal/db"
	"github.com/computersciencehouse/nominate/internal/util"
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

	oidcClient := util.NewOIDCClient(cfg.Auth.ClientID, cfg.Auth.ClientSecret)

	queries := db.New(pool)

	server := api.NewServer(queries, cfg, pool, oidcClient)

	r := gin.Default()

	r.SetFuncMap(template.FuncMap{
		"timeLeft": util.TimeLeft,
	})
	r.StaticFS("/static", http.Dir("static"))
	r.LoadHTMLGlob("templates/*")

	server.RegisterRoutes(r, csh)

	// Start the server
	r.Run(":" + cfg.Port)
	log.Printf("Listening on port %s", cfg.Port)
}
