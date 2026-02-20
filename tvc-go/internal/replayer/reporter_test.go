package replayer

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tvc-org/tvc/internal/diffing"
	"github.com/tvc-org/tvc/internal/models"
)

func TestReporter_GenerateSummary(t *testing.T) {
	reporter := NewReporter()

	results := []Result{
		{OriginalLog: models.TrafficLog{ID: uuid.New(), Path: "/api/users"}, StatusCode: 200, LatencyMs: 10},
		{OriginalLog: models.TrafficLog{ID: uuid.New(), Path: "/api/users"}, StatusCode: 200, LatencyMs: 20},
		{OriginalLog: models.TrafficLog{ID: uuid.New(), Path: "/api/orders"}, StatusCode: 500, LatencyMs: 30},
		{OriginalLog: models.TrafficLog{ID: uuid.New(), Path: "/api/orders"}, Error: assert.AnError},
	}

	comparisons := []ComparisonResult{
		{StatusMatch: true, BodyMatch: true, Severity: diffing.SeverityInfo, DriftScore: 0},
		{StatusMatch: true, BodyMatch: false, Severity: diffing.SeverityWarning, DriftScore: 0.3, Diffs: []diffing.Diff{{Path: "name"}}},
		{StatusMatch: false, BodyMatch: false, Severity: diffing.SeverityBreaking, DriftScore: 0.8},
		{Severity: diffing.SeverityError, DriftScore: 1.0},
	}

	summary := reporter.GenerateSummary("test", "http://staging", results, comparisons, 5*time.Second)

	assert.Equal(t, "test", summary.SessionName)
	assert.Equal(t, 4, summary.TotalRequests)
	assert.Equal(t, 1, summary.Successful)
	assert.Equal(t, 2, summary.Mismatched)
	assert.Equal(t, 1, summary.Failed)
	assert.InDelta(t, 33.3, summary.MatchRate, 0.5)
	assert.Greater(t, summary.AvgDriftScore, 0.0)
	assert.Greater(t, summary.AvgLatencyMs, 0.0)

	assert.Equal(t, 1, summary.SeverityCounts["info"])
	assert.Equal(t, 1, summary.SeverityCounts["warning"])
	assert.Equal(t, 1, summary.SeverityCounts["breaking"])
	assert.Equal(t, 1, summary.SeverityCounts["error"])

	assert.NotEmpty(t, summary.TopDriftingPaths)
}

func TestReporter_FormatText(t *testing.T) {
	reporter := NewReporter()

	summary := ReportSummary{
		SessionName:    "test-run",
		TargetURL:      "http://staging.example.com",
		TotalRequests:  100,
		Successful:     85,
		Failed:         5,
		Mismatched:     10,
		MatchRate:      89.5,
		AvgDriftScore:  0.12,
		AvgLatencyMs:   45.5,
		Duration:       3 * time.Second,
		SeverityCounts: map[string]int{"info": 85, "warning": 8, "error": 5, "breaking": 2},
	}

	output := reporter.FormatText(summary)

	assert.Contains(t, output, "TVC Replay Report")
	assert.Contains(t, output, "test-run")
	assert.Contains(t, output, "staging.example.com")
	assert.Contains(t, output, "100")
	assert.Contains(t, output, "89.5%")
}

func TestReporter_FormatJSON(t *testing.T) {
	reporter := NewReporter()

	summary := ReportSummary{
		SessionName:    "json-test",
		TotalRequests:  10,
		Successful:     8,
		SeverityCounts: map[string]int{"info": 8, "error": 2},
	}

	output, err := reporter.FormatJSON(summary)
	require.NoError(t, err)
	assert.Contains(t, output, `"session_name": "json-test"`)
	assert.Contains(t, output, `"total_requests": 10`)
}

func TestReporter_AllMatched(t *testing.T) {
	reporter := NewReporter()

	results := []Result{
		{OriginalLog: models.TrafficLog{Path: "/api"}, StatusCode: 200, LatencyMs: 10},
	}
	comparisons := []ComparisonResult{
		{StatusMatch: true, BodyMatch: true, Severity: diffing.SeverityInfo, DriftScore: 0},
	}

	summary := reporter.GenerateSummary("clean", "http://target", results, comparisons, time.Second)

	assert.Equal(t, 100.0, summary.MatchRate)
	assert.Equal(t, 0.0, summary.AvgDriftScore)
	assert.Equal(t, 1, summary.Successful)
	assert.Equal(t, 0, summary.Mismatched)
	assert.Equal(t, 0, summary.Failed)
}

func TestProgressBar(t *testing.T) {
	assert.Equal(t, "[##########..........]", progressBar(50, 100, 20))
	assert.Equal(t, "[....................]", progressBar(0, 100, 20))
	assert.Equal(t, "[####################]", progressBar(100, 100, 20))
	assert.Contains(t, progressBar(0, 0, 20), "[")
}
