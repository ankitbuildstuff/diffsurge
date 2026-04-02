package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/diffsurge-org/diffsurge/internal/diffing"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	schemaFileOld        string
	schemaFileNew        string
	schemaFormat         string
	schemaOutput         string
	schemaBreakingOnly   bool
	schemaFailOnBreaking bool
	schemaPushFile       string
	schemaPushVersion    string
	schemaPushType       string
	schemaPushGitCommit  string
	schemaPushGitBranch  string
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

var schemaPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Upload a schema file to the dashboard",
	Long: `Upload an OpenAPI or GraphQL schema file to the Diffsurge dashboard.

Supports JSON and YAML files. YAML files are automatically converted to JSON before upload.

Examples:
  surge schema push --file openapi.yaml --version v2.0.0
  surge schema push --file schema.json --version v1.0.0 --type openapi
  surge schema push --file openapi.yaml --version v2.1.0 --git-commit abc1234 --git-branch main`,
	RunE: runSchemaPush,
}

func init() {
	schemaDiffCmd.Flags().StringVar(&schemaFileOld, "old", "", "Path to the old schema file (required)")
	schemaDiffCmd.Flags().StringVar(&schemaFileNew, "new", "", "Path to the new schema file (required)")
	schemaDiffCmd.Flags().StringVar(&schemaFormat, "format", "text", "Output format: text, json")
	schemaDiffCmd.Flags().StringVar(&schemaOutput, "output", "", "Write output to file")
	schemaDiffCmd.Flags().BoolVar(&schemaBreakingOnly, "breaking-only", false, "Show only breaking changes")
	schemaDiffCmd.Flags().BoolVar(&schemaFailOnBreaking, "fail-on-breaking", false, "Exit with code 1 if breaking changes found")

	_ = schemaDiffCmd.MarkFlagRequired("old")
	_ = schemaDiffCmd.MarkFlagRequired("new")

	schemaPushCmd.Flags().StringVar(&schemaPushFile, "file", "", "Path to the schema file (required)")
	schemaPushCmd.Flags().StringVar(&schemaPushVersion, "version", "", "Schema version label, e.g. v2.0.0 (required)")
	schemaPushCmd.Flags().StringVar(&schemaPushType, "type", "openapi", "Schema type: openapi, graphql")
	schemaPushCmd.Flags().StringVar(&schemaPushGitCommit, "git-commit", "", "Git commit SHA (optional)")
	schemaPushCmd.Flags().StringVar(&schemaPushGitBranch, "git-branch", "", "Git branch name (optional)")

	_ = schemaPushCmd.MarkFlagRequired("file")
	_ = schemaPushCmd.MarkFlagRequired("version")

	schemaCmd.AddCommand(schemaDiffCmd)
	schemaCmd.AddCommand(schemaPushCmd)
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

func runSchemaPush(cmd *cobra.Command, args []string) error {
	if cliCfg == nil || cliCfg.APIKey == "" {
		return fmt.Errorf("API key not configured.\n\nSet SURGE_API_KEY in your .env file or environment:\n  echo 'SURGE_API_KEY=diffsurge_live_...' >> .env")
	}
	if cliCfg.ProjectID == "" {
		return fmt.Errorf("project ID not configured.\n\nSet SURGE_PROJECT_ID or use --project-id flag")
	}

	// Read the schema file
	data, err := os.ReadFile(schemaPushFile)
	if err != nil {
		return fmt.Errorf("reading schema file: %w", err)
	}

	// Parse the schema content — convert YAML to JSON-compatible structure if needed
	var schemaContent interface{}
	ext := strings.ToLower(filepath.Ext(schemaPushFile))
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &schemaContent); err != nil {
			return fmt.Errorf("parsing YAML schema: %w", err)
		}
		// Convert yaml.Node map keys from interface{} to string for JSON compat
		schemaContent = yamlToJSONCompat(schemaContent)
	case ".json":
		if err := json.Unmarshal(data, &schemaContent); err != nil {
			return fmt.Errorf("parsing JSON schema: %w", err)
		}
	default:
		// Try JSON first, fall back to YAML
		if err := json.Unmarshal(data, &schemaContent); err != nil {
			if err2 := yaml.Unmarshal(data, &schemaContent); err2 != nil {
				return fmt.Errorf("could not parse schema file as JSON or YAML")
			}
			schemaContent = yamlToJSONCompat(schemaContent)
		}
	}

	payload := map[string]interface{}{
		"version":        schemaPushVersion,
		"schema_type":    schemaPushType,
		"schema_content": schemaContent,
	}
	if schemaPushGitCommit != "" {
		payload["git_commit"] = schemaPushGitCommit
	}
	if schemaPushGitBranch != "" {
		payload["git_branch"] = schemaPushGitBranch
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling payload: %w", err)
	}

	client := NewAPIClient(cliCfg.APIURL, cliCfg.APIKey)
	path := fmt.Sprintf("/api/v1/projects/%s/schemas", cliCfg.ProjectID)

	fmt.Fprintf(cmd.OutOrStdout(), "→ Uploading schema %s (%s) from %s...\n", schemaPushVersion, schemaPushType, schemaPushFile)

	resp, err := client.Post(path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		ID string `json:"id"`
	}
	respBody, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(respBody, &result)

	fmt.Fprintf(cmd.OutOrStdout(), "✓ Schema %s uploaded successfully\n", schemaPushVersion)
	if result.ID != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "  Schema ID: %s\n", result.ID)
	}

	return nil
}

// yamlToJSONCompat recursively converts YAML-parsed structures to JSON-compatible types.
// YAML unmarshaling in Go can produce map[interface{}]interface{} which json.Marshal cannot handle.
func yamlToJSONCompat(v interface{}) interface{} {
	switch val := v.(type) {
	case map[interface{}]interface{}:
		m := make(map[string]interface{}, len(val))
		for k, v := range val {
			m[fmt.Sprintf("%v", k)] = yamlToJSONCompat(v)
		}
		return m
	case map[string]interface{}:
		m := make(map[string]interface{}, len(val))
		for k, v := range val {
			m[k] = yamlToJSONCompat(v)
		}
		return m
	case []interface{}:
		for i, v := range val {
			val[i] = yamlToJSONCompat(v)
		}
		return val
	default:
		return v
	}
}
