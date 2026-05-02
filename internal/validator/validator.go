package validator

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// LabelKey is the required MegaTech cost centre label.
const LabelKey = "metadata.megatech.inc/cost-centre"

// labelPattern matches the required format CC-NNN-YYYY.
var labelPattern = regexp.MustCompile(`^CC-(\d{3})-(\d{4})$`)

// resource is a minimal representation of a Kubernetes resource used for label validation.
type resource struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   metadata `yaml:"metadata"`
}

// identity returns a human-readable identifier for the resource.
func (r resource) identity() string {
	if r.Metadata.Namespace != "" {
		return fmt.Sprintf("%s/%s/%s/%s", r.APIVersion, r.Kind, r.Metadata.Namespace, r.Metadata.Name)
	}
	return fmt.Sprintf("%s/%s/%s", r.APIVersion, r.Kind, r.Metadata.Name)
}

type metadata struct {
	Name      string            `yaml:"name"`
	Namespace string            `yaml:"namespace"`
	Labels    map[string]string `yaml:"labels"`
}

// ResourceResult holds the validation outcome for a single Kubernetes resource.
type ResourceResult struct {
	Identity string
	Errors   []string
}

// Valid reports whether the resource passed all validation checks.
func (r ResourceResult) Valid() bool {
	return len(r.Errors) == 0
}

// Result holds the overall validation outcome for all processed resources.
type Result struct {
	Resources []ResourceResult
}

// Total returns the total number of resources processed.
func (r Result) Total() int { return len(r.Resources) }

// ValidCount returns the number of resources that passed validation.
func (r Result) ValidCount() int {
	count := 0
	for _, res := range r.Resources {
		if res.Valid() {
			count++
		}
	}
	return count
}

// InvalidCount returns the number of resources that failed validation.
func (r Result) InvalidCount() int {
	return r.Total() - r.ValidCount()
}

// Validate reads YAML documents from r and validates each Kubernetes resource for the
// required MegaTech cost centre label. year is the expected cost centre year; pass 0
// to use the current calendar year.
func Validate(r io.Reader, year int) (Result, error) {
	if year == 0 {
		year = time.Now().Year()
	}

	decoder := yaml.NewDecoder(r)
	var result Result

	for {
		var res resource
		err := decoder.Decode(&res)
		if err == io.EOF {
			break
		}
		if err != nil {
			return result, fmt.Errorf("parsing YAML: %w", err)
		}

		// skip empty documents (e.g. trailing --- with no content)
		if res.APIVersion == "" && res.Kind == "" {
			continue
		}

		errs := validateResource(res, year)
		result.Resources = append(result.Resources, ResourceResult{
			Identity: res.identity(),
			Errors:   errs,
		})
	}

	return result, nil
}

// validateResource checks a single resource for the required cost centre label and its validity.
func validateResource(r resource, year int) []string {
	labelValue, ok := r.Metadata.Labels[LabelKey]
	if !ok {
		return []string{fmt.Sprintf("missing required label %q", LabelKey)}
	}

	matches := labelPattern.FindStringSubmatch(labelValue)
	if matches == nil {
		return []string{fmt.Sprintf("label value %q does not match required format CC-NNN-YYYY", labelValue)}
	}

	var errs []string

	nnn, _ := strconv.Atoi(matches[1])
	if nnn < 50 || nnn > 150 {
		errs = append(errs, fmt.Sprintf("cost centre number %03d is out of valid range (050-150)", nnn))
	}

	yyyy, _ := strconv.Atoi(matches[2])
	if yyyy != year {
		errs = append(errs, fmt.Sprintf("cost centre year %d does not match current year %d", yyyy, year))
	}

	return errs
}
