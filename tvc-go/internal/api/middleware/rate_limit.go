package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/tvc-org/tvc/internal/api/response"
	"github.com/tvc-org/tvc/pkg/logger"
)

// RateLimiter defines the interface for rate limiting storage
type RateLimiter interface {
	CheckRateLimit(ctx context.Context, key string, limit int64, window time.Duration) (allowed bool, current int64, err error)
}

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	// Default limits
	DefaultLimit  int64
	DefaultWindow time.Duration

	// Tier-specific limits (requests per window)
	FreeLimit       int64
	ProLimit        int64
	EnterpriseLimit int64

	// Special endpoint limits
	AuthLimit       int64  // Login attempts
	UploadLimit     int64  // Schema uploads
	ReplayLimit     int64  // Replay starts
}

// DefaultRateLimitConfig returns sensible defaults
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		DefaultLimit:    100,
		DefaultWindow:   1 * time.Minute,
		FreeLimit:       100,   // 100 req/min for free tier
		ProLimit:        1000,  // 1000 req/min for pro tier
		EnterpriseLimit: 10000, // 10000 req/min for enterprise
		AuthLimit:       5,     // 5 login attempts per minute
		UploadLimit:     10,    // 10 schema uploads per hour
		ReplayLimit:     5,     // 5 replay starts per minute
	}
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(limiter RateLimiter, config *RateLimitConfig, log *logger.Logger) func(http.Handler) http.Handler {
	if config == nil {
		config = DefaultRateLimitConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			uid := GetUserID(r.Context())
			authenticated := uid != uuid.Nil
			userID := uid.String()
			tier := GetTierFromContext(r.Context())

			var key string
			var limit int64
			window := config.DefaultWindow

			switch {
			case r.URL.Path == "/api/v1/auth/login":
				key = fmt.Sprintf("auth:%s", GetClientIP(r))
				limit = config.AuthLimit

			case r.URL.Path == "/api/v1/auth/signup":
				key = fmt.Sprintf("signup:%s", GetClientIP(r))
				limit = config.AuthLimit

			case r.Method == "POST" && contains(r.URL.Path, "/schemas"):
				key = fmt.Sprintf("upload:%s", userID)
				limit = config.UploadLimit
				window = 1 * time.Hour

			case r.Method == "POST" && contains(r.URL.Path, "/replays") && contains(r.URL.Path, "/start"):
				key = fmt.Sprintf("replay:%s", userID)
				limit = config.ReplayLimit

			default:
				if authenticated {
					key = fmt.Sprintf("api:user:%s", userID)
					limit = getLimitForTier(tier, config)
				} else {
					key = fmt.Sprintf("api:ip:%s", GetClientIP(r))
					limit = config.FreeLimit / 10
				}
			}

			// Skip rate limiting for health checks
			if r.URL.Path == "/health" || r.URL.Path == "/ready" {
				next.ServeHTTP(w, r)
				return
			}

			// Check rate limit
			ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
			defer cancel()

			allowed, current, err := limiter.CheckRateLimit(ctx, key, limit, window)
			if err != nil {
				log.Error().Err(err).Str("key", key).Msg("Rate limit check failed")
				// On error, allow the request but log it
				next.ServeHTTP(w, r)
				return
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(limit, 10))
			w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(max(0, limit-current), 10))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(window).Unix(), 10))

			if !allowed {
				// Rate limit exceeded
				log.Warn().
					Str("key", key).
					Int64("limit", limit).
					Int64("current", current).
					Str("path", r.URL.Path).
					Msg("Rate limit exceeded")

				w.Header().Set("Retry-After", strconv.FormatInt(int64(window.Seconds()), 10))
				
				response.RateLimited(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getLimitForTier returns the rate limit for a subscription tier
func getLimitForTier(tier string, config *RateLimitConfig) int64 {
	switch tier {
	case "free":
		return config.FreeLimit
	case "pro":
		return config.ProLimit
	case "enterprise":
		return config.EnterpriseLimit
	default:
		return config.DefaultLimit
	}
}

// GetTierFromContext retrieves subscription tier from context
func GetTierFromContext(ctx context.Context) string {
	if tier, ok := ctx.Value(ContextKeyTier).(string); ok {
		return tier
	}
	return "free"
}

// GetClientIP extracts the real client IP from request
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (proxy/load balancer)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		return splitFirst(xff, ",")
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fallback to RemoteAddr
	return splitFirst(r.RemoteAddr, ":")
}

// Helper functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func splitFirst(s, sep string) string {
	for i := 0; i < len(s); i++ {
		if len(s[i:]) >= len(sep) && s[i:i+len(sep)] == sep {
			return s[:i]
		}
	}
	return s
}

const (
	ContextKeyTier contextKey = "tier"
)
