package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// runValidateCostCentre redirects os.Stdin and os.Stdout, executes the
// validate-cost-centre command with any additional args, and returns the
// captured stdout and the command error.
//
// Tests using this helper must not be run in parallel because they mutate
// global state (os.Stdin, os.Stdout, rootCmd args, yearFlag).
func runValidateCostCentre(t *testing.T, stdin string, args ...string) (string, error) {
	t.Helper()

	// Redirect os.Stdin.
	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe (stdin): %v", err)
	}
	oldStdin := os.Stdin
	os.Stdin = stdinR
	t.Cleanup(func() { os.Stdin = oldStdin })
	go func() {
		defer stdinW.Close()
		io.WriteString(stdinW, stdin) //nolint:errcheck
	}()

	// Redirect os.Stdout so fmt.Printf output can be captured.
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe (stdout): %v", err)
	}
	oldStdout := os.Stdout
	os.Stdout = stdoutW
	t.Cleanup(func() { os.Stdout = oldStdout })

	// Reset the year flag to its default before parsing to avoid state leakage
	// between sequential test cases.
	_ = validateCostCentreCmd.Flags().Set("year", strconv.Itoa(time.Now().Year()))

	rootCmd.SetArgs(append([]string{"validate-cost-centre"}, args...))
	_, cmdErr := rootCmd.ExecuteC()

	// Close the write end so io.Copy reaches EOF.
	stdoutW.Close()
	var buf bytes.Buffer
	io.Copy(&buf, stdoutR) //nolint:errcheck

	return buf.String(), cmdErr
}

func TestValidateCostCentre_allValid(t *testing.T) {
	year := time.Now().Year()
	input := fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  name: payments-ns
  labels:
    metadata.megatech.inc/cost-centre: CC-071-%d
`, year)

	out, err := runValidateCostCentre(t, input)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !strings.Contains(out, "VALID") {
		t.Errorf("expected VALID in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Total:   1") {
		t.Errorf("expected 'Total:   1' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Valid:   1") {
		t.Errorf("expected 'Valid:   1' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Invalid: 0") {
		t.Errorf("expected 'Invalid: 0' in output, got:\n%s", out)
	}
}

func TestValidateCostCentre_wrongYear(t *testing.T) {
	pastYear := time.Now().Year() - 1
	input := fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  name: payments-ns
  labels:
    metadata.megatech.inc/cost-centre: CC-071-%d
`, pastYear)

	out, err := runValidateCostCentre(t, input)

	if !errors.Is(err, errValidationFailed) {
		t.Fatalf("expected errValidationFailed, got: %v", err)
	}
	if !strings.Contains(out, "INVALID") {
		t.Errorf("expected INVALID in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Invalid: 1") {
		t.Errorf("expected 'Invalid: 1' in output, got:\n%s", out)
	}
}

func TestValidateCostCentre_missingLabel(t *testing.T) {
	input := `apiVersion: v1
kind: Namespace
metadata:
  name: payments-ns
`

	out, err := runValidateCostCentre(t, input)

	if !errors.Is(err, errValidationFailed) {
		t.Fatalf("expected errValidationFailed, got: %v", err)
	}
	if !strings.Contains(out, "INVALID") {
		t.Errorf("expected INVALID in output, got:\n%s", out)
	}
	if !strings.Contains(out, "missing required label") {
		t.Errorf("expected 'missing required label' in output, got:\n%s", out)
	}
}

func TestValidateCostCentre_multipleManifests_mixedResults(t *testing.T) {
	year := time.Now().Year()
	input := fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  name: valid-ns
  labels:
    metadata.megatech.inc/cost-centre: CC-100-%d
---
apiVersion: v1
kind: Namespace
metadata:
  name: missing-label-ns
---
apiVersion: v1
kind: Namespace
metadata:
  name: wrong-year-ns
  labels:
    metadata.megatech.inc/cost-centre: CC-071-%d
`, year, year-1)

	out, err := runValidateCostCentre(t, input)

	if !errors.Is(err, errValidationFailed) {
		t.Fatalf("expected errValidationFailed, got: %v", err)
	}
	if !strings.Contains(out, "Total:   3") {
		t.Errorf("expected 'Total:   3' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Valid:   1") {
		t.Errorf("expected 'Valid:   1' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Invalid: 2") {
		t.Errorf("expected 'Invalid: 2' in output, got:\n%s", out)
	}
}

func TestValidateCostCentre_emptyInput(t *testing.T) {
	out, err := runValidateCostCentre(t, "")

	if err != nil {
		t.Fatalf("expected no error for empty input, got: %v", err)
	}
	if !strings.Contains(out, "Total:   0") {
		t.Errorf("expected 'Total:   0' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Valid:   0") {
		t.Errorf("expected 'Valid:   0' in output, got:\n%s", out)
	}
}

func TestValidateCostCentre_nonKubernetesYAMLIsIgnored(t *testing.T) {
	// YAML without apiVersion/kind (e.g. Helm values) should be silently skipped.
	input := `image:
  repository: myapp
  tag: latest
`

	out, err := runValidateCostCentre(t, input)

	if err != nil {
		t.Fatalf("expected no error for non-K8s YAML, got: %v", err)
	}
	if !strings.Contains(out, "Total:   0") {
		t.Errorf("expected 'Total:   0' in output, got:\n%s", out)
	}
}

func TestValidateCostCentre_yearFlagOverride(t *testing.T) {
	// Manifest is valid for 2025; --year 2025 should accept it.
	input := `apiVersion: v1
kind: Namespace
metadata:
  name: historical-ns
  labels:
    metadata.megatech.inc/cost-centre: CC-071-2025
`

	out, err := runValidateCostCentre(t, input, "--year", "2025")

	if err != nil {
		t.Fatalf("expected no error with --year 2025, got: %v", err)
	}
	if !strings.Contains(out, "VALID") {
		t.Errorf("expected VALID in output, got:\n%s", out)
	}
}

func TestValidateCostCentre_invalidYearFlag(t *testing.T) {
	year := time.Now().Year()
	input := fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  name: ns
  labels:
    metadata.megatech.inc/cost-centre: CC-071-%d
`, year)

	_, err := runValidateCostCentre(t, input, "--year", "not-a-year")

	if err == nil {
		t.Fatal("expected error for invalid --year flag, got nil")
	}
}
