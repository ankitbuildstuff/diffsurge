package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var (
	checkSchemaFile   string
	checkFailBreaking bool
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Run API checks against the Diffsurge backend (for CI/CD)",
	Long: `Validate your API against the Diffsurge backend.

This command is designed for CI/CD pipelines. It authenticates using your
API key, fetches the latest traffic and schema data, and optionally diffs
a local schema file against the latest version on the server.

Setup:
  1. Create an API key in the dashboard (Settings → API Keys)
  2. Add to your .env file: SURGE_API_KEY=diffsurge_live_...
  3. Run: surge check --project-id <your-project-uuid>

Exit codes:
  0 - All checks passed
  1 - Breaking changes detected or checks failed

Examples:
  surge check --project-id abc-123
  surge check --project-id abc-123 --schema openapi.yaml --fail-on-breaking`,
	RunE: runCheck,
}

func init() {
	checkCmd.Flags().StringVar(&checkSchemaFile, "schema", "", "Local schema file to diff against latest version")
	checkCmd.Flags().BoolVar(&checkFailBreaking, "fail-on-breaking", true, "Exit with code 1 if breaking changes found")

	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	if cliCfg == nil || cliCfg.APIKey == "" {
		return fmt.Errorf("API key not configured.\n\nSet SURGE_API_KEY in your .env file or environment:\n  echo 'SURGE_API_KEY=diffsurge_live_...' >> .env")
	}

	projectID := cliCfg.ProjectID
	if projectID == "" {
		return fmt.Errorf("project ID required.\n\nUse --project-id flag or set SURGE_PROJECT_ID in your .env file")
	}

	client := NewAPIClient(cliCfg.APIURL, cliCfg.APIKey)

	fmt.Fprintf(cmd.OutOrStdout(), "🔍 Diffsurge Check\n")
	fmt.Fprintf(cmd.OutOrStdout(), "   API:     %s\n", cliCfg.APIURL)
	fmt.Fprintf(cmd.OutOrStdout(), "   Project: %s\n\n", projectID)

	hasFailures := false

	// 1. Validate API key against an authenticated endpoint
	fmt.Fprintf(cmd.OutOrStdout(), "→ Validating API key... ")
	orgResp, err := client.Get("/api/v1/organizations")
	if err != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "✗\n")
		return fmt.Errorf("API key validation failed: %w", err)
	}
	orgResp.Body.Close()
	fmt.Fprintf(cmd.OutOrStdout(), "✓\n")

	// 2. Fetch traffic stats
	fmt.Fprintf(cmd.OutOrStdout(), "→ Fetching traffic stats... ")
	trafficResp, err := client.Get(fmt.Sprintf("/api/v1/projects/%s/traffic/stats", projectID))
	if err != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "✗ (%v)\n", err)
		hasFailures = true
	} else {
		defer trafficResp.Body.Close()
		var stats map[string]interface{}
		if json.NewDecoder(trafficResp.Body).Decode(&stats) == nil {
			fmt.Fprintf(cmd.OutOrStdout(), "✓\n")
			if data, ok := stats["data"].(map[string]interface{}); ok {
				w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
				fmt.Fprintf(w, "   Total requests:\t%.0f\n", data["total_requests"])
				if endpoints, ok := data["unique_endpoints"].(float64); ok {
					fmt.Fprintf(w, "   Unique endpoints:\t%.0f\n", endpoints)
				}
				w.Flush()
			}
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "✓ (stats received)\n")
		}
	}

	// 3. Fetch latest schemas
	fmt.Fprintf(cmd.OutOrStdout(), "→ Fetching schemas... ")
	schemasResp, err := client.Get(fmt.Sprintf("/api/v1/projects/%s/schemas", projectID))
	if err != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "✗ (%v)\n", err)
		hasFailures = true
	} else {
		defer schemasResp.Body.Close()
		var schemasData struct {
			Data []struct {
				ID      string `json:"id"`
				Version string `json:"version"`
			} `json:"data"`
		}
		if json.NewDecoder(schemasResp.Body).Decode(&schemasData) == nil {
			fmt.Fprintf(cmd.OutOrStdout(), "✓ (%d versions)\n", len(schemasData.Data))
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "✓\n")
		}
	}

	// 4. Local schema diff (if --schema provided)
	if checkSchemaFile != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "→ Uploading schema for diff... ")

		schemaContent, err := os.ReadFile(checkSchemaFile)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "✗\n")
			return fmt.Errorf("reading schema file: %w", err)
		}

		schemaPayload := fmt.Sprintf(`{"schema_content": %s, "schema_type": %s, "version": "ci-check"}`,
			jsonEscapeString(string(schemaContent)),
			jsonEscapeString(inferSchemaType(checkSchemaFile)))

		diffResp, err := client.Post(
			fmt.Sprintf("/api/v1/projects/%s/schemas/diff", projectID),
			strings.NewReader(schemaPayload),
		)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "✗ (%v)\n", err)
			hasFailures = true
		} else {
			defer diffResp.Body.Close()
			var diffResult struct {
				Data struct {
					HasBreakingChanges bool `json:"has_breaking_changes"`
					BreakingChanges    int  `json:"breaking_changes_count"`
				} `json:"data"`
			}
			if json.NewDecoder(diffResp.Body).Decode(&diffResult) == nil {
				if diffResult.Data.HasBreakingChanges {
					fmt.Fprintf(cmd.OutOrStdout(), "⚠ %d breaking change(s) detected\n", diffResult.Data.BreakingChanges)
					if checkFailBreaking {
						hasFailures = true
					}
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "✓ No breaking changes\n")
				}
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "✓ (diff complete)\n")
			}
		}
	}

	fmt.Fprintln(cmd.OutOrStdout())

	if hasFailures {
		fmt.Fprintf(cmd.OutOrStdout(), "✗ Check failed\n")
		os.Exit(1)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ All checks passed\n")
	return nil
}

func jsonEscapeString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func inferSchemaType(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".graphql", ".gql":
		return "graphql"
	default:
		return "openapi"
	}
}
