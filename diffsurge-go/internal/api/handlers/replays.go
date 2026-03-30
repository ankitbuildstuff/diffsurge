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

type ReplayHandler struct {
	store storage.Repository
	log   *logger.Logger
}

func NewReplayHandler(store storage.Repository, log *logger.Logger) *ReplayHandler {
	return &ReplayHandler{store: store, log: log}
}

type createReplayRequest struct {
	Name                string                 `json:"name"`
	Description         string                 `json:"description,omitempty"`
	SourceEnvironmentID string                 `json:"source_environment_id"`
	TargetEnvironmentID string                 `json:"target_environment_id"`
	SampleSize          int                    `json:"sample_size"`
	TrafficFilter       map[string]interface{} `json:"traffic_filter,omitempty"`
}

func (h *ReplayHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.PathUUID(r, "id")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	sessions, err := h.store.ListReplaySessions(r.Context(), projectID)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to list replay sessions")
		response.InternalError(w)
		return
	}

	if sessions == nil {
		sessions = []models.ReplaySession{}
	}

	statusFilter := request.QueryString(r, "status", "")
	if statusFilter != "" {
		filtered := make([]models.ReplaySession, 0)
		for _, s := range sessions {
			if s.Status == statusFilter {
				filtered = append(filtered, s)
			}
		}
		sessions = filtered
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": sessions,
	})
}

func (h *ReplayHandler) Create(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.PathUUID(r, "id")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	var req createReplayRequest
	if err := request.ParseJSON(r, 0, &req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	var errs []response.FieldError
	if req.Name == "" {
		errs = append(errs, response.FieldError{Field: "name", Message: "Name is required"})
	}
	if req.SourceEnvironmentID == "" {
		errs = append(errs, response.FieldError{Field: "source_environment_id", Message: "Source environment is required"})
	}
	if req.TargetEnvironmentID == "" {
		errs = append(errs, response.FieldError{Field: "target_environment_id", Message: "Target environment is required"})
	}
	if req.SampleSize < 1 || req.SampleSize > 10000 {
		errs = append(errs, response.FieldError{Field: "sample_size", Message: "Sample size must be between 1 and 10000"})
	}
	if len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}

	sourceEnvID, err := uuid.Parse(req.SourceEnvironmentID)
	if err != nil {
		response.BadRequest(w, "invalid source_environment_id format")
		return
	}
	targetEnvID, err := uuid.Parse(req.TargetEnvironmentID)
	if err != nil {
		response.BadRequest(w, "invalid target_environment_id format")
		return
	}

	if _, err := h.store.GetEnvironment(r.Context(), sourceEnvID); err != nil {
		response.NotFound(w, "Source environment")
		return
	}
	if _, err := h.store.GetEnvironment(r.Context(), targetEnvID); err != nil {
		response.NotFound(w, "Target environment")
		return
	}

	desc := req.Description
	session := &models.ReplaySession{
		ID:                  uuid.New(),
		ProjectID:           projectID,
		SourceEnvironmentID: sourceEnvID,
		TargetEnvironmentID: targetEnvID,
		Name:                req.Name,
		Description:         &desc,
		TrafficFilter:       req.TrafficFilter,
		SampleSize:          req.SampleSize,
		Status:              "pending",
		CreatedBy:           middleware.GetUserID(r.Context()),
		CreatedAt:           time.Now(),
	}

	if err := h.store.CreateReplaySession(r.Context(), session); err != nil {
		h.log.Error().Err(err).Msg("failed to create replay session")
		response.InternalError(w)
		return
	}

	response.Created(w, session)
}

func (h *ReplayHandler) Get(w http.ResponseWriter, r *http.Request) {
	replayID, err := request.PathUUID(r, "replayId")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	session, err := h.store.GetReplaySession(r.Context(), replayID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, "Replay session")
			return
		}
		h.log.Error().Err(err).Msg("failed to get replay session")
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": session,
	})
}

func (h *ReplayHandler) Start(w http.ResponseWriter, r *http.Request) {
	replayID, err := request.PathUUID(r, "replayId")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	session, err := h.store.GetReplaySession(r.Context(), replayID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, "Replay session")
			return
		}
		h.log.Error().Err(err).Msg("failed to get replay session")
		response.InternalError(w)
		return
	}

	if session.Status != "pending" && session.Status != "failed" {
		response.Conflict(w, "Replay session is already "+session.Status)
		return
	}

	now := time.Now()
	session.Status = "running"
	session.StartedAt = &now

	if err := h.store.UpdateReplaySession(r.Context(), session); err != nil {
		h.log.Error().Err(err).Msg("failed to update replay session status")
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusAccepted, map[string]interface{}{
		"data":    session,
		"message": "Replay session started",
	})
}

func (h *ReplayHandler) Stop(w http.ResponseWriter, r *http.Request) {
	replayID, err := request.PathUUID(r, "replayId")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	session, err := h.store.GetReplaySession(r.Context(), replayID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, "Replay session")
			return
		}
		h.log.Error().Err(err).Msg("failed to get replay session")
		response.InternalError(w)
		return
	}

	if session.Status != "running" {
		response.Conflict(w, "Replay session is not running")
		return
	}

	now := time.Now()
	session.Status = "cancelled"
	session.CompletedAt = &now

	if err := h.store.UpdateReplaySession(r.Context(), session); err != nil {
		h.log.Error().Err(err).Msg("failed to stop replay session")
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data":    session,
		"message": "Replay session stopped",
	})
}

func (h *ReplayHandler) Results(w http.ResponseWriter, r *http.Request) {
	replayID, err := request.PathUUID(r, "replayId")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	results, err := h.store.GetReplayResults(r.Context(), replayID)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to get replay results")
		response.InternalError(w)
		return
	}

	if results == nil {
		results = []models.ReplayResult{}
	}

	severityFilter := request.QueryString(r, "severity", "")
	if severityFilter != "" {
		filtered := make([]models.ReplayResult, 0)
		for _, res := range results {
			if res.Severity == severityFilter {
				filtered = append(filtered, res)
			}
		}
		results = filtered
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": results,
	})
}
