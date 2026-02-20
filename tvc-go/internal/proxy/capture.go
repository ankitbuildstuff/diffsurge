package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/tvc-org/tvc/internal/models"
	"github.com/tvc-org/tvc/pkg/logger"
)

type CaptureStats struct {
	Captured  int64
	Dropped   int64
	QueueSize int
}

type TrafficStore interface {
	SaveTrafficLog(log *models.TrafficLog) error
}

type TrafficCapture struct {
	buffer   chan *models.TrafficLog
	workers  int
	store    TrafficStore
	sampler  Sampler
	log      *logger.Logger
	wg       sync.WaitGroup
	captured atomic.Int64
	dropped  atomic.Int64
	mu       sync.Mutex
	closed   bool
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

func (tc *TrafficCapture) Start() {
	for i := 0; i < tc.workers; i++ {
		tc.wg.Add(1)
		go tc.worker(i)
	}
	tc.log.Info().Int("workers", tc.workers).Int("queue_size", cap(tc.buffer)).Msg("Traffic capture started")
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
	return CaptureStats{
		Captured:  tc.captured.Load(),
		Dropped:   tc.dropped.Load(),
		QueueSize: len(tc.buffer),
	}
}

func (tc *TrafficCapture) worker(id int) {
	defer tc.wg.Done()
	for log := range tc.buffer {
		if tc.store != nil {
			if err := tc.store.SaveTrafficLog(log); err != nil {
				tc.log.Error().Err(err).Int("worker", id).Msg("Failed to save traffic log")
				continue
			}
		}
		tc.captured.Add(1)
	}
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
	select {
	case tc.buffer <- trafficLog:
	default:
		tc.dropped.Add(1)
	}
	tc.mu.Unlock()
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
