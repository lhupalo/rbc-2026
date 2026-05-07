package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lhupalo/rbc-2026/internal/detector"
	"github.com/lhupalo/rbc-2026/internal/handlers"
)

func main() {
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}

	log.Printf("carregando dataset de %s...", dataDir)
	det, err := detector.Load(dataDir)
	if err != nil {
		log.Fatalf("falha ao carregar detector: %v", err)
	}
	log.Println("dataset carregado, iniciando servidor")

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	handlers.Register(r, det)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: 500 * time.Millisecond,
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      1500 * time.Millisecond,
		IdleTimeout:       30 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("servidor encerrado: %v", err)
	}
}
