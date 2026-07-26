# Task 017: Return Job ID from Async API Actions

## Priority
P1

## Assigned To
Feature

## Description
All long-running API handlers (app build/deploy/destroy, platform up/down, component up/down, scenario up/down, service up/down, runtime activate/deactivate) return HTTP 202 Accepted with an empty body. The WebSocket stream emits `action_start` and `action_end` events with an `actionID`, but the HTTP response contains no reference to that ID. This means a client cannot correlate its 202 response with the WebSocket stream — it must guess which event corresponds to its request. Include the `actionID` in the 202 response body.

## Files to Modify
- `cmd/labctl/internal/api/handlers.go`
- `cmd/labctl/internal/executor/executor.go`

## Implementation Notes
1. In `executor.go`: `RunScriptStreamed` and `RunCommandStreamed` already generate an `actionID` (UUID or counter). Expose that ID as the **return value** of both methods: change their signatures from `error` to `(string, error)` — where the string is the `actionID`.

2. In `handlers.go`: Capture the returned `actionID` from streamed execution calls. For async goroutine-based handlers (where execution starts after the 202 is sent), generate the ID **before** launching the goroutine and include it in the response immediately.

3. The 202 response body should be:
   ```json
   {"jobId": "abc-123-def", "status": "accepted"}
   ```

4. All existing `action_start` / `action_end` WebSocket events must continue to carry this same `jobId` field (already called `actionID` internally — rename or alias consistently).

5. Do not change the HTTP method, path, or status code for any handler.

## Acceptance Criteria
- [ ] Every 202 response from an action handler includes a `jobId` field in a JSON body.
- [ ] The `jobId` in the HTTP response matches the `actionID` field in the corresponding WebSocket `action_start` event.
- [ ] Existing WebSocket behavior (event structure, timing) is unchanged.
- [ ] `go test ./internal/executor/... ./internal/api/...` passes.

## Testing Instructions
Trigger `POST /api/apps/go-api/build` and capture the `jobId` from the response. Open a WebSocket to `/api/ws` and confirm an `action_start` event with the same `jobId` arrives.

## Dependencies
Task 003 (stream action events) should be completed first, but this task is independently implementable if the executor's streaming paths are already in place.
