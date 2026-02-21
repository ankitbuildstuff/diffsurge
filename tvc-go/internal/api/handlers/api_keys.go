package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tvc-org/tvc/internal/api/request"
	"github.com/tvc-org/tvc/internal/api/response"
	"github.com/tvc-org/tvc/internal/models"
	"github.com/tvc-org/tvc/internal/storage"
	"github.com/tvc-org/tvc/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

type APIKeyHandler struct {
	store storage.Repository
	log   *logger.Logger
}

func NewAPIKeyHandler(store storage.Repository, log *logger.Logger) *APIKeyHandler {
	return &APIKeyHandler{store: store, log: log}
}

type createAPIKeyRequest struct {
	Name      string     `json:"name"`
	ProjectID *uuid.UUID `json:"project_id"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type apiKeyResponse struct {
	Key    string         `json:"key"`
	APIKey *models.APIKey `json:"api_key"`
}

// generateAPIKey creates a new API key with format tvc_live_{32_random_bytes_base64}
func generateAPIKey() (string, string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", err
	}
	
	keySecret := base64.RawURLEncoding.EncodeToString(randomBytes)
	fullKey := "tvc_live_" + keySecret
	prefix := fullKey[:16] // Store prefix for display (tvc_live_ + first 6 chars)
	
	return fullKey, prefix, nil
}

func (h *APIKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	orgID, err := request.PathUUID(r, "id")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	var req createAPIKeyRequest
	if err := request.ParseJSON(r, 0, &req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	if req.Name == "" || len(req.Name) > 100 {
		response.ValidationError(w, []response.FieldError{
			{Field: "name", Message: "Name is required (1-100 characters)"},
		})
		return
	}

	// Generate API key
	fullKey, prefix, err := generateAPIKey()
	if err != nil {
		h.log.Error().Err(err).Msg("failed to generate API key")
		response.InternalError(w)
		return
	}

	// Hash the key for storage
	keyHash, err := bcrypt.GenerateFromPassword([]byte(fullKey), bcrypt.DefaultCost)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to hash API key")
		response.InternalError(w)
		return
	}

	apiKey := &models.APIKey{
		ID:             uuid.New(),
		OrganizationID: orgID,
		ProjectID:      req.ProjectID,
		Name:           req.Name,
		KeyPrefix:      prefix,
		KeyHash:        string(keyHash),
		ExpiresAt:      req.ExpiresAt,
		CreatedAt:      time.Now(),
		// CreatedBy would come from auth context in production
	}

	if err := h.store.CreateAPIKey(r.Context(), apiKey); err != nil {
		h.log.Error().Err(err).Msg("failed to create API key")
		response.InternalError(w)
		return
	}

	// Return the full key only once (never stored or shown again)
	response.Created(w, apiKeyResponse{
		Key:    fullKey,
		APIKey: apiKey,
	})
}

func (h *APIKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID, err := request.PathUUID(r, "id")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	keys, err := h.store.ListAPIKeys(r.Context(), orgID)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to list API keys")
		response.InternalError(w)
		return
	}

	if keys == nil {
		keys = []models.APIKey{}
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": keys,
	})
}

func (h *APIKeyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	orgID, err := request.PathUUID(r, "id")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	keyID, err := request.PathUUID(r, "keyId")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	// Verify the key belongs to this org
	key, err := h.store.GetAPIKey(r.Context(), keyID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, "API Key")
			return
		}
		h.log.Error().Err(err).Msg("failed to get API key")
		response.InternalError(w)
		return
	}

	if key.OrganizationID != orgID {
		response.NotFound(w, "API Key")
		return
	}

	if err := h.store.DeleteAPIKey(r.Context(), keyID); err != nil {
		h.log.Error().Err(err).Msg("failed to delete API key")
		response.InternalError(w)
		return
	}

	response.NoContent(w)
}
