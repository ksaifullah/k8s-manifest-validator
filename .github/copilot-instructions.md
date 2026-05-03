# go cli Copilot Instructions

## Project Overview

A CLI tool to validate Kubernetes manifests with required labels.

## Background

At MegaTech, a fictional technology company, all Kubernetes resources must include a cost centre label so infrastructure costs can be attributed to the correct team.
The required label key is: `metadata.megatech.inc/cost-centre`
The format for valid cost centre values is: `CC-NNN-YYYY`
Where:

- NNN is a 3-digit number between 050 and 150 (inclusive)
- YYYY must match the current calendar year

For example, in 2025:

- `CC-071-2025` -> valid
- `CC-071-2024` -> invalid
- `CC-200-2025` -> invalid

## Requirements

The CLI must:

- Read Kubernetes manifests from stdin
- Support multiple YAML documents separated by ---
- Validate each resource for:
  - Presence of the required label
  - Correct format
  - Valid cost centre range
  - Correct year
- Output a summary to stdout including:
  - Total resources processed
  - Number of valid resources
  - Number of invalid resources

The output should be in a human-readable format, clearly indicating any validation errors for each resource.

## CLI Setup

Please scaffold the project using Cobra: `cobra-cli init`
The CLI should:

- Follow basic CLI best practices
- Exit with a non-zero status code if invalid resources are detected
- Be runnable locally via: `go run main.go`

## Assumptions

Where requirements are not explicit:

- Make reasonable assumptions
- Document them briefly in code comments or a short README.md

## Testing

We expect:

- Unit tests for the validation logic
- Reasonable coverage of edge cases

You do not need to build complex integration tests.

## Evolution

Update this instructions file as we make new assumptions and implement featurtes.
