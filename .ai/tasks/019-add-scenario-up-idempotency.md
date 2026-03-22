# Task 019: Add Scenario Up Idempotency Check

## Priority
P1

## Assigned To
Feature

## Description
Calling `labctl scenario up <name>` (or `POST /api/scenarios/<name>/up`) on an already-active scenario re-runs every component installation from scratch. For Helm components, this triggers `helm upgrade --install` again (often harmless but slow). For `grafana-dashboard` components, it creates duplicate ConfigMaps. For `manifest` components, it re-applies resources which may reset manually-edited state. The `Up()` method must detect that a scenario is already active and skip re-installation, or at minimum require an explicit `--force` flag to override.

## Files to Modify
- `cmd/labctl/internal/scenario/engine.go`
- `cmd/labctl/cmd/scenario.go`

## Implementation Notes
In `engine.go`, at the start of the `Up()` method, call `isActive(name)`. If the scenario is already marked active:
- Return a sentinel error (e.g., `ErrAlreadyActive`) rather than re-installing any components.
- Do NOT return a generic error string — use a typed error so callers can detect and handle it.

In `cmd/scenario.go`, when receiving `ErrAlreadyActive` from `scenes.Up()`:
- Print: `Scenario <name> is already active. Use --force to reinstall.`
- Exit with code 0 (this is not a failure — the desired state is already achieved).
- Do not add the `--force` flag implementation in this task; just add the message and graceful exit.

For the API handler in `handlers.go`, no change is needed in this task — the goroutine will simply return early via the sentinel error and optionally log it.

## Acceptance Criteria
- [ ] Running `labctl scenario up <name>` when the scenario is already active prints the "already active" message and exits 0.
- [ ] No component installers are invoked when the scenario is already active.
- [ ] `ErrAlreadyActive` is a typed error (not a string comparison).
- [ ] Running `labctl scenario up <name>` on an inactive scenario still works normally.
- [ ] `go test ./internal/scenario/...` passes.

## Testing Instructions
Activate a scenario: `labctl scenario up observability-sre`. Run it again and confirm the output says "already active" without any helm/kubectl output. Confirm `echo $?` is 0.

## Dependencies
None
