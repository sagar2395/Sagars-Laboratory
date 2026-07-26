# Task 022: Add Structured Error Responses to API Handlers

## Priority
P2

## Assigned To
Feature

## Description
When `labctl` API handlers encounter errors (invalid input, unavailable resource, internal failure), they return bare strings or empty bodies rather than structured JSON. This makes error handling in the UI and CLI client fragile — the client cannot distinguish "resource not found" from "cluster unreachable" without string parsing. All error responses must follow a consistent JSON envelope.

## Files to Modify
- `cmd/labctl/internal/api/handlers.go`
- `cmd/labctl/internal/api/server.go`

## Implementation Notes
1. Define a shared error response type in `handlers.go` (or a new `errors.go` file in the same package):
   ```go
   type ErrorResponse struct {
       Error   string `json:"error"`
       Code    string `json:"code,omitempty"`
   }
   ```

2. Create a helper function:
   ```go
   func writeError(w http.ResponseWriter, status int, code, message string) {
       w.Header().Set("Content-Type", "application/json")
       w.WriteHeader(status)
       json.NewEncoder(w).Encode(ErrorResponse{Error: message, Code: code})
   }
   ```

3. Replace every `http.Error(...)` call and every `w.WriteHeader(4xx/5xx)` + bare write with `writeError(...)`.

4. Use consistent error codes:
   - `"not_found"` — resource doesn't exist
   - `"invalid_input"` — path parameter or request body validation failed (see task 008)
   - `"internal_error"` — unexpected server failure
   - `"unavailable"` — cluster or dependency not reachable

5. Do not change any HTTP status codes already in use. Only change the response body format.

## Acceptance Criteria
- [ ] All 4xx and 5xx responses from API handlers return `Content-Type: application/json` with an `ErrorResponse` body.
- [ ] The `code` field is present and non-empty for all error types.
- [ ] Existing 2xx response shapes are unchanged.
- [ ] `go test ./internal/api/...` passes.

## Testing Instructions
Trigger a 404 by calling `GET /api/scenarios/nonexistent`. Confirm response body is `{"error":"...","code":"not_found"}`. Trigger a 500 by temporarily making the cluster unreachable and calling `GET /api/status`.

## Dependencies
Task 008 (input validation) adds validation errors that benefit from this structure; independently implementable.
