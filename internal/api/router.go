package api

import (
	"net/http"

	cshAuth "github.com/computersciencehouse/csh-auth"
	"github.com/computersciencehouse/nominate/internal/config"
	"github.com/computersciencehouse/nominate/internal/db"
	"github.com/computersciencehouse/nominate/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	Queries *db.Queries
	Cfg     *config.Config
	Pool    *pgxpool.Pool
	OIDC    *util.OIDCClient
}

// Initialize a new server, and pass an object containing all the database queries for use by the server
func NewServer(queries *db.Queries, cfg *config.Config, pool *pgxpool.Pool, oidc *util.OIDCClient) *Server {
	return &Server{
		Queries: queries,
		Cfg:     cfg,
		Pool:    pool,
		OIDC:    oidc,
	}
}

// Register API routes to the server
func (s *Server) RegisterRoutes(router *gin.Engine, csh cshAuth.CSHAuth) {

	// Routes to handle CSH auth
	auth := router.Group("/auth")
	auth.GET("/login", csh.AuthRequest)
	auth.GET("/callback", csh.AuthCallback)
	auth.GET("/logout", csh.AuthLogout)

	position := router.Group("/positions")
	position.GET("", csh.AuthWrapper(s.listPositions))
	position.POST("", csh.AuthWrapper(s.createPosition))

	period := router.Group("/periods")
	period.GET("", csh.AuthWrapper(s.listOpenPeriods))
	period.GET("/all", csh.AuthWrapper(s.listAllPeriods))
	period.GET("/:id", csh.AuthWrapper(s.getPeriodById))
	period.POST("/create", csh.AuthWrapper(s.createPeriod))
	period.GET("/:id/nominations", s.listNominationsForPeriod)
	period.GET("/:id/nominated_users", s.listNominatedUsersForPeriod)

	nomination := router.Group("/nomination")
	nomination.POST("", s.createNomination)

	decision := router.Group("/decision")
	decision.GET("", csh.AuthWrapper(s.listDecisionsForUser))
	decision.POST("/:id/respond", csh.AuthWrapper(s.respondToDecision))

	router.GET("/", csh.AuthWrapper(s.GetHomepage))

	router.GET("/nominate", csh.AuthWrapper(s.GetNominatePage))
	router.POST("/nominate", csh.AuthWrapper(s.PostNominate))

	// Health check
	router.GET("/health", csh.AuthWrapper(s.healthCheck))
}

// Health check route - used to ensure the server is still running
func (s *Server) healthCheck(c *gin.Context) {
	c.String(200, "ok")
}

func (s *Server) GetHomepage(c *gin.Context) {
	user := util.GetUserData(c)

	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"Username": user.Username,
		"FullName": user.FullName,
		"EBoard":   util.IsEboard(s.Cfg, user),
		"Groups":   user.Groups,
	})
}
