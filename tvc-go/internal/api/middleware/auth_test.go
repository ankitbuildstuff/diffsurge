package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestExtractToken_Bearer(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer my-token-123")
	assert.Equal(t, "my-token-123", extractToken(r))
}

func TestExtractToken_APIKey(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-API-Key", "tvc_live_abc123")
	assert.Equal(t, "tvc_live_abc123", extractToken(r))
}

func TestExtractToken_Empty(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	assert.Equal(t, "", extractToken(r))
}

func TestGetUserID(t *testing.T) {
	uid := uuid.New()
	ctx := context.WithValue(context.Background(), UserIDKey, uid)
	assert.Equal(t, uid, GetUserID(ctx))
	assert.Equal(t, uuid.Nil, GetUserID(context.Background()))
}

func TestGetUserEmail(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserEmailKey, "test@example.com")
	assert.Equal(t, "test@example.com", GetUserEmail(ctx))
	assert.Equal(t, "", GetUserEmail(context.Background()))
}

func TestGetUserRole(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserRoleKey, "admin")
	assert.Equal(t, "admin", GetUserRole(ctx))
	assert.Equal(t, "", GetUserRole(context.Background()))
}

func TestHasMinRole(t *testing.T) {
	assert.True(t, hasMinRole("owner", "viewer"))
	assert.True(t, hasMinRole("admin", "admin"))
	assert.True(t, hasMinRole("owner", "owner"))
	assert.False(t, hasMinRole("viewer", "admin"))
	assert.False(t, hasMinRole("member", "admin"))
	assert.False(t, hasMinRole("unknown", "viewer"))
	assert.False(t, hasMinRole("admin", "unknown"))
}
