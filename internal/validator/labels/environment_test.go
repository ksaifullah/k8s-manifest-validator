package labels

import (
	"strings"
	"testing"

	"github.com/ksaifullah/k8s-manifest-validator/internal/manifest"
)

func TestEnvironmentValidator_labelAbsent(t *testing.T) {
	// The environment label is optional; absence must not produce an error.
	v := EnvironmentValidator()
	m := makeManifest("test", nil)
	result := v(m)
	if len(result.Errors) != 0 {
		t.Errorf("expected no errors when label is absent, got: %v", result.Errors)
	}
}

func TestEnvironmentValidator_validValues(t *testing.T) {
	v := EnvironmentValidator()
	for _, value := range []string{"production", "staging", "development"} {
		t.Run(value, func(t *testing.T) {
			m := makeManifest("test", map[string]string{environmentLabelKey: value})
			result := v(m)
			if len(result.Errors) != 0 {
				t.Errorf("value %q: expected no errors, got: %v", value, result.Errors)
			}
		})
	}
}

func TestEnvironmentValidator_invalidValue(t *testing.T) {
	cases := []string{"prod", "dev", "PRODUCTION", "Production", "test", ""}
	v := EnvironmentValidator()
	for _, value := range cases {
		t.Run(value, func(t *testing.T) {
			m := manifest.Manifest{
				APIVersion: "v1",
				Kind:       "Namespace",
				Metadata: manifest.Metadata{
					Name:   "test",
					Labels: map[string]string{environmentLabelKey: value},
				},
			}
			result := v(m)
			if len(result.Errors) != 1 {
				t.Fatalf("value %q: expected 1 error, got %d: %v", value, len(result.Errors), result.Errors)
			}
			if !strings.Contains(result.Errors[0], "invalid value") {
				t.Errorf("value %q: unexpected error: %s", value, result.Errors[0])
			}
		})
	}
}
