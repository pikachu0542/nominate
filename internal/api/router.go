package api

import (
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
func (s *Server) RegisterRoutes(r *gin.Engine, csh cshAuth.CSHAuth) {

	// ---------------------------------------------------
	//			    CSH AUTH ROUTES
	// ---------------------------------------------------
	auth := r.Group("/auth")
	auth.GET("/login", csh.AuthRequest)
	auth.GET("/callback", csh.AuthCallback)
	auth.GET("/logout", csh.AuthLogout)

	r.GET("/", csh.AuthWrapper(s.listOpenPeriods))

	// Health check route
	r.GET("/health", csh.AuthWrapper(s.healthCheck))

	// ---------------------------------------------------
	//			    POSITION ROUTES
	// ---------------------------------------------------
	positions := r.Group("/position")
	positions.GET("/list", csh.AuthWrapper(s.GetPositionsPage))
	positions.POST("/create", csh.AuthWrapper(s.createPosition))

	// ---------------------------------------------------
	//			    NOMINATION PERIOD ROUTES
	// ---------------------------------------------------
	period := r.Group("/period")
	period.GET("/all", csh.AuthWrapper(s.listAllPeriods))
	period.GET("/:id", csh.AuthWrapper(s.getPeriodById))
	period.GET("/new", csh.AuthWrapper(s.GetCreatePeriodPage))
	period.POST("/new", csh.AuthWrapper(s.CreatePeriod))
	period.GET("/:id/nominations", csh.AuthWrapper(s.listNominationsForPeriod))
	period.GET("/:id/users", csh.AuthWrapper(s.listNominatedUsersForPeriod))
	period.GET("/:id/results", csh.AuthWrapper(s.GetPeriodResultsPage))

	// ---------------------------------------------------
	//			    NOMINATION ROUTES
	// ---------------------------------------------------
	nominate := r.Group("/nominate")
	nominate.GET("", csh.AuthWrapper(s.GetNominatePage))
	nominate.POST("", csh.AuthWrapper(s.PostNomination))

	// ---------------------------------------------------
	//			    DECISION ROUTES
	// ---------------------------------------------------
	decision := r.Group("/decision")
	decision.GET("/list", csh.AuthWrapper(s.listDecisionsForUser))
	decision.POST("/:id/respond", csh.AuthWrapper(s.respondToDecision))
}

// Health check route - used to ensure the server is still running
func (s *Server) healthCheck(c *gin.Context) {
	c.String(200, "ok")
}
