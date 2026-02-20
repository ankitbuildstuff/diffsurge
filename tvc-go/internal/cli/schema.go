package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tvc-org/tvc/internal/diffing"
)

var (
	schemaFileOld     string
	schemaFileNew     string
	schemaFormat      string
	schemaOutput      string
	schemaBreakingOnly bool
	schemaFailOnBreaking bool
)

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Schema management commands",
	Long:  "Commands for working with API schemas (OpenAPI, Swagger, GraphQL).",
}

var schemaDiffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Compare two OpenAPI/Swagger schema files",
	Long: `Compare two OpenAPI schema files and detect breaking changes.

Supports OpenAPI 3.0 and 3.1 specifications. Detects breaking changes such as:
  - Removed endpoints
  - Removed required fields
  - Changed field types
  - Added required parameters
  - Removed response codes

Exit codes:
  0 - No breaking changes (or no changes)
  1 - Breaking changes detected (with --fail-on-breaking)
  2 - Error occurred`,
	RunE: runSchemaDiff,
}

func init() {
	schemaDiffCmd.Flags().StringVar(&schemaFileOld, "file-old", "", "Path to the old schema file (required)")
	schemaDiffCmd.Flags().StringVar(&schemaFileNew, "file-new", "", "Path to the new schema file (required)")
	schemaDiffCmd.Flags().StringVar(&schemaFormat, "format", "text", "Output format: text, json")
	schemaDiffCmd.Flags().StringVar(&schemaOutput, "output", "", "Write output to file")
	schemaDiffCmd.Flags().BoolVar(&schemaBreakingOnly, "breaking-only", false, "Show only breaking changes")
	schemaDiffCmd.Flags().BoolVar(&schemaFailOnBreaking, "fail-on-breaking", false, "Exit with code 1 if breaking changes found")

	_ = schemaDiffCmd.MarkFlagRequired("file-old")
	_ = schemaDiffCmd.MarkFlagRequired("file-new")

	schemaCmd.AddCommand(schemaDiffCmd)
}

func runSchemaDiff(cmd *cobra.Command, args []string) error {
	comparer := diffing.NewSchemaComparer()

	diffs, breakingChanges, err := comparer.CompareFiles(schemaFileOld, schemaFileNew)
	if err != nil {
		return fmt.Errorf("schema diff failed: %w", err)
	}

	var output string

	if schemaBreakingOnly {
		formatter := diffing.NewBreakingChangeFormatter(isTerminal())
		output, err = formatter.FormatBreakingChanges(breakingChanges)
	} else {
		switch schemaFormat {
		case "json":
			formatter := diffing.NewJSONFormatter(true)
			output, err = formatter.Format(diffs)
		default:
			formatter := diffing.NewTextFormatter(isTerminal())
			output, err = formatter.Format(diffs)

			if len(breakingChanges) > 0 {
				bcFormatter := diffing.NewBreakingChangeFormatter(isTerminal())
				bcOutput, bcErr := bcFormatter.FormatBreakingChanges(breakingChanges)
				if bcErr != nil {
					return fmt.Errorf("formatting breaking changes: %w", bcErr)
				}
				output += "\n" + bcOutput
			}
		}
	}

	if err != nil {
		return fmt.Errorf("formatting output: %w", err)
	}

	if schemaOutput != "" {
		if writeErr := os.WriteFile(schemaOutput, []byte(output), 0644); writeErr != nil {
			return fmt.Errorf("writing output file: %w", writeErr)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Output written to %s\n", schemaOutput)
	} else {
		fmt.Fprint(cmd.OutOrStdout(), output)
	}

	if schemaFailOnBreaking && len(breakingChanges) > 0 {
		os.Exit(1)
	}

	return nil
}
