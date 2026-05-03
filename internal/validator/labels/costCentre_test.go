package labels

import (
	"strings"
	"testing"

	"github.com/ksaifullah/k8s-manifest-validator/internal/manifest"
)

const testYear = 2026

func makeManifest(name string, labels map[string]string) manifest.Manifest {
	return manifest.Manifest{
		APIVersion: "v1",
		Kind:       "Namespace",
		Metadata: manifest.Metadata{
			Name:   name,
			Labels: labels,
		},
	}
}

func TestCostCentreValidator_valid(t *testing.T) {
	v := CostCentreValidator(testYear)
	m := makeManifest("test", map[string]string{labelKey: "CC-071-2026"})
	result := v(m)
	if len(result.Errors) != 0 {
		t.Errorf("expected no errors, got: %v", result.Errors)
	}
}

func TestCostCentreValidator_missingLabel(t *testing.T) {
	v := CostCentreValidator(testYear)
	m := makeManifest("test", nil)
	result := v(m)
	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(result.Errors), result.Errors)
	}
	if !strings.Contains(result.Errors[0], "missing required label") {
		t.Errorf("unexpected error: %s", result.Errors[0])
	}
}

func TestCostCentreValidator_invalidFormat(t *testing.T) {
	cases := []struct {
		name  string
		value string
	}{
		{"no prefix", "071-2026"},
		{"lowercase", "cc-071-2026"},
		{"missing year", "CC-071"},
		{"extra segment", "CC-071-2026-extra"},
		{"short nnn", "CC-71-2026"},
		{"non-numeric nnn", "CC-ABC-2026"},
	}

	v := CostCentreValidator(testYear)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := makeManifest("test", map[string]string{labelKey: tc.value})
			result := v(m)
			if len(result.Errors) != 1 {
				t.Fatalf("expected 1 error for value %q, got %d: %v", tc.value, len(result.Errors), result.Errors)
			}
			if !strings.Contains(result.Errors[0], "does not match required format") {
				t.Errorf("unexpected error: %s", result.Errors[0])
			}
		})
	}
}

func TestCostCentreValidator_nnnBoundaries(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"exactly 050 is valid", "CC-050-2026", false},
		{"exactly 150 is valid", "CC-150-2026", false},
		{"049 is invalid", "CC-049-2026", true},
		{"151 is invalid", "CC-151-2026", true},
		{"000 is invalid", "CC-000-2026", true},
	}

	v := CostCentreValidator(testYear)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := makeManifest("test", map[string]string{labelKey: tc.value})
			result := v(m)
			gotErr := false
			for _, e := range result.Errors {
				if strings.Contains(e, "out of valid range") {
					gotErr = true
				}
			}
			if gotErr != tc.wantErr {
				t.Errorf("value %q: wantErr=%v, got errors: %v", tc.value, tc.wantErr, result.Errors)
			}
		})
	}
}

func TestCostCentreValidator_wrongYear(t *testing.T) {
	v := CostCentreValidator(testYear)
	m := makeManifest("test", map[string]string{labelKey: "CC-071-2025"})
	result := v(m)
	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(result.Errors), result.Errors)
	}
	if !strings.Contains(result.Errors[0], "does not match expected year") {
		t.Errorf("unexpected error: %s", result.Errors[0])
	}
}

func TestCostCentreValidator_multipleErrors(t *testing.T) {
	// NNN out of range AND wrong year — both errors should be reported.
	v := CostCentreValidator(testYear)
	m := makeManifest("test", map[string]string{labelKey: "CC-200-2024"})
	result := v(m)
	if len(result.Errors) != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", len(result.Errors), result.Errors)
	}
}
