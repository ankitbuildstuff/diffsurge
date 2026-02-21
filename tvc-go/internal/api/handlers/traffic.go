package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/tvc-org/tvc/internal/api/request"
	"github.com/tvc-org/tvc/internal/api/response"
	"github.com/tvc-org/tvc/internal/models"
	"github.com/tvc-org/tvc/internal/storage"
	"github.com/tvc-org/tvc/pkg/logger"
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
		ProjectID: projectID,
		Methods:   request.QueryStringSlice(r, "methods"),
		Paths:     request.QueryStringSlice(r, "paths"),
		StatusCodes: request.QueryIntSlice(r, "status_codes"),
		Limit:     pagination.Limit + 1, // fetch one extra to determine has_more
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
	switch period {
	case "1h":
		startTime = time.Now().Add(-1 * time.Hour)
	case "24h":
		startTime = time.Now().Add(-24 * time.Hour)
	case "7d":
		startTime = time.Now().Add(-7 * 24 * time.Hour)
	case "30d":
		startTime = time.Now().Add(-30 * 24 * time.Hour)
	default:
		startTime = time.Now().Add(-24 * time.Hour)
	}

	filter := storage.TrafficFilter{
		ProjectID: projectID,
		StartTime: &startTime,
		Limit:     10000,
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
