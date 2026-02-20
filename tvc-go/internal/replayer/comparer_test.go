package replayer

import (
	"testing"

	"github.com/tvc-org/tvc/internal/diffing"
	"github.com/tvc-org/tvc/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestComparer_IdenticalResponses(t *testing.T) {
	comparer := NewComparer(ComparerConfig{})

	original := models.TrafficLog{
		StatusCode: 200,
		ResponseBody: map[string]interface{}{
			"id": float64(1), "name": "Alice",
		},
	}

	replayed := Result{
		StatusCode: 200,
		ResponseBody: map[string]interface{}{
			"id": float64(1), "name": "Alice",
		},
	}

	result := comparer.Compare(original, replayed)

	assert.True(t, result.StatusMatch)
	assert.True(t, result.BodyMatch)
	assert.Empty(t, result.Diffs)
	assert.Equal(t, diffing.SeverityInfo, result.Severity)
	assert.Equal(t, 0.0, result.DriftScore)
}

func TestComparer_StatusCodeMismatch(t *testing.T) {
	comparer := NewComparer(ComparerConfig{})

	original := models.TrafficLog{
		StatusCode:   200,
		ResponseBody: map[string]interface{}{"ok": true},
	}

	replayed := Result{
		StatusCode:   500,
		ResponseBody: map[string]interface{}{"error": "internal"},
	}

	result := comparer.Compare(original, replayed)

	assert.False(t, result.StatusMatch)
	assert.Equal(t, diffing.SeverityBreaking, result.Severity)
	assert.Greater(t, result.DriftScore, 0.0)
}

func TestComparer_BodyDiff(t *testing.T) {
	comparer := NewComparer(ComparerConfig{})

	original := models.TrafficLog{
		StatusCode: 200,
		ResponseBody: map[string]interface{}{
			"id": float64(1), "name": "Alice", "email": "alice@test.com",
		},
	}

	replayed := Result{
		StatusCode: 200,
		ResponseBody: map[string]interface{}{
			"id": float64(1), "name": "Alice", "email": "alice@new.com", "role": "admin",
		},
	}

	result := comparer.Compare(original, replayed)

	assert.True(t, result.StatusMatch)
	assert.False(t, result.BodyMatch)
	assert.NotEmpty(t, result.Diffs)
}

func TestComparer_NilBodies(t *testing.T) {
	comparer := NewComparer(ComparerConfig{})

	original := models.TrafficLog{StatusCode: 204, ResponseBody: nil}
	replayed := Result{StatusCode: 204, ResponseBody: nil}

	result := comparer.Compare(original, replayed)

	assert.True(t, result.StatusMatch)
	assert.True(t, result.BodyMatch)
}

func TestComparer_OriginalHasBody_ReplayedNil(t *testing.T) {
	comparer := NewComparer(ComparerConfig{})

	original := models.TrafficLog{
		StatusCode:   200,
		ResponseBody: map[string]interface{}{"data": "value"},
	}
	replayed := Result{StatusCode: 200, ResponseBody: nil}

	result := comparer.Compare(original, replayed)

	assert.True(t, result.StatusMatch)
	assert.False(t, result.BodyMatch)
}

func TestComparer_WithIgnorePaths(t *testing.T) {
	comparer := NewComparer(ComparerConfig{
		IgnorePaths: []string{"timestamp", "request_id"},
	})

	original := models.TrafficLog{
		StatusCode: 200,
		ResponseBody: map[string]interface{}{
			"data":       "value",
			"timestamp":  "2026-01-01",
			"request_id": "abc-123",
		},
	}

	replayed := Result{
		StatusCode: 200,
		ResponseBody: map[string]interface{}{
			"data":       "value",
			"timestamp":  "2026-02-20",
			"request_id": "xyz-789",
		},
	}

	result := comparer.Compare(original, replayed)

	assert.True(t, result.StatusMatch)
	assert.True(t, result.BodyMatch)
	assert.Empty(t, result.Diffs)
}

func TestComparer_CompareAll(t *testing.T) {
	comparer := NewComparer(ComparerConfig{})

	originals := []models.TrafficLog{
		{StatusCode: 200, ResponseBody: map[string]interface{}{"ok": true}},
		{StatusCode: 200, ResponseBody: map[string]interface{}{"ok": true}},
		{StatusCode: 200, ResponseBody: map[string]interface{}{"ok": true}},
	}

	results := []Result{
		{StatusCode: 200, ResponseBody: map[string]interface{}{"ok": true}},
		{StatusCode: 500, ResponseBody: map[string]interface{}{"error": "fail"}},
		{Error: assert.AnError},
	}

	comparisons := comparer.CompareAll(originals, results)

	assert.Len(t, comparisons, 3)
	assert.Equal(t, diffing.SeverityInfo, comparisons[0].Severity)
	assert.Equal(t, diffing.SeverityBreaking, comparisons[1].Severity)
	assert.Equal(t, diffing.SeverityError, comparisons[2].Severity)
}

func TestScoreSeverity(t *testing.T) {
	tests := []struct {
		name        string
		statusMatch bool
		bodyMatch   bool
		diffs       []diffing.Diff
		expected    diffing.Severity
	}{
		{"all match", true, true, nil, diffing.SeverityInfo},
		{"status mismatch", false, true, nil, diffing.SeverityBreaking},
		{"body diffs - info", true, false, []diffing.Diff{
			{Severity: diffing.SeverityInfo},
		}, diffing.SeverityInfo},
		{"body diffs - warning", true, false, []diffing.Diff{
			{Severity: diffing.SeverityWarning},
		}, diffing.SeverityWarning},
		{"body diffs - breaking", true, false, []diffing.Diff{
			{Severity: diffing.SeverityBreaking},
		}, diffing.SeverityBreaking},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, scoreSeverity(tt.statusMatch, tt.bodyMatch, tt.diffs))
		})
	}
}

func TestCalculateDriftScore(t *testing.T) {
	original := models.TrafficLog{StatusCode: 200}

	score := calculateDriftScore(true, nil, original)
	assert.Equal(t, 0.0, score)

	score = calculateDriftScore(false, nil, original)
	assert.Equal(t, 0.5, score)

	score = calculateDriftScore(true, []diffing.Diff{
		{Severity: diffing.SeverityBreaking},
	}, original)
	assert.Greater(t, score, 0.0)

	score = calculateDriftScore(false, []diffing.Diff{
		{Severity: diffing.SeverityBreaking},
	}, original)
	assert.LessOrEqual(t, score, 1.0)
}
