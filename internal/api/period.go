package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/computersciencehouse/nominate/internal/db"
	"github.com/computersciencehouse/nominate/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func (s *Server) createPeriod(c *gin.Context) {
	user := util.GetUserData(c)

	if !util.IsChair(s.Cfg, user) {
		c.HTML(http.StatusForbidden, "forbidden.tmpl", gin.H{
			"AllowedGroups": "eboard-chairman",
		})
		return
	}

	var req struct {
		PositionID int32     `json:"position_id" binding:"required"`
		OpensAt    time.Time `json:"opens_at" binding:"required"`
		ClosesAt   time.Time `json:"closes_at" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate open and close date
	if !req.ClosesAt.After(req.OpensAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "closes_at must be after opens_at"})
		return
	}

	// TODO: ensure close date isnt in the past

	ctx := c.Request.Context()

	_, err := s.Queries.GetPosition(ctx, req.PositionID)
	if err != nil {
		// The position doesnt exist
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "position not found"})
			return
		}

		// Some other error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to look up specified position"})
		return
	}

	period, err := s.Queries.CreateNominationPeriod(ctx, db.CreateNominationPeriodParams{
		PositionID: req.PositionID,
		OpensAt:    req.OpensAt,
		ClosesAt:   req.ClosesAt,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create nomination period"})
		return
	}

	c.JSON(http.StatusCreated, period)
}

// func (s *Server) getPeriodById(c *gin.Context) {
// 	user := util.GetUserData(c)

// }

// func (s *Server) listOpenPeriods(c *gin.Context) {
// 	user := util.GetUserData(c)

// }

// Retreives a list of every existing nomination period
func (s *Server) listAllPeriods(c *gin.Context) {
	user := util.GetUserData(c)

	if !util.IsEboard(s.Cfg, user) {
		// TODO: Create a forbidden template page and replace the JSON with that
		c.HTML(http.StatusForbidden, "forbidden.tmpl", gin.H{
			"AllowedGroups": "eboard",
		})
		return
	}

	periods, err := s.Queries.ListAllNominationPeriods(c.Request.Context())
	if err != nil {
		// TODO: Create a page for internal server error and replace the JSON with it
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list all nomination periods"})
		return
	}

	// TODO: Create a history page that lists all periods and display that instead
	c.JSON(http.StatusOK, periods)
}

// func (s *Server) listPeriodsPendingConsolidation(c *gin.Context) {
// 	user := util.GetUserData(c)

// }

// func (s *Server) setPeriodConsolidated(c *gin.Context) {
// 	user := util.GetUserData(c)

// }
