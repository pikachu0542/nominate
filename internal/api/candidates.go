package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (s *Server) listCandidatesForPeriod(c *gin.Context) {
	periodID, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	candidates, err := s.Queries.ListCandidatesWithMembers(c.Request.Context(), int32(periodID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list candidates"})
		return
	}

	c.JSON(http.StatusOK, candidates)
}
