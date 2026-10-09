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

func (s *Server) listDecisionsForUser(c *gin.Context) {
	user := util.GetUserData(c)

	decisions, err := s.Queries.ListDecisionsForUser(c.Request.Context(), user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list decisions"})
		return
	}

	c.JSON(http.StatusOK, decisions)
}

type respondToDecisionRequest struct {
	Status string `json:"status" binding:"required,oneof=accepted declined"`
}

func (s *Server) respondToDecision(c *gin.Context) {
	user := util.GetUserData(c)

	decisionID, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req respondToDecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	decision, err := s.Queries.RespondToDecision(c.Request.Context(), db.RespondToDecisionParams{
		ID:       int32(decisionID),
		Status:   db.DecisionStatus(req.Status),
		Username: user.Username,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusConflict, gin.H{"error": "Invalid decision"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to respond to decision"})
		return
	}

	c.JSON(http.StatusOK, decision)
}
