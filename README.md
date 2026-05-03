# k8s-manifest-validator

A CLI tool that validates Kubernetes manifests for required MegaTech cost centre labels.

## Overview

At MegaTech, every Kubernetes resource must carry a `metadata.megatech.inc/cost-centre` label so infrastructure costs can be attributed to the correct team. This tool reads one or more manifest files from stdin, validates each resource against MegaTech's labelling policy, and reports a pass/fail summary.

## Prerequisites

- Go 1.22 or later

## Installation

### Build from source

```sh
git clone https://github.com/ksaifullah/k8s-manifest-validator
cd k8s-manifest-validator
go build -o k8s-manifest-validator .
```

### Run without installing

```sh
go run main.go <command> [flags]
```

## Usage

### validate-cost-centre

Reads Kubernetes manifests from stdin and validates the `metadata.megatech.inc/cost-centre` label on every resource.

```sh
k8s-manifest-validator validate-cost-centre [flags]
```

#### Flags

| Flag     | Short | Default      | Description                                     |
| -------- | ----- | ------------ | ----------------------------------------------- |
| `--year` | `-y`  | current year | Expected 4-digit cost centre year (e.g. `2026`) |
| `--help` | `-h`  |              | Show help for the command                       |

#### Examples

Validate a file using stdin redirection:

```sh
k8s-manifest-validator validate-cost-centre < sample.yaml
```

Pipe the file instead:

```sh
cat sample.yaml | k8s-manifest-validator validate-cost-centre
```

Validate against a specific year (useful in CI for future-dated releases):

```sh
k8s-manifest-validator validate-cost-centre --year 2027 < sample.yaml
```

Validate several files combined:

```sh
cat deployment.yaml service.yaml namespace.yaml | k8s-manifest-validator validate-cost-centre
```

## Label format

The `metadata.megatech.inc/cost-centre` label must be present on every resource and its value must follow the pattern `CC-NNN-YYYY`:

| Segment | Description                                        |
| ------- | -------------------------------------------------- |
| `CC`    | Literal prefix, uppercase                          |
| `NNN`   | 3-digit number between `050` and `150` (inclusive) |
| `YYYY`  | 4-digit calendar year matching the current year    |

Valid examples (2026):

```text
CC-050-2026
CC-100-2026
CC-150-2026
```

Invalid examples:

| Value         | Reason                              |
| ------------- | ----------------------------------- |
| `CC-049-2026` | Number below minimum (050)          |
| `CC-151-2026` | Number above maximum (150)          |
| `CC-100-2025` | Wrong year                          |
| `cc-100-2026` | Wrong case (must be uppercase `CC`) |
| `CC-10-2026`  | Number is not 3 digits              |

## Example output

The following sample manifest contains resources with various label issues. Save it as a file or pipe it directly:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  labels:
    metadata.megatech.inc/cost-centre: CC-071-2025
  name: payments-service
---
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    metadata.megatech.inc/cost-centre: CC-071-2024
  name: payments-service
  namespace: payments-service
---
apiVersion: v1
kind: Namespace
metadata:
  name: notification-service
---
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    metadata.megatech.inc/cost-centre: CC-113-2025
  namespace: notification-service
```

Run the validator against it (2026 is the default year):

```sh
$ k8s-manifest-validator validate-cost-centre << 'EOF'
apiVersion: v1
kind: Namespace
metadata:
  labels:
    metadata.megatech.inc/cost-centre: CC-071-2025
  name: payments-service
---
...
EOF
```

Or save the YAML to a file and pipe it:

```sh
$ k8s-manifest-validator validate-cost-centre < sample.yaml
  INVALID  v1/Namespace/payments-service
           - cost centre year 2025 does not match expected year 2026
  INVALID  v1/ServiceAccount/payments-service/payments-service
           - cost centre year 2024 does not match expected year 2026
  INVALID  v1/Namespace/notification-service
           - missing required label "metadata.megatech.inc/cost-centre"
  INVALID  v1/ServiceAccount/notification-service/
           - cost centre year 2025 does not match expected year 2026

Summary:
  Total:   4
  Valid:   0
  Invalid: 4
```

## Exit codes

| Code | Meaning                                                       |
| ---- | ------------------------------------------------------------- |
| `0`  | All resources passed validation                               |
| `1`  | One or more resources failed validation, or an error occurred |

## Running tests

```sh
go test ./...
```
