package middleware

import (
	"net/http"
	"strings"
)

// SecurityConfig holds configuration for security headers
type SecurityConfig struct {
	// Content Security Policy
	ContentSecurityPolicy string

	// Allowed origins for CORS (if empty, CORS headers won't be set)
	AllowedOrigins []string

	// Enable Strict Transport Security (HSTS)
	EnableHSTS bool
	HSTSMaxAge int // in seconds

	// Custom headers to add
	CustomHeaders map[string]string
}

// DefaultSecurityConfig returns sensible security defaults
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		// Strict CSP for production
		ContentSecurityPolicy: "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' https://js.stripe.com; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"font-src 'self' data:; " +
			"connect-src 'self' https://*.supabase.co https://api.stripe.com; " +
			"frame-src 'self' https://js.stripe.com; " +
			"object-src 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self'; " +
			"frame-ancestors 'none'; " +
			"upgrade-insecure-requests;",

		EnableHSTS: true,
		HSTSMaxAge: 31536000, // 1 year

		CustomHeaders: make(map[string]string),
	}
}

// DevelopmentSecurityConfig returns relaxed settings for development
func DevelopmentSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		// More permissive CSP for development
		ContentSecurityPolicy: "default-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
			"connect-src 'self' http://localhost:* ws://localhost:* https://*.supabase.co; " +
			"img-src 'self' data: https:;",

		EnableHSTS: false, // Don't enforce HTTPS in development

		AllowedOrigins: []string{"http://localhost:3000", "http://localhost:3001"},

		CustomHeaders: make(map[string]string),
	}
}

// SecurityHeadersMiddleware adds security headers to all responses
func SecurityHeadersMiddleware(config *SecurityConfig) func(http.Handler) http.Handler {
	if config == nil {
		config = DefaultSecurityConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Prevent MIME type sniffing
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// Prevent clickjacking
			w.Header().Set("X-Frame-Options", "DENY")

			// Disable XSS filtering (rely on CSP instead)
			// Modern browsers deprecated X-XSS-Protection
			w.Header().Set("X-XSS-Protection", "0")

			// Control referrer information
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Content Security Policy
			if config.ContentSecurityPolicy != "" {
				w.Header().Set("Content-Security-Policy", config.ContentSecurityPolicy)
			}

			// Permissions Policy (formerly Feature-Policy)
			w.Header().Set("Permissions-Policy",
				"camera=(), microphone=(), geolocation=(), payment=(self), usb=()")

			// Strict Transport Security (HTTPS only)
			if config.EnableHSTS && r.TLS != nil {
				hstsValue := "max-age=" + string(rune(config.HSTSMaxAge)) + "; includeSubDomains; preload"
				w.Header().Set("Strict-Transport-Security", hstsValue)
			}

			// Expect-CT (Certificate Transparency)
			if r.TLS != nil {
				w.Header().Set("Expect-CT", "max-age=86400, enforce")
			}

			// Remove server identification
			w.Header().Set("Server", "TVC")

			// Add custom headers
			for key, value := range config.CustomHeaders {
				w.Header().Set(key, value)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware(allowedOrigins []string, allowCredentials bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			if origin != "" && isOriginAllowed(origin, allowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			// Allow credentials (cookies, auth headers)
			if allowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers",
					"Authorization, Content-Type, X-Request-ID, X-API-Key, X-Organization-ID")
				w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// NoSniffMiddleware prevents MIME type sniffing
func NoSniffMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}

// NoFrameMiddleware prevents the page from being embedded in frames
func NoFrameMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

// Helper function to check if origin is in allowed list
func isOriginAllowed(origin string, allowed []string) bool {
	// If no origins specified, allow none
	if len(allowed) == 0 {
		return false
	}

	// Check exact matches
	for _, allowedOrigin := range allowed {
		if allowedOrigin == "*" {
			return true
		}
		if allowedOrigin == origin {
			return true
		}
		// Support wildcard subdomains (e.g., "https://*.example.com")
		if strings.Contains(allowedOrigin, "*") {
			if matchWildcard(origin, allowedOrigin) {
				return true
			}
		}
	}

	return false
}

// matchWildcard checks if a string matches a wildcard pattern
func matchWildcard(s, pattern string) bool {
	// Simple wildcard matching for origins
	// Pattern like "https://*.example.com" should match "https://api.example.com"

	if !strings.Contains(pattern, "*") {
		return s == pattern
	}

	parts := strings.Split(pattern, "*")
	if len(parts) != 2 {
		return false
	}

	return strings.HasPrefix(s, parts[0]) && strings.HasSuffix(s, parts[1])
}

// SecureRedirect ensures redirects are safe and don't lead to open redirect vulnerabilities
func SecureRedirect(w http.ResponseWriter, r *http.Request, url string, code int) {
	// Only allow relative URLs or same-origin URLs
	if strings.HasPrefix(url, "/") {
		// Relative URL is safe
		http.Redirect(w, r, url, code)
		return
	}

	// Check if URL is same-origin
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		// Extract origin from request
		requestOrigin := getRequestOrigin(r)
		if strings.HasPrefix(url, requestOrigin) {
			http.Redirect(w, r, url, code)
			return
		}
	}

	// Unsafe redirect attempt - redirect to home instead
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// getRequestOrigin extracts the origin from a request
func getRequestOrigin(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	// Check X-Forwarded-Proto header
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}

	return scheme + "://" + r.Host
}
