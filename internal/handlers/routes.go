package handlers

import "github.com/gin-gonic/gin"

func Register(r *gin.Engine) {
	r.GET("/ready", Ready)
	r.POST("/fraud-score", FraudScore)
}
