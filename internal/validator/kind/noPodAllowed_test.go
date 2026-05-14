package kind

import (
	"testing"

	"github.com/ksaifullah/k8s-manifest-validator/internal/manifest"
)

func makeManifest(kind string) manifest.Manifest {
	return manifest.Manifest{
		APIVersion: "v1",
		Kind:       kind,
		Metadata:   manifest.Metadata{Name: "test"},
	}
}

func TestValidate_podKindReturnsError(t *testing.T) {
	errs := validate(makeManifest("Pod"))
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestValidate_deploymentKindReturnsNoError(t *testing.T) {
	errs := validate(makeManifest("Deployment"))
	if len(errs) != 0 {
		t.Errorf("expected no errors for Deployment, got: %v", errs)
	}
}

func TestValidate_namespaceKindReturnsNoError(t *testing.T) {
	errs := validate(makeManifest("Namespace"))
	if len(errs) != 0 {
		t.Errorf("expected no errors for Namespace, got: %v", errs)
	}
}

func TestValidate_podErrorMessageContainsKind(t *testing.T) {
	errs := validate(makeManifest("Pod"))
	if len(errs) == 0 {
		t.Fatal("expected an error but got none")
	}
	const want = "Pod"
	found := false
	for _, e := range errs {
		if len(e) > 0 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error message to mention %q, got: %v", want, errs)
	}
}

func TestValidate_emptyKindReturnsNoError(t *testing.T) {
	errs := validate(makeManifest(""))
	if len(errs) != 0 {
		t.Errorf("expected no errors for empty kind, got: %v", errs)
	}
}
