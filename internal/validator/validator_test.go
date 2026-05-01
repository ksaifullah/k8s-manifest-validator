package validator_test

import (
	"testing"

	"github.com/ksaifullah/go-cli-k8s-manifest-label-validator/internal/manifest"
	"github.com/ksaifullah/go-cli-k8s-manifest-label-validator/internal/validator"
)

func resource(kind, name string, labels map[string]string) manifest.Resource {
	if labels == nil {
		labels = map[string]string{}
	}
	return manifest.Resource{
		Kind:       kind,
		Name:       name,
		Labels:     labels,
		SourceFile: "test.yaml",
	}
}

func TestValidate_AllPresent_NoPattern(t *testing.T) {
	v, err := validator.New([]validator.Rule{
		{Key: "app"},
		{Key: "env"},
	})
	if err != nil {
		t.Fatal(err)
	}
	res := []manifest.Resource{
		resource("Deployment", "my-app", map[string]string{"app": "my-app", "env": "prod"}),
	}
	result := v.Validate(res)
	if !result.Valid() {
		t.Errorf("expected valid, got violations: %v", result.Violations)
	}
}

func TestValidate_MissingLabel(t *testing.T) {
	v, err := validator.New([]validator.Rule{
		{Key: "app"},
		{Key: "env"},
	})
	if err != nil {
		t.Fatal(err)
	}
	res := []manifest.Resource{
		resource("Deployment", "my-app", map[string]string{"app": "my-app"}),
	}
	result := v.Validate(res)
	if result.Valid() {
		t.Error("expected violations for missing label")
	}
	if len(result.Violations) != 1 {
		t.Errorf("expected 1 violation, got %d", len(result.Violations))
	}
}

func TestValidate_PatternMatch(t *testing.T) {
	v, err := validator.New([]validator.Rule{
		{Key: "env", Pattern: `^(prod|staging|dev)$`},
	})
	if err != nil {
		t.Fatal(err)
	}

	valid := []manifest.Resource{
		resource("Deployment", "a", map[string]string{"env": "prod"}),
		resource("Service", "b", map[string]string{"env": "staging"}),
	}
	if result := v.Validate(valid); !result.Valid() {
		t.Errorf("expected valid, got: %v", result.Violations)
	}

	invalid := []manifest.Resource{
		resource("Deployment", "a", map[string]string{"env": "production"}),
	}
	if result := v.Validate(invalid); result.Valid() {
		t.Error("expected violation for pattern mismatch")
	}
}

func TestValidate_ExactMatch(t *testing.T) {
	v, err := validator.New([]validator.Rule{
		{Key: "team", Pattern: `^platform$`},
	})
	if err != nil {
		t.Fatal(err)
	}
	good := []manifest.Resource{resource("Deployment", "a", map[string]string{"team": "platform"})}
	if result := v.Validate(good); !result.Valid() {
		t.Errorf("unexpected violations: %v", result.Violations)
	}

	bad := []manifest.Resource{resource("Deployment", "a", map[string]string{"team": "infra"})}
	if result := v.Validate(bad); result.Valid() {
		t.Error("expected violation for wrong team label value")
	}
}

func TestValidate_InvalidPattern(t *testing.T) {
	_, err := validator.New([]validator.Rule{
		{Key: "env", Pattern: `[invalid`},
	})
	if err == nil {
		t.Error("expected error for invalid regex pattern")
	}
}

func TestValidate_MultipleResources(t *testing.T) {
	v, err := validator.New([]validator.Rule{
		{Key: "app"},
		{Key: "env"},
	})
	if err != nil {
		t.Fatal(err)
	}
	resources := []manifest.Resource{
		resource("Deployment", "a", map[string]string{"app": "a", "env": "prod"}),
		resource("Service", "b", map[string]string{"app": "b"}), // missing env
		resource("ConfigMap", "c", map[string]string{}),         // missing both
	}
	result := v.Validate(resources)
	if result.Valid() {
		t.Error("expected violations")
	}
	// 1 missing env (resource b) + 2 missing labels (resource c) = 3
	if len(result.Violations) != 3 {
		t.Errorf("expected 3 violations, got %d", len(result.Violations))
	}
}

func TestValidate_NoRules(t *testing.T) {
	v, err := validator.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	resources := []manifest.Resource{
		resource("Deployment", "a", map[string]string{}),
	}
	result := v.Validate(resources)
	if !result.Valid() {
		t.Error("expected no violations when no rules are defined")
	}
}

func TestViolation_String(t *testing.T) {
	v, err := validator.New([]validator.Rule{{Key: "app"}})
	if err != nil {
		t.Fatal(err)
	}
	result := v.Validate([]manifest.Resource{
		resource("Deployment", "my-app", map[string]string{}),
	})
	if result.Valid() {
		t.Fatal("expected violation")
	}
	s := result.Violations[0].String()
	if s == "" {
		t.Error("expected non-empty violation string")
	}
}
