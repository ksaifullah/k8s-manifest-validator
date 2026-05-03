package manifest

import (
	"strings"
	"testing"
)

func Test_Parse_validManifests(t *testing.T) {
	input := `apiVersion: v1
kind: Namespace
metadata:
  name: payments
  labels:
    app: payments
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
  namespace: payments
  labels:
    app: api
`
	got, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 manifests, got %d", len(got))
	}
	if got[0].APIVersion != "v1" || got[0].Kind != "Namespace" || got[0].Metadata.Name != "payments" {
		t.Errorf("unexpected first manifest: %+v", got[0])
	}
	if got[1].APIVersion != "apps/v1" || got[1].Kind != "Deployment" || got[1].Metadata.Namespace != "payments" {
		t.Errorf("unexpected second manifest: %+v", got[1])
	}
}

func Test_Parse_emptyDocumentsSkipped(t *testing.T) {
	// leading/trailing --- produce empty documents that should be skipped
	input := `---
apiVersion: v1
kind: Namespace
metadata:
  name: test-ns
---
`
	got, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 manifest, got %d", len(got))
	}
}

func Test_Parse_nonKubernetesYAMLSkipped(t *testing.T) {
	// YAML without apiVersion or kind (e.g. Helm values) should be skipped
	input := `image:
  repository: myapp
  tag: latest
replicaCount: 2
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
`
	got, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 manifest (non-K8s doc skipped), got %d", len(got))
	}
	if got[0].Kind != "ConfigMap" {
		t.Errorf("expected ConfigMap, got %s", got[0].Kind)
	}
}

func Test_Parse_emptyInput(t *testing.T) {
	got, err := Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 manifests, got %d", len(got))
	}
}

func Test_Parse_invalidYAML(t *testing.T) {
	_, err := Parse(strings.NewReader("this: is: not: valid: yaml: :"))
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

func Test_Parse_labelsPreserved(t *testing.T) {
	input := `apiVersion: v1
kind: Namespace
metadata:
  name: billing
  labels:
    metadata.megatech.inc/cost-centre: CC-071-2026
    app: billing
`
	got, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 manifest, got %d", len(got))
	}
	if got[0].Metadata.Labels["metadata.megatech.inc/cost-centre"] != "CC-071-2026" {
		t.Errorf("label not preserved: %v", got[0].Metadata.Labels)
	}
}

func Test_Manifest_Identity_withNamespace(t *testing.T) {
	m := Manifest{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
		Metadata:   Metadata{Name: "api", Namespace: "payments"},
	}
	want := "apps/v1/Deployment/payments/api"
	if got := m.Identity(); got != want {
		t.Errorf("Identity() = %q, want %q", got, want)
	}
}

func Test_Manifest_Identity_withoutNamespace(t *testing.T) {
	m := Manifest{
		APIVersion: "v1",
		Kind:       "Namespace",
		Metadata:   Metadata{Name: "payments"},
	}
	want := "v1/Namespace/payments"
	if got := m.Identity(); got != want {
		t.Errorf("Identity() = %q, want %q", got, want)
	}
}
