package labels

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/ksaifullah/k8s-manifest-validator/internal/manifest"
	"github.com/ksaifullah/k8s-manifest-validator/internal/validator"
)

const costCentreLabelKey = "metadata.megatech.inc/cost-centre"

var labelPattern = regexp.MustCompile(`^CC-(\d{3})-(\d{4})$`)

// CostCentreValidator returns a validator function for the cost centre label. year is
// the expected cost centre year as an integer.
func CostCentreValidator(year int) validator.ValidatorFunc {
	return func(m manifest.Manifest) validator.ValidationResult {
		errs := validate(m, year)
		return validator.ValidationResult{
			Errors: errs,
		}
	}
}

// validate checks a single manifest for the required cost centre label and its validity.
func validate(m manifest.Manifest, year int) []string {
	labelValue, ok := m.Metadata.Labels[costCentreLabelKey]
	if !ok {
		return []string{fmt.Sprintf("missing required label %q", costCentreLabelKey)}
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
		errs = append(errs, fmt.Sprintf("cost centre year %d does not match expected year %d", yyyy, year))
	}

	return errs
}
