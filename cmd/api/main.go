package main

import (
	"github.com/gin-gonic/gin"
	"github.com/lhupalo/rbc-2026/internal/handlers"
)

func main() {
	r := gin.Default()
	handlers.Register(r)
	_ = r.Run(":8080")
}
