package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/tvc-org/tvc/internal/models"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(connStr string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *PostgresStore) Close() error {
	return s.db.Close()
}

func (s *PostgresStore) CreateOrganization(ctx context.Context, org *models.Organization) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO organizations (id, name, slug) VALUES ($1, $2, $3)`,
		org.ID, org.Name, org.Slug)
	if err != nil {
		return fmt.Errorf("creating organization: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetOrganization(ctx context.Context, id uuid.UUID) (*models.Organization, error) {
	var org models.Organization
	err := s.db.QueryRowContext(ctx,
		"SELECT id, name, slug, created_at, updated_at FROM organizations WHERE id = $1", id).
		Scan(&org.ID, &org.Name, &org.Slug, &org.CreatedAt, &org.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("organization not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("getting organization: %w", err)
	}
	return &org, nil
}

func (s *PostgresStore) ListUserOrganizations(ctx context.Context, userID uuid.UUID) ([]models.Organization, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT o.id, o.name, o.slug, o.created_at, o.updated_at
		FROM organizations o
		INNER JOIN user_organizations uo ON o.id = uo.organization_id
		WHERE uo.user_id = $1
		ORDER BY o.name`, userID)
	if err != nil {
		return nil, fmt.Errorf("listing user organizations: %w", err)
	}
	defer rows.Close()

	var orgs []models.Organization
	for rows.Next() {
		var o models.Organization
		if err := rows.Scan(&o.ID, &o.Name, &o.Slug, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning organization: %w", err)
		}
		orgs = append(orgs, o)
	}
	return orgs, rows.Err()
}

func (s *PostgresStore) UpdateOrganization(ctx context.Context, org *models.Organization) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE organizations SET name = $1, slug = $2, updated_at = NOW() WHERE id = $3`,
		org.Name, org.Slug, org.ID)
	if err != nil {
		return fmt.Errorf("updating organization: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking update result: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("organization not found")
	}

	return nil
}

func (s *PostgresStore) DeleteOrganization(ctx context.Context, id uuid.UUID) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM organizations WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting organization: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking delete result: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("organization not found")
	}

	return nil
}

func (s *PostgresStore) AddOrganizationMember(ctx context.Context, orgID, userID uuid.UUID, role string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO user_organizations (user_id, organization_id, role) VALUES ($1, $2, $3)`,
		userID, orgID, role)
	if err != nil {
		return fmt.Errorf("adding organization member: %w", err)
	}
	return nil
}

func (s *PostgresStore) RemoveOrganizationMember(ctx context.Context, orgID, userID uuid.UUID) error {
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM user_organizations WHERE organization_id = $1 AND user_id = $2`,
		orgID, userID)
	if err != nil {
		return fmt.Errorf("removing organization member: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking delete result: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("member not found")
	}

	return nil
}

func (s *PostgresStore) ListOrganizationMembers(ctx context.Context, orgID uuid.UUID) ([]models.OrganizationMember, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT u.id, u.email, COALESCE(u.full_name, ''), uo.role, uo.created_at
		FROM users u
		INNER JOIN user_organizations uo ON u.id = uo.user_id
		WHERE uo.organization_id = $1
		ORDER BY uo.created_at ASC`, orgID)
	if err != nil {
		return nil, fmt.Errorf("listing organization members: %w", err)
	}
	defer rows.Close()

	var members []models.OrganizationMember
	for rows.Next() {
		var m models.OrganizationMember
		if err := rows.Scan(&m.UserID, &m.Email, &m.FullName, &m.Role, &m.JoinedAt); err != nil {
			return nil, fmt.Errorf("scanning member: %w", err)
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (s *PostgresStore) UpdateOrganizationMemberRole(ctx context.Context, orgID, userID uuid.UUID, role string) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE user_organizations SET role = $1 WHERE organization_id = $2 AND user_id = $3`,
		role, orgID, userID)
	if err != nil {
		return fmt.Errorf("updating member role: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking update result: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("member not found")
	}

	return nil
}

func (s *PostgresStore) CreateAPIKey(ctx context.Context, key *models.APIKey) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO api_keys (id, organization_id, project_id, name, key_prefix, key_hash, expires_at, created_by) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		key.ID, key.OrganizationID, key.ProjectID, key.Name, key.KeyPrefix, key.KeyHash, key.ExpiresAt, key.CreatedBy)
	if err != nil {
		return fmt.Errorf("creating API key: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetAPIKey(ctx context.Context, id uuid.UUID) (*models.APIKey, error) {
	var key models.APIKey
	err := s.db.QueryRowContext(ctx,
		`SELECT id, organization_id, project_id, name, key_prefix, key_hash, last_used_at, expires_at, created_at, created_by
		FROM api_keys WHERE id = $1`, id).
		Scan(&key.ID, &key.OrganizationID, &key.ProjectID, &key.Name, &key.KeyPrefix, 
			&key.KeyHash, &key.LastUsedAt, &key.ExpiresAt, &key.CreatedAt, &key.CreatedBy)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("API key not found")
	}
	if err != nil {
		return nil, fmt.Errorf("getting API key: %w", err)
	}
	return &key, nil
}

func (s *PostgresStore) GetAPIKeyByHash(ctx context.Context, keyPrefix string) (*models.APIKey, error) {
	var key models.APIKey
	err := s.db.QueryRowContext(ctx,
		`SELECT id, organization_id, project_id, name, key_prefix, key_hash, last_used_at, expires_at, created_at, created_by
		FROM api_keys WHERE key_prefix = $1 AND (expires_at IS NULL OR expires_at > NOW())`, keyPrefix).
		Scan(&key.ID, &key.OrganizationID, &key.ProjectID, &key.Name, &key.KeyPrefix, 
			&key.KeyHash, &key.LastUsedAt, &key.ExpiresAt, &key.CreatedAt, &key.CreatedBy)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("API key not found or expired")
	}
	if err != nil {
		return nil, fmt.Errorf("getting API key: %w", err)
	}
	return &key, nil
}

func (s *PostgresStore) ListAPIKeys(ctx context.Context, orgID uuid.UUID) ([]models.APIKey, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, organization_id, project_id, name, key_prefix, last_used_at, expires_at, created_at, created_by
		FROM api_keys WHERE organization_id = $1 ORDER BY created_at DESC`, orgID)
	if err != nil {
		return nil, fmt.Errorf("listing API keys: %w", err)
	}
	defer rows.Close()

	var keys []models.APIKey
	for rows.Next() {
		var k models.APIKey
		if err := rows.Scan(&k.ID, &k.OrganizationID, &k.ProjectID, &k.Name, &k.KeyPrefix, 
			&k.LastUsedAt, &k.ExpiresAt, &k.CreatedAt, &k.CreatedBy); err != nil {
			return nil, fmt.Errorf("scanning API key: %w", err)
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

func (s *PostgresStore) DeleteAPIKey(ctx context.Context, id uuid.UUID) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM api_keys WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting API key: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking delete result: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("API key not found")
	}

	return nil
}

func (s *PostgresStore) UpdateAPIKeyLastUsed(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE api_keys SET last_used_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("updating API key last used: %w", err)
	}
	return nil
}

// ── Audit Logs ──────────────────────────────────────────────────────────────

func (s *PostgresStore) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO audit_logs (
			id, organization_id, user_id, action, resource_type, resource_id,
			details, ip_address, user_agent, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		log.ID, log.OrganizationID, log.UserID, log.Action, log.ResourceType,
		log.ResourceID, log.Details, log.IPAddress, log.UserAgent, log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("creating audit log: %w", err)
	}
	return nil
}

func (s *PostgresStore) ListAuditLogs(ctx context.Context, filter models.AuditLogFilter) ([]models.AuditLog, error) {
	query := `
		SELECT id, organization_id, user_id, action, resource_type, resource_id,
		       details, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE organization_id = $1
	`
	args := []interface{}{filter.OrganizationID}
	argIdx := 2

	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIdx)
		args = append(args, *filter.UserID)
		argIdx++
	}

	if filter.Action != nil {
		query += fmt.Sprintf(" AND action = $%d", argIdx)
		args = append(args, *filter.Action)
		argIdx++
	}

	if filter.ResourceType != nil {
		query += fmt.Sprintf(" AND resource_type = $%d", argIdx)
		args = append(args, *filter.ResourceType)
		argIdx++
	}

	if filter.StartTime != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argIdx)
		args = append(args, *filter.StartTime)
		argIdx++
	}

	if filter.EndTime != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argIdx)
		args = append(args, *filter.EndTime)
		argIdx++
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filter.Limit)
		argIdx++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, filter.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying audit logs: %w", err)
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var log models.AuditLog
		err := rows.Scan(
			&log.ID, &log.OrganizationID, &log.UserID, &log.Action,
			&log.ResourceType, &log.ResourceID, &log.Details,
			&log.IPAddress, &log.UserAgent, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning audit log: %w", err)
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating audit logs: %w", err)
	}

	return logs, nil
}

// ── Traffic ─────────────────────────────────────────────────────────────────

func (s *PostgresStore) SaveTrafficLog(log *models.TrafficLog) error {
	queryParams, _ := json.Marshal(log.QueryParams)
	reqHeaders, _ := json.Marshal(log.RequestHeaders)
	reqBody, _ := json.Marshal(log.RequestBody)
	respHeaders, _ := json.Marshal(log.ResponseHeaders)
	respBody, _ := json.Marshal(log.ResponseBody)

	_, err := s.db.Exec(`
		INSERT INTO traffic_logs (id, project_id, environment_id, method, path, query_params,
			request_headers, request_body, status_code, response_headers, response_body,
			timestamp, latency_ms, ip_address, user_agent, pii_redacted)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`,
		log.ID, log.ProjectID, log.EnvironmentID, log.Method, log.Path,
		queryParams, reqHeaders, reqBody, log.StatusCode, respHeaders, respBody,
		log.Timestamp, log.LatencyMs, log.IPAddress, log.UserAgent, log.PIIRedacted,
	)
	if err != nil {
		return fmt.Errorf("inserting traffic log: %w", err)
	}
	return nil
}

func (s *PostgresStore) FetchTraffic(ctx context.Context, filter TrafficFilter) ([]models.TrafficLog, error) {
	query := `SELECT id, project_id, environment_id, method, path, query_params,
		request_headers, request_body, status_code, response_headers, response_body,
		timestamp, latency_ms, ip_address, user_agent, pii_redacted
		FROM traffic_logs WHERE 1=1`

	var args []interface{}
	argIdx := 1

	if filter.ProjectID != uuid.Nil {
		query += fmt.Sprintf(" AND project_id = $%d", argIdx)
		args = append(args, filter.ProjectID)
		argIdx++
	}

	if filter.EnvironmentID != nil && *filter.EnvironmentID != uuid.Nil {
		query += fmt.Sprintf(" AND environment_id = $%d", argIdx)
		args = append(args, *filter.EnvironmentID)
		argIdx++
	}

	if filter.StartTime != nil && !filter.StartTime.IsZero() {
		query += fmt.Sprintf(" AND timestamp >= $%d", argIdx)
		args = append(args, *filter.StartTime)
		argIdx++
	}

	if filter.EndTime != nil && !filter.EndTime.IsZero() {
		query += fmt.Sprintf(" AND timestamp <= $%d", argIdx)
		args = append(args, *filter.EndTime)
		argIdx++
	}

	if len(filter.Methods) > 0 {
		placeholders := make([]string, len(filter.Methods))
		for i, m := range filter.Methods {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, m)
			argIdx++
		}
		query += fmt.Sprintf(" AND method IN (%s)", strings.Join(placeholders, ","))
	}

	if len(filter.StatusCodes) > 0 {
		placeholders := make([]string, len(filter.StatusCodes))
		for i, sc := range filter.StatusCodes {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, sc)
			argIdx++
		}
		query += fmt.Sprintf(" AND status_code IN (%s)", strings.Join(placeholders, ","))
	}

	query += " ORDER BY timestamp DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filter.Limit)
		argIdx++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, filter.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying traffic logs: %w", err)
	}
	defer rows.Close()

	var logs []models.TrafficLog
	for rows.Next() {
		var log models.TrafficLog
		var queryParams, reqHeaders, reqBody, respHeaders, respBody []byte

		err := rows.Scan(
			&log.ID, &log.ProjectID, &log.EnvironmentID, &log.Method, &log.Path,
			&queryParams, &reqHeaders, &reqBody, &log.StatusCode, &respHeaders, &respBody,
			&log.Timestamp, &log.LatencyMs, &log.IPAddress, &log.UserAgent, &log.PIIRedacted,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning traffic log: %w", err)
		}

		json.Unmarshal(queryParams, &log.QueryParams)     //nolint:errcheck
		json.Unmarshal(reqHeaders, &log.RequestHeaders)   //nolint:errcheck
		json.Unmarshal(reqBody, &log.RequestBody)         //nolint:errcheck
		json.Unmarshal(respHeaders, &log.ResponseHeaders) //nolint:errcheck
		json.Unmarshal(respBody, &log.ResponseBody)       //nolint:errcheck

		logs = append(logs, log)
	}

	return logs, rows.Err()
}

func (s *PostgresStore) GetTrafficLog(ctx context.Context, id uuid.UUID) (*models.TrafficLog, error) {
	logs, err := s.FetchTraffic(ctx, TrafficFilter{Limit: 1})
	if err != nil {
		return nil, err
	}
	if len(logs) == 0 {
		return nil, fmt.Errorf("traffic log not found: %s", id)
	}
	return &logs[0], nil
}

func (s *PostgresStore) DeleteTraffic(ctx context.Context, filter TrafficFilter) (int64, error) {
	query := "DELETE FROM traffic_logs WHERE 1=1"
	var args []interface{}
	argIdx := 1

	if filter.ProjectID != uuid.Nil {
		query += fmt.Sprintf(" AND project_id = $%d", argIdx)
		args = append(args, filter.ProjectID)
		argIdx++
	}

	if !filter.EndTime.IsZero() {
		query += fmt.Sprintf(" AND timestamp <= $%d", argIdx)
		args = append(args, filter.EndTime)
	}

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("deleting traffic logs: %w", err)
	}
	return result.RowsAffected()
}

func (s *PostgresStore) ListProjects(ctx context.Context, orgID uuid.UUID) ([]models.Project, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, organization_id, name, slug, description, config, created_at, updated_at FROM projects WHERE organization_id = $1 ORDER BY created_at DESC", orgID)
	if err != nil {
		return nil, fmt.Errorf("listing projects: %w", err)
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		var configJSON []byte
		if err := rows.Scan(&p.ID, &p.OrganizationID, &p.Name, &p.Slug, &p.Description, &configJSON, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning project: %w", err)
		}
		json.Unmarshal(configJSON, &p.Config) //nolint:errcheck
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (s *PostgresStore) GetProject(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	var p models.Project
	var configJSON []byte
	err := s.db.QueryRowContext(ctx,
		"SELECT id, organization_id, name, slug, description, config, created_at, updated_at FROM projects WHERE id = $1", id).
		Scan(&p.ID, &p.OrganizationID, &p.Name, &p.Slug, &p.Description, &configJSON, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("getting project: %w", err)
	}
	json.Unmarshal(configJSON, &p.Config) //nolint:errcheck
	return &p, nil
}

func (s *PostgresStore) CreateProject(ctx context.Context, project *models.Project) error {
	configJSON, _ := json.Marshal(project.Config)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO projects (id, organization_id, name, slug, description, config)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		project.ID, project.OrganizationID, project.Name, project.Slug, project.Description, configJSON)
	if err != nil {
		return fmt.Errorf("creating project: %w", err)
	}
	return nil
}

func (s *PostgresStore) UpdateProject(ctx context.Context, project *models.Project) error {
	configJSON, _ := json.Marshal(project.Config)
	_, err := s.db.ExecContext(ctx,
		`UPDATE projects SET name = $2, description = $3, config = $4, updated_at = NOW() WHERE id = $1`,
		project.ID, project.Name, project.Description, configJSON)
	if err != nil {
		return fmt.Errorf("updating project: %w", err)
	}
	return nil
}

func (s *PostgresStore) DeleteProject(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM projects WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("deleting project: %w", err)
	}
	return nil
}

func (s *PostgresStore) ListEnvironments(ctx context.Context, projectID uuid.UUID) ([]models.Environment, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, project_id, name, base_url, is_source, created_at FROM environments WHERE project_id = $1", projectID)
	if err != nil {
		return nil, fmt.Errorf("listing environments: %w", err)
	}
	defer rows.Close()

	var envs []models.Environment
	for rows.Next() {
		var e models.Environment
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.Name, &e.BaseURL, &e.IsSource, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning environment: %w", err)
		}
		envs = append(envs, e)
	}
	return envs, rows.Err()
}

func (s *PostgresStore) GetEnvironment(ctx context.Context, id uuid.UUID) (*models.Environment, error) {
	var e models.Environment
	err := s.db.QueryRowContext(ctx,
		"SELECT id, project_id, name, base_url, is_source, created_at FROM environments WHERE id = $1", id).
		Scan(&e.ID, &e.ProjectID, &e.Name, &e.BaseURL, &e.IsSource, &e.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("environment not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("getting environment: %w", err)
	}
	return &e, nil
}

func (s *PostgresStore) CreateEnvironment(ctx context.Context, env *models.Environment) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO environments (id, project_id, name, base_url, is_source) VALUES ($1, $2, $3, $4, $5)`,
		env.ID, env.ProjectID, env.Name, env.BaseURL, env.IsSource)
	if err != nil {
		return fmt.Errorf("creating environment: %w", err)
	}
	return nil
}

func (s *PostgresStore) UpdateEnvironment(ctx context.Context, env *models.Environment) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE environments SET name = $1, base_url = $2, is_source = $3 WHERE id = $4`,
		env.Name, env.BaseURL, env.IsSource, env.ID)
	if err != nil {
		return fmt.Errorf("updating environment: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking update result: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("environment not found")
	}

	return nil
}

func (s *PostgresStore) DeleteEnvironment(ctx context.Context, id uuid.UUID) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM environments WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting environment: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking delete result: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("environment not found")
	}

	return nil
}

func (s *PostgresStore) ListReplaySessions(ctx context.Context, projectID uuid.UUID) ([]models.ReplaySession, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, project_id, source_environment_id, target_environment_id, name, description,
		status, total_requests, successful_requests, failed_requests, mismatched_responses,
		created_at, started_at, completed_at
		FROM replay_sessions WHERE project_id = $1 ORDER BY created_at DESC`, projectID)
	if err != nil {
		return nil, fmt.Errorf("listing replay sessions: %w", err)
	}
	defer rows.Close()

	var sessions []models.ReplaySession
	for rows.Next() {
		var s models.ReplaySession
		if err := rows.Scan(
			&s.ID, &s.ProjectID, &s.SourceEnvironmentID, &s.TargetEnvironmentID,
			&s.Name, &s.Description, &s.Status, &s.TotalRequests, &s.SuccessfulRequests,
			&s.FailedRequests, &s.MismatchedResponses, &s.CreatedAt, &s.StartedAt, &s.CompletedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning replay session: %w", err)
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

func (s *PostgresStore) GetReplaySession(ctx context.Context, id uuid.UUID) (*models.ReplaySession, error) {
	var session models.ReplaySession
	err := s.db.QueryRowContext(ctx,
		`SELECT id, project_id, source_environment_id, target_environment_id, name, description,
		status, total_requests, successful_requests, failed_requests, mismatched_responses,
		created_at, started_at, completed_at
		FROM replay_sessions WHERE id = $1`, id).
		Scan(&session.ID, &session.ProjectID, &session.SourceEnvironmentID, &session.TargetEnvironmentID,
			&session.Name, &session.Description, &session.Status, &session.TotalRequests,
			&session.SuccessfulRequests, &session.FailedRequests, &session.MismatchedResponses,
			&session.CreatedAt, &session.StartedAt, &session.CompletedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("replay session not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("getting replay session: %w", err)
	}
	return &session, nil
}

func (s *PostgresStore) CreateReplaySession(ctx context.Context, session *models.ReplaySession) error {
	filterJSON, _ := json.Marshal(session.TrafficFilter)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO replay_sessions (id, project_id, source_environment_id, target_environment_id,
		name, description, traffic_filter, start_time, end_time, sample_size, status, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		session.ID, session.ProjectID, session.SourceEnvironmentID, session.TargetEnvironmentID,
		session.Name, session.Description, filterJSON, session.StartTime, session.EndTime,
		session.SampleSize, session.Status, session.CreatedBy)
	if err != nil {
		return fmt.Errorf("creating replay session: %w", err)
	}
	return nil
}

func (s *PostgresStore) UpdateReplaySession(ctx context.Context, session *models.ReplaySession) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE replay_sessions SET status = $2, total_requests = $3, successful_requests = $4,
		failed_requests = $5, mismatched_responses = $6, started_at = $7, completed_at = $8
		WHERE id = $1`,
		session.ID, session.Status, session.TotalRequests, session.SuccessfulRequests,
		session.FailedRequests, session.MismatchedResponses, session.StartedAt, session.CompletedAt)
	if err != nil {
		return fmt.Errorf("updating replay session: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetPendingSessions(ctx context.Context, limit int) ([]PendingSession, error) {
	query := `
		SELECT rs.id, rs.name, rs.project_id, rs.source_environment_id, rs.sample_size, rs.created_at,
		       e.base_url, rs.traffic_start_time, rs.traffic_end_time
		FROM replay_sessions rs
		JOIN environments e ON e.id = rs.target_environment_id
		WHERE rs.status = 'pending'
		ORDER BY rs.created_at ASC
		LIMIT $1
	`

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("querying pending sessions: %w", err)
	}
	defer rows.Close()

	var sessions []PendingSession
	for rows.Next() {
		var ps PendingSession
		if err := rows.Scan(
			&ps.ID, &ps.Name, &ps.ProjectID, &ps.SourceEnvID, &ps.SampleSize, &ps.CreatedAt,
			&ps.TargetURL, &ps.TrafficStartTime, &ps.TrafficEndTime,
		); err != nil {
			return nil, fmt.Errorf("scanning pending session: %w", err)
		}
		sessions = append(sessions, ps)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating pending sessions: %w", err)
	}

	return sessions, nil
}

func (s *PostgresStore) SaveReplayResult(ctx context.Context, result *models.ReplayResult) error {
	diffJSON, _ := json.Marshal(result.DiffReport)
	respBody, _ := json.Marshal(result.TargetResponseBody)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO replay_results (id, replay_session_id, original_traffic_log_id,
		target_status_code, target_response_body, target_latency_ms, status_match,
		body_match, diff_report, severity, error_message)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		result.ID, result.ReplaySessionID, result.OriginalTrafficLogID,
		result.TargetStatusCode, respBody, result.TargetLatencyMs,
		result.StatusMatch, result.BodyMatch, diffJSON, result.Severity, result.ErrorMessage)
	if err != nil {
		return fmt.Errorf("saving replay result: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetReplayResults(ctx context.Context, sessionID uuid.UUID) ([]models.ReplayResult, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, replay_session_id, original_traffic_log_id, target_status_code,
		target_response_body, target_latency_ms, status_match, body_match,
		diff_report, severity, error_message, timestamp
		FROM replay_results WHERE replay_session_id = $1 ORDER BY timestamp`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("getting replay results: %w", err)
	}
	defer rows.Close()

	var results []models.ReplayResult
	for rows.Next() {
		var r models.ReplayResult
		var respBody, diffReport []byte
		if err := rows.Scan(
			&r.ID, &r.ReplaySessionID, &r.OriginalTrafficLogID, &r.TargetStatusCode,
			&respBody, &r.TargetLatencyMs, &r.StatusMatch, &r.BodyMatch,
			&diffReport, &r.Severity, &r.ErrorMessage, &r.Timestamp,
		); err != nil {
			return nil, fmt.Errorf("scanning replay result: %w", err)
		}
		json.Unmarshal(respBody, &r.TargetResponseBody) //nolint:errcheck
		json.Unmarshal(diffReport, &r.DiffReport)       //nolint:errcheck
		results = append(results, r)
	}
	return results, rows.Err()
}

func (s *PostgresStore) ListSchemaVersions(ctx context.Context, projectID uuid.UUID) ([]models.SchemaVersion, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, project_id, version, schema_type, schema_content, git_commit, git_branch, created_at
		FROM schema_versions WHERE project_id = $1 ORDER BY created_at DESC`, projectID)
	if err != nil {
		return nil, fmt.Errorf("listing schema versions: %w", err)
	}
	defer rows.Close()

	var versions []models.SchemaVersion
	for rows.Next() {
		var v models.SchemaVersion
		var content []byte
		if err := rows.Scan(&v.ID, &v.ProjectID, &v.Version, &v.SchemaType, &content, &v.GitCommit, &v.GitBranch, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning schema version: %w", err)
		}
		json.Unmarshal(content, &v.SchemaContent) //nolint:errcheck
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

func (s *PostgresStore) SaveSchemaVersion(ctx context.Context, schema *models.SchemaVersion) error {
	content, _ := json.Marshal(schema.SchemaContent)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO schema_versions (id, project_id, version, schema_type, schema_content, git_commit, git_branch, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		schema.ID, schema.ProjectID, schema.Version, schema.SchemaType, content, schema.GitCommit, schema.GitBranch, schema.CreatedBy)
	if err != nil {
		return fmt.Errorf("saving schema version: %w", err)
	}
	return nil
}

func (s *PostgresStore) SaveSchemaDiff(ctx context.Context, diff *models.SchemaDiff) error {
	report, _ := json.Marshal(diff.DiffReport)
	breaking, _ := json.Marshal(diff.BreakingChanges)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO schema_diffs (id, project_id, from_version_id, to_version_id, diff_report, has_breaking_changes, breaking_changes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		diff.ID, diff.ProjectID, diff.FromVersionID, diff.ToVersionID, report, diff.HasBreakingChanges, breaking)
	if err != nil {
		return fmt.Errorf("saving schema diff: %w", err)
	}
	return nil
}
