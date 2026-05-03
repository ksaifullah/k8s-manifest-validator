package validator

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/ksaifullah/k8s-manifest-validator/internal/manifest"
)

// LabelKey is the required MegaTech cost centre label.
const LabelKey = "metadata.megatech.inc/cost-centre"

// labelPattern matches the required format CC-NNN-YYYY.
var labelPattern = regexp.MustCompile(`^CC-(\d{3})-(\d{4})$`)

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
func Validate(manifests []manifest.Manifest, year int) Result {
	if year == 0 {
		year = time.Now().Year()
	}

	var result Result

	for _, m := range manifests {
		errs := validateResource(m, year)
		result.Resources = append(result.Resources, ResourceResult{
			Identity: m.Identity(),
			Errors:   errs,
		})
	}

	return result
}

// validateResource checks a single manifest for the required cost centre label and its validity.
func validateResource(m manifest.Manifest, year int) []string {
	labelValue, ok := m.Metadata.Labels[LabelKey]
	if !ok {
		return []string{fmt.Sprintf("missing required label %q", LabelKey)}
	}

	matches := labelPattern.FindStringSubmatch(labelValue)
	if matches == nil {
		return []string{fmt.Sprintf("label value %q does not match required format CC-NNN-YYYY", labelValue)}
	}

	var errs []string

	nnn, _ := strconv.Atoi(matches[1])
	if nnn < 50 || nnn > 150 {
		errs = append(errs, fmt.Sprintf("cost centre number %03d is out of valid range (050-150)", nnn))
	}

	yyyy, _ := strconv.Atoi(matches[2])
	if yyyy != year {
		errs = append(errs, fmt.Sprintf("cost centre year %d does not match current year %d", yyyy, year))
	}

	return errs
}
