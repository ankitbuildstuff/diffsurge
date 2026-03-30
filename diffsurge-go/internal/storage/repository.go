package storage

import (
	"context"
	"time"

	"github.com/diffsurge-org/diffsurge/internal/models"
	"github.com/google/uuid"
)

type TrafficFilter struct {
	ProjectID     uuid.UUID
	EnvironmentID *uuid.UUID
	StartTime     *time.Time
	EndTime       *time.Time
	Paths         []string
	Methods       []string
	StatusCodes   []int
	Limit         int
	Offset        int
}

// PendingSession represents a replay session waiting to be executed
type PendingSession struct {
	ID               uuid.UUID
	Name             string
	ProjectID        uuid.UUID
	SourceEnvID      uuid.UUID
	TargetURL        string
	SampleSize       int
	TrafficStartTime *time.Time
	TrafficEndTime   *time.Time
	CreatedAt        time.Time
}

type Repository interface {
	// Organizations
	CreateOrganization(ctx context.Context, org *models.Organization) error
	GetOrganization(ctx context.Context, id uuid.UUID) (*models.Organization, error)
	UpdateOrganization(ctx context.Context, org *models.Organization) error
	DeleteOrganization(ctx context.Context, id uuid.UUID) error
	ListUserOrganizations(ctx context.Context, userID uuid.UUID) ([]models.Organization, error)

	// Organization Members
	AddOrganizationMember(ctx context.Context, orgID, userID uuid.UUID, role string) error
	RemoveOrganizationMember(ctx context.Context, orgID, userID uuid.UUID) error
	ListOrganizationMembers(ctx context.Context, orgID uuid.UUID) ([]models.OrganizationMember, error)
	UpdateOrganizationMemberRole(ctx context.Context, orgID, userID uuid.UUID, role string) error

	// API Keys
	CreateAPIKey(ctx context.Context, key *models.APIKey) error
	GetAPIKey(ctx context.Context, id uuid.UUID) (*models.APIKey, error)
	GetAPIKeyByHash(ctx context.Context, keyHash string) (*models.APIKey, error)
	ListAPIKeys(ctx context.Context, orgID uuid.UUID) ([]models.APIKey, error)
	DeleteAPIKey(ctx context.Context, id uuid.UUID) error
	UpdateAPIKeyLastUsed(ctx context.Context, id uuid.UUID) error

	// Audit Logs
	CreateAuditLog(ctx context.Context, log *models.AuditLog) error
	ListAuditLogs(ctx context.Context, filter models.AuditLogFilter) ([]models.AuditLog, error)

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
	UpdateEnvironment(ctx context.Context, env *models.Environment) error
	DeleteEnvironment(ctx context.Context, id uuid.UUID) error

	// Replay Sessions
	ListReplaySessions(ctx context.Context, projectID uuid.UUID) ([]models.ReplaySession, error)
	GetReplaySession(ctx context.Context, id uuid.UUID) (*models.ReplaySession, error)
	CreateReplaySession(ctx context.Context, session *models.ReplaySession) error
	UpdateReplaySession(ctx context.Context, session *models.ReplaySession) error
	GetPendingSessions(ctx context.Context, limit int) ([]PendingSession, error)

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
