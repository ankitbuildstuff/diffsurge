package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/diffsurge-org/diffsurge/internal/api/request"
	"github.com/diffsurge-org/diffsurge/internal/api/response"
	"github.com/diffsurge-org/diffsurge/internal/models"
	"github.com/diffsurge-org/diffsurge/internal/storage"
	"github.com/diffsurge-org/diffsurge/pkg/logger"
	"github.com/google/uuid"
)

type TrafficHandler struct {
	store storage.Repository
	log   *logger.Logger
}

func NewTrafficHandler(store storage.Repository, log *logger.Logger) *TrafficHandler {
	return &TrafficHandler{store: store, log: log}
}

func (h *TrafficHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.PathUUID(r, "id")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	pagination := request.ParsePagination(r)

	filter := storage.TrafficFilter{
		ProjectID:   projectID,
		Methods:     request.QueryStringSlice(r, "methods"),
		Paths:       request.QueryStringSlice(r, "paths"),
		StatusCodes: request.QueryIntSlice(r, "status_codes"),
		Limit:       pagination.Limit + 1, // fetch one extra to determine has_more
	}
	if pagination.Cursor != nil {
		filter.CursorTime = &pagination.Cursor.Timestamp
		if pagination.Cursor.ID != "" {
			cursorID, err := uuid.Parse(pagination.Cursor.ID)
			if err == nil {
				filter.CursorID = &cursorID
			}
		}
	}

	if start := request.QueryTime(r, "start_time"); start != nil {
		filter.StartTime = start
	}
	if end := request.QueryTime(r, "end_time"); end != nil {
		filter.EndTime = end
	}

	logs, err := h.store.FetchTraffic(r.Context(), filter)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to fetch traffic")
		response.InternalError(w)
		return
	}

	hasMore := len(logs) > pagination.Limit
	if hasMore {
		logs = logs[:pagination.Limit]
	}

	if logs == nil {
		logs = []models.TrafficLog{}
	}

	var nextCursor string
	if hasMore && len(logs) > 0 {
		last := logs[len(logs)-1]
		nextCursor = request.EncodeCursor(request.Cursor{
			Timestamp: last.Timestamp,
			ID:        last.ID.String(),
		})
	}

	response.Paginated(w, logs, response.PaginationMeta{
		NextCursor: nextCursor,
		HasMore:    hasMore,
	})
}

type createTrafficRequest struct {
	Method          string                 `json:"method"`
	Path            string                 `json:"path"`
	StatusCode      int                    `json:"status_code"`
	LatencyMs       int                    `json:"latency_ms"`
	QueryParams     map[string]interface{} `json:"query_params,omitempty"`
	RequestHeaders  map[string]interface{} `json:"request_headers,omitempty"`
	RequestBody     map[string]interface{} `json:"request_body,omitempty"`
	ResponseHeaders map[string]interface{} `json:"response_headers,omitempty"`
	ResponseBody    map[string]interface{} `json:"response_body,omitempty"`
	IPAddress       string                 `json:"ip_address,omitempty"`
	UserAgent       string                 `json:"user_agent,omitempty"`
	Timestamp       *time.Time             `json:"timestamp,omitempty"`
	EnvironmentID   *uuid.UUID             `json:"environment_id,omitempty"`
}

func (h *TrafficHandler) Create(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.PathUUID(r, "id")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	var req createTrafficRequest
	if err := request.ParseJSON(r, 0, &req); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	var errs []response.FieldError
	if req.Method == "" {
		errs = append(errs, response.FieldError{Field: "method", Message: "Method is required"})
	}
	if req.Path == "" {
		errs = append(errs, response.FieldError{Field: "path", Message: "Path is required"})
	}
	if req.StatusCode == 0 {
		errs = append(errs, response.FieldError{Field: "status_code", Message: "Status code is required"})
	}
	if len(errs) > 0 {
		response.ValidationError(w, errs)
		return
	}

	ts := time.Now()
	if req.Timestamp != nil {
		ts = *req.Timestamp
	}

	envID := uuid.Nil
	if req.EnvironmentID != nil && *req.EnvironmentID != uuid.Nil {
		envID = *req.EnvironmentID
	} else {
		// Fetch existing environments for the project
		envs, err := h.store.ListEnvironments(r.Context(), projectID)
		if err == nil && len(envs) > 0 {
			// Find the source environment if possible, otherwise use the first one
			envID = envs[0].ID
			for _, e := range envs {
				if e.IsSource {
					envID = e.ID
					break
				}
			}
		} else {
			// Create a default environment if none exists
			defaultEnv := &models.Environment{
				ID:        uuid.New(),
				ProjectID: projectID,
				Name:      "Default",
				BaseURL:   "http://localhost",
				IsSource:  true,
				CreatedAt: time.Now(),
			}
			if err := h.store.CreateEnvironment(r.Context(), defaultEnv); err == nil {
				envID = defaultEnv.ID
			}
		}
	}

	log := &models.TrafficLog{
		ID:              uuid.New(),
		ProjectID:       projectID,
		EnvironmentID:   envID,
		Method:          req.Method,
		Path:            req.Path,
		QueryParams:     req.QueryParams,
		RequestHeaders:  req.RequestHeaders,
		RequestBody:     req.RequestBody,
		StatusCode:      req.StatusCode,
		ResponseHeaders: req.ResponseHeaders,
		ResponseBody:    req.ResponseBody,
		Timestamp:       ts,
		LatencyMs:       req.LatencyMs,
		IPAddress:       req.IPAddress,
		UserAgent:       req.UserAgent,
	}

	if err := h.store.SaveTrafficLog(log); err != nil {
		h.log.Error().Err(err).Str("project_id", projectID.String()).Msg("failed to save traffic log")
		response.InternalError(w)
		return
	}

	response.Created(w, log)
}

func (h *TrafficHandler) Get(w http.ResponseWriter, r *http.Request) {
	logID, err := request.PathUUID(r, "logId")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	log, err := h.store.GetTrafficLog(r.Context(), logID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, "Traffic log")
			return
		}
		h.log.Error().Err(err).Msg("failed to get traffic log")
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": log,
	})
}

type trafficStatsResponse struct {
	TotalRequests int64            `json:"total_requests"`
	ErrorCount    int64            `json:"error_count"`
	ErrorRate     float64          `json:"error_rate"`
	AvgLatencyMs  float64          `json:"avg_latency_ms"`
	ByMethod      map[string]int64 `json:"by_method"`
	ByStatus      map[int]int64    `json:"by_status"`
	Period        string           `json:"period"`
}

func (h *TrafficHandler) Stats(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.PathUUID(r, "id")
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	period := request.QueryString(r, "period", "24h")
	var startTime time.Time
	useStartTime := true
	switch period {
	case "1h":
		startTime = time.Now().Add(-1 * time.Hour)
	case "24h":
		startTime = time.Now().Add(-24 * time.Hour)
	case "7d":
		startTime = time.Now().Add(-7 * 24 * time.Hour)
	case "30d":
		startTime = time.Now().Add(-30 * 24 * time.Hour)
	case "all":
		useStartTime = false
	default:
		startTime = time.Now().Add(-24 * time.Hour)
	}

	filter := storage.TrafficFilter{
		ProjectID: projectID,
		Limit:     10000,
	}
	if useStartTime {
		filter.StartTime = &startTime
	}

	logs, err := h.store.FetchTraffic(r.Context(), filter)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to fetch traffic for stats")
		response.InternalError(w)
		return
	}

	stats := computeStats(logs, period)
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": stats,
	})
}

func computeStats(logs []models.TrafficLog, period string) trafficStatsResponse {
	stats := trafficStatsResponse{
		ByMethod: make(map[string]int64),
		ByStatus: make(map[int]int64),
		Period:   period,
	}

	if len(logs) == 0 {
		return stats
	}

	var totalLatency int64
	for _, l := range logs {
		stats.TotalRequests++
		totalLatency += int64(l.LatencyMs)
		stats.ByMethod[l.Method]++
		stats.ByStatus[l.StatusCode]++
		if l.StatusCode >= 400 {
			stats.ErrorCount++
		}
	}

	if stats.TotalRequests > 0 {
		stats.AvgLatencyMs = float64(totalLatency) / float64(stats.TotalRequests)
		stats.ErrorRate = float64(stats.ErrorCount) / float64(stats.TotalRequests)
	}

	return stats
}
