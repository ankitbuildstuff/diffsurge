package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AuditAction represents the type of action performed
type AuditAction string

const (
	AuditActionCreate AuditAction = "create"
	AuditActionUpdate AuditAction = "update"
	AuditActionDelete AuditAction = "delete"
	AuditActionInvite AuditAction = "invite"
	AuditActionRemove AuditAction = "remove"
	AuditActionLogin  AuditAction = "login"
	AuditActionLogout AuditAction = "logout"
	AuditActionAccess AuditAction = "access"
)

// AuditLog represents an audit trail entry
type AuditLog struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	OrganizationID uuid.UUID       `json:"organization_id" db:"organization_id"`
	UserID         *uuid.UUID      `json:"user_id,omitempty" db:"user_id"`
	Action         AuditAction     `json:"action" db:"action"`
	ResourceType   string          `json:"resource_type" db:"resource_type"`
	ResourceID     *uuid.UUID      `json:"resource_id,omitempty" db:"resource_id"`
	Details        json.RawMessage `json:"details,omitempty" db:"details"`
	IPAddress      *string         `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent      *string         `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
}

// AuditLogFilter represents filtering options for audit log queries
type AuditLogFilter struct {
	OrganizationID uuid.UUID    `json:"organization_id"`
	UserID         *uuid.UUID   `json:"user_id,omitempty"`
	Action         *AuditAction `json:"action,omitempty"`
	ResourceType   *string      `json:"resource_type,omitempty"`
	StartTime      *time.Time   `json:"start_time,omitempty"`
	EndTime        *time.Time   `json:"end_time,omitempty"`
	Limit          int          `json:"limit"`
	Offset         int          `json:"offset"`
}

// Scan implements the sql.Scanner interface for json.RawMessage in Details field
func (a *AuditLog) Scan(src interface{}) error {
	return nil
}

// Value implements the driver.Valuer interface for json.RawMessage in Details field
func (a AuditLog) Value() (driver.Value, error) {
	return json.Marshal(a)
}
