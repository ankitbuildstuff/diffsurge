package cli

import (
	"fmt"
	"os"

	"github.com/diffsurge-org/diffsurge/internal/config"
	"github.com/spf13/cobra"
)

var (
	// CLI config loaded from .env / env vars
	cliCfg *config.CLIConfig

	// Flag overrides
	flagAPIKey    string
	flagAPIURL    string
	flagProjectID string
)

var rootCmd = &cobra.Command{
	Use:   "surge",
	Short: "Diffsurge - Catch breaking API changes before your users do",
	Long: `Diffsurge is a developer tool for API governance.

It provides schema diffing, traffic capture & replay, and breaking change detection
to help teams ship API changes with confidence.

Usage:
  surge diff --old api-v1.json --new api-v2.json
  surge schema diff --old v1.yaml --new v2.yaml --fail-on-breaking
  surge replay --source traffic.json --target http://staging.example.com
  surge check --project-id <uuid>
  surge whoami`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Load CLI config from .env and env vars
		cliCfg = config.LoadCLI()

		// Override with flags if provided
		if flagAPIKey != "" {
			cliCfg.APIKey = flagAPIKey
		}
		if flagAPIURL != "" {
			cliCfg.APIURL = flagAPIURL
		}
		if flagProjectID != "" {
			cliCfg.ProjectID = flagProjectID
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Persistent flags available to all subcommands
	rootCmd.PersistentFlags().StringVar(&flagAPIKey, "api-key", "", "API key (overrides SURGE_API_KEY env var)")
	rootCmd.PersistentFlags().StringVar(&flagAPIURL, "api-url", "", "API base URL (overrides SURGE_API_URL env var)")
	rootCmd.PersistentFlags().StringVar(&flagProjectID, "project-id", "", "Default project ID (overrides SURGE_PROJECT_ID env var)")

	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(schemaCmd)
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of Diffsurge",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("surge version %s\n", config.Version)
	},
}
