package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lhupalo/rbc-2026/internal/detector"
	"github.com/lhupalo/rbc-2026/internal/models"
)

type fraudScoreHandler struct {
	det *detector.Detector
}

func newFraudScoreHandler(det *detector.Detector) *fraudScoreHandler {
	return &fraudScoreHandler{det: det}
}

func (h *fraudScoreHandler) handle(c *gin.Context) {
	var req models.FraudScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	approved, score := h.det.Score(&req)
	c.JSON(http.StatusOK, models.FraudScoreResponse{
		Approved:   approved,
		FraudScore: score,
	})
}
