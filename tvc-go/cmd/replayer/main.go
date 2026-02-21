package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/tvc-org/tvc/internal/config"
	"github.com/tvc-org/tvc/internal/models"
	"github.com/tvc-org/tvc/internal/replayer"
	"github.com/tvc-org/tvc/internal/storage"
	"github.com/tvc-org/tvc/pkg/logger"
)

const (
	pollInterval   = 5 * time.Second
	sessionTimeout = 30 * time.Minute
	maxConcurrent  = 3 // Max number of concurrent replay sessions
	workerPoolSize = 10
	requestTimeout = 30 * time.Second
	maxRetries     = 3
)

func main() {
	log := logger.New("info", "json")

	log.Info().
		Str("version", config.Version).
		Msg("TVC Replay Engine starting")

	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Validate required configuration
	if cfg.Storage.PostgresURL == "" {
		log.Fatal().Msg("PostgreSQL URL is required")
	}

	// Connect to database
	repo, err := storage.NewPostgresStore(cfg.Storage.PostgresURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer repo.Close()

	log.Info().Msg("Connected to database")

	// Optional: Connect to Redis for metrics/caching
	var redisStore *storage.RedisStore
	if cfg.Storage.RedisURL != "" {
		redisStore, err = storage.NewRedisStore(cfg.Storage.RedisURL)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to connect to Redis, continuing without it")
		} else {
			defer redisStore.Close()
			log.Info().Msg("Connected to Redis")
		}
	}

	// Create replayer components
	replayEngine := replayer.New(replayer.Config{
		Workers:    workerPoolSize,
		Timeout:    requestTimeout,
		MaxRetries: maxRetries,
	}, log)

	comparer := replayer.NewComparer(replayer.ComparerConfig{
		IgnorePaths:      []string{"metadata.timestamp", "metadata.request_id"},
		TreatArraysAsSet: true,
	})

	sessionRunner := replayer.NewSessionRunner(replayEngine, comparer, repo, log)

	// Create service
	service := &ReplayerService{
		repo:           repo,
		redis:          redisStore,
		sessionRunner:  sessionRunner,
		log:            log,
		maxConcurrent:  maxConcurrent,
		activeSessions: make(map[string]context.CancelFunc),
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start polling for replay sessions
	go service.Poll(ctx, pollInterval)

	log.Info().Msg("Replayer service is running. Press Ctrl+C to stop.")

	// Wait for shutdown signal
	<-sigChan
	log.Info().Msg("Shutdown signal received, stopping gracefully...")

	// Cancel all active sessions
	service.Stop()

	// Wait a bit for cleanup
	time.Sleep(2 * time.Second)

	log.Info().Msg("Replayer service stopped")
}

// ReplayerService manages the lifecycle of replay sessions
type ReplayerService struct {
	repo           storage.Repository
	redis          *storage.RedisStore
	sessionRunner  *replayer.SessionRunner
	log            *logger.Logger
	maxConcurrent  int
	activeSessions map[string]context.CancelFunc
	mu             sync.Mutex
}

// Poll continuously checks for pending replay sessions
func (s *ReplayerService) Poll(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkAndExecute(ctx)
		}
	}
}

// checkAndExecute finds pending sessions and executes them
func (s *ReplayerService) checkAndExecute(ctx context.Context) {
	s.mu.Lock()
	activeCount := len(s.activeSessions)
	s.mu.Unlock()

	if activeCount >= s.maxConcurrent {
		s.log.Debug().
			Int("active", activeCount).
			Int("max", s.maxConcurrent).
			Msg("Max concurrent sessions reached, waiting")
		return
	}

	// Fetch pending sessions from database
	sessions, err := s.repo.GetPendingSessions(ctx, s.maxConcurrent-activeCount)
	if err != nil {
		s.log.Error().Err(err).Msg("Failed to fetch pending sessions")
		return
	}

	if len(sessions) == 0 {
		return
	}

	s.log.Info().Int("count", len(sessions)).Msg("Found pending replay sessions")

	for i := range sessions {
		s.executeSession(ctx, &sessions[i])
	}
}

// executeSession starts a replay session in a goroutine
func (s *ReplayerService) executeSession(parentCtx context.Context, session *storage.PendingSession) {
	sessionID := session.ID.String()

	s.mu.Lock()
	if _, exists := s.activeSessions[sessionID]; exists {
		s.mu.Unlock()
		return // Already executing
	}

	// Create cancellable context for this session
	sessionCtx, cancel := context.WithTimeout(parentCtx, sessionTimeout)
	s.activeSessions[sessionID] = cancel
	s.mu.Unlock()

	s.log.Info().
		Str("session_id", sessionID).
		Str("name", session.Name).
		Msg("Starting replay session execution")

	go func() {
		defer func() {
			// Cleanup
			s.mu.Lock()
			delete(s.activeSessions, sessionID)
			s.mu.Unlock()
			cancel()

			s.log.Info().
				Str("session_id", sessionID).
				Msg("Session execution completed")
		}()

		// Build run configuration
		runConfig := replayer.RunConfig{
			SessionID:   session.ID,
			SessionName: session.Name,
			ProjectID:   session.ProjectID,
			SourceEnvID: session.SourceEnvID,
			TargetURL:   session.TargetURL,
			Filter: storage.TrafficFilter{
				ProjectID:     session.ProjectID,
				EnvironmentID: &session.SourceEnvID,
				StartTime:     session.TrafficStartTime,
				EndTime:       session.TrafficEndTime,
				Limit:         session.SampleSize,
			},
			FilterConfig: replayer.FilterConfig{
				StripSensitiveHeaders: true,
			},
		}

		// Execute the session
		result, err := s.sessionRunner.Run(sessionCtx, runConfig)
		if err != nil {
			s.log.Error().
				Err(err).
				Str("session_id", sessionID).
				Msg("Session execution failed")

			// Mark session as failed
			now := time.Now()
			failedSession := &models.ReplaySession{
				ID:          session.ID,
				Status:      "failed",
				CompletedAt: &now,
			}
			_ = s.repo.UpdateReplaySession(context.Background(), failedSession)
			return
		}

		s.log.Info().
			Str("session_id", sessionID).
			Int("total", result.Summary.TotalRequests).
			Int("successful", result.Summary.Successful).
			Int("failed", result.Summary.Failed).
			Int("mismatched", result.Summary.Mismatched).
			Msg("Session completed successfully")

		// Publish completion event if Redis is available
		if s.redis != nil {
			completion := map[string]interface{}{
				"session_id": sessionID,
				"status":     "completed",
				"summary":    result.Summary,
				"timestamp":  time.Now(),
			}
			if err := s.redis.Publish(context.Background(), "replay:completed", completion); err != nil {
				s.log.Error().Err(err).Msg("Failed to publish completion event")
			}
		}
	}()
}

// Stop cancels all active sessions
func (s *ReplayerService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.log.Info().
		Int("active_sessions", len(s.activeSessions)).
		Msg("Stopping all active sessions")

	for sessionID, cancel := range s.activeSessions {
		s.log.Info().Str("session_id", sessionID).Msg("Cancelling session")
		cancel()
	}

	s.activeSessions = make(map[string]context.CancelFunc)
}
