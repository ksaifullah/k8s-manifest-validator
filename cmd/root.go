// Package cmd contains the CLI commands for k8s-manifest-validator.
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// errValidationFailed is the sentinel returned by RunE when one or more resources
// fail validation. Using a sentinel lets Execute() exit with code 1 without
// printing a duplicate error message — the per-resource output is sufficient.
var errValidationFailed = errors.New("validation failed")

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "k8s-manifest-validator",
	Short: "Validate Kubernetes manifests for MegaTech compliance",
	Long: `k8s-manifest-validator is a command-line tool for validating Kubernetes manifests at MegaTech.

Use the 'validate-cost-centre' command to check manifests for required cost centre labels and compliance with MegaTech policies.`,
	SilenceErrors: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		// errValidationFailed is already communicated via per-resource output; suppress it.
		// All other errors are genuine failures and should be surfaced to the user.
		if !errors.Is(err, errValidationFailed) {
			fmt.Fprintln(os.Stderr, "Error:", err)
		}
		os.Exit(1)
	}
}
