package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tvc-org/tvc/internal/config"
)

var rootCmd = &cobra.Command{
	Use:   "surge",
	Short: "Driftsurge - Catch breaking API changes before your users do",
	Long: `Driftsurge is a developer tool for API governance.

It provides schema diffing, traffic capture & replay, and breaking change detection
to help teams ship API changes with confidence.

Usage:
  surge diff --old api-v1.json --new api-v2.json
  surge schema diff --old v1.yaml --new v2.yaml --fail-on-breaking
  surge replay --source traffic.json --target http://staging.example.com`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(schemaCmd)
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of Driftsurge",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("surge version %s\n", config.Version)
	},
}
