package storage

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/diffsurge-org/diffsurge/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPostgresStore_CreateReplaySession(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := &PostgresStore{db: db}
	now := time.Now().UTC()
	desc := "CLI replay against https://api.diffsurge.com"
	startedAt := now.Add(-30 * time.Second)
	completedAt := now

	session := &models.ReplaySession{
		ID:                  uuid.New(),
		ProjectID:           uuid.New(),
		SourceEnvironmentID: uuid.New(),
		TargetEnvironmentID: uuid.New(),
		Name:                "CLI Replay",
		Description:         &desc,
		TrafficFilter: map[string]interface{}{
			"method": "GET",
		},
		SampleSize:          10,
		Status:              "completed",
		TotalRequests:       10,
		SuccessfulRequests:  8,
		FailedRequests:      2,
		MismatchedResponses: 3,
		CreatedAt:           now,
		StartedAt:           &startedAt,
		CompletedAt:         &completedAt,
	}

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO replay_sessions (id, project_id, source_environment_id, target_environment_id,
		name, description, traffic_filter, start_time, end_time, sample_size, status,
		total_requests, successful_requests, failed_requests, mismatched_responses,
		created_by, created_at, started_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`)).
		WithArgs(
			session.ID,
			session.ProjectID,
			session.SourceEnvironmentID,
			session.TargetEnvironmentID,
			session.Name,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			session.StartTime,
			session.EndTime,
			session.SampleSize,
			session.Status,
			session.TotalRequests,
			session.SuccessfulRequests,
			session.FailedRequests,
			session.MismatchedResponses,
			nil,
			session.CreatedAt,
			session.StartedAt,
			session.CompletedAt,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = store.CreateReplaySession(context.Background(), session)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
