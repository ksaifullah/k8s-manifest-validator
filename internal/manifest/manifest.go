// Package manifest provides types and functions for parsing Kubernetes manifests.
package manifest

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// Resource represents a Kubernetes resource with its kind, API version, name, and labels.
type Resource struct {
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	Name       string            // populated from metadata.name
	Labels     map[string]string // populated from metadata.labels
	SourceFile string            // file the resource was parsed from
}

// rawResource is used for intermediate unmarshalling.
type rawResource struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name   string            `yaml:"name"`
		Labels map[string]string `yaml:"labels"`
	} `yaml:"metadata"`
}

// ParseFile parses all Kubernetes resources from a single YAML/JSON file.
// Multi-document YAML files (separated by "---") are supported.
func ParseFile(path string) ([]Resource, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening file %s: %w", path, err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", path, err)
	}

	return ParseBytes(data, path)
}

// ParseBytes parses all Kubernetes resources from a byte slice.
// source is used to populate Resource.SourceFile.
func ParseBytes(data []byte, source string) ([]Resource, error) {
	var resources []Resource

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	for {
		var raw rawResource
		err := decoder.Decode(&raw)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", source, err)
		}
		// Skip empty documents
		if raw.Kind == "" && raw.APIVersion == "" {
			continue
		}
		res := Resource{
			APIVersion: raw.APIVersion,
			Kind:       raw.Kind,
			Name:       raw.Metadata.Name,
			Labels:     raw.Metadata.Labels,
			SourceFile: source,
		}
		if res.Labels == nil {
			res.Labels = map[string]string{}
		}
		resources = append(resources, res)
	}
	return resources, nil
}
