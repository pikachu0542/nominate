package api

import (
	"net/http"
	"time"

	"github.com/computersciencehouse/nominate/internal/db"
	"github.com/computersciencehouse/nominate/internal/util"
	"github.com/gin-gonic/gin"
)

func (s *Server) PostNominate(c *gin.Context) {
	var req submitNominationRequest
	if err := c.ShouldBind(&req); err != nil {
		c.Redirect(http.StatusSeeOther, "/nominate?error=invalid+submission")
		return
	}

	ctx := c.Request.Context()

	period, err := s.Queries.GetNominationPeriod(ctx, req.PeriodID)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/nominate?error=period+not+found")
		return
	}

	now := time.Now()
	if now.Before(period.OpensAt) || !now.Before(period.ClosesAt) {
		c.Redirect(http.StatusSeeOther, "/nominate?error=period+is+not+open")
		return
	}

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/nominate?error=something+went+wrong")
		return
	}
	defer tx.Rollback(ctx)

	qtx := s.Queries.WithTx(tx)

	nomination, err := qtx.CreateNomination(ctx, db.CreateNominationParams{
		PositionID: period.PositionID,
		PeriodID:   period.ID,
	})
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/nominate?error=something+went+wrong")
		return
	}

	for _, username := range req.Usernames {
		if err := qtx.AddNominatedUser(ctx, db.AddNominatedUserParams{
			Username:     username,
			NominationID: nomination.ID,
		}); err != nil {
			c.Redirect(http.StatusSeeOther, "/nominate?error=something+went+wrong")
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		c.Redirect(http.StatusSeeOther, "/nominate?error=something+went+wrong")
		return
	}

	c.Redirect(http.StatusSeeOther, "/nominate?success=1")
}

func (s *Server) GetNominatePage(c *gin.Context) {
	periods, err := s.Queries.ListOpenPeriodsWithPosition(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to retrieve nomination periods: %v", err)
		return
	}

	members := s.OIDC.GetOIDCGroup(s.OIDC.FindOIDCGroupID("active"))
	user := util.GetUserData(c)

	c.HTML(http.StatusOK, "nominate.tmpl", gin.H{
		"Username": user.Username,
		"Periods":  periods,
		"Members":  members,
		"Success":  c.Query("success") != "",
		"Error":    c.Query("error"),
	})
}
