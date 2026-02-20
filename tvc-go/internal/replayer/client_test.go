package replayer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tvc-org/tvc/internal/models"
	"github.com/tvc-org/tvc/pkg/logger"
)

func TestReplayer_BasicReplay(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id": 1, "name": "Alice",
		})
	}))
	defer server.Close()

	log := logger.New("error", "console")
	r := New(Config{
		Workers:   2,
		Timeout:   5 * time.Second,
		TargetURL: server.URL,
	}, log)

	traffic := []models.TrafficLog{
		{ID: uuid.New(), Method: "GET", Path: "/api/users/1", StatusCode: 200},
	}

	results, err := r.ReplayTraffic(context.Background(), traffic)
	require.NoError(t, err)
	require.Len(t, results, 1)

	assert.Nil(t, results[0].Error)
	assert.Equal(t, http.StatusOK, results[0].StatusCode)
	assert.NotNil(t, results[0].ResponseBody)
}

func TestReplayer_ConcurrentRequests(t *testing.T) {
	var reqCount atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"path":"%s"}`, r.URL.Path)
	}))
	defer server.Close()

	log := logger.New("error", "console")
	r := New(Config{
		Workers:   5,
		Timeout:   5 * time.Second,
		TargetURL: server.URL,
	}, log)

	traffic := make([]models.TrafficLog, 50)
	for i := range traffic {
		traffic[i] = models.TrafficLog{
			ID:     uuid.New(),
			Method: "GET",
			Path:   fmt.Sprintf("/api/items/%d", i),
		}
	}

	results, err := r.ReplayTraffic(context.Background(), traffic)
	require.NoError(t, err)
	assert.Len(t, results, 50)
	assert.Equal(t, int64(50), reqCount.Load())

	for _, result := range results {
		assert.Nil(t, result.Error)
		assert.Equal(t, http.StatusOK, result.StatusCode)
	}
}

func TestReplayer_RateLimiting(t *testing.T) {
	var reqCount atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCount.Add(1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	log := logger.New("error", "console")
	r := New(Config{
		Workers:   5,
		RateLimit: 100,
		Timeout:   5 * time.Second,
		TargetURL: server.URL,
	}, log)

	traffic := make([]models.TrafficLog, 10)
	for i := range traffic {
		traffic[i] = models.TrafficLog{ID: uuid.New(), Method: "GET", Path: "/test"}
	}

	start := time.Now()
	results, err := r.ReplayTraffic(context.Background(), traffic)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Len(t, results, 10)
	assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(80), "Rate limiting should throttle requests")
}

func TestReplayer_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	log := logger.New("error", "console")
	r := New(Config{
		Workers:    1,
		RateLimit:  1,
		Timeout:    10 * time.Second,
		MaxRetries: 0,
		TargetURL:  server.URL,
	}, log)

	traffic := make([]models.TrafficLog, 50)
	for i := range traffic {
		traffic[i] = models.TrafficLog{ID: uuid.New(), Method: "GET", Path: "/slow"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	results, err := r.ReplayTraffic(ctx, traffic)

	// Either returns error or some results have context errors
	hasErrors := err != nil
	for _, res := range results {
		if res.Error != nil {
			hasErrors = true
			break
		}
	}
	assert.True(t, hasErrors, "Should have errors from context cancellation")
}

func TestReplayer_PreservesHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "custom-value", r.Header.Get("X-Custom"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	log := logger.New("error", "console")
	r := New(Config{
		Workers:   1,
		Timeout:   5 * time.Second,
		TargetURL: server.URL,
	}, log)

	traffic := []models.TrafficLog{
		{
			ID:     uuid.New(),
			Method: "POST",
			Path:   "/api/test",
			RequestHeaders: map[string]interface{}{
				"Content-Type": "application/json",
				"X-Custom":     "custom-value",
			},
			RequestBody: map[string]interface{}{"key": "value"},
		},
	}

	results, err := r.ReplayTraffic(context.Background(), traffic)
	require.NoError(t, err)
	assert.Nil(t, results[0].Error)
}

func TestReplayer_PreservesQueryParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	log := logger.New("error", "console")
	r := New(Config{
		Workers:   1,
		Timeout:   5 * time.Second,
		TargetURL: server.URL,
	}, log)

	traffic := []models.TrafficLog{
		{
			ID:     uuid.New(),
			Method: "GET",
			Path:   "/api/users",
			QueryParams: map[string]interface{}{
				"page": "1",
			},
		},
	}

	results, err := r.ReplayTraffic(context.Background(), traffic)
	require.NoError(t, err)
	assert.Nil(t, results[0].Error)
}

func TestReplayer_EmptyTraffic(t *testing.T) {
	log := logger.New("error", "console")
	r := New(Config{Workers: 1, TargetURL: "http://localhost"}, log)

	results, err := r.ReplayTraffic(context.Background(), nil)
	require.NoError(t, err)
	assert.Nil(t, results)
}

func TestReplayer_ServerErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal server error"}`))
	}))
	defer server.Close()

	log := logger.New("error", "console")
	r := New(Config{
		Workers:   1,
		Timeout:   5 * time.Second,
		TargetURL: server.URL,
	}, log)

	traffic := []models.TrafficLog{
		{ID: uuid.New(), Method: "GET", Path: "/fail"},
	}

	results, err := r.ReplayTraffic(context.Background(), traffic)
	require.NoError(t, err)
	assert.Nil(t, results[0].Error)
	assert.Equal(t, http.StatusInternalServerError, results[0].StatusCode)
}

func TestIsRetryable(t *testing.T) {
	assert.True(t, isRetryable(fmt.Errorf("connection refused")))
	assert.True(t, isRetryable(fmt.Errorf("i/o timeout")))
	assert.True(t, isRetryable(fmt.Errorf("unexpected EOF")))
	assert.False(t, isRetryable(fmt.Errorf("bad request")))
	assert.False(t, isRetryable(nil))
}

func TestReplayToModel(t *testing.T) {
	sessionID := uuid.New()
	result := Result{
		OriginalLog: models.TrafficLog{ID: uuid.New()},
		StatusCode:  200,
		LatencyMs:   42,
	}
	comp := ComparisonResult{
		StatusMatch: true,
		BodyMatch:   true,
		Severity:    "info",
	}

	model := ReplayToModel(sessionID, result, comp)

	assert.Equal(t, sessionID, model.ReplaySessionID)
	assert.Equal(t, 200, model.TargetStatusCode)
	assert.Equal(t, 42, model.TargetLatencyMs)
	assert.True(t, model.StatusMatch)
	assert.True(t, model.BodyMatch)
	assert.Empty(t, model.ErrorMessage)
}
