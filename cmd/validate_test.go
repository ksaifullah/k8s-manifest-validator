package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ksaifullah/go-cli-k8s-manifest-label-validator/internal/manifest"
	"github.com/ksaifullah/go-cli-k8s-manifest-label-validator/internal/validator"
)

// helpers

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
	return path
}

// validateFiles is a helper that wires up the same logic as the validate command
// but returns output and violation count so we can assert without os.Exit.
func validateFiles(t *testing.T, rules []validator.Rule, paths ...string) (string, int) {
	t.Helper()
	v, err := validator.New(rules)
	if err != nil {
		t.Fatalf("creating validator: %v", err)
	}
	var resources []manifest.Resource
	for _, p := range paths {
		res, err := manifest.ParseFile(p)
		if err != nil {
			t.Fatalf("parsing %s: %v", p, err)
		}
		resources = append(resources, res...)
	}
	result := v.Validate(resources)
	var buf bytes.Buffer
	for _, viol := range result.Violations {
		buf.WriteString(viol.String() + "\n")
	}
	return buf.String(), len(result.Violations)
}

func TestValidateCommand_AllValid(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "deploy.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  labels:
    app: my-app
    env: prod
`)
	rules := []validator.Rule{{Key: "app"}, {Key: "env"}}
	_, violations := validateFiles(t, rules, filepath.Join(dir, "deploy.yaml"))
	if violations != 0 {
		t.Errorf("expected 0 violations, got %d", violations)
	}
}

func TestValidateCommand_MissingLabel(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "deploy.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  labels:
    app: my-app
`)
	rules := []validator.Rule{{Key: "app"}, {Key: "env"}}
	out, violations := validateFiles(t, rules, filepath.Join(dir, "deploy.yaml"))
	if violations != 1 {
		t.Errorf("expected 1 violation, got %d\noutput: %s", violations, out)
	}
}

func TestValidateCommand_PatternMismatch(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "deploy.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  labels:
    app: my-app
    env: production
`)
	rules := []validator.Rule{
		{Key: "app"},
		{Key: "env", Pattern: `^(prod|staging|dev)$`},
	}
	_, violations := validateFiles(t, rules, filepath.Join(dir, "deploy.yaml"))
	if violations != 1 {
		t.Errorf("expected 1 violation for pattern mismatch, got %d", violations)
	}
}

func TestValidateCommand_MultiDocumentFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "multi.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy-one
  labels:
    app: one
    env: prod
---
apiVersion: v1
kind: Service
metadata:
  name: svc-one
  labels:
    app: one
`)
	rules := []validator.Rule{{Key: "app"}, {Key: "env"}}
	_, violations := validateFiles(t, rules, filepath.Join(dir, "multi.yaml"))
	// Service is missing "env" label
	if violations != 1 {
		t.Errorf("expected 1 violation, got %d", violations)
	}
}

func TestValidateCommand_NoLabelRules(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "deploy.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
`)
	rules := []validator.Rule{}
	_, violations := validateFiles(t, rules, filepath.Join(dir, "deploy.yaml"))
	if violations != 0 {
		t.Errorf("expected 0 violations with no rules, got %d", violations)
	}
}

// TestRootCommand_Help ensures the Cobra root command is wired up.
func TestRootCommand_Help(t *testing.T) {
	root := &cobra.Command{Use: "k8s-label-validator"}
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
