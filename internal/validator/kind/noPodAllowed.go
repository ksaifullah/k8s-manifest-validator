// Package labels contains ValidatorFunc implementations for Kubernetes resource label checks.
package kind

import (
	"fmt"

	"github.com/ksaifullah/k8s-manifest-validator/internal/manifest"
	"github.com/ksaifullah/k8s-manifest-validator/internal/validator"
)

// NoPodValidator returns a validator function that checks if a manifest has no pods.
func NoPodValidator() validator.ValidatorFunc {
	return func(m manifest.Manifest) validator.ValidationResult {
		errs := validate(m)
		return validator.ValidationResult{
			Errors: errs,
		}
	}
}

// validate checks a single manifest for the required cost centre label and its validity.
func validate(m manifest.Manifest) []string {
	// Check if the manifest is of kind Pod
	if m.Kind == "Pod" {
		return []string{fmt.Sprintf("resource of kind %q is not allowed", m.Kind)}
	}

	return nil
}
