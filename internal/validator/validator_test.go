package validator

import (
	"testing"

	"github.com/ksaifullah/k8s-manifest-validator/internal/manifest"
)

// alwaysValid is a stub ValidatorFunc that never reports errors.
func alwaysValid(_ manifest.Manifest) ValidationResult {
	return ValidationResult{}
}

func Test_Validate_emptySlice(t *testing.T) {
	result := Validate(nil, []ValidatorFunc{alwaysValid})
	if result.Total() != 0 {
		t.Errorf("expected 0 resources, got %d", result.Total())
	}
}

func Test_Validate_allValid(t *testing.T) {
	manifests := []manifest.Manifest{
		{APIVersion: "v1", Kind: "Namespace", Metadata: manifest.Metadata{Name: "ns-a"}},
		{APIVersion: "v1", Kind: "Namespace", Metadata: manifest.Metadata{Name: "ns-b"}},
	}
	result := Validate(manifests, []ValidatorFunc{alwaysValid})
	if result.Total() != 2 {
		t.Errorf("expected 2 resources, got %d", result.Total())
	}
	if result.ValidCount() != 2 {
		t.Errorf("expected 2 valid, got %d", result.ValidCount())
	}
	if result.InvalidCount() != 0 {
		t.Errorf("expected 0 invalid, got %d", result.InvalidCount())
	}
}

func Test_Validate_mixedResults(t *testing.T) {
	manifests := []manifest.Manifest{
		{APIVersion: "v1", Kind: "Namespace", Metadata: manifest.Metadata{Name: "valid-ns"}},
		{APIVersion: "v1", Kind: "Namespace", Metadata: manifest.Metadata{Name: "invalid-ns"}},
		{APIVersion: "v1", Kind: "Namespace", Metadata: manifest.Metadata{Name: "also-invalid"}},
	}
	validators := []ValidatorFunc{
		func(m manifest.Manifest) ValidationResult {
			if m.Metadata.Name == "valid-ns" {
				return ValidationResult{}
			}
			return ValidationResult{Errors: []string{"invalid resource"}}
		},
	}

	result := Validate(manifests, validators)
	if result.Total() != 3 {
		t.Errorf("expected 3 resources, got %d", result.Total())
	}
	if result.ValidCount() != 1 {
		t.Errorf("expected 1 valid, got %d", result.ValidCount())
	}
	if result.InvalidCount() != 2 {
		t.Errorf("expected 2 invalid, got %d", result.InvalidCount())
	}
}

func Test_Validate_multipleValidatorsAccumulateErrors(t *testing.T) {
	// Each validator contributes its own error; both should appear in the result.
	m := manifest.Manifest{
		APIVersion: "v1",
		Kind:       "Namespace",
		Metadata:   manifest.Metadata{Name: "test"},
	}
	errA := func(_ manifest.Manifest) ValidationResult {
		return ValidationResult{Errors: []string{"error from A"}}
	}
	errB := func(_ manifest.Manifest) ValidationResult {
		return ValidationResult{Errors: []string{"error from B"}}
	}

	result := Validate([]manifest.Manifest{m}, []ValidatorFunc{errA, errB})
	if result.Total() != 1 {
		t.Errorf("expected 1 resource, got %d", result.Total())
	}
	if len(result.Resources[0].Errors) != 2 {
		t.Errorf("expected 2 accumulated errors, got %d: %v", len(result.Resources[0].Errors), result.Resources[0].Errors)
	}
}

func Test_Validate_noValidators(t *testing.T) {
	// With no validators every resource should be considered valid.
	m := manifest.Manifest{
		APIVersion: "v1",
		Kind:       "Namespace",
		Metadata:   manifest.Metadata{Name: "test"},
	}
	result := Validate([]manifest.Manifest{m}, nil)
	if result.Total() != 1 {
		t.Errorf("expected 1 resource, got %d", result.Total())
	}
	if !result.Resources[0].Valid() {
		t.Errorf("expected resource to be valid with no validators")
	}
}

func Test_ResourceResult_Valid(t *testing.T) {
	valid := ResourceResult{Identity: "test", Errors: nil}
	if !valid.Valid() {
		t.Error("expected Valid() == true when there are no errors")
	}
	invalid := ResourceResult{Identity: "test", Errors: []string{"some error"}}
	if invalid.Valid() {
		t.Error("expected Valid() == false when errors are present")
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
