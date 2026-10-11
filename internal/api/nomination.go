package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/computersciencehouse/nominate/internal/db"
	"github.com/computersciencehouse/nominate/internal/util"
	"github.com/gin-gonic/gin"
)

// Input data for submitting a nomination
type submitNominationRequest struct {
	PeriodID  int32    `json:"period_id" binding:"required"`
	Usernames []string `json:"usernames" binding:"required,min=1,dive,required"`
}

// Creates a new nomination in the database
func (s *Server) PostNomination(c *gin.Context) {
	var req submitNominationRequest
	if err := c.ShouldBind(&req); err != nil {
		c.Redirect(http.StatusSeeOther, "/nominate?error=Invalid+submission")
		return
	}

	ctx := c.Request.Context()

	period, err := s.Queries.GetPeriod(ctx, req.PeriodID)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/nominate?error=Period+not+found")
		return
	}

	now := time.Now()
	if now.Before(period.OpensAt) || !now.Before(period.ClosesAt) {
		c.Redirect(http.StatusSeeOther, "/nominate?error=Period+is+not+open")
		return
	}

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/nominate?error=Failed+to+connect+to+database")
		return
	}
	defer tx.Rollback(ctx)

	qtx := s.Queries.WithTx(tx)

	nomination, err := qtx.CreateNomination(ctx, db.CreateNominationParams{
		PositionID: period.PositionID,
		PeriodID:   period.ID,
	})
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/nominate?error=Failed+to+create+period+in+database")
		return
	}

	for _, username := range req.Usernames {
		if err := qtx.AddNominatedUser(ctx, db.AddNominatedUserParams{
			Username:     username,
			NominationID: nomination.ID,
		}); err != nil {
			c.Redirect(http.StatusSeeOther, "/nominate?error=Failed+to+update+nominated+users")
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		c.Redirect(http.StatusSeeOther, "/nominate?error=Failed+to+commit+transaction+to+database")
		return
	}

	c.Redirect(http.StatusSeeOther, "/nominate?success=1")
}

// Displays the nomination submission page
func (s *Server) GetNominatePage(c *gin.Context) {
	user := util.GetUserData(c)

	periods, err := s.Queries.ListOpenPeriodsWithPosition(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to retrieve nomination periods: %v", err)
		return
	}

	members := s.OIDC.GetOIDCGroup(s.OIDC.FindOIDCGroupID("active"))

	c.HTML(http.StatusOK, "nominate.tmpl", gin.H{
		"Username": user.Username,
		"FullName": user.FullName,
		"Periods":  periods,
		"Members":  members,
		"Success":  c.Query("success") != "",
		"Error":    c.Query("error"),
	})
}

// Retreives a list of all nominations submitted in a given period
func (s *Server) listNominationsForPeriod(c *gin.Context) {
	periodID, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	nominations, err := s.Queries.ListNominationsForPeriod(c.Request.Context(), int32(periodID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list nominations"})
		return
	}
	c.JSON(http.StatusOK, nominations)
}

// Retreives a list of users that have been nominated in a given period
func (s *Server) listNominatedUsersForPeriod(c *gin.Context) {
	periodID, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	nominatedUsers, err := s.Queries.ListNominatedUsersForPeriod(c.Request.Context(), int32(periodID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list nominated users"})
		return
	}
	c.JSON(http.StatusOK, nominatedUsers)
}
