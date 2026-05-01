/*
Copyright © 2026 NAME HERE <khalid@outlook.com.au>
*/
package cmd

import (
	"fmt"

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
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("validate called")
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
