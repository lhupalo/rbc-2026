package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/lhupalo/rbc-2026/internal/detector"
)

func Register(r *gin.Engine, det *detector.Detector) {
	fs := newFraudScoreHandler(det)
	r.GET("/ready", Ready)
	r.POST("/fraud-score", fs.handle)
}
