// Package validator provides label validation logic for Kubernetes resources.
package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ksaifullah/go-cli-k8s-manifest-label-validator/internal/manifest"
)

// Rule defines a validation rule for a label.
type Rule struct {
	// Key is the label key that must be present.
	Key string
	// Pattern is an optional regular expression the label value must match.
	// An empty pattern means the label only needs to be present (any value is accepted).
	Pattern string
}

// Violation represents a single label validation violation on a resource.
type Violation struct {
	Resource manifest.Resource
	Message  string
}

func (v Violation) String() string {
	ref := resourceRef(v.Resource)
	return fmt.Sprintf("%s [%s]: %s", v.Resource.SourceFile, ref, v.Message)
}

// Result holds the outcome of validating a set of resources.
type Result struct {
	Violations []Violation
}

// Valid returns true when there are no violations.
func (r Result) Valid() bool {
	return len(r.Violations) == 0
}

// Validator validates Kubernetes resources against a set of label rules.
type Validator struct {
	rules   []Rule
	compiled map[string]*regexp.Regexp
}

// New creates a Validator from the provided rules.
// Returns an error if any rule pattern fails to compile.
func New(rules []Rule) (*Validator, error) {
	compiled := make(map[string]*regexp.Regexp, len(rules))
	for _, r := range rules {
		if r.Pattern == "" {
			continue
		}
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern for label %q: %w", r.Key, err)
		}
		compiled[r.Key] = re
	}
	return &Validator{rules: rules, compiled: compiled}, nil
}

// Validate checks resources against all rules and returns the aggregated result.
func (v *Validator) Validate(resources []manifest.Resource) Result {
	var result Result
	for _, res := range resources {
		for _, rule := range v.rules {
			val, present := res.Labels[rule.Key]
			if !present {
				result.Violations = append(result.Violations, Violation{
					Resource: res,
					Message:  fmt.Sprintf("missing required label %q", rule.Key),
				})
				continue
			}
			if re, ok := v.compiled[rule.Key]; ok {
				if !re.MatchString(val) {
					result.Violations = append(result.Violations, Violation{
						Resource: res,
						Message: fmt.Sprintf(
							"label %q value %q does not match required pattern %q",
							rule.Key, val, rule.Pattern,
						),
					})
				}
			}
		}
	}
	return result
}

// resourceRef builds a short human-readable reference for a resource.
func resourceRef(res manifest.Resource) string {
	parts := []string{}
	if res.Kind != "" {
		parts = append(parts, res.Kind)
	}
	if res.Name != "" {
		parts = append(parts, res.Name)
	}
	if len(parts) == 0 {
		return "unknown"
	}
	return strings.Join(parts, "/")
}
