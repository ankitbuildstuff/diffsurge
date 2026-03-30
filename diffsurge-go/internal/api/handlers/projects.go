package handlers

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/diffsurge-org/diffsurge/internal/api/request"
	"github.com/diffsurge-org/diffsurge/internal/api/response"
	"github.com/diffsurge-org/diffsurge/internal/models"
	"github.com/diffsurge-org/diffsurge/internal/storage"
	"github.com/diffsurge-org/diffsurge/pkg/logger"
	"github.com/google/uuid"
)

type ProjectHandler struct {
	store storage.Repository
	log   *logger.Logger
}

func NewProjectHandler(store storage.Repository, log *logger.Logger) *ProjectHandler {
	return &ProjectHandler{store: store, log: log}
}

type createProjectRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
	OrgID       string `json:"organization_id"`
}

type updateProjectRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	orgIDStr := request.QueryString(r, "organization_id", "")
	if orgIDStr == "" {
		response.BadRequest(w, "organization_id query parameter is required")
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		response.BadRequest(w, "invalid organization_id format")
		return
	}

	projects, err := h.store.ListProjects(r.Context(), orgID)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to list projects")
		response.InternalError(w)
		return
	}

	if projects == nil {
		projects = []models.Project{}
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": projects,
	})
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createProjectRequest
	if err := request.ParseJSON(r, 0, &req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	var errs []response.FieldError
	if req.Name == "" || len(req.Name) > 255 {
		errs = append(errs, response.FieldError{Field: "name", Message: "Name is required (1-255 characters)"})
	}
	if req.OrgID == "" {
		errs = append(errs, response.FieldError{Field: "organization_id", Message: "Organization ID is required"})
	}
	if req.Slug != "" && !slugPattern.MatchString(req.Slug) {
		errs = append(errs, response.FieldError{Field: "slug", Message: "Slug must be lowercase alphanumeric with dashes"})
	}
	if len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}

	orgID, err := uuid.Parse(req.OrgID)
	if err != nil {
		response.BadRequest(w, "invalid organization_id format")
		return
	}

	if req.Slug == "" {
		req.Slug = generateSlug(req.Name)
	}

	project := &models.Project{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           req.Name,
		Slug:           req.Slug,
		Description:    req.Description,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := h.store.CreateProject(r.Context(), project); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			response.Conflict(w, "A project with this slug already exists in the organization")
			return
		}
		h.log.Error().Err(err).Msg("failed to create project")
		response.InternalError(w)
		return
	}

	response.Created(w, project)
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := request.PathUUID(r, "id")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	project, err := h.store.GetProject(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, "Project")
			return
		}
		h.log.Error().Err(err).Msg("failed to get project")
		response.InternalError(w)
		return
	}

	envs, err := h.store.ListEnvironments(r.Context(), id)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to list environments for project")
		envs = []models.Environment{}
	}
	if envs == nil {
		envs = []models.Environment{}
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data":         project,
		"environments": envs,
	})
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := request.PathUUID(r, "id")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	project, err := h.store.GetProject(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, "Project")
			return
		}
		h.log.Error().Err(err).Msg("failed to get project")
		response.InternalError(w)
		return
	}

	var req updateProjectRequest
	if err := request.ParseJSON(r, 0, &req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	if req.Name != nil {
		if *req.Name == "" || len(*req.Name) > 255 {
			response.ValidationError(w, []response.FieldError{
				{Field: "name", Message: "Name must be 1-255 characters"},
			})
			return
		}
		project.Name = *req.Name
	}
	if req.Description != nil {
		project.Description = *req.Description
	}

	if err := h.store.UpdateProject(r.Context(), project); err != nil {
		h.log.Error().Err(err).Msg("failed to update project")
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, project)
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := request.PathUUID(r, "id")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	if _, err := h.store.GetProject(r.Context(), id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, "Project")
			return
		}
		h.log.Error().Err(err).Msg("failed to get project")
		response.InternalError(w)
		return
	}

	if err := h.store.DeleteProject(r.Context(), id); err != nil {
		h.log.Error().Err(err).Msg("failed to delete project")
		response.InternalError(w)
		return
	}

	response.NoContent(w)
}

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		if r == ' ' || r == '-' || r == '_' {
			return '-'
		}
		return -1
	}, slug)

	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	slug = strings.Trim(slug, "-")

	if len(slug) > 100 {
		slug = slug[:100]
	}
	return slug
}
