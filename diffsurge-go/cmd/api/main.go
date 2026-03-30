package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/diffsurge-org/diffsurge/internal/api"
	"github.com/diffsurge-org/diffsurge/internal/api/middleware"
	"github.com/diffsurge-org/diffsurge/internal/config"
	"github.com/diffsurge-org/diffsurge/internal/storage"
	"github.com/diffsurge-org/diffsurge/pkg/logger"
)

func main() {
	// healthcheck subcommand: used by Docker HEALTHCHECK, avoids the need for
	// wget/curl in the container image.
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		port := "8080"
		if p := os.Getenv("PORT"); p != "" {
			port = p
		}
		resp, err := http.Get("http://localhost:" + port + "/api/v1/health")
		if err != nil || resp.StatusCode >= 500 {
			os.Exit(1)
		}
		os.Exit(0)
	}

	log := logger.Default()

	cfg, err := config.Load("")
	if err != nil {
		log.Warn().Err(err).Msg("Config not found, using defaults")
	}

	port := 8080
	if envPort := os.Getenv("PORT"); envPort != "" {
		if _, err := fmt.Sscanf(envPort, "%d", &port); err != nil {
			log.Warn().Err(err).Str("PORT", envPort).Msg("Invalid PORT, using default")
		}
	} else if cfg != nil && cfg.Server.Port != 0 {
		port = cfg.Server.Port
	}
	addr := fmt.Sprintf(":%d", port)

	// Database connection
	pgURL := os.Getenv("DIFFSURGE_STORAGE_POSTGRES_URL")
	if pgURL == "" && cfg != nil {
		pgURL = cfg.Storage.PostgresURL
	}
	if pgURL == "" {
		log.Fatal().Msg("DIFFSURGE_STORAGE_POSTGRES_URL is required (set via env var or config file)")
	}

	store, err := storage.NewPostgresStore(pgURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer store.Close()
	log.Info().Msg("Connected to PostgreSQL")

	// Auth config from environment (accept both docker-compose and .env naming)
	authCfg := middleware.AuthConfig{
		SupabaseURL:    getEnvFallback("SUPABASE_URL", "NEXT_PUBLIC_SUPABASE_URL"),
		SupabaseSecret: getEnvFallback("SUPABASE_SERVICE_KEY", "SUPABASE_SERVICE_ROLE_KEY"),
		JWTSecret:      os.Getenv("SUPABASE_JWT_SECRET"),
	}

	if authCfg.JWTSecret == "" {
		log.Warn().Msg("SUPABASE_JWT_SECRET is not set — JWT auth will fail")
	}
	if authCfg.SupabaseURL == "" {
		log.Warn().Msg("SUPABASE_URL / NEXT_PUBLIC_SUPABASE_URL is not set — JWKS auth will be unavailable")
	}

	deps := api.ServerDeps{
		Store:      store,
		Log:        log,
		AuthConfig: authCfg,
	}

	router := api.NewRouter(deps)

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

	log.Info().Str("addr", addr).Msg("Starting Diffsurge API server")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("API server failed")
	}

	<-ctx.Done()
	log.Info().Msg("API server stopped")
}

// getEnvFallback returns the value of the primary env var, or the fallback if primary is empty.
func getEnvFallback(primary, fallback string) string {
	if v := os.Getenv(primary); v != "" {
		return v
	}
	return os.Getenv(fallback)
}
