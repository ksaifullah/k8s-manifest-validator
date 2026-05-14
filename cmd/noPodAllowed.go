package cmd

import (
	"fmt"
	"os"

	"github.com/ksaifullah/k8s-manifest-validator/internal/manifest"
	"github.com/ksaifullah/k8s-manifest-validator/internal/validator"
	"github.com/ksaifullah/k8s-manifest-validator/internal/validator/kind"
	"github.com/spf13/cobra"
)

// validateNoPodAllowed represents the no-pod-allowed command.
var validateNoPodAllowed = &cobra.Command{
	Use:   "no-pod-allowed",
	Short: "Validate Kubernetes manifests to ensure no Pods are allowed",
	Long:  `Reads Kubernetes manifests from stdin and validates that no resources of kind Pod are present.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// silence usage output on runtime errors so only the error message is shown
		cmd.SilenceUsage = true

		manifests, err := manifest.Parse(os.Stdin)
		if err != nil {
			return fmt.Errorf("reading manifests: %w", err)
		}

		result := validator.Validate(manifests, []validator.ValidatorFunc{kind.NoPodValidator()})

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
			return errValidationFailed
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateNoPodAllowed)
}
