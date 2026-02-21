package api

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tvc-org/tvc/internal/api/handlers"
	"github.com/tvc-org/tvc/internal/api/middleware"
	"github.com/tvc-org/tvc/internal/storage"
	"github.com/tvc-org/tvc/pkg/logger"
)

type ServerDeps struct {
	Store      storage.Repository
	Log        *logger.Logger
	AuthConfig middleware.AuthConfig
}

func NewRouter(deps ServerDeps) http.Handler {
	mux := http.NewServeMux()

	projects := handlers.NewProjectHandler(deps.Store, deps.Log)
	traffic := handlers.NewTrafficHandler(deps.Store, deps.Log)
	environments := handlers.NewEnvironmentHandler(deps.Store, deps.Log)
	replays := handlers.NewReplayHandler(deps.Store, deps.Log)
	schemas := handlers.NewSchemaHandler(deps.Store, deps.Log)
	orgs := handlers.NewOrganizationHandler(deps.Store, deps.Log)
	apiKeys := handlers.NewAPIKeyHandler(deps.Store, deps.Log)
	audit := handlers.NewAuditLogHandler(deps.Store, deps.Log)
	health := handlers.NewHealthHandler(deps.Store)

	// Health endpoints (no auth)
	mux.HandleFunc("GET /api/v1/health", health.Health)
	mux.HandleFunc("GET /api/v1/ready", health.Ready)

	// Metrics endpoint (no auth - should be restricted via firewall in production)
	mux.Handle("GET /metrics", promhttp.Handler())

	// Projects
	mux.HandleFunc("GET /api/v1/projects", projects.List)
	mux.HandleFunc("POST /api/v1/projects", projects.Create)
	mux.HandleFunc("GET /api/v1/projects/{id}", projects.Get)
	mux.HandleFunc("PUT /api/v1/projects/{id}", projects.Update)
	mux.HandleFunc("DELETE /api/v1/projects/{id}", projects.Delete)

	// Traffic
	mux.HandleFunc("GET /api/v1/projects/{id}/traffic", traffic.List)
	mux.HandleFunc("GET /api/v1/projects/{id}/traffic/stats", traffic.Stats)
	mux.HandleFunc("GET /api/v1/projects/{id}/traffic/{logId}", traffic.Get)

	// Environments
	mux.HandleFunc("GET /api/v1/projects/{id}/environments", environments.List)
	mux.HandleFunc("POST /api/v1/projects/{id}/environments", environments.Create)
	mux.HandleFunc("GET /api/v1/projects/{id}/environments/{envId}", environments.Get)
	mux.HandleFunc("PUT /api/v1/projects/{id}/environments/{envId}", environments.Update)
	mux.HandleFunc("DELETE /api/v1/projects/{id}/environments/{envId}", environments.Delete)

	// Replays
	mux.HandleFunc("GET /api/v1/projects/{id}/replays", replays.List)
	mux.HandleFunc("POST /api/v1/projects/{id}/replays", replays.Create)
	mux.HandleFunc("GET /api/v1/projects/{id}/replays/{replayId}", replays.Get)
	mux.HandleFunc("POST /api/v1/projects/{id}/replays/{replayId}/start", replays.Start)
	mux.HandleFunc("POST /api/v1/projects/{id}/replays/{replayId}/stop", replays.Stop)
	mux.HandleFunc("GET /api/v1/projects/{id}/replays/{replayId}/results", replays.Results)

	// Schemas
	mux.HandleFunc("GET /api/v1/projects/{id}/schemas", schemas.List)
	mux.HandleFunc("POST /api/v1/projects/{id}/schemas", schemas.Upload)
	mux.HandleFunc("POST /api/v1/projects/{id}/schemas/diff", schemas.Diff)

	// Organizations
	mux.HandleFunc("GET /api/v1/organizations", orgs.List)
	mux.HandleFunc("POST /api/v1/organizations", orgs.Create)
	mux.HandleFunc("GET /api/v1/organizations/{id}", orgs.Get)
	mux.HandleFunc("PUT /api/v1/organizations/{id}", orgs.Update)
	mux.HandleFunc("DELETE /api/v1/organizations/{id}", orgs.Delete)
	mux.HandleFunc("GET /api/v1/organizations/{id}/members", orgs.ListMembers)
	mux.HandleFunc("POST /api/v1/organizations/{id}/members", orgs.AddMember)
	mux.HandleFunc("DELETE /api/v1/organizations/{id}/members/{userId}", orgs.RemoveMember)

	// API Keys
	mux.HandleFunc("GET /api/v1/organizations/{id}/api-keys", apiKeys.List)
	mux.HandleFunc("POST /api/v1/organizations/{id}/api-keys", apiKeys.Create)
	mux.HandleFunc("DELETE /api/v1/organizations/{id}/api-keys/{keyId}", apiKeys.Delete)

	// Audit Logs
	mux.HandleFunc("GET /api/v1/organizations/{id}/audit-logs", audit.List)

	// Build middleware chain (applied in reverse order)
	auth := middleware.NewAuth(deps.AuthConfig, deps.Log, deps.Store)
	var handler http.Handler = mux
	handler = middleware.AuthExempt(auth, handler,
		"/api/v1/health",
		"/api/v1/ready",
		"/metrics",
	)
	handler = middleware.PrometheusMiddleware()(handler)
	handler = middleware.Recovery(deps.Log)(handler)
	handler = middleware.Logging(deps.Log)(handler)
	handler = middleware.RequestID(handler)
	handler = middleware.CORS(middleware.DefaultCORSConfig())(handler)

	return handler
}
