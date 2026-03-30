package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/diffsurge-org/diffsurge/internal/api/middleware"
	"github.com/diffsurge-org/diffsurge/internal/api/request"
	"github.com/diffsurge-org/diffsurge/internal/api/response"
	"github.com/diffsurge-org/diffsurge/internal/models"
	"github.com/diffsurge-org/diffsurge/internal/storage"
	"github.com/diffsurge-org/diffsurge/pkg/logger"
	"github.com/google/uuid"
)

type SchemaHandler struct {
	store storage.Repository
	log   *logger.Logger
}

func NewSchemaHandler(store storage.Repository, log *logger.Logger) *SchemaHandler {
	return &SchemaHandler{store: store, log: log}
}

type uploadSchemaRequest struct {
	Version       string      `json:"version"`
	SchemaType    string      `json:"schema_type"`
	SchemaContent interface{} `json:"schema_content"`
	GitCommit     string      `json:"git_commit,omitempty"`
	GitBranch     string      `json:"git_branch,omitempty"`
}

type schemaDiffRequest struct {
	FromVersionID string `json:"from_version_id"`
	ToVersionID   string `json:"to_version_id"`
}

func (h *SchemaHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.PathUUID(r, "id")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	versions, err := h.store.ListSchemaVersions(r.Context(), projectID)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to list schema versions")
		response.InternalError(w)
		return
	}

	if versions == nil {
		versions = []models.SchemaVersion{}
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": versions,
	})
}

func (h *SchemaHandler) Upload(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.PathUUID(r, "id")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	var req uploadSchemaRequest
	if err := request.ParseJSON(r, request.SchemaMaxBodySize, &req); err != nil {
		if strings.Contains(err.Error(), "too large") {
			response.PayloadTooLarge(w, "10MB")
			return
		}
		response.BadRequest(w, err.Error())
		return
	}

	var errs []response.FieldError
	if req.Version == "" {
		errs = append(errs, response.FieldError{Field: "version", Message: "Version is required"})
	}
	if req.SchemaType == "" {
		errs = append(errs, response.FieldError{Field: "schema_type", Message: "Schema type is required"})
	} else if req.SchemaType != "openapi" && req.SchemaType != "graphql" {
		errs = append(errs, response.FieldError{Field: "schema_type", Message: "Schema type must be 'openapi' or 'graphql'"})
	}
	if req.SchemaContent == nil {
		errs = append(errs, response.FieldError{Field: "schema_content", Message: "Schema content is required"})
	}
	if len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}

	schema := &models.SchemaVersion{
		ID:            uuid.New(),
		ProjectID:     projectID,
		Version:       req.Version,
		SchemaType:    req.SchemaType,
		SchemaContent: req.SchemaContent,
		GitCommit:     req.GitCommit,
		GitBranch:     req.GitBranch,
		CreatedBy:     middleware.GetUserID(r.Context()),
		CreatedAt:     time.Now(),
	}

	if err := h.store.SaveSchemaVersion(r.Context(), schema); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			response.Conflict(w, "Schema version already exists for this project")
			return
		}
		h.log.Error().Err(err).Msg("failed to save schema version")
		response.InternalError(w)
		return
	}

	response.Created(w, schema)
}

func (h *SchemaHandler) Diff(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.PathUUID(r, "id")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	var req schemaDiffRequest
	if err := request.ParseJSON(r, 0, &req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	if req.FromVersionID == "" || req.ToVersionID == "" {
		response.ValidationError(w, []response.FieldError{
			{Field: "from_version_id", Message: "From version ID is required"},
			{Field: "to_version_id", Message: "To version ID is required"},
		})
		return
	}

	fromID, err := uuid.Parse(req.FromVersionID)
	if err != nil {
		response.BadRequest(w, "invalid from_version_id format")
		return
	}
	toID, err := uuid.Parse(req.ToVersionID)
	if err != nil {
		response.BadRequest(w, "invalid to_version_id format")
		return
	}

	// Placeholder: actual schema diff logic would use internal/diffing
	diff := &models.SchemaDiff{
		ID:                 uuid.New(),
		ProjectID:          projectID,
		FromVersionID:      fromID,
		ToVersionID:        toID,
		DiffReport:         map[string]interface{}{"status": "diff computed"},
		HasBreakingChanges: false,
		CreatedAt:          time.Now(),
	}

	if err := h.store.SaveSchemaDiff(r.Context(), diff); err != nil {
		h.log.Error().Err(err).Msg("failed to save schema diff")
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": diff,
	})
}
