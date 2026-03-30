package cli

import (
	"fmt"
	"os"

	"github.com/diffsurge-org/diffsurge/internal/diffing"
	"github.com/spf13/cobra"
)

var (
	diffFileOld    string
	diffFileNew    string
	diffFormat     string
	diffOutput     string
	diffIgnore     []string
	diffArrayAsSet bool
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Compare two JSON files and show differences",
	Long: `Compare two JSON files and produce a detailed diff report.

Supports multiple output formats (text, json) and can be used in CI/CD
pipelines with appropriate exit codes.

Exit codes:
  0 - No differences found
  1 - Differences detected
  2 - Error occurred`,
	RunE: runDiff,
}

func init() {
	diffCmd.Flags().StringVar(&diffFileOld, "old", "", "Path to the old/original JSON file (required)")
	diffCmd.Flags().StringVar(&diffFileNew, "new", "", "Path to the new/modified JSON file (required)")
	diffCmd.Flags().StringVar(&diffFormat, "format", "text", "Output format: text, json")
	diffCmd.Flags().StringVar(&diffOutput, "output", "", "Write output to file instead of stdout")
	diffCmd.Flags().StringSliceVar(&diffIgnore, "ignore", nil, "JSON paths to ignore (comma-separated)")
	diffCmd.Flags().BoolVar(&diffArrayAsSet, "array-as-set", false, "Compare arrays as sets (ignore order)")

	_ = diffCmd.MarkFlagRequired("old")
	_ = diffCmd.MarkFlagRequired("new")
}

func runDiff(cmd *cobra.Command, args []string) error {
	cfg := diffing.Config{
		IgnorePaths:      diffIgnore,
		TreatArraysAsSet: diffArrayAsSet,
	}

	diffs, err := diffing.DiffJSONFiles(diffFileOld, diffFileNew, cfg)
	if err != nil {
		return fmt.Errorf("diff failed: %w", err)
	}

	var formatter diffing.Formatter
	switch diffFormat {
	case "json":
		formatter = diffing.NewJSONFormatter(true)
	default:
		formatter = diffing.NewTextFormatter(isTerminal())
	}

	output, err := formatter.Format(diffs)
	if err != nil {
		return fmt.Errorf("formatting output: %w", err)
	}

	if diffOutput != "" {
		if err := os.WriteFile(diffOutput, []byte(output), 0644); err != nil {
			return fmt.Errorf("writing output file: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Output written to %s\n", diffOutput)
	} else {
		fmt.Fprint(cmd.OutOrStdout(), output)
	}

	if len(diffs) > 0 {
		os.Exit(1)
	}

	return nil
}

func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
