package validator

import (
	"strings"
	"testing"

	"github.com/ksaifullah/k8s-manifest-validator/internal/manifest"
)

const testYear = 2026

func Test_validateResource_valid(t *testing.T) {
	m := manifest.Manifest{
		APIVersion: "v1",
		Kind:       "Namespace",
		Metadata: manifest.Metadata{
			Name:   "payments-service",
			Labels: map[string]string{LabelKey: "CC-071-2026"},
		},
	}
	errs := validateResource(m, testYear)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func Test_validateResource_missingLabel(t *testing.T) {
	m := manifest.Manifest{
		APIVersion: "v1",
		Kind:       "Namespace",
		Metadata:   manifest.Metadata{Name: "payments-service"},
	}
	errs := validateResource(m, testYear)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "missing required label") {
		t.Errorf("unexpected error message: %s", errs[0])
	}
}

func Test_validateResource_invalidFormat(t *testing.T) {
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

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := manifest.Manifest{
				APIVersion: "v1",
				Kind:       "Namespace",
				Metadata: manifest.Metadata{
					Name:   "test",
					Labels: map[string]string{LabelKey: tc.value},
				},
			}
			errs := validateResource(m, testYear)
			if len(errs) != 1 {
				t.Fatalf("expected 1 error for value %q, got %d: %v", tc.value, len(errs), errs)
			}
			if !strings.Contains(errs[0], "does not match required format") {
				t.Errorf("unexpected error message: %s", errs[0])
			}
		})
	}
}

func Test_validateResource_nnnBoundaries(t *testing.T) {
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

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := manifest.Manifest{
				APIVersion: "v1",
				Kind:       "Namespace",
				Metadata: manifest.Metadata{
					Name:   "test",
					Labels: map[string]string{LabelKey: tc.value},
				},
			}
			errs := validateResource(m, testYear)
			gotErr := false
			for _, e := range errs {
				if strings.Contains(e, "out of valid range") {
					gotErr = true
				}
			}
			if gotErr != tc.wantErr {
				t.Errorf("value %q: wantErr=%v, got errors: %v", tc.value, tc.wantErr, errs)
			}
		})
	}
}

func Test_validateResource_wrongYear(t *testing.T) {
	m := manifest.Manifest{
		APIVersion: "v1",
		Kind:       "Namespace",
		Metadata: manifest.Metadata{
			Name:   "test",
			Labels: map[string]string{LabelKey: "CC-071-2025"},
		},
	}
	errs := validateResource(m, testYear)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "does not match current year") {
		t.Errorf("unexpected error message: %s", errs[0])
	}
}

func Test_validateResource_multipleErrors(t *testing.T) {
	// NNN out of range AND wrong year — both errors should be reported
	m := manifest.Manifest{
		APIVersion: "v1",
		Kind:       "Namespace",
		Metadata: manifest.Metadata{
			Name:   "test",
			Labels: map[string]string{LabelKey: "CC-200-2024"},
		},
	}
	errs := validateResource(m, testYear)
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", len(errs), errs)
	}
}

func Test_Validate_multipleManifests(t *testing.T) {
	manifests := []manifest.Manifest{
		{
			APIVersion: "v1",
			Kind:       "Namespace",
			Metadata: manifest.Metadata{
				Name:   "valid-ns",
				Labels: map[string]string{LabelKey: "CC-071-2026"},
			},
		},
		{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
			Metadata: manifest.Metadata{
				Name:      "invalid-sa",
				Namespace: "some-ns",
				Labels:    map[string]string{LabelKey: "CC-071-2024"},
			},
		},
		{
			APIVersion: "v1",
			Kind:       "Namespace",
			Metadata:   manifest.Metadata{Name: "missing-label-ns"},
		},
	}

	result := Validate(manifests, testYear)

	if result.Total() != 3 {
		t.Errorf("expected 3 resources, got %d", result.Total())
	}
	if result.ValidCount() != 1 {
		t.Errorf("expected 1 valid, got %d", result.ValidCount())
	}
	if result.InvalidCount() != 2 {
		t.Errorf("expected 2 invalid, got %d", result.InvalidCount())
	}
	if !result.Resources[0].Valid() {
		t.Errorf("expected resource[0] to be valid")
	}
	if result.Resources[1].Valid() {
		t.Errorf("expected resource[1] to be invalid")
	}
	if result.Resources[2].Valid() {
		t.Errorf("expected resource[2] to be invalid")
	}
}

func Test_Validate_emptySlice(t *testing.T) {
	result := Validate(nil, testYear)
	if result.Total() != 0 {
		t.Errorf("expected 0 resources, got %d", result.Total())
	}
}

func Test_Result_counts(t *testing.T) {
	r := Result{
		Resources: []ResourceResult{
			{Identity: "a", Errors: nil},
			{Identity: "b", Errors: []string{"err"}},
			{Identity: "c", Errors: nil},
		},
	}
	if r.Total() != 3 {
		t.Errorf("Total: expected 3, got %d", r.Total())
	}
	if r.ValidCount() != 2 {
		t.Errorf("ValidCount: expected 2, got %d", r.ValidCount())
	}
	if r.InvalidCount() != 1 {
		t.Errorf("InvalidCount: expected 1, got %d", r.InvalidCount())
	}
}
