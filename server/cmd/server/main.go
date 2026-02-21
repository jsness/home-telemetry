//go:generate swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
// @title Home Telemetry API
// @version 0.1.0
// @description Ingest and query home telemetry metrics.
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"home-telemetry/server/internal/api"
	"home-telemetry/server/internal/config"
	"home-telemetry/server/internal/db"
	"home-telemetry/server/internal/store"

	_ "home-telemetry/server/docs"
)

func main() {
	cfg := config.LoadFromEnv()
	logger := log.New(os.Stdout, "", log.LstdFlags|log.LUTC)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	stores := store.New(pool)
	h := api.NewHandler(cfg, stores, logger)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           h.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Printf("listening on %s", cfg.HTTPAddr)
	if !cfg.TLSEnabled {
		logger.Fatalf("TLS required. Set TLS_CERT and TLS_KEY to enable HTTPS.")
	}

	if err := srv.ListenAndServeTLS(cfg.TLSCertPath, cfg.TLSKeyPath); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("server: %v", err)
	}
}
