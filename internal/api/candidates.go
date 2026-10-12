package api

import (
	"net/http"
	"strconv"

	"github.com/computersciencehouse/nominate/internal/util"
	"github.com/gin-gonic/gin"
)

type candidateMember struct {
	Username string
	Status   string // Valid values: Accepted, Declined, Pending
}

type candidateResult struct {
	ID      int32
	Members []candidateMember
	Overall string // Valid values: Accepted, Declined, Pending
}

// Checks the status of the candidate as a whole
// All members must accept for it to be accepted.
// If any member declines, the whole candidate is declined
func checkOverallStatus(members []candidateMember) string {
	allAccepted := true

	for _, m := range members {
		if m.Status == "Declined" {
			return "Declined"
		}
		if m.Status != "Accepted" {
			allAccepted = false
		}
	}

	if allAccepted {
		return "Accepted"
	}
	return "Pending"
}

// Page that shows each candidate for a given period, the members of the candidate, the decision status, and the deadline to respond
func (s *Server) GetPeriodResultsPage(c *gin.Context) {
	user := util.GetUserData(c)

	// Chair is the only one that should be able to see the list of candidates and their responses
	if !util.IsChair(s.Cfg, user) {
		c.HTML(http.StatusForbidden, "forbidden.tmpl", gin.H{
			"FullName":      user.FullName,
			"Username":      user.Username,
			"EBoard":        util.IsEboard(s.Cfg, user),
			"Chair":         util.IsChair(s.Cfg, user),
			"AllowedGroups": "eboard",
		})
		return
	}

	// Get the period id from the url
	periodId, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse period id to an integer"})
		return
	}

	ctx := c.Request.Context()

	period, err := s.Queries.GetPeriodById(ctx, int32(periodId))
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"FullName": user.FullName,
			"Username": user.Username,
			"EBoard":   util.IsEboard(s.Cfg, user),
			"Chair":    util.IsChair(s.Cfg, user),
			"Error":    err.Error(),
		})
		return
	}
	results, err := s.Queries.ListCandidateResultsForPeriod(ctx, int32(periodId))
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{
			"FullName": user.FullName,
			"Username": user.Username,
			"EBoard":   util.IsEboard(s.Cfg, user),
			"Chair":    util.IsChair(s.Cfg, user),
			"Error":    err.Error(),
		})
		return
	}

	var candidates []candidateResult
	for _, r := range results {
		if len(candidates) == 0 || candidates[len(candidates)-1].ID != r.CandidateID {
			candidates = append(candidates, candidateResult{ID: r.CandidateID})
		}
		last := &candidates[len(candidates)-1]
		last.Members = append(last.Members, candidateMember{
			Username: r.Username,
			Status:   r.DecisionStatus,
		})
	}

	// Determines and sets the decision status for each candidate based on the status of each member in the candidate
	for i := range candidates {
		candidates[i].Overall = checkOverallStatus(candidates[i].Members)
	}

	c.HTML(http.StatusOK, "results.tmpl", gin.H{
		"FullName":   user.FullName,
		"Username":   user.Username,
		"EBoard":     util.IsEboard(s.Cfg, user),
		"Chair":      util.IsChair(s.Cfg, user),
		"Period":     period,
		"Candidates": candidates,
	})

}

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
