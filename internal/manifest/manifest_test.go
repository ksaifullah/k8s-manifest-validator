package manifest_test

import (
	"testing"

	"github.com/ksaifullah/go-cli-k8s-manifest-label-validator/internal/manifest"
)

func TestParseBytes_SingleDocument(t *testing.T) {
	data := []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  labels:
    app: my-app
    env: prod
`)
	resources, err := manifest.ParseBytes(data, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}
	r := resources[0]
	if r.Kind != "Deployment" {
		t.Errorf("expected Kind=Deployment, got %q", r.Kind)
	}
	if r.Name != "my-app" {
		t.Errorf("expected Name=my-app, got %q", r.Name)
	}
	if r.Labels["app"] != "my-app" {
		t.Errorf("expected label app=my-app, got %q", r.Labels["app"])
	}
	if r.Labels["env"] != "prod" {
		t.Errorf("expected label env=prod, got %q", r.Labels["env"])
	}
}

func TestParseBytes_MultiDocument(t *testing.T) {
	data := []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy-one
---
apiVersion: v1
kind: Service
metadata:
  name: svc-one
  labels:
    app: svc
`)
	resources, err := manifest.ParseBytes(data, "multi.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(resources))
	}
	if resources[0].Kind != "Deployment" {
		t.Errorf("expected first Kind=Deployment, got %q", resources[0].Kind)
	}
	if resources[1].Kind != "Service" {
		t.Errorf("expected second Kind=Service, got %q", resources[1].Kind)
	}
}

func TestParseBytes_NoLabels(t *testing.T) {
	data := []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
`)
	resources, err := manifest.ParseBytes(data, "no-labels.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}
	if resources[0].Labels == nil {
		t.Error("expected non-nil Labels map")
	}
	if len(resources[0].Labels) != 0 {
		t.Errorf("expected empty labels map, got %v", resources[0].Labels)
	}
}

func TestParseBytes_EmptyDocument(t *testing.T) {
	data := []byte(`---
---`)
	resources, err := manifest.ParseBytes(data, "empty.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources) != 0 {
		t.Errorf("expected 0 resources for empty documents, got %d", len(resources))
	}
}

func TestParseFile_ValidDeployment(t *testing.T) {
	resources, err := manifest.ParseFile("../../testdata/valid/deployment.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}
	if resources[0].Kind != "Deployment" {
		t.Errorf("expected Kind=Deployment, got %q", resources[0].Kind)
	}
}

func TestParseFile_MultiDocument(t *testing.T) {
	resources, err := manifest.ParseFile("../../testdata/invalid/missing_labels.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(resources))
	}
}

func TestParseFile_NotFound(t *testing.T) {
	_, err := manifest.ParseFile("nonexistent.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestParseBytes_InvalidYAML(t *testing.T) {
	data := []byte(`{invalid yaml: [}`)
	_, err := manifest.ParseBytes(data, "bad.yaml")
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}
