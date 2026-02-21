package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tvc-org/tvc/internal/models"
)

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "My Project", "my-project"},
		{"with special chars", "Hello World! @#$", "hello-world"},
		{"underscores", "foo_bar_baz", "foo-bar-baz"},
		{"multiple dashes", "a--b---c", "a-b-c"},
		{"trim dashes", "-hello-", "hello"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, generateSlug(tt.input))
		})
	}
}

func TestComputeStats(t *testing.T) {
	logs := []models.TrafficLog{
		{Method: "GET", StatusCode: 200, LatencyMs: 100},
		{Method: "POST", StatusCode: 201, LatencyMs: 200},
		{Method: "GET", StatusCode: 404, LatencyMs: 50},
		{Method: "GET", StatusCode: 500, LatencyMs: 300},
	}

	stats := computeStats(logs, "24h")
	assert.Equal(t, int64(4), stats.TotalRequests)
	assert.Equal(t, int64(2), stats.ErrorCount)
	assert.Equal(t, 0.5, stats.ErrorRate)
	assert.InDelta(t, 162.5, stats.AvgLatencyMs, 0.1)
	assert.Equal(t, int64(3), stats.ByMethod["GET"])
	assert.Equal(t, int64(1), stats.ByMethod["POST"])
	assert.Equal(t, int64(1), stats.ByStatus[200])
	assert.Equal(t, int64(1), stats.ByStatus[404])
	assert.Equal(t, "24h", stats.Period)
}

func TestComputeStats_Empty(t *testing.T) {
	stats := computeStats(nil, "1h")
	assert.Equal(t, int64(0), stats.TotalRequests)
	assert.Equal(t, "1h", stats.Period)
}
