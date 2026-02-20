package proxy

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/tvc-org/tvc/internal/config"
	"github.com/tvc-org/tvc/pkg/logger"
)

type Server struct {
	config     *config.ProxyConfig
	httpServer *http.Server
	log        *logger.Logger
	capture    *TrafficCapture
	router     *Router
}

func NewServer(cfg *config.ProxyConfig, log *logger.Logger, capture *TrafficCapture) *Server {
	s := &Server{
		config:  cfg,
		log:     log,
		capture: capture,
		router:  NewRouter(cfg.Routes, log),
	}

	handler := s.buildHandler()

	s.httpServer = &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return s
}

func (s *Server) buildHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/metrics", s.metricsHandler)

	proxyHandler := s.router.Handler()
	proxyHandler = s.capture.Middleware(proxyHandler)

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" || r.URL.Path == "/metrics" {
			mux.ServeHTTP(w, r)
			return
		}
		proxyHandler.ServeHTTP(w, r)
	})

	handler = RecoveryMiddleware(s.log)(handler)
	handler = RequestIDMiddleware(handler)
	handler = LoggingMiddleware(s.log)(handler)
	handler = CORSMiddleware(handler)

	return handler
}

func (s *Server) Start(ctx context.Context) error {
	if s.capture != nil {
		s.capture.Start()
	}

	errCh := make(chan error, 1)
	go func() {
		s.log.Info().Str("addr", s.config.ListenAddr).Msg("TVC proxy listening")
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		s.log.Info().Msg("Shutting down proxy...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if s.capture != nil {
			s.capture.Stop()
		}

		return s.httpServer.Shutdown(shutdownCtx)
	}
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"ok","version":"%s","uptime":"active"}`, config.Version)
}

func (s *Server) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	stats := s.capture.Stats()
	fmt.Fprintf(w, `{"captured":%d,"dropped":%d,"queue_size":%d}`,
		stats.Captured, stats.Dropped, stats.QueueSize)
}
