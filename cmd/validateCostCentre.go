package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/ksaifullah/k8s-manifest-validator/internal/manifest"
	"github.com/ksaifullah/k8s-manifest-validator/internal/validator"
	"github.com/ksaifullah/k8s-manifest-validator/internal/validator/labels"
	"github.com/spf13/cobra"
)

var yearFlag string

// validateCostCentreCmd represents the validate-cost-centre command.
var validateCostCentreCmd = &cobra.Command{
	Use:   "validate-cost-centre",
	Short: "Validate Kubernetes manifests for the MegaTech cost centre label",
	Long:  `Reads Kubernetes manifests from stdin and validates that each resource carries a valid metadata.megatech.inc/cost-centre label in the format CC-NNN-YYYY.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if matched, err := regexp.MatchString(`^\d{4}$`, yearFlag); err != nil || !matched {
			return fmt.Errorf("invalid --year flag: %q is not a valid 4-digit year", yearFlag)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// silence usage output on runtime errors so only the error message is shown
		cmd.SilenceUsage = true

		manifests, err := manifest.Parse(os.Stdin)
		if err != nil {
			return fmt.Errorf("reading manifests: %w", err)
		}

		year, _ := strconv.Atoi(yearFlag)
		result := validator.Validate(manifests, []validator.ValidatorFunc{labels.CostCentreValidator(year)})

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
	rootCmd.AddCommand(validateCostCentreCmd)
	validateCostCentreCmd.Flags().StringVarP(&yearFlag, "year", "y", strconv.Itoa(time.Now().Year()), "Expected cost centre year (4-digit). Defaults to the current calendar year.")
}
