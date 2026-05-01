# go-cli-k8s-manifest-label-validator

A Go CLI to validate labels of K8S manifests

## Overview

`k8s-label-validator` is a command-line tool that checks Kubernetes manifest files (YAML/JSON) to ensure resources contain required labels and that label values conform to specified patterns. It is designed for use in CI/CD pipelines to enforce labelling standards.

## Installation

```bash
go install github.com/ksaifullah/go-cli-k8s-manifest-label-validator@latest
```

Or build from source:

```bash
git clone https://github.com/ksaifullah/go-cli-k8s-manifest-label-validator.git
cd go-cli-k8s-manifest-label-validator
go build -o k8s-label-validator .
```

## Usage

```
k8s-label-validator validate [flags] <file|dir> [<file|dir>...]
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--label KEY` | `-l` | Label must be present (any value) |
| `--label KEY=VALUE` | `-l` | Label value must equal `VALUE` |
| `--label KEY=~PATTERN` | `-l` | Label value must match regex `PATTERN` |

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | All manifests are valid |
| `1` | One or more violations were found |
| `2` | Usage or runtime error |

### Examples

Check that all resources in a directory have `app` and `env` labels:

```bash
k8s-label-validator validate -l app -l env ./manifests/
```

Require `env` to be one of `prod`, `staging`, or `dev`:

```bash
k8s-label-validator validate -l "env=~^(prod|staging|dev)$" ./manifests/
```

Require an exact value:

```bash
k8s-label-validator validate -l "team=platform" ./deploy.yaml
```

Combine multiple rules on multiple files/directories:

```bash
k8s-label-validator validate \
  -l app \
  -l env \
  -l "env=~^(prod|staging|dev)$" \
  ./services/ ./jobs/
```

### Sample Output

**All valid:**

```
✓ All 3 resource(s) passed label validation
```

**Violations found:**

```
✗ manifests/deploy.yaml [Deployment/my-app]: missing required label "env"
✗ manifests/svc.yaml [Service/my-svc]: label "env" value "production" does not match required pattern "^(prod|staging|dev)$"

2 violation(s) found in 4 resource(s)
```

## Features

- Supports single-file and multi-document YAML (`---` separated) files
- Recursively scans directories for `.yaml`, `.yml`, and `.json` files
- Three rule types: presence check, exact match, regex pattern match
- Clear, human-readable output showing file, resource kind/name, and violation details
- Non-zero exit code on any violation for easy CI integration

## Development

```bash
# Run tests
go test ./...

# Build
go build -o k8s-label-validator .
```
