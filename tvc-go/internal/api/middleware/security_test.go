package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecurityHeadersMiddleware_DefaultHeaders(t *testing.T) {
	config := DefaultSecurityConfig()
	middleware := SecurityHeadersMiddleware(config)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Check all security headers are set
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "0", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
	assert.NotEmpty(t, w.Header().Get("Content-Security-Policy"))
	assert.NotEmpty(t, w.Header().Get("Permissions-Policy"))
	assert.Equal(t, "TVC", w.Header().Get("Server"))
}

func TestSecurityHeadersMiddleware_CustomHeaders(t *testing.T) {
	config := DefaultSecurityConfig()
	config.CustomHeaders["X-Custom-Header"] = "CustomValue"

	middleware := SecurityHeadersMiddleware(config)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, "CustomValue", w.Header().Get("X-Custom-Header"))
}

func TestSecurityHeadersMiddleware_DevelopmentMode(t *testing.T) {
	config := DevelopmentSecurityConfig()
	middleware := SecurityHeadersMiddleware(config)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// HSTS should not be set in development
	assert.Empty(t, w.Header().Get("Strict-Transport-Security"))

	// But other headers should still be present
	assert.NotEmpty(t, w.Header().Get("Content-Security-Policy"))
}

func TestCORSMiddleware_AllowedOrigin(t *testing.T) {
	allowedOrigins := []string{"https://app.example.com", "https://admin.example.com"}
	middleware := CORSMiddleware(allowedOrigins, true)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://app.example.com")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, "https://app.example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCORSMiddleware_DisallowedOrigin(t *testing.T) {
	allowedOrigins := []string{"https://app.example.com"}
	middleware := CORSMiddleware(allowedOrigins, true)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://evil.com")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSMiddleware_Preflight(t *testing.T) {
	allowedOrigins := []string{"https://app.example.com"}
	middleware := CORSMiddleware(allowedOrigins, true)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Handler should not be called for OPTIONS request")
	}))

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://app.example.com")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Methods"))
	assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))
}

func TestCORSMiddleware_WildcardOrigin(t *testing.T) {
	allowedOrigins := []string{"https://*.example.com"}
	middleware := CORSMiddleware(allowedOrigins, true)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		origin  string
		allowed bool
	}{
		{"https://app.example.com", true},
		{"https://api.example.com", true},
		{"https://admin.example.com", true},
		{"https://example.com", false},    // No subdomain
		{"https://evil.com", false},       // Different domain
		{"http://app.example.com", false}, // Different protocol
	}

	for _, tt := range tests {
		t.Run(tt.origin, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Origin", tt.origin)

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if tt.allowed {
				assert.Equal(t, tt.origin, w.Header().Get("Access-Control-Allow-Origin"))
			} else {
				assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
			}
		})
	}
}

func TestCORSMiddleware_AllowAll(t *testing.T) {
	allowedOrigins := []string{"*"}
	middleware := CORSMiddleware(allowedOrigins, false)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://any-origin.com")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, "https://any-origin.com", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestNoSniffMiddleware(t *testing.T) {
	middleware := NoSniffMiddleware

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
}

func TestNoFrameMiddleware(t *testing.T) {
	middleware := NoFrameMiddleware

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
}

func TestIsOriginAllowed(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		allowedOrigins []string
		expected       bool
	}{
		{
			name:           "Exact match",
			origin:         "https://app.example.com",
			allowedOrigins: []string{"https://app.example.com"},
			expected:       true,
		},
		{
			name:           "Wildcard match",
			origin:         "https://api.example.com",
			allowedOrigins: []string{"https://*.example.com"},
			expected:       true,
		},
		{
			name:           "Wildcard all",
			origin:         "https://any.com",
			allowedOrigins: []string{"*"},
			expected:       true,
		},
		{
			name:           "No match",
			origin:         "https://evil.com",
			allowedOrigins: []string{"https://app.example.com"},
			expected:       false,
		},
		{
			name:           "Empty allowed list",
			origin:         "https://app.example.com",
			allowedOrigins: []string{},
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isOriginAllowed(tt.origin, tt.allowedOrigins)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMatchWildcard(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		pattern  string
		expected bool
	}{
		{
			name:     "No wildcard - exact match",
			s:        "https://app.example.com",
			pattern:  "https://app.example.com",
			expected: true,
		},
		{
			name:     "No wildcard - no match",
			s:        "https://api.example.com",
			pattern:  "https://app.example.com",
			expected: false,
		},
		{
			name:     "Wildcard subdomain match",
			s:        "https://api.example.com",
			pattern:  "https://*.example.com",
			expected: true,
		},
		{
			name:     "Wildcard subdomain no match",
			s:        "https://example.com",
			pattern:  "https://*.example.com",
			expected: false,
		},
		{
			name:     "Wildcard path match",
			s:        "https://example.com/api/v1",
			pattern:  "https://example.com/*",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchWildcard(tt.s, tt.pattern)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecureRedirect_RelativeURL(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	SecureRedirect(w, req, "/dashboard", http.StatusFound)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/dashboard", w.Header().Get("Location"))
}

func TestSecureRedirect_SameOriginURL(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	req.Host = "example.com"
	w := httptest.NewRecorder()

	SecureRedirect(w, req, "http://example.com/dashboard", http.StatusFound)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "http://example.com/dashboard", w.Header().Get("Location"))
}

func TestSecureRedirect_DifferentOriginURL(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	req.Host = "example.com"
	w := httptest.NewRecorder()

	// Attempt to redirect to different origin - should redirect to home instead
	SecureRedirect(w, req, "http://evil.com/phishing", http.StatusFound)

	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/", w.Header().Get("Location"))
}

func TestGetRequestOrigin(t *testing.T) {
	tests := []struct {
		name     string
		setupReq func() *http.Request
		expected string
	}{
		{
			name: "HTTP request",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Host = "example.com"
				return req
			},
			expected: "http://example.com",
		},
		{
			name: "HTTPS request with X-Forwarded-Proto",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Host = "example.com"
				req.Header.Set("X-Forwarded-Proto", "https")
				return req
			},
			expected: "https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupReq()
			origin := getRequestOrigin(req)
			assert.Equal(t, tt.expected, origin)
		})
	}
}

// Benchmark tests
func BenchmarkSecurityHeadersMiddleware(b *testing.B) {
	config := DefaultSecurityConfig()
	middleware := SecurityHeadersMiddleware(config)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkCORSMiddleware(b *testing.B) {
	allowedOrigins := []string{"https://app.example.com", "https://api.example.com"}
	middleware := CORSMiddleware(allowedOrigins, true)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://app.example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}
