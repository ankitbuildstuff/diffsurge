package middleware

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tvc-org/tvc/internal/api/response"
	"github.com/tvc-org/tvc/internal/storage"
	"github.com/tvc-org/tvc/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

const (
	UserIDKey   contextKey = "user_id"
	UserEmailKey contextKey = "user_email"
	UserRoleKey  contextKey = "user_role"
)

type AuthConfig struct {
	SupabaseURL    string
	SupabaseSecret string
	JWTSecret      string
}

type JWKSCache struct {
	mu      sync.RWMutex
	keys    map[string]*rsa.PublicKey
	fetched time.Time
	ttl     time.Duration
	url     string
}

type Auth struct {
	config    AuthConfig
	log       *logger.Logger
	jwksCache *JWKSCache
	store     storage.Repository
}

func NewAuth(cfg AuthConfig, log *logger.Logger, store storage.Repository) *Auth {
	a := &Auth{
		config: cfg,
		log:    log,
		store:  store,
	}

	if cfg.SupabaseURL != "" {
		a.jwksCache = &JWKSCache{
			keys: make(map[string]*rsa.PublicKey),
			ttl:  6 * time.Hour,
			url:  strings.TrimRight(cfg.SupabaseURL, "/") + "/auth/v1/.well-known/jwks.json",
		}
	}

	return a
}

func (a *Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			response.Unauthorized(w, "Missing authorization token")
			return
		}

		// Check if it's an API key (starts with tvc_live_)
		if strings.HasPrefix(token, "tvc_live_") {
			if err := a.validateAPIKey(r, token); err != nil {
				a.log.Debug().Err(err).Msg("API key validation failed")
				response.Unauthorized(w, "Invalid or expired API key")
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		// Otherwise validate as JWT
		claims, err := a.validateToken(token)
		if err != nil {
			a.log.Debug().Err(err).Msg("token validation failed")
			response.Unauthorized(w, "Invalid or expired token")
			return
		}

		ctx := r.Context()
		if sub, ok := claims["sub"].(string); ok {
			if uid, err := uuid.Parse(sub); err == nil {
				ctx = context.WithValue(ctx, UserIDKey, uid)
			}
		}
		if email, ok := claims["email"].(string); ok {
			ctx = context.WithValue(ctx, UserEmailKey, email)
		}
		if role, ok := claims["role"].(string); ok {
			ctx = context.WithValue(ctx, UserRoleKey, role)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *Auth) validateAPIKey(r *http.Request, fullKey string) error {
	if len(fullKey) < 16 {
		return fmt.Errorf("invalid API key format")
	}

	prefix := fullKey[:16]
	
	// Lookup key by prefix
	apiKey, err := a.store.GetAPIKeyByHash(r.Context(), prefix)
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

	// Update last used timestamp (async, don't wait)
	go a.store.UpdateAPIKeyLastUsed(context.Background(), apiKey.ID)

	// Set org context from API key
	ctx := context.WithValue(r.Context(), UserIDKey, apiKey.OrganizationID)
	ctx = context.WithValue(ctx, UserRoleKey, "api_key")
	*r = *r.WithContext(ctx)

	return nil
}

// RequireRole creates middleware that enforces a minimum role level.
// Role hierarchy: owner > admin > member > viewer
func RequireRole(minRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := GetUserRole(r.Context())
			if !hasMinRole(role, minRole) {
				response.Forbidden(w, "")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (a *Auth) validateToken(tokenStr string) (jwt.MapClaims, error) {
	// Try HMAC (HS256) with JWT secret first — simpler setup for development
	if a.config.JWTSecret != "" {
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); ok {
				return []byte(a.config.JWTSecret), nil
			}
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		})
		if err == nil && token.Valid {
			return claims, nil
		}
	}

	// Try RSA (RS256) with JWKS from Supabase
	if a.jwksCache != nil {
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			kid, _ := t.Header["kid"].(string)
			key, err := a.getPublicKey(kid)
			if err != nil {
				return nil, err
			}
			return key, nil
		})
		if err == nil && token.Valid {
			return claims, nil
		}
		return nil, err
	}

	return nil, fmt.Errorf("no valid signing key configured")
}

func (a *Auth) getPublicKey(kid string) (*rsa.PublicKey, error) {
	a.jwksCache.mu.RLock()
	if key, ok := a.jwksCache.keys[kid]; ok && time.Since(a.jwksCache.fetched) < a.jwksCache.ttl {
		a.jwksCache.mu.RUnlock()
		return key, nil
	}
	a.jwksCache.mu.RUnlock()

	return a.refreshJWKS(kid)
}

type jwksResponse struct {
	Keys []jwkKey `json:"keys"`
}

type jwkKey struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func (a *Auth) refreshJWKS(kid string) (*rsa.PublicKey, error) {
	a.jwksCache.mu.Lock()
	defer a.jwksCache.mu.Unlock()

	if key, ok := a.jwksCache.keys[kid]; ok && time.Since(a.jwksCache.fetched) < a.jwksCache.ttl {
		return key, nil
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(a.jwksCache.url)
	if err != nil {
		return nil, fmt.Errorf("fetching JWKS: %w", err)
	}
	defer resp.Body.Close()

	var jwks jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("decoding JWKS: %w", err)
	}

	a.jwksCache.keys = make(map[string]*rsa.PublicKey)
	for _, k := range jwks.Keys {
		if k.Kty != "RSA" {
			continue
		}
		pubKey, err := parseRSAPublicKey(k)
		if err != nil {
			continue
		}
		a.jwksCache.keys[k.Kid] = pubKey
	}
	a.jwksCache.fetched = time.Now()

	key, ok := a.jwksCache.keys[kid]
	if !ok {
		return nil, fmt.Errorf("key %s not found in JWKS", kid)
	}
	return key, nil
}

func parseRSAPublicKey(k jwkKey) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, err
	}

	n := new(big.Int).SetBytes(nBytes)
	e := 0
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}

	return &rsa.PublicKey{N: n, E: e}, nil
}

func extractToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	if key := r.Header.Get("X-API-Key"); key != "" {
		return key
	}
	return ""
}

func GetUserID(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(UserIDKey).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

func GetUserEmail(ctx context.Context) string {
	if email, ok := ctx.Value(UserEmailKey).(string); ok {
		return email
	}
	return ""
}

func GetUserRole(ctx context.Context) string {
	if role, ok := ctx.Value(UserRoleKey).(string); ok {
		return role
	}
	return ""
}

// AuthExempt wraps auth middleware but exempts specific paths from authentication.
func AuthExempt(auth *Auth, inner http.Handler, exemptPaths ...string) http.Handler {
	exempt := make(map[string]bool, len(exemptPaths))
	for _, p := range exemptPaths {
		exempt[p] = true
	}

	authedHandler := auth.Middleware(inner)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if exempt[r.URL.Path] {
			inner.ServeHTTP(w, r)
			return
		}
		authedHandler.ServeHTTP(w, r)
	})
}

var roleHierarchy = map[string]int{
	"viewer": 0,
	"member": 1,
	"admin":  2,
	"owner":  3,
}

func hasMinRole(userRole, minRole string) bool {
	userLevel, ok := roleHierarchy[userRole]
	if !ok {
		return false
	}
	minLevel, ok := roleHierarchy[minRole]
	if !ok {
		return false
	}
	return userLevel >= minLevel
}
