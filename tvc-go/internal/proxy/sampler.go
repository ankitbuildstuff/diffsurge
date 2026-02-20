package proxy

import (
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Sampler interface {
	ShouldSample(r *http.Request) bool
}

// PercentageSampler samples a configurable percentage of requests.
type PercentageSampler struct {
	rate float64
	rng  *rand.Rand
	mu   sync.Mutex
}

func NewPercentageSampler(rate float64) *PercentageSampler {
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}
	return &PercentageSampler{
		rate: rate,
		rng:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *PercentageSampler) ShouldSample(r *http.Request) bool {
	if s.rate >= 1.0 {
		return true
	}
	if s.rate <= 0 {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.rng.Float64() < s.rate
}

// AlwaysSampler captures every request.
type AlwaysSampler struct{}

func NewAlwaysSampler() *AlwaysSampler {
	return &AlwaysSampler{}
}

func (s *AlwaysSampler) ShouldSample(r *http.Request) bool {
	return true
}

// PathSampler only samples requests matching specific path prefixes.
type PathSampler struct {
	paths    []string
	fallback Sampler
}

func NewPathSampler(paths []string, fallback Sampler) *PathSampler {
	return &PathSampler{
		paths:    paths,
		fallback: fallback,
	}
}

func (s *PathSampler) ShouldSample(r *http.Request) bool {
	for _, p := range s.paths {
		if strings.HasPrefix(r.URL.Path, p) {
			return true
		}
	}
	if s.fallback != nil {
		return s.fallback.ShouldSample(r)
	}
	return false
}

// CompositeSampler chains multiple samplers (all must agree).
type CompositeSampler struct {
	samplers []Sampler
}

func NewCompositeSampler(samplers ...Sampler) *CompositeSampler {
	return &CompositeSampler{samplers: samplers}
}

func (s *CompositeSampler) ShouldSample(r *http.Request) bool {
	for _, sampler := range s.samplers {
		if !sampler.ShouldSample(r) {
			return false
		}
	}
	return true
}
