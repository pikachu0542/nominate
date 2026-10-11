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

// The input data needed to open nomination period(s)
type createPeriodForm struct {
	PositionIDs []int32 `form:"position_ids" binding:"required,min=1"`
	ClosesAt    string  `form:"closes_at" binding:"required"`
}

// Displays the page for creating a nomination period
func (s *Server) GetCreatePeriodPage(c *gin.Context) {
	user := util.GetUserData(c)
	if !util.IsChair(s.Cfg, user) {
		c.String(http.StatusForbidden, "Only the chairperson can open nominations.")
		return
	}

	positions, err := s.Queries.ListPositions(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to load positions")
		return
	}

	c.HTML(http.StatusOK, "create_period.tmpl", gin.H{
		"Username":  user.Username,
		"FullName":  user.FullName,
		"EBoard":    true,
		"Groups":    user.Groups,
		"Positions": positions,
		"Success":   c.Query("success") != "",
		"Error":     c.Query("error"),
	})
}

// Create a new nomination period in the database
func (s *Server) CreatePeriod(c *gin.Context) {
	user := util.GetUserData(c)

	if !util.IsChair(s.Cfg, user) {
		c.String(http.StatusForbidden, "Only the chairperson can open nominations.")
		return
	}

	var req createPeriodForm
	if err := c.ShouldBind(&req); err != nil {
		c.Redirect(http.StatusSeeOther, "/period/new?error=Invalid+form+input")
		return
	}

	opensAt := time.Now()

	closesAt, err := time.ParseInLocation("2006-01-02T15:04", req.ClosesAt, time.Local)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/period/new?error=Invalid+closing+time")
		return
	}
	if !closesAt.After(opensAt) {
		c.Redirect(http.StatusSeeOther, "/period/new?error=Close+time+must+be+after+open+time")
		return
	}

	ctx := c.Request.Context()

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/period/new?error=Internal+server+error+occurred")
		return
	}
	defer tx.Rollback(ctx)
	qtx := s.Queries.WithTx(tx)

	for _, positionID := range req.PositionIDs {
		if _, err := qtx.GetPosition(ctx, positionID); err != nil {
			c.Redirect(http.StatusSeeOther, "/period/new?error=Invalid+position")
			return
		}
		if _, err := qtx.CreatePeriod(ctx, db.CreatePeriodParams{
			PositionID: positionID,
			OpensAt:    opensAt,
			ClosesAt:   closesAt,
		}); err != nil {
			c.Redirect(http.StatusSeeOther, "/period/new?error=Failed+to+create+nomination+period")
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		c.Redirect(http.StatusSeeOther, "/period/new?error=Failed+to+commit+transaction+to+database")
		return
	}

	c.Redirect(http.StatusSeeOther, "/period/new?success=1")
}

// Retreives a specific nomination period using its unique id
func (s *Server) getPeriodById(c *gin.Context) {
	// user := util.GetUserData(c)

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	period, err := s.Queries.GetPeriod(c.Request.Context(), int32(id))
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
	user := util.GetUserData(c)

	periods, err := s.Queries.ListOpenPeriodsWithPosition(c.Request.Context())
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{
			"Error": err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "periods.tmpl", gin.H{
		"Username":   user.Username,
		"FullName":   user.FullName,
		"EBoard":     util.IsEboard(s.Cfg, user),
		"Groups":     user.Groups,
		"Title":      "Open Nomination Periods",
		"Periods":    periods,
		"ShowStatus": false,
	})
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

	periods, err := s.Queries.ListAllPeriods(c.Request.Context())
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{
			"Error": err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "periods.tmpl", gin.H{
		"Username":   user.Username,
		"FullName":   user.FullName,
		"EBoard":     true,
		"Groups":     user.Groups,
		"Title":      "All Nomination Periods",
		"Periods":    periods,
		"ShowStatus": true,
	})
}
