package replayer

import (
	"github.com/tvc-org/tvc/internal/diffing"
	"github.com/tvc-org/tvc/internal/models"
)

type ComparisonResult struct {
	StatusMatch bool             `json:"status_match"`
	BodyMatch   bool             `json:"body_match"`
	Diffs       []diffing.Diff   `json:"diffs,omitempty"`
	Severity    diffing.Severity `json:"severity"`
	DriftScore  float64          `json:"drift_score"`
}

type ComparerConfig struct {
	IgnorePaths      []string
	IgnoreHeaders    []string
	TreatArraysAsSet bool
}

type Comparer struct {
	engine *diffing.Engine
	config ComparerConfig
}

func NewComparer(cfg ComparerConfig) *Comparer {
	return &Comparer{
		engine: diffing.NewEngine(diffing.Config{
			IgnorePaths:      cfg.IgnorePaths,
			TreatArraysAsSet: cfg.TreatArraysAsSet,
		}),
		config: cfg,
	}
}

func (c *Comparer) Compare(original models.TrafficLog, replayed Result) ComparisonResult {
	statusMatch := original.StatusCode == replayed.StatusCode

	var bodyDiffs []diffing.Diff
	bodyMatch := true

	if original.ResponseBody != nil || replayed.ResponseBody != nil {
		var origBody interface{} = original.ResponseBody
		var replayBody interface{} = replayed.ResponseBody

		diffs, err := c.engine.Compare(origBody, replayBody)
		if err == nil && len(diffs) > 0 {
			bodyDiffs = diffs
			bodyMatch = false
		}
	}

	severity := scoreSeverity(statusMatch, bodyMatch, bodyDiffs)
	driftScore := calculateDriftScore(statusMatch, bodyDiffs, original)

	return ComparisonResult{
		StatusMatch: statusMatch,
		BodyMatch:   bodyMatch,
		Diffs:       bodyDiffs,
		Severity:    severity,
		DriftScore:  driftScore,
	}
}

func (c *Comparer) CompareAll(original []models.TrafficLog, results []Result) []ComparisonResult {
	comparisons := make([]ComparisonResult, len(results))
	for i, result := range results {
		if result.Error != nil {
			comparisons[i] = ComparisonResult{
				Severity:   diffing.SeverityError,
				DriftScore: 1.0,
			}
			continue
		}
		comparisons[i] = c.Compare(original[i], result)
	}
	return comparisons
}

func scoreSeverity(statusMatch, bodyMatch bool, diffs []diffing.Diff) diffing.Severity {
	if statusMatch && bodyMatch {
		return diffing.SeverityInfo
	}

	if !statusMatch {
		return diffing.SeverityBreaking
	}

	maxSeverity := diffing.SeverityInfo
	for _, d := range diffs {
		if severityRank(d.Severity) > severityRank(maxSeverity) {
			maxSeverity = d.Severity
		}
	}

	// Promote if many diffs exist
	if len(diffs) > 10 && maxSeverity == diffing.SeverityInfo {
		return diffing.SeverityWarning
	}

	return maxSeverity
}

func severityRank(s diffing.Severity) int {
	switch s {
	case diffing.SeverityInfo:
		return 0
	case diffing.SeverityWarning:
		return 1
	case diffing.SeverityError:
		return 2
	case diffing.SeverityBreaking:
		return 3
	default:
		return 0
	}
}

func calculateDriftScore(statusMatch bool, diffs []diffing.Diff, original models.TrafficLog) float64 {
	if statusMatch && len(diffs) == 0 {
		return 0.0
	}

	score := 0.0

	if !statusMatch {
		score += 0.5
	}

	if len(diffs) > 0 {
		breakingCount := 0
		for _, d := range diffs {
			switch d.Severity {
			case diffing.SeverityBreaking:
				breakingCount += 4
			case diffing.SeverityError:
				breakingCount += 2
			case diffing.SeverityWarning:
				breakingCount++
			}
		}
		// Normalize: cap at 0.5 for body drift
		bodyScore := float64(breakingCount) / float64(len(diffs)*4)
		if bodyScore > 0.5 {
			bodyScore = 0.5
		}
		score += bodyScore
	}

	if score > 1.0 {
		score = 1.0
	}
	return score
}
