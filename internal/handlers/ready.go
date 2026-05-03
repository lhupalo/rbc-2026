package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Ready(c *gin.Context) {
	c.Status(http.StatusOK)
}
