package handlers

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/diffsurge-org/diffsurge/internal/api/request"
	"github.com/diffsurge-org/diffsurge/internal/api/response"
	"github.com/diffsurge-org/diffsurge/internal/models"
	"github.com/diffsurge-org/diffsurge/internal/storage"
	"github.com/diffsurge-org/diffsurge/pkg/logger"
	"github.com/google/uuid"
)

type EnvironmentHandler struct {
	store storage.Repository
	log   *logger.Logger
}

func NewEnvironmentHandler(store storage.Repository, log *logger.Logger) *EnvironmentHandler {
	return &EnvironmentHandler{store: store, log: log}
}

type createEnvironmentRequest struct {
	Name     string `json:"name"`
	BaseURL  string `json:"base_url"`
	IsSource bool   `json:"is_source"`
}

func (h *EnvironmentHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.PathUUID(r, "id")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	envs, err := h.store.ListEnvironments(r.Context(), projectID)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to list environments")
		response.InternalError(w)
		return
	}

	if envs == nil {
		envs = []models.Environment{}
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": envs,
	})
}

func (h *EnvironmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.PathUUID(r, "id")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	if _, err := h.store.GetProject(r.Context(), projectID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, "Project")
			return
		}
		h.log.Error().Err(err).Msg("failed to verify project")
		response.InternalError(w)
		return
	}

	var req createEnvironmentRequest
	if err := request.ParseJSON(r, 0, &req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	var errs []response.FieldError
	if req.Name == "" || len(req.Name) > 100 {
		errs = append(errs, response.FieldError{Field: "name", Message: "Name is required (1-100 characters)"})
	}
	if req.BaseURL == "" {
		errs = append(errs, response.FieldError{Field: "base_url", Message: "Base URL is required"})
	} else if _, err := url.ParseRequestURI(req.BaseURL); err != nil {
		errs = append(errs, response.FieldError{Field: "base_url", Message: "Base URL must be a valid URL"})
	}
	if len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}

	env := &models.Environment{
		ID:        uuid.New(),
		ProjectID: projectID,
		Name:      req.Name,
		BaseURL:   strings.TrimRight(req.BaseURL, "/"),
		IsSource:  req.IsSource,
		CreatedAt: time.Now(),
	}

	if err := h.store.CreateEnvironment(r.Context(), env); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			response.Conflict(w, "An environment with this name already exists in the project")
			return
		}
		h.log.Error().Err(err).Msg("failed to create environment")
		response.InternalError(w)
		return
	}

	response.Created(w, env)
}

func (h *EnvironmentHandler) Get(w http.ResponseWriter, r *http.Request) {
	envID, err := request.PathUUID(r, "envId")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	env, err := h.store.GetEnvironment(r.Context(), envID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, "Environment")
			return
		}
		h.log.Error().Err(err).Msg("failed to get environment")
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": env,
	})
}

type updateEnvironmentRequest struct {
	Name     string `json:"name"`
	BaseURL  string `json:"base_url"`
	IsSource *bool  `json:"is_source"`
}

func (h *EnvironmentHandler) Update(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.PathUUID(r, "id")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	envID, err := request.PathUUID(r, "envId")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	env, err := h.store.GetEnvironment(r.Context(), envID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, "Environment")
			return
		}
		h.log.Error().Err(err).Msg("failed to get environment")
		response.InternalError(w)
		return
	}

	// Verify environment belongs to the project
	if env.ProjectID != projectID {
		response.NotFound(w, "Environment")
		return
	}

	var req updateEnvironmentRequest
	if err := request.ParseJSON(r, 0, &req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	var errs []response.FieldError
	if req.Name != "" {
		if len(req.Name) > 100 {
			errs = append(errs, response.FieldError{Field: "name", Message: "Name must be 1-100 characters"})
		} else {
			env.Name = req.Name
		}
	}
	if req.BaseURL != "" {
		if _, err := url.ParseRequestURI(req.BaseURL); err != nil {
			errs = append(errs, response.FieldError{Field: "base_url", Message: "Base URL must be a valid URL"})
		} else {
			env.BaseURL = strings.TrimRight(req.BaseURL, "/")
		}
	}
	if req.IsSource != nil {
		env.IsSource = *req.IsSource
	}
	if len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}

	if err := h.store.UpdateEnvironment(r.Context(), env); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			response.Conflict(w, "An environment with this name already exists in the project")
			return
		}
		h.log.Error().Err(err).Msg("failed to update environment")
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": env,
	})
}

func (h *EnvironmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.PathUUID(r, "id")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	envID, err := request.PathUUID(r, "envId")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	env, err := h.store.GetEnvironment(r.Context(), envID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, "Environment")
			return
		}
		h.log.Error().Err(err).Msg("failed to get environment")
		response.InternalError(w)
		return
	}

	// Verify environment belongs to the project
	if env.ProjectID != projectID {
		response.NotFound(w, "Environment")
		return
	}

	if err := h.store.DeleteEnvironment(r.Context(), envID); err != nil {
		h.log.Error().Err(err).Msg("failed to delete environment")
		response.InternalError(w)
		return
	}

	response.NoContent(w)
}
