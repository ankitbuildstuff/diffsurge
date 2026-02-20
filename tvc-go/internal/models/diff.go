package models

import (
	"time"

	"github.com/google/uuid"
)

type SchemaVersion struct {
	ID            uuid.UUID   `json:"id" db:"id"`
	ProjectID     uuid.UUID   `json:"project_id" db:"project_id"`
	Version       string      `json:"version" db:"version"`
	SchemaType    string      `json:"schema_type" db:"schema_type"`
	SchemaContent interface{} `json:"schema_content" db:"schema_content"`
	GitCommit     string      `json:"git_commit,omitempty" db:"git_commit"`
	GitBranch     string      `json:"git_branch,omitempty" db:"git_branch"`
	CreatedBy     uuid.UUID   `json:"created_by,omitempty" db:"created_by"`
	CreatedAt     time.Time   `json:"created_at" db:"created_at"`
}

type SchemaDiff struct {
	ID                 uuid.UUID   `json:"id" db:"id"`
	ProjectID          uuid.UUID   `json:"project_id" db:"project_id"`
	FromVersionID      uuid.UUID   `json:"from_version_id" db:"from_version_id"`
	ToVersionID        uuid.UUID   `json:"to_version_id" db:"to_version_id"`
	DiffReport         interface{} `json:"diff_report" db:"diff_report"`
	HasBreakingChanges bool        `json:"has_breaking_changes" db:"has_breaking_changes"`
	BreakingChanges    interface{} `json:"breaking_changes,omitempty" db:"breaking_changes"`
	CreatedAt          time.Time   `json:"created_at" db:"created_at"`
}
