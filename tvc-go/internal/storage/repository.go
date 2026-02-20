package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tvc-org/tvc/internal/models"
)

type TrafficFilter struct {
	ProjectID     uuid.UUID
	EnvironmentID uuid.UUID
	StartTime     time.Time
	EndTime       time.Time
	Paths         []string
	Methods       []string
	StatusCodes   []int
	Limit         int
	Offset        int
}

type Repository interface {
	// Traffic
	SaveTrafficLog(log *models.TrafficLog) error
	FetchTraffic(ctx context.Context, filter TrafficFilter) ([]models.TrafficLog, error)
	GetTrafficLog(ctx context.Context, id uuid.UUID) (*models.TrafficLog, error)
	DeleteTraffic(ctx context.Context, filter TrafficFilter) (int64, error)

	// Projects
	ListProjects(ctx context.Context, orgID uuid.UUID) ([]models.Project, error)
	GetProject(ctx context.Context, id uuid.UUID) (*models.Project, error)
	CreateProject(ctx context.Context, project *models.Project) error
	UpdateProject(ctx context.Context, project *models.Project) error
	DeleteProject(ctx context.Context, id uuid.UUID) error

	// Environments
	ListEnvironments(ctx context.Context, projectID uuid.UUID) ([]models.Environment, error)
	GetEnvironment(ctx context.Context, id uuid.UUID) (*models.Environment, error)
	CreateEnvironment(ctx context.Context, env *models.Environment) error

	// Replay Sessions
	ListReplaySessions(ctx context.Context, projectID uuid.UUID) ([]models.ReplaySession, error)
	GetReplaySession(ctx context.Context, id uuid.UUID) (*models.ReplaySession, error)
	CreateReplaySession(ctx context.Context, session *models.ReplaySession) error
	UpdateReplaySession(ctx context.Context, session *models.ReplaySession) error

	// Replay Results
	SaveReplayResult(ctx context.Context, result *models.ReplayResult) error
	GetReplayResults(ctx context.Context, sessionID uuid.UUID) ([]models.ReplayResult, error)

	// Schema
	ListSchemaVersions(ctx context.Context, projectID uuid.UUID) ([]models.SchemaVersion, error)
	SaveSchemaVersion(ctx context.Context, schema *models.SchemaVersion) error
	SaveSchemaDiff(ctx context.Context, diff *models.SchemaDiff) error

	// Health
	Ping(ctx context.Context) error
	Close() error
}
