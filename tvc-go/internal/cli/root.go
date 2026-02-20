package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tvc-org/tvc/internal/config"
)

var rootCmd = &cobra.Command{
	Use:   "tvc",
	Short: "TVC - Traffic Version Control for APIs",
	Long: `TVC (Traffic Version Control) is a developer tool for API governance.

It provides schema diffing, traffic capture & replay, and breaking change detection
to help teams ship API changes with confidence.

Usage:
  tvc diff --file-old old.json --file-new new.json
  tvc schema diff --file-old v1.yaml --file-new v2.yaml --breaking-only`,
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
	Short: "Print the version of TVC",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("tvc version %s\n", config.Version)
	},
}
