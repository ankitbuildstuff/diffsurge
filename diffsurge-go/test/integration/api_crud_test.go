package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/diffsurge-org/diffsurge/internal/api"
	"github.com/diffsurge-org/diffsurge/internal/api/middleware"
	"github.com/diffsurge-org/diffsurge/internal/models"
	"github.com/diffsurge-org/diffsurge/internal/storage"
	"github.com/diffsurge-org/diffsurge/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// setupTestDB creates a PostgreSQL container and returns the connection string
func setupTestDB(t *testing.T) (string, func()) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	connStr := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable",
		host, port.Port())

	cleanup := func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}

	return connStr, cleanup
}

// runMigrations applies the database schema
func runMigrations(t *testing.T, connStr string) {
	ctx := context.Background()

	// Connect directly for migrations
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	defer db.Close()

	// Create tables (simplified schema for testing)
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS organizations (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(100) UNIQUE NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`CREATE TABLE IF NOT EXISTS projects (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(100) NOT NULL,
			description TEXT,
			config JSONB DEFAULT '{}',
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			UNIQUE (organization_id, slug)
		)`,
		`CREATE TABLE IF NOT EXISTS environments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			name VARCHAR(100) NOT NULL,
			base_url VARCHAR(500) NOT NULL,
			is_source BOOLEAN NOT NULL DEFAULT false,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`CREATE TABLE IF NOT EXISTS traffic_logs (
			id UUID DEFAULT gen_random_uuid(),
			project_id UUID NOT NULL,
			environment_id UUID NOT NULL,
			method VARCHAR(10) NOT NULL,
			path TEXT NOT NULL,
			query_params JSONB,
			request_headers JSONB,
			request_body JSONB,
			status_code INTEGER NOT NULL,
			response_headers JSONB,
			response_body JSONB,
			timestamp TIMESTAMPTZ NOT NULL DEFAULT now(),
			latency_ms INTEGER,
			ip_address INET,
			user_agent TEXT,
			pii_redacted BOOLEAN DEFAULT false,
			PRIMARY KEY (id, timestamp)
		)`,
	}

	for _, migration := range migrations {
		_, err := db.ExecContext(ctx, migration)
		require.NoError(t, err)
	}
}

const testJWTSecret = "test-secret"

func generateTestJWT(userID uuid.UUID) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  userID.String(),
		"role": "owner",
		"exp":  time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, _ := token.SignedString([]byte(testJWTSecret))
	return tokenStr
}

func addAuth(req *http.Request, token string) {
	req.Header.Set("Authorization", "Bearer "+token)
}

// TestAPICRUDLifecycle tests the full CRUD lifecycle for projects and environments
func TestAPICRUDLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup test database
	connStr, cleanup := setupTestDB(t)
	defer cleanup()

	// Run migrations first
	runMigrations(t, connStr)

	// Connect to database
	store, err := storage.NewPostgresStore(connStr)
	require.NoError(t, err)
	defer store.Close()

	// Setup HTTP server
	log := logger.New("info", "text")
	deps := api.ServerDeps{
		Store: store,
		Log:   log,
		AuthConfig: middleware.AuthConfig{
			JWTSecret: testJWTSecret,
		},
	}
	router := api.NewRouter(deps)

	// Create test organization and auth token
	ctx := context.Background()
	userID := uuid.New()
	authToken := generateTestJWT(userID)

	org := &models.Organization{
		ID:        uuid.New(),
		Name:      "Test Org",
		Slug:      "test-org",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = store.CreateOrganization(ctx, org)
	require.NoError(t, err)

	t.Run("Project CRUD", func(t *testing.T) {
		// CREATE: Create a new project
		createReq := map[string]interface{}{
			"name":            "Test Project",
			"description":     "Integration test project",
			"organization_id": org.ID.String(),
		}
		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(body))
		addAuth(req, authToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var createdProject models.Project
		err := json.NewDecoder(w.Body).Decode(&createdProject)
		require.NoError(t, err)
		projectID := createdProject.ID
		require.NotEqual(t, uuid.Nil, projectID, "Created project should have a valid ID")

		// READ: Get the project
		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/projects/%s", projectID), nil)
		addAuth(req, authToken)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var getResp struct {
			Data models.Project `json:"data"`
		}
		err = json.NewDecoder(w.Body).Decode(&getResp)
		require.NoError(t, err)
		assert.Equal(t, "Test Project", getResp.Data.Name)

		// UPDATE: Update the project
		updateReq := map[string]interface{}{
			"name": "Updated Test Project",
		}
		body, _ = json.Marshal(updateReq)
		req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/projects/%s", projectID), bytes.NewReader(body))
		addAuth(req, authToken)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// LIST: List projects
		req = httptest.NewRequest(http.MethodGet,
			fmt.Sprintf("/api/v1/projects?organization_id=%s", org.ID), nil)
		addAuth(req, authToken)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var listResp struct {
			Data []models.Project `json:"data"`
		}
		err = json.NewDecoder(w.Body).Decode(&listResp)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(listResp.Data), 1)

		// DELETE: Delete the project
		req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/projects/%s", projectID), nil)
		addAuth(req, authToken)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify deletion
		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/projects/%s", projectID), nil)
		addAuth(req, authToken)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Environment CRUD", func(t *testing.T) {
		// Create a project first
		project := &models.Project{
			ID:             uuid.New(),
			OrganizationID: org.ID,
			Name:           "Test Project for Env",
			Slug:           "test-project-for-env",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		err := store.CreateProject(ctx, project)
		require.NoError(t, err)

		// CREATE: Create environment
		createReq := map[string]interface{}{
			"name":      "staging",
			"base_url":  "https://staging.example.com",
			"is_source": false,
		}
		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost,
			fmt.Sprintf("/api/v1/projects/%s/environments", project.ID),
			bytes.NewReader(body))
		addAuth(req, authToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		// LIST: List environments
		req = httptest.NewRequest(http.MethodGet,
			fmt.Sprintf("/api/v1/projects/%s/environments", project.ID), nil)
		addAuth(req, authToken)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var listResp struct {
			Data []models.Environment `json:"data"`
		}
		err = json.NewDecoder(w.Body).Decode(&listResp)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(listResp.Data), 1)
		envID := listResp.Data[0].ID

		// UPDATE: Update environment
		updateReq := map[string]interface{}{
			"name": "production",
		}
		body, _ = json.Marshal(updateReq)
		req = httptest.NewRequest(http.MethodPut,
			fmt.Sprintf("/api/v1/projects/%s/environments/%s", project.ID, envID),
			bytes.NewReader(body))
		addAuth(req, authToken)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// DELETE: Delete environment
		req = httptest.NewRequest(http.MethodDelete,
			fmt.Sprintf("/api/v1/projects/%s/environments/%s", project.ID, envID), nil)
		addAuth(req, authToken)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}
