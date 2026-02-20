package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tvc-org/tvc/internal/api"
	"github.com/tvc-org/tvc/internal/config"
	"github.com/tvc-org/tvc/pkg/logger"
)

func main() {
	log := logger.Default()

	cfg, err := config.Load("")
	if err != nil {
		log.Warn().Err(err).Msg("Config not found, using defaults")
	}

	port := 8081
	if cfg != nil && cfg.Server.Port != 0 {
		port = cfg.Server.Port
	}
	addr := fmt.Sprintf(":%d", port)

	router := api.NewRouter()

	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh

		log.Info().Msg("Shutting down API server...")
		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("API server forced shutdown")
		}
	}()

	log.Info().Str("addr", addr).Msg("Starting TVC API server")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("API server failed")
	}

	<-ctx.Done()
	log.Info().Msg("API server stopped")
}
