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
func (s *Server) RegisterRoutes(r *gin.Engine, csh cshAuth.CSHAuth) {

	// ---------------------------------------------------
	//			    CSH AUTH ROUTES
	// ---------------------------------------------------
	r.GET("/auth/login", csh.AuthRequest)
	r.GET("/auth/callback", csh.AuthCallback)
	r.GET("/auth/logout", csh.AuthLogout)

	r.GET("/", csh.AuthWrapper(s.GetHomepage))

	// Health check route
	r.GET("/health", csh.AuthWrapper(s.healthCheck))

	// ---------------------------------------------------
	//			    POSITION ROUTES
	// ---------------------------------------------------
	r.GET("/position/list", csh.AuthWrapper(s.listPositions))
	r.POST("/position/create", csh.AuthWrapper(s.createPosition))

	// ---------------------------------------------------
	//			    NOMINATION PERIOD ROUTES
	// ---------------------------------------------------
	r.GET("/period/open", csh.AuthWrapper(s.listOpenPeriods))
	r.GET("/period/all", csh.AuthWrapper(s.listAllPeriods))
	r.GET("/period/:id", csh.AuthWrapper(s.getPeriodById))
	r.GET("/period/new", csh.AuthWrapper(s.GetCreatePeriodPage))
	r.POST("/period/new", csh.AuthWrapper(s.CreatePeriod))
	r.GET("/period/:id/nominations", csh.AuthWrapper(s.listNominationsForPeriod))
	r.GET("/period/:id/users", csh.AuthWrapper(s.listNominatedUsersForPeriod))

	// ---------------------------------------------------
	//			    NOMINATION ROUTES
	// ---------------------------------------------------
	r.GET("/nominate", csh.AuthWrapper(s.GetNominatePage))
	r.POST("/nominate", csh.AuthWrapper(s.PostNomination))

	// ---------------------------------------------------
	//			    DECISION ROUTES
	// ---------------------------------------------------
	r.GET("/decisions", csh.AuthWrapper(s.listDecisionsForUser))
	r.POST("/decision/:id/respond", csh.AuthWrapper(s.respondToDecision))
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
