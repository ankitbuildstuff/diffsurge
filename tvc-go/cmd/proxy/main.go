package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tvc-org/tvc/internal/config"
	"github.com/tvc-org/tvc/pkg/logger"
)

func main() {
	log := logger.Default()

	cfg, err := config.Load("")
	if err != nil {
		log.Warn().Err(err).Msg("Using default configuration")
	}

	addr := ":8080"
	if cfg != nil && cfg.Proxy.ListenAddr != "" {
		addr = cfg.Proxy.ListenAddr
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","version":"%s"}`, config.Version)
	})

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
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

		log.Info().Msg("Shutting down proxy server...")
		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("Proxy server forced shutdown")
		}
	}()

	log.Info().Str("addr", addr).Msg("Starting TVC proxy server")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("Proxy server failed")
	}

	<-ctx.Done()
	log.Info().Msg("Proxy server stopped")
}
