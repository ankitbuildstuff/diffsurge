package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/diffsurge-org/diffsurge/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRateLimiter is a mock implementation of RateLimiter
type MockRateLimiter struct {
	mock.Mock
}

func (m *MockRateLimiter) CheckRateLimit(ctx context.Context, key string, limit int64, window time.Duration) (bool, int64, error) {
	args := m.Called(ctx, key, limit, window)
	return args.Bool(0), args.Get(1).(int64), args.Error(2)
}

func TestRateLimitMiddleware_AllowsWithinLimit(t *testing.T) {
	limiter := new(MockRateLimiter)
	config := DefaultRateLimitConfig()
	log := logger.New("error", "json")

	// Mock rate limiter to allow request (current: 1, limit: 100)
	limiter.On("CheckRateLimit", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(true, int64(1), nil)

	middleware := RateLimitMiddleware(limiter, config, log)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))

	req := httptest.NewRequest("GET", "/api/v1/projects", nil)
	// Add org context
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	ctx := context.WithValue(req.Context(), UserIDKey, userID)
	ctx = context.WithValue(ctx, ContextKeyTier, "pro")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "success", w.Body.String())

	// Check rate limit headers
	assert.Equal(t, "1000", w.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "999", w.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
}

func TestRateLimitMiddleware_BlocksWhenExceeded(t *testing.T) {
	limiter := new(MockRateLimiter)
	config := DefaultRateLimitConfig()
	log := logger.New("error", "json")

	// Mock rate limiter to deny request (current: 101, limit: 100)
	limiter.On("CheckRateLimit", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(false, int64(101), nil)

	middleware := RateLimitMiddleware(limiter, config, log)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))

	req := httptest.NewRequest("GET", "/api/v1/projects", nil)
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")
	ctx := context.WithValue(req.Context(), UserIDKey, userID)
	ctx = context.WithValue(ctx, ContextKeyTier, "free")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "Rate limit exceeded")

	// Check rate limit headers
	assert.Equal(t, "100", w.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "0", w.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, w.Header().Get("Retry-After"))
}

func TestRateLimitMiddleware_TierSpecificLimits(t *testing.T) {
	tests := []struct {
		name          string
		tier          string
		expectedLimit int64
	}{
		{
			name:          "Free tier",
			tier:          "free",
			expectedLimit: 100,
		},
		{
			name:          "Pro tier",
			tier:          "pro",
			expectedLimit: 1000,
		},
		{
			name:          "Enterprise tier",
			tier:          "enterprise",
			expectedLimit: 10000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := new(MockRateLimiter)
			config := DefaultRateLimitConfig()
			log := logger.New("error", "json")

			var capturedLimit int64
			limiter.On("CheckRateLimit", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Run(func(args mock.Arguments) {
					capturedLimit = args.Get(2).(int64)
				}).
				Return(true, int64(1), nil)

			middleware := RateLimitMiddleware(limiter, config, log)

			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/api/v1/projects", nil)
			userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440003")
			ctx := context.WithValue(req.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, ContextKeyTier, tt.tier)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedLimit, capturedLimit)
		})
	}
}

func TestRateLimitMiddleware_AuthEndpointsStricterLimit(t *testing.T) {
	limiter := new(MockRateLimiter)
	config := DefaultRateLimitConfig()
	log := logger.New("error", "json")

	var capturedKey string
	var capturedLimit int64

	limiter.On("CheckRateLimit", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			capturedKey = args.Get(1).(string)
			capturedLimit = args.Get(2).(int64)
		}).
		Return(true, int64(1), nil)

	middleware := RateLimitMiddleware(limiter, config, log)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	req.RemoteAddr = "192.168.1.1:1234"

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Contains(t, capturedKey, "auth:")
	assert.Contains(t, capturedKey, "192.168.1.1")
	assert.Equal(t, config.AuthLimit, capturedLimit)
}

func TestRateLimitMiddleware_SkipsHealthChecks(t *testing.T) {
	limiter := new(MockRateLimiter)
	config := DefaultRateLimitConfig()
	log := logger.New("error", "json")

	middleware := RateLimitMiddleware(limiter, config, log)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("healthy"))
	}))

	tests := []string{"/health", "/ready"}

	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "healthy", w.Body.String())

			// Verify rate limiter was not called
			limiter.AssertNotCalled(t, "CheckRateLimit")
		})
	}
}

func TestGetClientIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1")
	req.RemoteAddr = "192.168.1.1:1234"

	ip := GetClientIP(req)
	assert.Equal(t, "203.0.113.1", ip)
}

func TestGetClientIP_XRealIP(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Real-IP", "203.0.113.1")
	req.RemoteAddr = "192.168.1.1:1234"

	ip := GetClientIP(req)
	assert.Equal(t, "203.0.113.1", ip)
}

func TestGetClientIP_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"

	ip := GetClientIP(req)
	assert.Equal(t, "192.168.1.1", ip)
}

func TestGetLimitForTier(t *testing.T) {
	config := DefaultRateLimitConfig()

	tests := []struct {
		tier     string
		expected int64
	}{
		{"free", config.FreeLimit},
		{"pro", config.ProLimit},
		{"enterprise", config.EnterpriseLimit},
		{"unknown", config.DefaultLimit},
	}

	for _, tt := range tests {
		t.Run(tt.tier, func(t *testing.T) {
			limit := getLimitForTier(tt.tier, config)
			assert.Equal(t, tt.expected, limit)
		})
	}
}

func TestRateLimitMiddleware_UnauthenticatedRequest(t *testing.T) {
	limiter := new(MockRateLimiter)
	config := DefaultRateLimitConfig()
	log := logger.New("error", "json")

	var capturedKey string
	var capturedLimit int64

	limiter.On("CheckRateLimit", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			capturedKey = args.Get(1).(string)
			capturedLimit = args.Get(2).(int64)
		}).
		Return(true, int64(1), nil)

	middleware := RateLimitMiddleware(limiter, config, log)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/projects", nil)
	req.RemoteAddr = "192.168.1.1:1234"

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Unauthenticated requests should be rate limited by IP with stricter limit
	assert.Contains(t, capturedKey, "api:ip:")
	assert.Equal(t, config.FreeLimit/10, capturedLimit)
}

func TestRateLimitMiddleware_SchemaUpload(t *testing.T) {
	limiter := new(MockRateLimiter)
	config := DefaultRateLimitConfig()
	log := logger.New("error", "json")

	var capturedLimit int64
	var capturedWindow time.Duration

	limiter.On("CheckRateLimit", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			capturedLimit = args.Get(2).(int64)
			capturedWindow = args.Get(3).(time.Duration)
		}).
		Return(true, int64(1), nil)

	middleware := RateLimitMiddleware(limiter, config, log)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/v1/projects/123/schemas", nil)
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440004")
	ctx := context.WithValue(req.Context(), UserIDKey, userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Schema uploads should have slower rate limit (per hour)
	assert.Equal(t, config.UploadLimit, capturedLimit)
	assert.Equal(t, 1*time.Hour, capturedWindow)
}

func TestRateLimitMiddleware_ReplayStart(t *testing.T) {
	limiter := new(MockRateLimiter)
	config := DefaultRateLimitConfig()
	log := logger.New("error", "json")

	var capturedLimit int64

	limiter.On("CheckRateLimit", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			capturedLimit = args.Get(2).(int64)
		}).
		Return(true, int64(1), nil)

	middleware := RateLimitMiddleware(limiter, config, log)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/v1/projects/123/replays/456/start", nil)
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440005")
	ctx := context.WithValue(req.Context(), UserIDKey, userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Replay starts should have stricter limit
	assert.Equal(t, config.ReplayLimit, capturedLimit)
}

// Benchmark tests
func BenchmarkRateLimitMiddleware(b *testing.B) {
	limiter := new(MockRateLimiter)
	config := DefaultRateLimitConfig()
	log := logger.New("error", "json")

	limiter.On("CheckRateLimit", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(true, int64(1), nil)

	middleware := RateLimitMiddleware(limiter, config, log)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/projects", nil)
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440006")
	ctx := context.WithValue(req.Context(), UserIDKey, userID)
	ctx = context.WithValue(ctx, ContextKeyTier, "pro")
	req = req.WithContext(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}
