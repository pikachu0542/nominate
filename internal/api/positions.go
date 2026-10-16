package api

import (
	"net/http"
	"strings"

	"github.com/computersciencehouse/nominate/internal/util"
	"github.com/gin-gonic/gin"
)

type createPositionForm struct {
	PositionName string `form:"position-name" binding:"required"`
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

	var req createPositionForm
	if err := c.ShouldBind(&req); err != nil {
		c.Redirect(http.StatusSeeOther, "/position/list?error=Enter+a+position+name")
		return
	}

	name := strings.TrimSpace(req.PositionName)
	if name == "" {
		c.Redirect(http.StatusSeeOther, "/position/list?error=Enter+a+position+name")
		return
	}

	if _, err := s.Queries.CreatePosition(c.Request.Context(), name); err != nil {
		c.Redirect(http.StatusSeeOther, "/position/list?error=Failed+to+add+position")
		return
	}

	c.Redirect(http.StatusSeeOther, "/position/list?success=1")
}

func (s *Server) GetPositionsPage(c *gin.Context) {
	user := util.GetUserData(c)

	if !util.IsChair(s.Cfg, user) {
		c.HTML(http.StatusForbidden, "forbidden.tmpl", gin.H{
			"Username":     user.Username,
			"FullName":     user.FullName,
			"EBoard":       util.IsEboard(s.Cfg, user),
			"Chair":        util.IsChair(s.Cfg, user),
			"AllowedGroup": "eboard-chairman",
		})
		return
	}

	ctx := c.Request.Context()

	positions, err := s.Queries.ListPositionsWithOpenPeriod(ctx)
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

	c.HTML(http.StatusOK, "positions.tmpl", gin.H{
		"Username":  user.Username,
		"FullName":  user.FullName,
		"EBoard":    util.IsEboard(s.Cfg, user),
		"Chair":     util.IsChair(s.Cfg, user),
		"Positions": positions,
		"Success":   c.Query("success") != "",
		"Error":     c.Query("error"),
	})

}
