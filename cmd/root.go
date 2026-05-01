// Package cmd defines the CLI commands for the k8s-label-validator tool.
package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "k8s-label-validator",
	Short:        "Validate labels on Kubernetes manifests",
	SilenceErrors: true,
	Long: `k8s-label-validator is a CLI tool that checks Kubernetes manifest files
to ensure that resources contain the required labels and that label values
conform to the specified patterns.

Use the 'validate' subcommand to check one or more manifest files or directories.`,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
