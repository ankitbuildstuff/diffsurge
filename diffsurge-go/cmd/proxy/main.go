package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/diffsurge-org/diffsurge/internal/config"
	"github.com/diffsurge-org/diffsurge/internal/proxy"
	"github.com/diffsurge-org/diffsurge/internal/storage"
	"github.com/diffsurge-org/diffsurge/pkg/logger"
)

func main() {
	// healthcheck subcommand: used by Docker HEALTHCHECK, avoids the need for
	// wget/curl in the container image.
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		addr := "http://localhost:8080/health"
		resp, err := http.Get(addr)
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

	if cfg == nil {
		cfg = &config.Config{}
	}
	if cfg.Proxy.ListenAddr == "" {
		cfg.Proxy.ListenAddr = ":8080"
	}
	if cfg.Proxy.SamplingRate == 0 {
		cfg.Proxy.SamplingRate = 1.0
	}
	if cfg.Proxy.Buffer.QueueSize == 0 {
		cfg.Proxy.Buffer.QueueSize = 10000
	}
	if cfg.Proxy.Buffer.Workers == 0 {
		cfg.Proxy.Buffer.Workers = 10
	}

	var store proxy.TrafficStore
	var keyStore proxy.APIKeyStore
	if cfg.Storage.PostgresURL != "" {
		pgStore, err := storage.NewPostgresStore(cfg.Storage.PostgresURL)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to connect to database")
		}
		defer pgStore.Close()
		store = pgStore
		keyStore = pgStore
		log.Info().Msg("Connected to PostgreSQL")
	} else {
		log.Warn().Msg("No database configured, traffic will be captured but not persisted")
	}

	sampler := proxy.NewPercentageSampler(cfg.Proxy.SamplingRate)
	capture := proxy.NewTrafficCapture(
		cfg.Proxy.Buffer.QueueSize,
		cfg.Proxy.Buffer.Workers,
		store,
		sampler,
		log,
	)

	srv := proxy.NewServer(&cfg.Proxy, log, capture, keyStore)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	if err := srv.Start(ctx); err != nil {
		log.Fatal().Err(err).Msg("Proxy server error")
	}
}
