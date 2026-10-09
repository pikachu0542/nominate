package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/computersciencehouse/nominate/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type submitNominationRequest struct {
	PeriodID  int32    `json:"period_id" binding:"required"`
	Usernames []string `json:"usernames" binding:"required,min=1,dive,required"`
}

func (s *Server) createNomination(c *gin.Context) {
	// user := util.GetUserData(c)

	var req submitNominationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	period, err := s.Queries.GetNominationPeriod(ctx, req.PeriodID)
	if err != nil {
		// Unable to find specified period
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Nomination period not found"})
			return
		}

		// Other error
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{
			"Error": err.Error(),
		})
		return
	}

	now := time.Now()
	if now.Before(period.OpensAt) || now.After(period.ClosesAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nomination period is not open"})
		return
	}

	// Start transaction
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback(ctx)

	qtx := s.Queries.WithTx(tx)
	nomination, err := qtx.CreateNomination(ctx, db.CreateNominationParams{
		PositionID: period.PositionID,
		PeriodID:   period.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create nomination"})
		return
	}

	// Add each member of the nomination to the nominated_users table
	for _, username := range req.Usernames {
		if err := qtx.AddNominatedUser(ctx, db.AddNominatedUserParams{
			Username:     username,
			NominationID: nomination.ID,
		}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add nominated user"})
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit changes to the database"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"nomination_id": nomination.ID,
		"usernames":     req.Usernames,
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
