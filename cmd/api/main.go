package main

import (
	"github.com/gin-gonic/gin"
	"github.com/lhupalo/rbc-2026/internal/handlers"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	handlers.Register(r)
	_ = r.Run(":8080")
}
