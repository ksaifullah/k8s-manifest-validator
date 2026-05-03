package validator

import (
	"github.com/ksaifullah/k8s-manifest-validator/internal/manifest"
)

// ValidationResult holds the outcome of validating a single manifest's cost centre label.
type ValidationResult struct {
	Errors []string
}

type ValidatorFunc func(manifest.Manifest) ValidationResult

// ResourceResult holds the validation outcome for a single Kubernetes resource.
type ResourceResult struct {
	Identity string
	Errors   []string
}

// Valid reports whether the resource passed all validation checks.
func (r ResourceResult) Valid() bool {
	return len(r.Errors) == 0
}

// Result holds the overall validation outcome for all processed resources.
type Result struct {
	Resources []ResourceResult
}

// Total returns the total number of resources processed.
func (r Result) Total() int { return len(r.Resources) }

// ValidCount returns the number of resources that passed validation.
func (r Result) ValidCount() int {
	count := 0
	for _, res := range r.Resources {
		if res.Valid() {
			count++
		}
	}
	return count
}

// InvalidCount returns the number of resources that failed validation.
func (r Result) InvalidCount() int {
	return r.Total() - r.ValidCount()
}

// Validate checks each manifest in the provided slice for the required MegaTech
// cost centre label. year is the expected cost centre year; pass 0 to use the
// current calendar year.
func Validate(manifests []manifest.Manifest, validators []ValidatorFunc) Result {
	var result Result

	for _, m := range manifests {
		var errs []string
		for _, validator := range validators {
			validationResult := validator(m)
			errs = append(errs, validationResult.Errors...)
		}
		result.Resources = append(result.Resources, ResourceResult{
			Identity: m.Identity(),
			Errors:   errs,
		})
	}

	return result
}
