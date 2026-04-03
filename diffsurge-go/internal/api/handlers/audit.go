package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/diffsurge-org/diffsurge/internal/api/request"
	"github.com/diffsurge-org/diffsurge/internal/api/response"
	"github.com/diffsurge-org/diffsurge/internal/models"
	"github.com/diffsurge-org/diffsurge/internal/storage"
	"github.com/diffsurge-org/diffsurge/pkg/logger"
	"github.com/google/uuid"
)

type AuditLogHandler struct {
	store storage.Repository
	log   *logger.Logger
}

func NewAuditLogHandler(store storage.Repository, log *logger.Logger) *AuditLogHandler {
	return &AuditLogHandler{store: store, log: log}
}

// List returns audit logs with optional filtering
// Query params:
//   - user_id: filter by user UUID
//   - action: filter by action type (create, update, delete, etc.)
//   - resource_type: filter by resource type (project, environment, etc.)
//   - start_time: filter by start timestamp (RFC3339)
//   - end_time: filter by end timestamp (RFC3339)
//   - limit: max results (default 100, max 1000)
//   - offset: pagination offset (default 0)
func (h *AuditLogHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID, err := request.PathUUID(r, "id")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	ctx := r.Context()
	filter := models.AuditLogFilter{
		OrganizationID: orgID,
		Limit:          100,
		Offset:         0,
	}

	// Parse query parameters
	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			response.BadRequest(w, "Invalid user_id format")
			return
		}
		filter.UserID = &userID
	}

	if actionStr := r.URL.Query().Get("action"); actionStr != "" {
		action := models.AuditAction(actionStr)
		filter.Action = &action
	}

	if resourceType := r.URL.Query().Get("resource_type"); resourceType != "" {
		filter.ResourceType = &resourceType
	}

	if startTimeStr := r.URL.Query().Get("start_time"); startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			response.BadRequest(w, "Invalid start_time format (use RFC3339)")
			return
		}
		filter.StartTime = &startTime
	}

	if endTimeStr := r.URL.Query().Get("end_time"); endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			response.BadRequest(w, "Invalid end_time format (use RFC3339)")
			return
		}
		filter.EndTime = &endTime
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 1000 {
			response.BadRequest(w, "Invalid limit (must be 1-1000)")
			return
		}
		filter.Limit = limit
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			response.BadRequest(w, "Invalid offset (must be >= 0)")
			return
		}
		filter.Offset = offset
	}

	logs, err := h.store.ListAuditLogs(ctx, filter)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to list audit logs")
		response.InternalError(w)
		return
	}

	if logs == nil {
		logs = []models.AuditLog{}
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": logs,
	})
}
