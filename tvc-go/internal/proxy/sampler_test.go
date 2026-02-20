package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPercentageSampler_AlwaysSample(t *testing.T) {
	sampler := NewPercentageSampler(1.0)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	for i := 0; i < 100; i++ {
		assert.True(t, sampler.ShouldSample(req))
	}
}

func TestPercentageSampler_NeverSample(t *testing.T) {
	sampler := NewPercentageSampler(0.0)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	for i := 0; i < 100; i++ {
		assert.False(t, sampler.ShouldSample(req))
	}
}

func TestPercentageSampler_PartialRate(t *testing.T) {
	sampler := NewPercentageSampler(0.5)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	sampled := 0
	total := 10000
	for i := 0; i < total; i++ {
		if sampler.ShouldSample(req) {
			sampled++
		}
	}

	rate := float64(sampled) / float64(total)
	assert.InDelta(t, 0.5, rate, 0.05, "Sample rate should be approximately 50%%")
}

func TestPercentageSampler_ClampsNegative(t *testing.T) {
	sampler := NewPercentageSampler(-0.5)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	assert.False(t, sampler.ShouldSample(req))
}

func TestPercentageSampler_ClampsAboveOne(t *testing.T) {
	sampler := NewPercentageSampler(1.5)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	assert.True(t, sampler.ShouldSample(req))
}

func TestAlwaysSampler(t *testing.T) {
	sampler := NewAlwaysSampler()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	for i := 0; i < 100; i++ {
		assert.True(t, sampler.ShouldSample(req))
	}
}

func TestPathSampler_MatchingPath(t *testing.T) {
	sampler := NewPathSampler([]string{"/api/v1/users"}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/123", nil)
	assert.True(t, sampler.ShouldSample(req))
}

func TestPathSampler_NonMatchingPath(t *testing.T) {
	sampler := NewPathSampler([]string{"/api/v1/users"}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/123", nil)
	assert.False(t, sampler.ShouldSample(req))
}

func TestPathSampler_FallbackSampler(t *testing.T) {
	sampler := NewPathSampler([]string{"/api/v1/users"}, NewAlwaysSampler())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/123", nil)
	assert.True(t, sampler.ShouldSample(req))
}

func TestCompositeSampler_AllAgree(t *testing.T) {
	sampler := NewCompositeSampler(
		NewAlwaysSampler(),
		NewPercentageSampler(1.0),
	)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	assert.True(t, sampler.ShouldSample(req))
}

func TestCompositeSampler_OneRejects(t *testing.T) {
	sampler := NewCompositeSampler(
		NewAlwaysSampler(),
		NewPercentageSampler(0.0),
	)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	assert.False(t, sampler.ShouldSample(req))
}
