# Task 002: Add Scenario Preflight Validation

## Priority
P0

## Assigned To
Feature

## Description
Implement the prerequisite and compatibility checks promised by the scenario documentation before any scenario activation work begins. The scenario engine currently installs components immediately without validating runtime compatibility, required platform capabilities, required apps, or whether referenced files exist.

## Files to Modify
- `cmd/labctl/internal/scenario/engine.go`
- `cmd/labctl/internal/scenario/engine_test.go`
- `cmd/labctl/cmd/scenario.go`

## Implementation Notes
Add a preflight validation phase for `scenario up` that checks declared `runtimes`, `prerequisites.platform`, `prerequisites.apps`, and referenced component assets such as `valuesFile`, `path`, and `script`. Fail fast before any component install starts, and return actionable errors that identify the missing prerequisite or incompatible runtime. Keep activation idempotency and current state tracking behavior intact.

## Acceptance Criteria
- [ ] `labctl scenario up <name>` refuses to start when the active runtime is not listed in the scenario manifest.
- [ ] Missing platform prerequisites or required apps produce clear validation errors before any install occurs.
- [ ] Missing component files are detected during validation rather than midway through activation.
- [ ] Unit tests cover successful validation and representative failure cases.

## Testing Instructions
Run `go test ./internal/scenario/...` from `cmd/labctl`. Manually verify one happy-path scenario and one failure case such as an unmet app prerequisite.

## Dependencies
None
