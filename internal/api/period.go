package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/computersciencehouse/nominate/internal/db"
	"github.com/computersciencehouse/nominate/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// Create a new nomination period
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
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{
			"Error": err.Error(),
		})
		return
	}

	period, err := s.Queries.CreateNominationPeriod(ctx, db.CreateNominationPeriodParams{
		PositionID: req.PositionID,
		OpensAt:    req.OpensAt,
		ClosesAt:   req.ClosesAt,
	})

	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{
			"Error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, period)
}

func (s *Server) getPeriodById(c *gin.Context) {
	// user := util.GetUserData(c)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	period, err := s.Queries.GetNominationPeriod(c.Request.Context(), int32(id))
	if err != nil {
		// No period with the specified id
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "No period found with that id"})
			return
		}

		// Any other error
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{
			"Error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, period)

}

// Lists all periods that are currently accepting nominations
func (s *Server) listOpenPeriods(c *gin.Context) {
	// user := util.GetUserData(c)

	periods, err := s.Queries.ListOpenNominationPeriods(c.Request.Context())
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{
			"Error": err.Error(),
		})
	}
	c.JSON(http.StatusOK, periods)

}

// Retreives a list of every existing nomination period
func (s *Server) listAllPeriods(c *gin.Context) {
	user := util.GetUserData(c)

	if !util.IsEboard(s.Cfg, user) {
		c.HTML(http.StatusForbidden, "forbidden.tmpl", gin.H{
			"AllowedGroups": "eboard",
		})
		return
	}

	periods, err := s.Queries.ListAllNominationPeriods(c.Request.Context())
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{
			"Error": err.Error(),
		})
		return
	}

	// TODO: Create a history page that lists all periods and display that instead
	c.JSON(http.StatusOK, periods)
}
