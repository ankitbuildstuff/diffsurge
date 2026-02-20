package replayer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/tvc-org/tvc/internal/models"
	"github.com/tvc-org/tvc/pkg/logger"
)

type Config struct {
	Workers    int
	RateLimit  int
	Timeout    time.Duration
	MaxRetries int
	TargetURL  string
}

type Result struct {
	OriginalLog    models.TrafficLog
	StatusCode     int
	ResponseBody   map[string]interface{}
	ResponseHeaders map[string]interface{}
	LatencyMs      int
	Error          error
}

type Stats struct {
	Total      int64
	Completed  int64
	Failed     int64
	InFlight   int64
	AvgLatency float64
}

type Replayer struct {
	config     Config
	httpClient *http.Client
	log        *logger.Logger
	stats      Stats
	totalLatMs atomic.Int64
	mu         sync.Mutex
}

func New(cfg Config, log *logger.Logger) *Replayer {
	if cfg.Workers <= 0 {
		cfg.Workers = 10
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 2
	}

	transport := &http.Transport{
		MaxIdleConns:        cfg.Workers * 2,
		MaxIdleConnsPerHost: cfg.Workers * 2,
		MaxConnsPerHost:     cfg.Workers * 2,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
	}

	return &Replayer{
		config: cfg,
		httpClient: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: transport,
		},
		log: log,
	}
}

func (r *Replayer) ReplayTraffic(ctx context.Context, traffic []models.TrafficLog) ([]Result, error) {
	if len(traffic) == 0 {
		return nil, nil
	}

	results := make([]Result, len(traffic))
	jobs := make(chan indexedJob, len(traffic))

	var rateLimiter <-chan time.Time
	if r.config.RateLimit > 0 {
		interval := time.Second / time.Duration(r.config.RateLimit)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		rateLimiter = ticker.C
	}

	var wg sync.WaitGroup
	for i := 0; i < r.config.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				select {
				case <-ctx.Done():
					results[job.index] = Result{
						OriginalLog: job.log,
						Error:       ctx.Err(),
					}
					return
				default:
				}

				result := r.replayOne(ctx, job.log)
				results[job.index] = result
			}
		}()
	}

	for i, log := range traffic {
		if rateLimiter != nil {
			select {
			case <-rateLimiter:
			case <-ctx.Done():
				close(jobs)
				wg.Wait()
				return results, ctx.Err()
			}
		}

		select {
		case jobs <- indexedJob{index: i, log: log}:
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return results, ctx.Err()
		}
	}

	close(jobs)
	wg.Wait()

	return results, nil
}

func (r *Replayer) replayOne(ctx context.Context, log models.TrafficLog) Result {
	atomic.AddInt64(&r.stats.InFlight, 1)
	defer atomic.AddInt64(&r.stats.InFlight, -1)

	var lastErr error
	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * 100 * time.Millisecond
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return Result{OriginalLog: log, Error: ctx.Err()}
			}
		}

		result, err := r.doRequest(ctx, log)
		if err == nil {
			atomic.AddInt64(&r.stats.Completed, 1)
			r.totalLatMs.Add(int64(result.LatencyMs))
			return result
		}

		lastErr = err
		if !isRetryable(err) {
			break
		}
	}

	atomic.AddInt64(&r.stats.Failed, 1)
	return Result{OriginalLog: log, Error: lastErr}
}

func (r *Replayer) doRequest(ctx context.Context, log models.TrafficLog) (Result, error) {
	url := r.config.TargetURL + log.Path
	if log.QueryParams != nil && len(log.QueryParams) > 0 {
		url += "?"
		first := true
		for k, v := range log.QueryParams {
			if !first {
				url += "&"
			}
			url += fmt.Sprintf("%s=%v", k, v)
			first = false
		}
	}

	var bodyReader io.Reader
	if log.RequestBody != nil {
		bodyJSON, err := json.Marshal(log.RequestBody)
		if err != nil {
			return Result{}, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyJSON)
	}

	req, err := http.NewRequestWithContext(ctx, log.Method, url, bodyReader)
	if err != nil {
		return Result{}, fmt.Errorf("creating request: %w", err)
	}

	if log.RequestHeaders != nil {
		for k, v := range log.RequestHeaders {
			if s, ok := v.(string); ok {
				req.Header.Set(k, s)
			}
		}
	}

	if req.Header.Get("Content-Type") == "" && bodyReader != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	start := time.Now()
	resp, err := r.httpClient.Do(req)
	latency := time.Since(start)

	if err != nil {
		return Result{}, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{}, fmt.Errorf("reading response body: %w", err)
	}

	respHeaders := make(map[string]interface{})
	for k, v := range resp.Header {
		if len(v) == 1 {
			respHeaders[k] = v[0]
		} else {
			respHeaders[k] = v
		}
	}

	var parsedBody map[string]interface{}
	json.Unmarshal(respBody, &parsedBody) //nolint:errcheck

	return Result{
		OriginalLog: log,
		StatusCode:  resp.StatusCode,
		ResponseBody: parsedBody,
		ResponseHeaders: respHeaders,
		LatencyMs:   int(latency.Milliseconds()),
	}, nil
}

func (r *Replayer) GetStats() Stats {
	completed := atomic.LoadInt64(&r.stats.Completed)
	totalLat := r.totalLatMs.Load()
	avgLat := float64(0)
	if completed > 0 {
		avgLat = float64(totalLat) / float64(completed)
	}

	return Stats{
		Total:      atomic.LoadInt64(&r.stats.Total),
		Completed:  completed,
		Failed:     atomic.LoadInt64(&r.stats.Failed),
		InFlight:   atomic.LoadInt64(&r.stats.InFlight),
		AvgLatency: avgLat,
	}
}

type indexedJob struct {
	index int
	log   models.TrafficLog
}

func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	retryablePatterns := []string{
		"connection refused",
		"connection reset",
		"i/o timeout",
		"temporary failure",
		"EOF",
	}
	for _, p := range retryablePatterns {
		if contains(errStr, p) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ReplayToModel converts raw results into models.ReplayResult for DB storage.
func ReplayToModel(sessionID uuid.UUID, result Result, comparison ComparisonResult) models.ReplayResult {
	diffReport := make(map[string]interface{})
	if len(comparison.Diffs) > 0 {
		diffsRaw := make([]interface{}, len(comparison.Diffs))
		for i, d := range comparison.Diffs {
			diffsRaw[i] = map[string]interface{}{
				"path":     d.Path,
				"type":     string(d.Type),
				"severity": string(d.Severity),
			}
		}
		diffReport["diffs"] = diffsRaw
		diffReport["total_diffs"] = len(comparison.Diffs)
	}

	errMsg := ""
	if result.Error != nil {
		errMsg = result.Error.Error()
	}

	return models.ReplayResult{
		ID:                   uuid.New(),
		ReplaySessionID:      sessionID,
		OriginalTrafficLogID: result.OriginalLog.ID,
		TargetStatusCode:     result.StatusCode,
		TargetResponseBody:   result.ResponseBody,
		TargetLatencyMs:      result.LatencyMs,
		StatusMatch:          comparison.StatusMatch,
		BodyMatch:            comparison.BodyMatch,
		DiffReport:           diffReport,
		Severity:             string(comparison.Severity),
		ErrorMessage:         errMsg,
		Timestamp:            time.Now(),
	}
}
