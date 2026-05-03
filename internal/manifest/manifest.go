// Package manifest provides types and parsing logic for Kubernetes manifest files.
package manifest

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// Manifest is a parsed Kubernetes resource following the standard K8s structure.
type Manifest struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
}

// Metadata holds the Kubernetes resource metadata fields relevant to validation.
type Metadata struct {
	Name      string            `yaml:"name"`
	Namespace string            `yaml:"namespace"`
	Labels    map[string]string `yaml:"labels"`
}

// Identity returns a human-readable identifier for the manifest.
func (m Manifest) Identity() string {
	if m.Metadata.Namespace != "" {
		return fmt.Sprintf("%s/%s/%s/%s", m.APIVersion, m.Kind, m.Metadata.Namespace, m.Metadata.Name)
	}
	return fmt.Sprintf("%s/%s/%s", m.APIVersion, m.Kind, m.Metadata.Name)
}

// Parse reads YAML documents from r, identifies Kubernetes manifests (those with
// both apiVersion and kind set), and returns them. Non-Kubernetes YAML documents
// and empty documents are silently skipped.
func Parse(r io.Reader) ([]Manifest, error) {
	decoder := yaml.NewDecoder(r)
	var manifests []Manifest

	for {
		var m Manifest
		err := decoder.Decode(&m)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("parsing YAML: %w", err)
		}

		// skip empty documents and non-Kubernetes YAML (e.g. Helm values files)
		if m.APIVersion == "" || m.Kind == "" {
			continue
		}

		manifests = append(manifests, m)
	}

	return manifests, nil
}
