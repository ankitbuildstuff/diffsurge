package replayer

import (
	"context"
	"time"

	"github.com/diffsurge-org/diffsurge/internal/models"
	"github.com/diffsurge-org/diffsurge/internal/storage"
	"github.com/diffsurge-org/diffsurge/pkg/logger"
	"github.com/google/uuid"
)

// SessionRunner orchestrates a full replay session end-to-end:
// fetch traffic -> filter -> replay -> compare -> store results -> report.
type SessionRunner struct {
	replayer *Replayer
	comparer *Comparer
	reporter *Reporter
	repo     storage.Repository
	log      *logger.Logger
}

func NewSessionRunner(replayer *Replayer, comparer *Comparer, repo storage.Repository, log *logger.Logger) *SessionRunner {
	return &SessionRunner{
		replayer: replayer,
		comparer: comparer,
		reporter: NewReporter(),
		repo:     repo,
		log:      log,
	}
}

// RunConfig holds all parameters for a replay session run.
type RunConfig struct {
	SessionID    uuid.UUID
	SessionName  string
	ProjectID    uuid.UUID
	SourceEnvID  uuid.UUID
	TargetURL    string
	Filter       storage.TrafficFilter
	FilterConfig FilterConfig
}

// RunResult is returned after a session completes.
type RunResult struct {
	Summary     ReportSummary
	Results     []Result
	Comparisons []ComparisonResult
	Session     *models.ReplaySession
}

func (sr *SessionRunner) Run(ctx context.Context, cfg RunConfig) (*RunResult, error) {
	start := time.Now()

	sr.log.Info().
		Str("session_id", cfg.SessionID.String()).
		Str("target", cfg.TargetURL).
		Msg("Starting replay session")

	// 1. Fetch traffic
	traffic, err := sr.repo.FetchTraffic(ctx, cfg.Filter)
	if err != nil {
		return nil, err
	}
	if len(traffic) == 0 {
		sr.log.Warn().Msg("No traffic found matching filter")
		return &RunResult{
			Summary: ReportSummary{SessionName: cfg.SessionName, TargetURL: cfg.TargetURL},
		}, nil
	}

	sr.log.Info().Int("count", len(traffic)).Msg("Traffic fetched")

	// 2. Filter and sanitize for replay
	filtered := FilterTrafficForReplay(traffic, cfg.FilterConfig)
	sr.log.Info().Int("count", len(filtered)).Msg("Traffic filtered for replay")

	// 3. Update session status
	if sr.repo != nil {
		now := time.Now()
		session := &models.ReplaySession{
			ID:            cfg.SessionID,
			Status:        "running",
			TotalRequests: len(filtered),
			StartedAt:     &now,
		}
		_ = sr.repo.UpdateReplaySession(ctx, session)
	}

	// 4. Replay
	results, err := sr.replayer.ReplayTraffic(ctx, filtered)
	if err != nil {
		return nil, err
	}

	// 5. Compare
	comparisons := sr.comparer.CompareAll(filtered, results)

	// 6. Store results
	successful, failed, mismatched := 0, 0, 0
	for i, result := range results {
		comp := comparisons[i]
		model := ReplayToModel(cfg.SessionID, result, comp)

		if sr.repo != nil {
			if storeErr := sr.repo.SaveReplayResult(ctx, &model); storeErr != nil {
				sr.log.Error().Err(storeErr).Msg("Failed to save replay result")
			}
		}

		if result.Error != nil {
			failed++
		} else if comp.StatusMatch && comp.BodyMatch {
			successful++
		} else {
			mismatched++
		}
	}

	duration := time.Since(start)

	// 7. Generate report
	summary := sr.reporter.GenerateSummary(cfg.SessionName, cfg.TargetURL, results, comparisons, duration)

	// 8. Update session as completed
	if sr.repo != nil {
		now := time.Now()
		session := &models.ReplaySession{
			ID:                  cfg.SessionID,
			Status:              "completed",
			TotalRequests:       len(filtered),
			SuccessfulRequests:  successful,
			FailedRequests:      failed,
			MismatchedResponses: mismatched,
			CompletedAt:         &now,
		}
		_ = sr.repo.UpdateReplaySession(ctx, session)
	}

	sr.log.Info().
		Int("total", len(filtered)).
		Int("successful", successful).
		Int("failed", failed).
		Int("mismatched", mismatched).
		Dur("duration", duration).
		Msg("Replay session completed")

	return &RunResult{
		Summary:     summary,
		Results:     results,
		Comparisons: comparisons,
	}, nil
}
