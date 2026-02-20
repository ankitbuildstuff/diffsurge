package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/tvc-org/tvc/internal/models"
	"github.com/tvc-org/tvc/internal/replayer"
	"github.com/tvc-org/tvc/pkg/logger"
)

var (
	replayTarget     string
	replaySource     string
	replayWorkers    int
	replayRateLimit  int
	replayTimeout    string
	replayFormat     string
	replayOutput     string
	replayMaxRetries int
)

var replayCmd = &cobra.Command{
	Use:   "replay",
	Short: "Replay traffic against a target server",
	Long: `Replay captured traffic against a target server and compare responses.

Can replay from a JSON file of captured traffic or from the database.

Examples:
  tvc replay --source traffic.json --target http://staging.example.com
  tvc replay --source traffic.json --target http://localhost:8080 --workers 20 --rate-limit 500`,
	RunE: runReplay,
}

func init() {
	replayCmd.Flags().StringVar(&replayTarget, "target", "", "Target server URL to replay against (required)")
	replayCmd.Flags().StringVar(&replaySource, "source", "", "Source traffic JSON file (required)")
	replayCmd.Flags().IntVar(&replayWorkers, "workers", 10, "Number of concurrent workers")
	replayCmd.Flags().IntVar(&replayRateLimit, "rate-limit", 0, "Max requests per second (0 = unlimited)")
	replayCmd.Flags().StringVar(&replayTimeout, "timeout", "30s", "Per-request timeout")
	replayCmd.Flags().StringVar(&replayFormat, "format", "text", "Output format: text, json")
	replayCmd.Flags().StringVar(&replayOutput, "output", "", "Write report to file")
	replayCmd.Flags().IntVar(&replayMaxRetries, "max-retries", 2, "Max retries per request")

	_ = replayCmd.MarkFlagRequired("target")
	_ = replayCmd.MarkFlagRequired("source")

	rootCmd.AddCommand(replayCmd)
}

func runReplay(cmd *cobra.Command, args []string) error {
	timeout, err := time.ParseDuration(replayTimeout)
	if err != nil {
		return fmt.Errorf("invalid timeout: %w", err)
	}

	log := logger.New("info", "console")

	traffic, err := loadTrafficFromFile(replaySource)
	if err != nil {
		return fmt.Errorf("loading traffic: %w", err)
	}

	if len(traffic) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No traffic to replay.")
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Loaded %d requests from %s\n", len(traffic), replaySource)

	// Validate target is reachable
	if err := pingTarget(replayTarget); err != nil {
		return fmt.Errorf("target unreachable: %w", err)
	}

	filtered := replayer.FilterTrafficForReplay(traffic, replayer.FilterConfig{
		StripSensitiveHeaders: true,
	})

	r := replayer.New(replayer.Config{
		Workers:    replayWorkers,
		RateLimit:  replayRateLimit,
		Timeout:    timeout,
		MaxRetries: replayMaxRetries,
		TargetURL:  replayTarget,
	}, log)

	comparer := replayer.NewComparer(replayer.ComparerConfig{})
	reporter := replayer.NewReporter()

	ctx := context.Background()
	start := time.Now()

	fmt.Fprintf(cmd.OutOrStdout(), "Replaying %d requests to %s (workers: %d)...\n\n",
		len(filtered), replayTarget, replayWorkers)

	results, err := r.ReplayTraffic(ctx, filtered)
	if err != nil {
		return fmt.Errorf("replay failed: %w", err)
	}

	comparisons := comparer.CompareAll(filtered, results)
	duration := time.Since(start)

	summary := reporter.GenerateSummary("CLI Replay", replayTarget, results, comparisons, duration)

	var output string
	switch replayFormat {
	case "json":
		output, err = reporter.FormatJSON(summary)
		if err != nil {
			return fmt.Errorf("formatting report: %w", err)
		}
	default:
		output = reporter.FormatText(summary)
	}

	if replayOutput != "" {
		if writeErr := os.WriteFile(replayOutput, []byte(output), 0644); writeErr != nil {
			return fmt.Errorf("writing output: %w", writeErr)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Report written to %s\n", replayOutput)
	} else {
		fmt.Fprint(cmd.OutOrStdout(), output)
	}

	if summary.Mismatched > 0 || summary.Failed > 0 {
		os.Exit(1)
	}

	return nil
}

func loadTrafficFromFile(path string) ([]models.TrafficLog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	// Try parsing as array of traffic logs
	var logs []models.TrafficLog
	if err := json.Unmarshal(data, &logs); err == nil {
		for i := range logs {
			if logs[i].ID == uuid.Nil {
				logs[i].ID = uuid.New()
			}
		}
		return logs, nil
	}

	// Try parsing as object with "traffic" key
	var wrapper struct {
		Traffic []models.TrafficLog `json:"traffic"`
	}
	if err := json.Unmarshal(data, &wrapper); err == nil && len(wrapper.Traffic) > 0 {
		for i := range wrapper.Traffic {
			if wrapper.Traffic[i].ID == uuid.Nil {
				wrapper.Traffic[i].ID = uuid.New()
			}
		}
		return wrapper.Traffic, nil
	}

	return nil, fmt.Errorf("could not parse traffic from file (expected JSON array or {\"traffic\": [...]})")
}

func pingTarget(targetURL string) error {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(targetURL)
	if err != nil {
		// Try a HEAD request if GET fails
		resp, err = client.Head(targetURL)
		if err != nil {
			return err
		}
	}
	resp.Body.Close()
	return nil
}

