package replayer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

type ReportSummary struct {
	SessionName      string         `json:"session_name"`
	TargetURL        string         `json:"target_url"`
	TotalRequests    int            `json:"total_requests"`
	Successful       int            `json:"successful"`
	Failed           int            `json:"failed"`
	Mismatched       int            `json:"mismatched"`
	MatchRate        float64        `json:"match_rate"`
	AvgDriftScore    float64        `json:"avg_drift_score"`
	AvgLatencyMs     float64        `json:"avg_latency_ms"`
	Duration         time.Duration  `json:"duration"`
	SeverityCounts   map[string]int `json:"severity_counts"`
	TopDriftingPaths []PathDrift    `json:"top_drifting_paths,omitempty"`
}

type PathDrift struct {
	Path       string  `json:"path"`
	DriftScore float64 `json:"drift_score"`
	DiffCount  int     `json:"diff_count"`
}

type Reporter struct{}

func NewReporter() *Reporter {
	return &Reporter{}
}

func (r *Reporter) GenerateSummary(name, targetURL string, results []Result, comparisons []ComparisonResult, duration time.Duration) ReportSummary {
	summary := ReportSummary{
		SessionName:    name,
		TargetURL:      targetURL,
		TotalRequests:  len(results),
		Duration:       duration,
		SeverityCounts: make(map[string]int),
	}

	var totalDrift float64
	var totalLatency int64
	pathDrifts := make(map[string]*PathDrift)

	for i, result := range results {
		if result.Error != nil {
			summary.Failed++
			summary.SeverityCounts["error"]++
			continue
		}

		comp := comparisons[i]
		summary.SeverityCounts[string(comp.Severity)]++
		totalDrift += comp.DriftScore
		totalLatency += int64(result.LatencyMs)

		if comp.StatusMatch && comp.BodyMatch {
			summary.Successful++
		} else {
			summary.Mismatched++
		}

		path := result.OriginalLog.Path
		if _, exists := pathDrifts[path]; !exists {
			pathDrifts[path] = &PathDrift{Path: path}
		}
		pd := pathDrifts[path]
		pd.DriftScore += comp.DriftScore
		pd.DiffCount += len(comp.Diffs)
	}

	responded := summary.TotalRequests - summary.Failed
	if responded > 0 {
		summary.MatchRate = float64(summary.Successful) / float64(responded) * 100
		summary.AvgDriftScore = totalDrift / float64(responded)
		summary.AvgLatencyMs = float64(totalLatency) / float64(responded)
	}

	driftList := make([]PathDrift, 0, len(pathDrifts))
	for _, pd := range pathDrifts {
		driftList = append(driftList, *pd)
	}
	sort.Slice(driftList, func(i, j int) bool {
		return driftList[i].DriftScore > driftList[j].DriftScore
	})
	if len(driftList) > 10 {
		driftList = driftList[:10]
	}
	summary.TopDriftingPaths = driftList

	return summary
}

func (r *Reporter) FormatText(summary ReportSummary) string {
	var buf bytes.Buffer

	buf.WriteString("╔══════════════════════════════════════════════════╗\n")
	buf.WriteString("║            TVC Replay Report                    ║\n")
	buf.WriteString("╚══════════════════════════════════════════════════╝\n\n")

	buf.WriteString(fmt.Sprintf("  Session:    %s\n", summary.SessionName))
	buf.WriteString(fmt.Sprintf("  Target:     %s\n", summary.TargetURL))
	buf.WriteString(fmt.Sprintf("  Duration:   %s\n", summary.Duration.Round(time.Millisecond)))
	buf.WriteString(fmt.Sprintf("  Avg Latency: %.1fms\n\n", summary.AvgLatencyMs))

	buf.WriteString("  ── Results ───────────────────────────────────\n")
	buf.WriteString(fmt.Sprintf("  Total:      %d\n", summary.TotalRequests))
	buf.WriteString(fmt.Sprintf("  Matched:    %d  (%.1f%%)\n", summary.Successful, summary.MatchRate))
	buf.WriteString(fmt.Sprintf("  Mismatched: %d\n", summary.Mismatched))
	buf.WriteString(fmt.Sprintf("  Failed:     %d\n", summary.Failed))
	buf.WriteString(fmt.Sprintf("  Drift Score: %.3f\n\n", summary.AvgDriftScore))

	buf.WriteString("  ── Severity Breakdown ────────────────────────\n")
	severities := []string{"info", "warning", "error", "breaking"}
	for _, s := range severities {
		count := summary.SeverityCounts[s]
		if count > 0 {
			bar := progressBar(count, summary.TotalRequests, 20)
			buf.WriteString(fmt.Sprintf("  %-10s %s %d\n", s, bar, count))
		}
	}
	buf.WriteString("\n")

	if len(summary.TopDriftingPaths) > 0 {
		buf.WriteString("  ── Top Drifting Paths ────────────────────────\n")
		for i, pd := range summary.TopDriftingPaths {
			if pd.DriftScore > 0 {
				buf.WriteString(fmt.Sprintf("  %d. %s  (drift: %.3f, diffs: %d)\n",
					i+1, pd.Path, pd.DriftScore, pd.DiffCount))
			}
		}
		buf.WriteString("\n")
	}

	return buf.String()
}

func (r *Reporter) FormatJSON(summary ReportSummary) (string, error) {
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling report: %w", err)
	}
	return string(data), nil
}

func progressBar(value, total, width int) string {
	if total == 0 {
		return fmt.Sprintf("[%*s]", width, "")
	}
	filled := value * width / total
	if filled > width {
		filled = width
	}
	var runes []rune
	for i := 0; i < width; i++ {
		if i < filled {
			runes = append(runes, '#')
		} else {
			runes = append(runes, '.')
		}
	}
	return "[" + string(runes) + "]"
}
