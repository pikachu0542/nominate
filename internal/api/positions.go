package api

import (
	"net/http"

	"github.com/computersciencehouse/nominate/internal/util"
	"github.com/gin-gonic/gin"
)

// Lists all of the defined positions in the database
func (s *Server) listPositions(c *gin.Context) {
	user := util.GetUserData(c)

	positions, err := s.Queries.ListPositions(c.Request.Context())
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{
			"Username": user.Username,
			"FullName": user.FullName,
			"EBoard":   util.IsEboard(s.Cfg, user),
			"Chair":    util.IsChair(s.Cfg, user),
			"Error":    err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, positions)
}

// Creates a new position in the database with the provided name
// Only E-Board members can access this endpoint
// TODO: Decide whether I actually want this to be accessible in the app
func (s *Server) createPosition(c *gin.Context) {

	user := util.GetUserData(c)

	if !util.IsEboard(s.Cfg, user) {
		c.HTML(http.StatusForbidden, "forbidden.tmpl", gin.H{
			"Username":      user.Username,
			"FullName":      user.FullName,
			"EBoard":        util.IsEboard(s.Cfg, user),
			"Chair":         util.IsChair(s.Cfg, user),
			"AllowedGroups": "eboard",
		})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	position, err := s.Queries.CreatePosition(c.Request.Context(), req.Name)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{
			"Username": user.Username,
			"FullName": user.FullName,
			"EBoard":   util.IsEboard(s.Cfg, user),
			"Chair":    util.IsChair(s.Cfg, user),
			"Error":    err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, position)
}
