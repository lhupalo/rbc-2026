package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lhupalo/rbc-2026/internal/models"
)

func FraudScore(c *gin.Context) {
	var req models.FraudScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.FraudScoreResponse{
		Approved:   false,
		FraudScore: 0.0,
	})
}
