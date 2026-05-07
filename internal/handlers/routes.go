package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lhupalo/rbc-2026/internal/detector"
)

// maxInFlight limita a concorrência interna por instância.
// Com GOMAXPROCS=1 e 0.475 vCPU, manter uma janela pequena evita que a fila
// interna cresça: requests excedentes recebem 503 imediatamente (fast-fail)
// em vez de esperar até o timeout do cliente.
const maxInFlight = 12

func Register(r *gin.Engine, det *detector.Detector) {
	sem := make(chan struct{}, maxInFlight)

	admissionControl := func(c *gin.Context) {
		select {
		case sem <- struct{}{}:
			defer func() { <-sem }()
			c.Next()
		default:
			c.AbortWithStatus(http.StatusServiceUnavailable)
		}
	}

	fs := newFraudScoreHandler(det)
	r.GET("/ready", Ready)
	r.POST("/fraud-score", admissionControl, fs.handle)
}
