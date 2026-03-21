# Task 006: Expand CI Validation for Infra and Shell Assets

## Priority
P1

## Assigned To
DevOps

## Description
Add CI coverage for the parts of the repository that currently have the least automated validation: shell scripts, Terraform modules, and scenario/platform configuration assets. Right now CI is mostly focused on Go code and app Helm charts.

## Files to Modify
- `.github/workflows/ci.yaml`
- `.github/workflows/helm-validation.yaml`
- `docs/ci-cd.md`

## Implementation Notes
Add lightweight, fast checks that match the repo’s main failure modes: shell syntax and style validation for `engine/`, `platform/`, `runtimes/`, `services/`, and `bootstrap/`; Terraform formatting and validation for `foundation/terraform`; and YAML/config validation for scenario manifests or values files where practical. Keep the workflow maintainable and avoid adding network-heavy integration jobs.

## Acceptance Criteria
- [ ] CI validates repository shell scripts beyond the current Go-only jobs.
- [ ] Terraform modules and environments are checked for formatting and basic validity.
- [ ] Scenario and platform configuration changes trigger meaningful validation beyond app chart linting alone.
- [ ] `docs/ci-cd.md` documents the new validation coverage.

## Testing Instructions
Use GitHub Actions workflow validation locally if available, then run the equivalent commands for shell and Terraform checks in the repo. Confirm the workflow definitions parse cleanly.

## Dependencies
None
