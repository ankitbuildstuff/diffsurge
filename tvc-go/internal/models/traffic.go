package models

import (
	"time"

	"github.com/google/uuid"
)

type TrafficLog struct {
	ID              uuid.UUID              `json:"id" db:"id"`
	ProjectID       uuid.UUID              `json:"project_id" db:"project_id"`
	EnvironmentID   uuid.UUID              `json:"environment_id" db:"environment_id"`
	Method          string                 `json:"method" db:"method"`
	Path            string                 `json:"path" db:"path"`
	QueryParams     map[string]interface{} `json:"query_params,omitempty" db:"query_params"`
	RequestHeaders  map[string]interface{} `json:"request_headers,omitempty" db:"request_headers"`
	RequestBody     map[string]interface{} `json:"request_body,omitempty" db:"request_body"`
	StatusCode      int                    `json:"status_code" db:"status_code"`
	ResponseHeaders map[string]interface{} `json:"response_headers,omitempty" db:"response_headers"`
	ResponseBody    map[string]interface{} `json:"response_body,omitempty" db:"response_body"`
	Timestamp       time.Time              `json:"timestamp" db:"timestamp"`
	LatencyMs       int                    `json:"latency_ms" db:"latency_ms"`
	IPAddress       string                 `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent       string                 `json:"user_agent,omitempty" db:"user_agent"`
	PIIRedacted     bool                   `json:"pii_redacted" db:"pii_redacted"`
}

type Project struct {
	ID             uuid.UUID              `json:"id" db:"id"`
	OrganizationID uuid.UUID              `json:"organization_id" db:"organization_id"`
	Name           string                 `json:"name" db:"name"`
	Slug           string                 `json:"slug" db:"slug"`
	Description    string                 `json:"description,omitempty" db:"description"`
	Config         map[string]interface{} `json:"config,omitempty" db:"config"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
}

type Environment struct {
	ID        uuid.UUID `json:"id" db:"id"`
	ProjectID uuid.UUID `json:"project_id" db:"project_id"`
	Name      string    `json:"name" db:"name"`
	BaseURL   string    `json:"base_url" db:"base_url"`
	IsSource  bool      `json:"is_source" db:"is_source"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type ReplaySession struct {
	ID                  uuid.UUID              `json:"id" db:"id"`
	ProjectID           uuid.UUID              `json:"project_id" db:"project_id"`
	SourceEnvironmentID uuid.UUID              `json:"source_environment_id" db:"source_environment_id"`
	TargetEnvironmentID uuid.UUID              `json:"target_environment_id" db:"target_environment_id"`
	Name                string                 `json:"name" db:"name"`
	Description         *string                `json:"description,omitempty" db:"description"`
	TrafficFilter       map[string]interface{} `json:"traffic_filter,omitempty" db:"traffic_filter"`
	StartTime           *time.Time             `json:"start_time,omitempty" db:"start_time"`
	EndTime             *time.Time             `json:"end_time,omitempty" db:"end_time"`
	SampleSize          int                    `json:"sample_size" db:"sample_size"`
	Status              string                 `json:"status" db:"status"`
	TotalRequests       int                    `json:"total_requests" db:"total_requests"`
	SuccessfulRequests  int                    `json:"successful_requests" db:"successful_requests"`
	FailedRequests      int                    `json:"failed_requests" db:"failed_requests"`
	MismatchedResponses int                    `json:"mismatched_responses" db:"mismatched_responses"`
	CreatedBy           uuid.UUID              `json:"created_by" db:"created_by"`
	CreatedAt           time.Time              `json:"created_at" db:"created_at"`
	StartedAt           *time.Time             `json:"started_at,omitempty" db:"started_at"`
	CompletedAt         *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
}

type ReplayResult struct {
	ID                     uuid.UUID              `json:"id" db:"id"`
	ReplaySessionID        uuid.UUID              `json:"replay_session_id" db:"replay_session_id"`
	OriginalTrafficLogID   uuid.UUID              `json:"original_traffic_log_id" db:"original_traffic_log_id"`
	TargetStatusCode       int                    `json:"target_status_code" db:"target_status_code"`
	TargetResponseBody     map[string]interface{} `json:"target_response_body,omitempty" db:"target_response_body"`
	TargetLatencyMs        int                    `json:"target_latency_ms" db:"target_latency_ms"`
	StatusMatch            bool                   `json:"status_match" db:"status_match"`
	BodyMatch              bool                   `json:"body_match" db:"body_match"`
	DiffReport             map[string]interface{} `json:"diff_report,omitempty" db:"diff_report"`
	Severity               string                 `json:"severity" db:"severity"`
	ErrorMessage           string                 `json:"error_message,omitempty" db:"error_message"`
	Timestamp              time.Time              `json:"timestamp" db:"timestamp"`
}
