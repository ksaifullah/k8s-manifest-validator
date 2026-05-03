/*
Copyright © 2026 NAME HERE <khalid@outlook.com.au>
*/
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/ksaifullah/k8s-manifest-validator/internal/manifest"
	"github.com/ksaifullah/k8s-manifest-validator/internal/validator"
	"github.com/ksaifullah/k8s-manifest-validator/internal/validator/labels"
	"github.com/spf13/cobra"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate Kubernetes manifests for MegaTech cost centre labels",
	Long: `Validates Kubernetes manifests for the required MegaTech cost centre label (metadata.megatech.inc/cost-centre).

Reads manifests from stdin, supports multiple YAML documents, and checks each resource for:
  - Presence of the required label
  - Correct format: CC-NNN-YYYY (NNN: 050-150, YYYY: current year)
  - Valid cost centre range and year

Outputs a summary of total, valid, and invalid resources, with details of any validation errors.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// silence usage output on runtime errors so only the error message is shown
		cmd.SilenceUsage = true

		manifests, err := manifest.Parse(os.Stdin)
		if err != nil {
			return fmt.Errorf("reading manifests: %w", err)
		}

		result := validator.Validate(manifests, []validator.ValidatorFunc{labels.CostCentreValidator(time.Now().Year())})

		for _, r := range result.Resources {
			if r.Valid() {
				fmt.Printf("  VALID    %s\n", r.Identity)
			} else {
				fmt.Printf("  INVALID  %s\n", r.Identity)
				for _, e := range r.Errors {
					fmt.Printf("           - %s\n", e)
				}
			}
		}

		fmt.Printf("\nSummary:\n")
		fmt.Printf("  Total:   %d\n", result.Total())
		fmt.Printf("  Valid:   %d\n", result.ValidCount())
		fmt.Printf("  Invalid: %d\n", result.InvalidCount())

		if result.InvalidCount() > 0 {
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// validateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// validateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
