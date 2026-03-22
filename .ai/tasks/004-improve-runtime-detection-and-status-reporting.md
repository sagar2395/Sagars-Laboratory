# Task 004: Improve Runtime Detection and Status Reporting

## Priority
P1

## Assigned To
Bug

## Description
Fix runtime status reporting so AKS and EKS are not treated as active based on fragile context-name guesses. The current runtime manager uses best-effort string matching, which can misreport cloud runtime state in both CLI and API outputs.

## Files to Modify
- `cmd/labctl/internal/runtime/manager.go`
- `cmd/labctl/cmd/runtime.go`
- `cmd/labctl/internal/api/handlers.go`

## Implementation Notes
Use a more reliable source of truth than the runtime directory name alone when determining whether a runtime is active or current. Prefer kubeconfig context inspection and runtime-specific expectations derived from config or script outputs. Keep the fix narrowly scoped to status accuracy and do not redesign runtime activation flows.

## Acceptance Criteria
- [ ] Runtime status distinguishes `current` from merely discoverable or previously configured runtimes.
- [ ] AKS and EKS no longer report `active` solely because a context name happens to partially match.
- [ ] CLI and API runtime listings use the same status logic.
- [ ] The change includes focused tests for k3d and at least one cloud-runtime case.

## Testing Instructions
Run `go test ./internal/runtime/... ./internal/api/...` from `cmd/labctl`. Validate runtime listings against a kubeconfig fixture or mocked command output.

## Dependencies
None
