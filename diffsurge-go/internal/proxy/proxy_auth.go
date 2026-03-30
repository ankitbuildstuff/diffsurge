package proxy

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/diffsurge-org/diffsurge/internal/models"
	"github.com/diffsurge-org/diffsurge/pkg/logger"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Context keys for proxy auth — contextKey type is defined in middleware.go
const (
	ctxProjectID     contextKey = "proxy_project_id"
	ctxEnvironmentID contextKey = "proxy_environment_id"
	ctxOrgID         contextKey = "proxy_org_id"
)

// APIKeyStore defines the interface needed by the proxy auth middleware.
type APIKeyStore interface {
	GetAPIKeyByHash(ctx context.Context, keyPrefix string) (*models.APIKey, error)
	UpdateAPIKeyLastUsed(ctx context.Context, id uuid.UUID) error
}

// Auth validates API keys on proxy requests and resolves project/environment context.
type Auth struct {
	store APIKeyStore
	log   *logger.Logger
	// Fallback project/environment IDs from static config
	fallbackProjectID     uuid.UUID
	fallbackEnvironmentID uuid.UUID
}

// NewProxyAuth creates a new proxy auth middleware.
func NewProxyAuth(store APIKeyStore, log *logger.Logger) *Auth {
	return &Auth{
		store: store,
		log:   log,
	}
}

// SetFallback sets static fallback project/environment IDs for requests without an API key.
func (pa *Auth) SetFallback(projectID, environmentID uuid.UUID) {
	pa.fallbackProjectID = projectID
	pa.fallbackEnvironmentID = environmentID
}

// Middleware returns an http.Handler that validates API keys and injects project context.
func (pa *Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := extractAPIKey(r)

		if apiKey != "" && pa.store != nil {
			if err := pa.resolveAPIKey(r, apiKey); err != nil {
				pa.log.Debug().Err(err).Msg("Proxy API key validation failed")
				// Don't block traffic — just log and continue without project context
				// Traffic will still be captured but without project association
			}
		} else if pa.fallbackProjectID != uuid.Nil {
			// Use static fallback config
			ctx := r.Context()
			ctx = context.WithValue(ctx, ctxProjectID, pa.fallbackProjectID)
			if pa.fallbackEnvironmentID != uuid.Nil {
				ctx = context.WithValue(ctx, ctxEnvironmentID, pa.fallbackEnvironmentID)
			}
			*r = *r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

// resolveAPIKey validates the API key and sets project/environment context.
func (pa *Auth) resolveAPIKey(r *http.Request, fullKey string) error {
	if len(fullKey) < 16 {
		return fmt.Errorf("invalid API key format")
	}

	prefix := fullKey[:16]

	// Lookup key by prefix
	apiKey, err := pa.store.GetAPIKeyByHash(r.Context(), prefix)
	if err != nil {
		return fmt.Errorf("API key not found: %w", err)
	}

	// Verify the full key against stored hash
	if err := bcrypt.CompareHashAndPassword([]byte(apiKey.KeyHash), []byte(fullKey)); err != nil {
		return fmt.Errorf("API key mismatch")
	}

	// Check expiration
	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("API key expired")
	}

	// Update last used timestamp (async, don't block the request)
	go func() {
		_ = pa.store.UpdateAPIKeyLastUsed(context.Background(), apiKey.ID)
	}()

	// Set context with project info
	ctx := r.Context()
	ctx = context.WithValue(ctx, ctxOrgID, apiKey.OrganizationID)
	if apiKey.ProjectID != nil {
		ctx = context.WithValue(ctx, ctxProjectID, *apiKey.ProjectID)
	}
	*r = *r.WithContext(ctx)

	return nil
}

// extractAPIKey extracts the API key from request headers.
func extractAPIKey(r *http.Request) string {
	// Check X-API-Key header first
	if key := r.Header.Get("X-API-Key"); key != "" {
		return key
	}
	// Check Authorization: Bearer header
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		token := strings.TrimPrefix(auth, "Bearer ")
		if strings.HasPrefix(token, "diffsurge_live_") {
			return token
		}
	}
	return ""
}

// GetProjectIDFromContext retrieves the project ID from request context.
func GetProjectIDFromContext(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(ctxProjectID).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

// GetEnvironmentIDFromContext retrieves the environment ID from request context.
func GetEnvironmentIDFromContext(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(ctxEnvironmentID).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}
