package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ksaifullah/go-cli-k8s-manifest-label-validator/internal/manifest"
	"github.com/ksaifullah/go-cli-k8s-manifest-label-validator/internal/validator"
)

// validateCmd represents the validate command.
var validateCmd = &cobra.Command{
	Use:   "validate [flags] <file|dir> [<file|dir>...]",
	Short: "Validate labels on Kubernetes manifest files",
	Long: `Validate that Kubernetes manifests contain the required labels.

Accepts one or more paths to YAML/JSON manifest files or directories.
Directories are scanned recursively for files with .yaml, .yml, or .json extensions.

Required labels and optional value patterns are specified via --label flags:

  --label app                   # label "app" must be present (any value)
  --label app=myapp             # label "app" must equal "myapp"
  --label env=~^(prod|staging)$ # label "env" must match the regex

Exit code:
  0  all manifests are valid
  1  one or more violations were found
  2  usage or runtime error`,
	Args: cobra.MinimumNArgs(1),
	RunE: runValidate,
}

var labelFlags []string

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringArrayVarP(
		&labelFlags, "label", "l", nil,
		`label rule in one of these formats:
  KEY            label must be present (any value)
  KEY=VALUE      label value must equal VALUE
  KEY=~PATTERN   label value must match regex PATTERN`,
	)
}

// runValidate implements the validate subcommand.
func runValidate(cmd *cobra.Command, args []string) error {
	rules, err := parseRules(labelFlags)
	if err != nil {
		return fmt.Errorf("invalid label rule: %w", err)
	}

	v, err := validator.New(rules)
	if err != nil {
		return err
	}

	var resources []manifest.Resource
	for _, arg := range args {
		res, err := loadPath(arg)
		if err != nil {
			return err
		}
		resources = append(resources, res...)
	}

	if len(resources) == 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), "warning: no Kubernetes resources found")
		return nil
	}

	result := v.Validate(resources)
	if result.Valid() {
		fmt.Fprintf(cmd.OutOrStdout(), "✓ All %d resource(s) passed label validation\n", len(resources))
		return nil
	}

	for _, viol := range result.Violations {
		fmt.Fprintln(cmd.OutOrStderr(), "✗ "+viol.String())
	}
	fmt.Fprintf(cmd.OutOrStderr(), "\n%d violation(s) found in %d resource(s)\n",
		len(result.Violations), len(resources))
	os.Exit(1)
	return nil
}

// loadPath loads resources from a file or recursively from a directory.
func loadPath(path string) ([]manifest.Resource, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("accessing path %s: %w", path, err)
	}

	if info.IsDir() {
		return loadDir(path)
	}
	return manifest.ParseFile(path)
}

// loadDir walks a directory and loads all YAML/JSON manifest files.
func loadDir(dir string) ([]manifest.Resource, error) {
	var resources []manifest.Resource
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".yaml" && ext != ".yml" && ext != ".json" {
			return nil
		}
		res, err := manifest.ParseFile(path)
		if err != nil {
			return err
		}
		resources = append(resources, res...)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scanning directory %s: %w", dir, err)
	}
	return resources, nil
}

// parseRules converts --label flag values into validator.Rule slice.
// Supported formats:
//
//	KEY          - label must be present (any value)
//	KEY=VALUE    - label value must equal VALUE exactly
//	KEY=~PATTERN - label value must match regex PATTERN
func parseRules(flags []string) ([]validator.Rule, error) {
	rules := make([]validator.Rule, 0, len(flags))
	for _, f := range flags {
		idx := strings.IndexByte(f, '=')
		if idx < 0 {
			// bare key — just require presence
			rules = append(rules, validator.Rule{Key: f})
			continue
		}
		key := f[:idx]
		value := f[idx+1:]
		if key == "" {
			return nil, fmt.Errorf("label key cannot be empty in %q", f)
		}
		if strings.HasPrefix(value, "~") {
			// regex pattern
			rules = append(rules, validator.Rule{Key: key, Pattern: value[1:]})
		} else {
			// exact match — anchor the pattern
			rules = append(rules, validator.Rule{Key: key, Pattern: "^" + regexp.QuoteMeta(value) + "$"})
		}
	}
	return rules, nil
}
