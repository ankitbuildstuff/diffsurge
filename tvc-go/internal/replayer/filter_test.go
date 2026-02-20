package replayer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tvc-org/tvc/internal/models"
)

func TestFilterTrafficForReplay_StripsSensitiveHeaders(t *testing.T) {
	logs := []models.TrafficLog{
		{
			Method: "GET",
			Path:   "/api/users",
			RequestHeaders: map[string]interface{}{
				"Authorization": "Bearer secret-token",
				"Accept":        "application/json",
				"Cookie":        "session=abc123",
				"X-Api-Key":     "key-456",
			},
		},
	}

	filtered := FilterTrafficForReplay(logs, FilterConfig{
		StripSensitiveHeaders: true,
	})

	require.Len(t, filtered, 1)
	assert.NotContains(t, filtered[0].RequestHeaders, "Authorization")
	assert.NotContains(t, filtered[0].RequestHeaders, "Cookie")
	assert.NotContains(t, filtered[0].RequestHeaders, "X-Api-Key")
	assert.Contains(t, filtered[0].RequestHeaders, "Accept")
}

func TestFilterTrafficForReplay_PreservesHeadersWhenNotStripping(t *testing.T) {
	logs := []models.TrafficLog{
		{
			Method: "GET",
			Path:   "/api/users",
			RequestHeaders: map[string]interface{}{
				"Authorization": "Bearer secret-token",
				"Accept":        "application/json",
			},
		},
	}

	filtered := FilterTrafficForReplay(logs, FilterConfig{
		StripSensitiveHeaders: false,
	})

	require.Len(t, filtered, 1)
	assert.Contains(t, filtered[0].RequestHeaders, "Authorization")
}

func TestFilterTrafficForReplay_ReplacesHeaders(t *testing.T) {
	logs := []models.TrafficLog{
		{
			Method: "GET",
			Path:   "/api/users",
			RequestHeaders: map[string]interface{}{
				"Authorization": "Bearer old-token",
				"Accept":        "application/json",
			},
		},
	}

	filtered := FilterTrafficForReplay(logs, FilterConfig{
		HeaderReplacements: HeaderReplacements{
			"Authorization": "Bearer test-token",
		},
	})

	require.Len(t, filtered, 1)
	assert.Equal(t, "Bearer test-token", filtered[0].RequestHeaders["Authorization"])
}

func TestFilterTrafficForReplay_ExcludesPaths(t *testing.T) {
	logs := []models.TrafficLog{
		{Method: "GET", Path: "/api/users"},
		{Method: "GET", Path: "/health"},
		{Method: "GET", Path: "/metrics"},
		{Method: "GET", Path: "/api/orders"},
		{Method: "GET", Path: "/internal/debug/pprof"},
	}

	filtered := FilterTrafficForReplay(logs, FilterConfig{
		ExcludePaths: []string{"/health", "/metrics", "/internal/*"},
	})

	require.Len(t, filtered, 2)
	assert.Equal(t, "/api/users", filtered[0].Path)
	assert.Equal(t, "/api/orders", filtered[1].Path)
}

func TestFilterTrafficForReplay_DoesNotMutateOriginal(t *testing.T) {
	logs := []models.TrafficLog{
		{
			Method: "GET",
			Path:   "/api/users",
			RequestHeaders: map[string]interface{}{
				"Authorization": "Bearer secret",
			},
		},
	}

	_ = FilterTrafficForReplay(logs, FilterConfig{
		StripSensitiveHeaders: true,
	})

	assert.Contains(t, logs[0].RequestHeaders, "Authorization",
		"Original log should not be mutated")
}

func TestFilterTrafficForReplay_EmptyInput(t *testing.T) {
	filtered := FilterTrafficForReplay(nil, FilterConfig{})
	assert.Empty(t, filtered)
}

func TestFilterTrafficForReplay_NilHeaders(t *testing.T) {
	logs := []models.TrafficLog{
		{Method: "GET", Path: "/api/users", RequestHeaders: nil},
	}

	filtered := FilterTrafficForReplay(logs, FilterConfig{
		StripSensitiveHeaders: true,
	})

	require.Len(t, filtered, 1)
	assert.Nil(t, filtered[0].RequestHeaders)
}

func TestShouldExclude(t *testing.T) {
	tests := []struct {
		path     string
		excludes []string
		expected bool
	}{
		{"/health", []string{"/health"}, true},
		{"/api/users", []string{"/health"}, false},
		{"/internal/debug", []string{"/internal/*"}, true},
		{"/internal", []string{"/internal/*"}, false},
		{"/api/users", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.expected, shouldExclude(tt.path, tt.excludes))
		})
	}
}
