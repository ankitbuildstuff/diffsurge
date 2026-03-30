package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/diffsurge-org/diffsurge/internal/models"
	"github.com/diffsurge-org/diffsurge/pkg/logger"
	"github.com/google/uuid"
)

type CaptureStats struct {
	Captured   int64
	Dropped    int64
	QueueSize  int
	RedisQueue int64
}

type TrafficStore interface {
	SaveTrafficLog(log *models.TrafficLog) error
}

type PIIRedactor interface {
	RedactTrafficLog(log *models.TrafficLog) interface{}
}

// RedisQueue defines the interface for Redis-based traffic queueing
type RedisQueue interface {
	EnqueueTraffic(ctx context.Context, log *models.TrafficLog) error
	DequeueTraffic(ctx context.Context, timeout time.Duration) (*models.TrafficLog, error)
	GetQueueLength(ctx context.Context) (int64, error)
}

type TrafficCapture struct {
	buffer      chan *models.TrafficLog
	workers     int
	store       TrafficStore
	sampler     Sampler
	piiRedactor PIIRedactor
	redisQueue  RedisQueue
	log         *logger.Logger
	wg          sync.WaitGroup
	captured    atomic.Int64
	dropped     atomic.Int64
	mu          sync.Mutex
	closed      bool
	useRedis    bool
}

func NewTrafficCapture(queueSize, workers int, store TrafficStore, sampler Sampler, log *logger.Logger) *TrafficCapture {
	return &TrafficCapture{
		buffer:  make(chan *models.TrafficLog, queueSize),
		workers: workers,
		store:   store,
		sampler: sampler,
		log:     log,
	}
}

// SetRedisQueue enables Redis-based persistent queueing
func (tc *TrafficCapture) SetRedisQueue(redisQueue RedisQueue) {
	tc.redisQueue = redisQueue
	tc.useRedis = true
}

// SetPIIRedactor attaches a PII redactor that processes logs before storage.
func (tc *TrafficCapture) SetPIIRedactor(r PIIRedactor) {
	tc.piiRedactor = r
}

func (tc *TrafficCapture) Start() {
	for i := 0; i < tc.workers; i++ {
		tc.wg.Add(1)
		if tc.useRedis {
			go tc.redisWorker(i)
		} else {
			go tc.channelWorker(i)
		}
	}
	mode := "channel"
	if tc.useRedis {
		mode = "redis"
	}
	tc.log.Info().
		Int("workers", tc.workers).
		Int("queue_size", cap(tc.buffer)).
		Str("mode", mode).
		Msg("Traffic capture started")
}

func (tc *TrafficCapture) Stop() {
	tc.mu.Lock()
	if tc.closed {
		tc.mu.Unlock()
		return
	}
	tc.closed = true
	close(tc.buffer)
	tc.mu.Unlock()

	tc.wg.Wait()
	tc.log.Info().
		Int64("total_captured", tc.captured.Load()).
		Int64("total_dropped", tc.dropped.Load()).
		Msg("Traffic capture stopped")
}

func (tc *TrafficCapture) Stats() CaptureStats {
	stats := CaptureStats{
		Captured:  tc.captured.Load(),
		Dropped:   tc.dropped.Load(),
		QueueSize: len(tc.buffer),
	}

	if tc.useRedis && tc.redisQueue != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if length, err := tc.redisQueue.GetQueueLength(ctx); err == nil {
			stats.RedisQueue = length
		}
	}

	return stats
}

// channelWorker processes logs from the in-memory channel buffer
func (tc *TrafficCapture) channelWorker(id int) {
	defer tc.wg.Done()
	for log := range tc.buffer {
		tc.processLog(log, id)
	}
}

// redisWorker processes logs from Redis queue
func (tc *TrafficCapture) redisWorker(id int) {
	defer tc.wg.Done()
	ctx := context.Background()

	for {
		tc.mu.Lock()
		if tc.closed {
			tc.mu.Unlock()
			return
		}
		tc.mu.Unlock()

		log, err := tc.redisQueue.DequeueTraffic(ctx, 1*time.Second)
		if err != nil {
			tc.log.Error().Err(err).Int("worker", id).Msg("Failed to dequeue from Redis")
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if log == nil {
			// Timeout, no items in queue - continue polling
			continue
		}

		tc.processLog(log, id)
	}
}

// processLog handles PII redaction and storage for a single log
func (tc *TrafficCapture) processLog(log *models.TrafficLog, workerID int) {
	if tc.piiRedactor != nil {
		tc.piiRedactor.RedactTrafficLog(log)
	}

	if tc.store != nil {
		if err := tc.store.SaveTrafficLog(log); err != nil {
			tc.log.Error().Err(err).Int("worker", workerID).Msg("Failed to save traffic log")
			return
		}
	}

	tc.captured.Add(1)
}

func (tc *TrafficCapture) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !tc.sampler.ShouldSample(r) {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()

		var reqBody []byte
		if r.Body != nil {
			reqBody, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		recorder := &responseRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
			body:           &bytes.Buffer{},
		}

		next.ServeHTTP(recorder, r)

		latency := time.Since(start)

		go tc.enqueue(r, reqBody, recorder, latency)
	})
}

func (tc *TrafficCapture) enqueue(r *http.Request, reqBody []byte, rec *responseRecorder, latency time.Duration) {
	reqHeaders := make(map[string]interface{})
	for k, v := range r.Header {
		if len(v) == 1 {
			reqHeaders[k] = v[0]
		} else {
			reqHeaders[k] = v
		}
	}

	var reqBodyParsed map[string]interface{}
	if len(reqBody) > 0 {
		json.Unmarshal(reqBody, &reqBodyParsed) //nolint:errcheck
	}

	respHeaders := make(map[string]interface{})
	for k, v := range rec.Header() {
		if len(v) == 1 {
			respHeaders[k] = v[0]
		} else {
			respHeaders[k] = v
		}
	}

	var respBodyParsed map[string]interface{}
	if rec.body.Len() > 0 {
		json.Unmarshal(rec.body.Bytes(), &respBodyParsed) //nolint:errcheck
	}

	queryParams := make(map[string]interface{})
	for k, v := range r.URL.Query() {
		if len(v) == 1 {
			queryParams[k] = v[0]
		} else {
			queryParams[k] = v
		}
	}

	trafficLog := &models.TrafficLog{
		ID:              uuid.New(),
		ProjectID:       GetProjectIDFromContext(r.Context()),
		EnvironmentID:   GetEnvironmentIDFromContext(r.Context()),
		Method:          r.Method,
		Path:            r.URL.Path,
		QueryParams:     queryParams,
		RequestHeaders:  reqHeaders,
		RequestBody:     reqBodyParsed,
		StatusCode:      rec.status,
		ResponseHeaders: respHeaders,
		ResponseBody:    respBodyParsed,
		Timestamp:       time.Now(),
		LatencyMs:       int(latency.Milliseconds()),
		IPAddress:       r.RemoteAddr,
		UserAgent:       r.UserAgent(),
	}

	tc.mu.Lock()
	if tc.closed {
		tc.mu.Unlock()
		return
	}

	if tc.useRedis && tc.redisQueue != nil {
		tc.mu.Unlock()
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		if err := tc.redisQueue.EnqueueTraffic(ctx, trafficLog); err != nil {
			tc.log.Error().Err(err).Msg("Failed to enqueue to Redis, dropping traffic log")
			tc.dropped.Add(1)
		}
	} else {
		select {
		case tc.buffer <- trafficLog:
		default:
			tc.dropped.Add(1)
		}
		tc.mu.Unlock()
	}
}

type responseRecorder struct {
	http.ResponseWriter
	status      int
	body        *bytes.Buffer
	wroteHeader bool
}

func (r *responseRecorder) WriteHeader(code int) {
	if !r.wroteHeader {
		r.status = code
		r.wroteHeader = true
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.wroteHeader = true
	}
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}
