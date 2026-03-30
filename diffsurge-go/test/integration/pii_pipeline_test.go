package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/diffsurge-org/diffsurge/internal/models"
	"github.com/diffsurge-org/diffsurge/internal/pii"
	"github.com/diffsurge-org/diffsurge/internal/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPIIPipelineEndToEnd tests the full flow: Request → PII Redaction → Storage → Verification
func TestPIIPipelineEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Parallel()

	// Setup test database
	connStr, cleanup := setupTestDB(t)
	defer cleanup()

	// Run migrations
	runMigrations(t, connStr)

	store, err := storage.NewPostgresStore(connStr)
	require.NoError(t, err)
	defer store.Close()

	// Setup test data
	ctx := context.Background()
	org := &models.Organization{
		ID:        uuid.New(),
		Name:      "Test Org",
		Slug:      "test-org",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = store.CreateOrganization(ctx, org)
	require.NoError(t, err)

	project := &models.Project{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "Test Project",
		Slug:           "test-project",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	err = store.CreateProject(ctx, project)
	require.NoError(t, err)

	env := &models.Environment{
		ID:        uuid.New(),
		ProjectID: project.ID,
		Name:      "production",
		BaseURL:   "https://api.example.com",
		IsSource:  true,
		CreatedAt: time.Now(),
	}
	err = store.CreateEnvironment(ctx, env)
	require.NoError(t, err)

	// Configure PII redactor
	piiConfig := pii.Config{
		Enabled:          true,
		Mode:             pii.ModeMask,
		ScanRequestBody:  true,
		ScanResponseBody: true,
		ScanHeaders:      true,
		ScanQueryParams:  true,
		ScanURLPath:      true,
		Patterns: pii.PatternConfig{
			Email:      true,
			Phone:      true,
			CreditCard: true,
			SSN:        true,
			APIKey:     true,
		},
	}
	redactor := pii.NewRedactor(piiConfig)

	t.Run("PII Detection and Redaction in Request Body", func(t *testing.T) {
		// Create a request with PII data
		requestBodyWithPII := map[string]interface{}{
			"user": map[string]interface{}{
				"name":        "John Doe",
				"email":       "john.doe@example.com",
				"phone":       "555-123-4567",
				"ssn":         "123-45-6789",
				"credit_card": "4111-1111-1111-1111",
				"address":     "123 Main St",
			},
		}

		// Simulate backend response
		simulatedBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			resp := map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"user_id": "u_12345",
					"email":   "john.doe@example.com", // PII in response too
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer simulatedBackend.Close()

		// Create traffic log from the request
		trafficLog := &models.TrafficLog{
			ID:            uuid.New(),
			ProjectID:     project.ID,
			EnvironmentID: env.ID,
			Method:        "POST",
			Path:          "/api/users?api_key=secret123", // PII in query param
			StatusCode:    200,
			LatencyMs:     100,
			RequestHeaders: map[string]interface{}{
				"Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...", // Potential PII
				"Content-Type":  "application/json",
			},
			ResponseHeaders: map[string]interface{}{
				"Content-Type": "application/json",
			},
			QueryParams: map[string]interface{}{
				"api_key": "secret123", // Should be redacted
			},
			RequestBody: requestBodyWithPII,
			ResponseBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"user_id": "u_12345",
					"email":   "john.doe@example.com",
				},
			},
			Timestamp: time.Now(),
		}

		// Apply PII redaction
		scanResult := redactor.RedactTrafficLog(trafficLog)

		// Verify PII was detected
		assert.True(t, scanResult.Found, "PII should have been detected")
		assert.True(t, trafficLog.PIIRedacted, "TrafficLog should be marked as redacted")
		assert.Greater(t, len(scanResult.Detections), 0, "Should have detections")

		// Verify specific PII fields were redacted
		assert.True(t, containsRedactionMarker(trafficLog.RequestBody), "Request body should contain redaction markers")
		assert.True(t, containsRedactionMarker(trafficLog.ResponseBody), "Response body should contain redaction markers")

		// Save to database
		err := store.SaveTrafficLog(trafficLog)
		require.NoError(t, err, "Should save redacted traffic log")

		// Retrieve from database and verify redaction persisted
		retrieved, err := store.GetTrafficLog(ctx, trafficLog.ID)
		require.NoError(t, err)
		assert.True(t, retrieved.PIIRedacted, "Retrieved log should still be marked as redacted")
		assert.True(t, containsRedactionMarker(retrieved.RequestBody), "Persisted data should contain redaction markers")

		// Verify original PII data is NOT in stored log
		retrievedJSON, _ := json.Marshal(retrieved.RequestBody)
		assert.NotContains(t, string(retrievedJSON), "john.doe@example.com", "Original email should not be present")
		assert.NotContains(t, string(retrievedJSON), "123-45-6789", "Original SSN should not be present")
		assert.NotContains(t, string(retrievedJSON), "4111-1111-1111-1111", "Original credit card should not be present")
	})

	t.Run("No PII Detection in Safe Data", func(t *testing.T) {
		// Create a request with no PII
		safeRequestBody := map[string]interface{}{
			"product": map[string]interface{}{
				"id":    "prod_123",
				"name":  "Widget",
				"price": 29.99,
			},
		}

		trafficLog := &models.TrafficLog{
			ID:            uuid.New(),
			ProjectID:     project.ID,
			EnvironmentID: env.ID,
			Method:        "GET",
			Path:          "/api/products/123",
			StatusCode:    200,
			LatencyMs:     50,
			RequestBody:   safeRequestBody,
			ResponseBody: map[string]interface{}{
				"product": map[string]interface{}{
					"id":    "prod_123",
					"name":  "Widget",
					"price": 29.99,
				},
			},
			Timestamp: time.Now(),
		}

		// Apply PII redaction
		scanResult := redactor.RedactTrafficLog(trafficLog)

		// Verify NO PII was detected
		assert.False(t, scanResult.Found, "Should not detect PII in safe data")
		assert.False(t, trafficLog.PIIRedacted, "TrafficLog should not be marked as redacted")
		assert.Equal(t, 0, len(scanResult.Detections), "Should have no detections")

		// Save and verify
		err := store.SaveTrafficLog(trafficLog)
		require.NoError(t, err)

		retrieved, err := store.GetTrafficLog(ctx, trafficLog.ID)
		require.NoError(t, err)
		assert.False(t, retrieved.PIIRedacted, "Retrieved log should not be marked as redacted")
	})

	t.Run("PII in Headers and Query Params", func(t *testing.T) {
		trafficLog := &models.TrafficLog{
			ID:            uuid.New(),
			ProjectID:     project.ID,
			EnvironmentID: env.ID,
			Method:        "GET",
			Path:          "/api/users/search?email=user@example.com&ssn=123-45-6789",
			StatusCode:    200,
			LatencyMs:     75,
			RequestHeaders: map[string]interface{}{
				"X-User-Email": "admin@example.com",
				"X-API-Key":    "sk_live_1234567890abcdef",
			},
			QueryParams: map[string]interface{}{
				"email": "user@example.com",
				"ssn":   "123-45-6789",
			},
			Timestamp: time.Now(),
		}

		// Apply PII redaction
		scanResult := redactor.RedactTrafficLog(trafficLog)

		// Verify PII was detected in multiple locations
		assert.True(t, scanResult.Found, "Should detect PII in headers/params")
		assert.True(t, trafficLog.PIIRedacted, "Should be marked as redacted")

		// Check path was redacted
		assert.NotContains(t, trafficLog.Path, "user@example.com", "Path should not contain original email")
		assert.NotContains(t, trafficLog.Path, "123-45-6789", "Path should not contain original SSN")

		// Save and verify persistence
		err := store.SaveTrafficLog(trafficLog)
		require.NoError(t, err)

		retrieved, err := store.GetTrafficLog(ctx, trafficLog.ID)
		require.NoError(t, err)
		assert.True(t, retrieved.PIIRedacted)
		assert.NotContains(t, retrieved.Path, "user@example.com")
	})
}

// containsRedactionMarker checks if the data contains any redaction markers
func containsRedactionMarker(data interface{}) bool {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return false
	}

	// Check for common redaction markers
	markers := []string{
		"[REDACTED]",
		"***",
		"****",
		"[EMAIL_REDACTED]",
		"[SSN_REDACTED]",
		"[CARD_REDACTED]",
	}

	for _, marker := range markers {
		if bytes.Contains(jsonBytes, []byte(marker)) {
			return true
		}
	}

	return false
}

// BenchmarkPIIRedaction benchmarks the PII redaction performance
func BenchmarkPIIRedaction(b *testing.B) {
	config := pii.Config{
		Enabled:          true,
		Mode:             pii.ModeMask,
		ScanRequestBody:  true,
		ScanResponseBody: true,
		ScanHeaders:      true,
		ScanQueryParams:  true,
		ScanURLPath:      true,
		Patterns: pii.PatternConfig{
			Email:      true,
			Phone:      true,
			CreditCard: true,
		},
	}
	redactor := pii.NewRedactor(config)

	trafficLog := &models.TrafficLog{
		ID:         uuid.New(),
		ProjectID:  uuid.New(),
		Method:     "POST",
		Path:       "/api/users",
		StatusCode: 200,
		LatencyMs:  100,
		RequestBody: map[string]interface{}{
			"email":       "test@example.com",
			"phone":       "555-123-4567",
			"credit_card": "4111-1111-1111-1111",
		},
		ResponseBody: map[string]interface{}{
			"status":  "success",
			"user_id": "u_12345",
		},
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		redactor.RedactTrafficLog(trafficLog)
	}
}
