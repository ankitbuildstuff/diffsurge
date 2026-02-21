package models

import (
	"time"

	"github.com/google/uuid"
)

type Organization struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Slug      string    `json:"slug" db:"slug"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type UserOrganization struct {
	UserID         uuid.UUID `json:"user_id" db:"user_id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	Role           string    `json:"role" db:"role"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// OrganizationMember represents a member with their user details
type OrganizationMember struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name,omitempty"`
	Role      string    `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
}

type Subscription struct {
	ID                   uuid.UUID  `json:"id" db:"id"`
	OrganizationID       uuid.UUID  `json:"organization_id" db:"organization_id"`
	Tier                 string     `json:"tier" db:"tier"`
	Status               string     `json:"status" db:"status"`
	StripeCustomerID     string     `json:"stripe_customer_id,omitempty" db:"stripe_customer_id"`
	StripeSubscriptionID string     `json:"stripe_subscription_id,omitempty" db:"stripe_subscription_id"`
	MonthlyTrafficLimit  int        `json:"monthly_traffic_limit" db:"monthly_traffic_limit"`
	MonthlyReplayLimit   int        `json:"monthly_replay_limit" db:"monthly_replay_limit"`
	CurrentPeriodStart   *time.Time `json:"current_period_start,omitempty" db:"current_period_start"`
	CurrentPeriodEnd     *time.Time `json:"current_period_end,omitempty" db:"current_period_end"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
}

type APIKey struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	ProjectID      *uuid.UUID `json:"project_id,omitempty" db:"project_id"`
	Name           string     `json:"name" db:"name"`
	KeyPrefix      string     `json:"key_prefix" db:"key_prefix"`
	KeyHash        string     `json:"-" db:"key_hash"`
	LastUsedAt     *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	CreatedBy      uuid.UUID  `json:"created_by" db:"created_by"`
}
