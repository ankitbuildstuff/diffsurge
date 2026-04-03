package handlers

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/diffsurge-org/diffsurge/internal/api/middleware"
	"github.com/diffsurge-org/diffsurge/internal/models"
	"github.com/diffsurge-org/diffsurge/internal/storage"
	"github.com/diffsurge-org/diffsurge/pkg/logger"
	"github.com/google/uuid"
)

func writeAuditLog(
	r *http.Request,
	store storage.Repository,
	log *logger.Logger,
	orgID uuid.UUID,
	action models.AuditAction,
	resourceType string,
	resourceID *uuid.UUID,
	details map[string]interface{},
) {
	entry := &models.AuditLog{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         auditUserID(r),
		Action:         action,
		ResourceType:   resourceType,
		ResourceID:     resourceID,
		Details:        marshalAuditDetails(details),
		IPAddress:      requestIPAddress(r),
		UserAgent:      requestUserAgent(r),
		CreatedAt:      time.Now(),
	}

	if err := store.CreateAuditLog(r.Context(), entry); err != nil {
		log.Warn().Err(err).
			Str("organization_id", orgID.String()).
			Str("resource_type", resourceType).
			Msg("failed to write audit log")
	}
}

func auditUserID(r *http.Request) *uuid.UUID {
	if middleware.IsAPIKey(r.Context()) {
		return nil
	}

	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		return nil
	}

	return &userID
}

func marshalAuditDetails(details map[string]interface{}) json.RawMessage {
	if len(details) == 0 {
		return nil
	}

	encoded, err := json.Marshal(details)
	if err != nil {
		return nil
	}

	return encoded
}

func requestIPAddress(r *http.Request) *string {
	for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
		value := strings.TrimSpace(r.Header.Get(header))
		if value == "" {
			continue
		}

		if header == "X-Forwarded-For" {
			parts := strings.Split(value, ",")
			value = strings.TrimSpace(parts[0])
		}

		if value != "" {
			return &value
		}
	}

	if r.RemoteAddr == "" {
		return nil
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return &host
	}

	remote := r.RemoteAddr
	return &remote
}

func requestUserAgent(r *http.Request) *string {
	if ua := strings.TrimSpace(r.UserAgent()); ua != "" {
		return &ua
	}

	return nil
}
