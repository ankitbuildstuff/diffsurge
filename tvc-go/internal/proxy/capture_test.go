package proxy

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tvc-org/tvc/internal/models"
	"github.com/tvc-org/tvc/pkg/logger"
)

type mockStore struct {
	mu   sync.Mutex
	logs []*models.TrafficLog
}

func (m *mockStore) SaveTrafficLog(log *models.TrafficLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = append(m.logs, log)
	return nil
}

func (m *mockStore) Logs() []*models.TrafficLog {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.logs
}

func TestTrafficCapture_CapturesRequest(t *testing.T) {
	store := &mockStore{}
	log := logger.New("error", "console")
	capture := NewTrafficCapture(100, 2, store, NewAlwaysSampler(), log)
	capture.Start()
	defer capture.Stop()

	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, `{"name":"test"}`, string(body))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	handler := capture.Middleware(backend)

	body := strings.NewReader(`{"name":"test"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users?page=1", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, `{"status":"ok"}`, rec.Body.String())

	// Give async worker time to process
	time.Sleep(100 * time.Millisecond)

	logs := store.Logs()
	require.Len(t, logs, 1)

	assert.Equal(t, http.MethodPost, logs[0].Method)
	assert.Equal(t, "/api/v1/users", logs[0].Path)
	assert.Equal(t, http.StatusOK, logs[0].StatusCode)
	assert.GreaterOrEqual(t, logs[0].LatencyMs, 0)
}

func TestTrafficCapture_Respectssampling(t *testing.T) {
	store := &mockStore{}
	log := logger.New("error", "console")
	sampler := NewPercentageSampler(0.0)
	capture := NewTrafficCapture(100, 2, store, sampler, log)
	capture.Start()
	defer capture.Stop()

	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := capture.Middleware(backend)

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	time.Sleep(100 * time.Millisecond)

	assert.Empty(t, store.Logs())
}

func TestTrafficCapture_HandlesHighThroughput(t *testing.T) {
	store := &mockStore{}
	log := logger.New("error", "console")
	capture := NewTrafficCapture(1000, 5, store, NewAlwaysSampler(), log)
	capture.Start()
	defer capture.Stop()

	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	handler := capture.Middleware(backend)

	total := 100
	var wg sync.WaitGroup
	for i := 0; i < total; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
		}()
	}

	wg.Wait()
	time.Sleep(500 * time.Millisecond)

	logs := store.Logs()
	assert.Len(t, logs, total)
}

func TestTrafficCapture_Stats(t *testing.T) {
	store := &mockStore{}
	log := logger.New("error", "console")
	capture := NewTrafficCapture(100, 2, store, NewAlwaysSampler(), log)
	capture.Start()
	defer capture.Stop()

	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := capture.Middleware(backend)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	time.Sleep(200 * time.Millisecond)

	stats := capture.Stats()
	assert.Equal(t, int64(5), stats.Captured)
	assert.Equal(t, int64(0), stats.Dropped)
}

func TestTrafficCapture_PreservesResponseToClient(t *testing.T) {
	store := &mockStore{}
	log := logger.New("error", "console")
	capture := NewTrafficCapture(100, 2, store, NewAlwaysSampler(), log)
	capture.Start()
	defer capture.Stop()

	responseBody := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{"id": float64(1), "name": "Alice"},
			map[string]interface{}{"id": float64(2), "name": "Bob"},
		},
	}
	responseJSON, _ := json.Marshal(responseBody)

	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Custom", "value")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(responseJSON)
	})

	handler := capture.Middleware(backend)

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Let async goroutine finish before Stop() is called in defer
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, string(responseJSON), rec.Body.String())
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.Equal(t, "value", rec.Header().Get("X-Custom"))
}

func TestTrafficCapture_NilStore(t *testing.T) {
	log := logger.New("error", "console")
	capture := NewTrafficCapture(100, 2, nil, NewAlwaysSampler(), log)
	capture.Start()
	defer capture.Stop()

	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := capture.Middleware(backend)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, http.StatusOK, rec.Code)
}
