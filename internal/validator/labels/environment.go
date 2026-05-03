package labels

import (
	"github.com/ksaifullah/k8s-manifest-validator/internal/manifest"
	"github.com/ksaifullah/k8s-manifest-validator/internal/validator"
)

const environmentLabelKey = "environment"

func EnvironmentValidator() validator.ValidatorFunc {
	return func(m manifest.Manifest) validator.ValidationResult {
		// if there is no environment label, return empty result
		if _, ok := m.Metadata.Labels[environmentLabelKey]; !ok {
			return validator.ValidationResult{}
		}
		// if the label is present, validate that it is one of the allowed values
		value := m.Metadata.Labels[environmentLabelKey]
		switch value {
		case "production", "staging", "development":
			return validator.ValidationResult{}
		default:
			return validator.ValidationResult{Errors: []string{environmentLabelKey + " label has invalid value: " + value}}
		}
	}
}
