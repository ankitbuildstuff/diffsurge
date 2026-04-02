package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/diffsurge-org/diffsurge/internal/api/middleware"
	"github.com/diffsurge-org/diffsurge/internal/api/request"
	"github.com/diffsurge-org/diffsurge/internal/api/response"
	"github.com/diffsurge-org/diffsurge/internal/diffing"
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
	FromVersionID string      `json:"from_version_id"`
	ToVersionID   string      `json:"to_version_id"`
	SchemaContent interface{} `json:"schema_content,omitempty"`
	SchemaType    string      `json:"schema_type,omitempty"`
	Version       string      `json:"version,omitempty"`
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
		errMsg := err.Error()
		if strings.Contains(errMsg, "duplicate") || strings.Contains(errMsg, "unique") ||
			strings.Contains(errMsg, "UNIQUE") || strings.Contains(errMsg, "23505") ||
			strings.Contains(errMsg, "violates unique constraint") {
			response.Conflict(w, "Schema version '"+req.Version+"' already exists for this project")
			return
		}
		h.log.Error().Err(err).
			Str("project_id", projectID.String()).
			Str("version", req.Version).
			Str("schema_type", req.SchemaType).
			Str("created_by", schema.CreatedBy.String()).
			Msg("failed to save schema version")
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

	if req.SchemaContent == nil && (req.FromVersionID == "" || req.ToVersionID == "") {
		response.ValidationError(w, []response.FieldError{
			{Field: "from_version_id", Message: "Provide from_version_id and to_version_id, or supply schema_content for an inline diff"},
			{Field: "to_version_id", Message: "Provide from_version_id and to_version_id, or supply schema_content for an inline diff"},
		})
		return
	}

	versions, err := h.store.ListSchemaVersions(r.Context(), projectID)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to list schema versions for diff")
		response.InternalError(w)
		return
	}

	var (
		fromVersion *models.SchemaVersion
		toVersion   *models.SchemaVersion
		newContent  interface{}
	)

	if req.SchemaContent != nil {
		if len(versions) == 0 {
			response.NotFound(w, "Schema version")
			return
		}
		fromVersion = &versions[0]
		newContent = req.SchemaContent
	} else {
		fromVersion = findSchemaVersionByID(versions, req.FromVersionID)
		if fromVersion == nil {
			response.NotFound(w, "From schema version")
			return
		}

		toVersion = findSchemaVersionByID(versions, req.ToVersionID)
		if toVersion == nil {
			response.NotFound(w, "To schema version")
			return
		}

		newContent = toVersion.SchemaContent
	}

	schemaType := fromVersion.SchemaType
	if req.SchemaType != "" && req.SchemaType != schemaType {
		response.BadRequest(w, "schema_type must match the stored base schema type")
		return
	}
	if toVersion != nil && toVersion.SchemaType != schemaType {
		response.BadRequest(w, "schema versions must have the same schema_type")
		return
	}
	if schemaType != "openapi" {
		response.BadRequest(w, "schema diff currently supports only openapi schemas")
		return
	}

	comparer := diffing.NewSchemaComparer()
	diffs, breakingChanges, err := comparer.CompareContents(fromVersion.SchemaContent, newContent)
	if err != nil {
		response.BadRequest(w, "schema diff failed: "+err.Error())
		return
	}

	diffReport := buildSchemaDiffReport(diffs, breakingChanges)
	result := map[string]interface{}{
		"project_id":             projectID,
		"from_version_id":        fromVersion.ID,
		"diff_report":            diffReport,
		"has_breaking_changes":   len(breakingChanges) > 0,
		"breaking_changes":       formatBreakingChanges(breakingChanges),
		"breaking_changes_count": len(breakingChanges),
		"created_at":             time.Now(),
	}

	if toVersion != nil {
		persistedDiff := &models.SchemaDiff{
			ID:                 uuid.New(),
			ProjectID:          projectID,
			FromVersionID:      fromVersion.ID,
			ToVersionID:        toVersion.ID,
			DiffReport:         diffReport,
			HasBreakingChanges: len(breakingChanges) > 0,
			BreakingChanges:    formatBreakingChanges(breakingChanges),
			CreatedAt:          time.Now(),
		}

		if err := h.store.SaveSchemaDiff(r.Context(), persistedDiff); err != nil {
			h.log.Error().Err(err).Msg("failed to save schema diff")
			response.InternalError(w)
			return
		}

		result["id"] = persistedDiff.ID
		result["to_version_id"] = toVersion.ID
		result["created_at"] = persistedDiff.CreatedAt
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": result,
	})
}

func findSchemaVersionByID(versions []models.SchemaVersion, versionID string) *models.SchemaVersion {
	for i := range versions {
		if versions[i].ID.String() == versionID {
			return &versions[i]
		}
	}
	return nil
}

func buildSchemaDiffReport(diffs []diffing.Diff, breakingChanges []diffing.BreakingChange) map[string]interface{} {
	nonBreakingChanges := make([]map[string]string, 0, len(diffs))
	for _, diff := range diffs {
		if diff.Severity == diffing.SeverityBreaking {
			continue
		}
		nonBreakingChanges = append(nonBreakingChanges, map[string]string{
			"path":        diff.Path,
			"type":        string(diff.Type),
			"description": describeDiff(diff),
		})
	}

	formattedBreaking := formatBreakingChanges(breakingChanges)

	return map[string]interface{}{
		"summary": map[string]interface{}{
			"total_changes":        len(diffs),
			"breaking_changes":     len(formattedBreaking),
			"non_breaking_changes": len(nonBreakingChanges),
			"has_breaking_changes": len(formattedBreaking) > 0,
		},
		"breaking_changes":     formattedBreaking,
		"non_breaking_changes": nonBreakingChanges,
		"all_diffs":            diffs,
	}
}

func formatBreakingChanges(changes []diffing.BreakingChange) []map[string]string {
	formatted := make([]map[string]string, 0, len(changes))
	for _, change := range changes {
		formatted = append(formatted, map[string]string{
			"path":        change.Path,
			"type":        change.Type,
			"description": change.Description,
		})
	}
	return formatted
}

func describeDiff(diff diffing.Diff) string {
	switch diff.Type {
	case diffing.DiffTypeAdded:
		return "Added in the newer schema"
	case diffing.DiffTypeRemoved:
		return "Removed from the newer schema"
	case diffing.DiffTypeTypeChanged:
		return "Type changed between schema versions"
	case diffing.DiffTypeModified:
		return "Value changed between schema versions"
	default:
		return "Schema changed"
	}
}
