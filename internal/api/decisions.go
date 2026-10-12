package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/computersciencehouse/nominate/internal/db"
	"github.com/computersciencehouse/nominate/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type respondToDecisionForm struct {
	Status string `form:"status" binding:"required,oneof=accepted declined"`
}

func (s *Server) listDecisionsForUser(c *gin.Context) {
	user := util.GetUserData(c)

	decisions, err := s.Queries.ListDecisionsForUser(c.Request.Context(), user.Username)
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

	c.HTML(http.StatusOK, "decision.tmpl", gin.H{
		"Username":  user.Username,
		"FullName":  user.FullName,
		"EBoard":    util.IsEboard(s.Cfg, user),
		"Chair":     util.IsChair(s.Cfg, user),
		"Groups":    user.Groups,
		"Decisions": decisions,
		"Success":   c.Query("success") != "",
		"Error":     c.Query("error"),
	})
}

func (s *Server) respondToDecision(c *gin.Context) {
	user := util.GetUserData(c)

	decisionID, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/decision/list?error=Failed+to+parse+decision+id+to+integer")
		return
	}

	var req respondToDecisionForm
	if err := c.ShouldBind(&req); err != nil {
		c.Redirect(http.StatusSeeOther, "/decision/list?error=Invalid+response")
		return
	}

	_, err = s.Queries.RespondToDecision(c.Request.Context(), db.RespondToDecisionParams{
		ID:       int32(decisionID),
		Status:   db.DecisionStatus(req.Status),
		Username: user.Username,
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Redirect(http.StatusSeeOther, "/decision/list?error=Nomination+has+expired")
			return
		}
		c.Redirect(http.StatusSeeOther, "/decision/list?error=Internal+server+error")
		return
	}

	c.Redirect(http.StatusSeeOther, "/decision/list?success=1")
}
