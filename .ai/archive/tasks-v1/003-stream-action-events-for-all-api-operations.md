# Task 003: Stream Action Events for All API Operations

## Priority
P1

## Assigned To
Feature

## Description
Make API-triggered operations emit a consistent action lifecycle to the WebSocket stream. Scenario actions already broadcast completion results, but app, platform, service, and runtime actions mostly fire-and-forget, which leaves the UI without reliable completion or error feedback.

## Files to Modify
- `cmd/labctl/internal/api/handlers.go`
- `cmd/labctl/internal/executor/executor.go`
- `cmd/labctl/internal/executor/executor_test.go`
- `cmd/labctl/internal/api/server.go`

## Implementation Notes
Use the executor’s streamed execution path, or add a small helper around it, so every long-running API action emits start, output, and end events with exit status. Keep the API’s `202 Accepted` behavior, but make sure failures are visible to WebSocket subscribers and use consistent action labels across apps, platform components, services, and runtimes.

## Acceptance Criteria
- [ ] App build, deploy, and destroy actions produce `action_start` and `action_end` events with exit codes.
- [ ] Platform, service, and runtime API actions also emit completion events on success and failure.
- [ ] Errors are surfaced through the same WebSocket channel the UI already consumes.
- [ ] Tests cover at least one success path and one failure path for streamed execution.

## Testing Instructions
Run `go test ./internal/executor/... ./internal/api/...` from `cmd/labctl`. Manually verify that triggering an API action produces start and end events over `/api/ws`.

## Dependencies
None
