# Task 008: Add Request Timeout and Input Validation to API Handlers

## Priority
P0

## Assigned To
Bug

## Description
The `labctl` API server exposes HTTP handlers that accept untrusted input (app names, component names, scenario names from URL path segments) without validation or timeout enforcement. Long-running subprocess commands can hang indefinitely, and unvalidated path parameters risk command injection via shell execution. Both issues must be fixed before the API is considered safe for use.

## Files to Modify
- `cmd/labctl/internal/api/handlers.go`
- `cmd/labctl/internal/api/server.go`

## Implementation Notes
**Timeouts**: Wrap context passed to executor calls with a finite deadline (suggested: 10 minutes for build/deploy, 30 seconds for status queries). Use `context.WithTimeout` and handle `context.DeadlineExceeded` with an HTTP 504 response.

**Input validation**: Before using any path parameter (app name, component name, scenario name, runtime name) in a shell command or file path, validate it against an allowlist pattern. A safe restrictive pattern is `^[a-zA-Z0-9_-]{1,64}$`. Reject with HTTP 400 and a clear error message on mismatch.

Do not redesign the handler structure or change the API surface (paths, methods, response shapes). Fix only the missing safety checks. Follow existing error-handling patterns in the file.

## Acceptance Criteria
- [ ] All handler path parameters are validated against a safe allowlist before use.
- [ ] Requests that result in subprocesses running longer than the configured timeout receive HTTP 504.
- [ ] Invalid parameter values return HTTP 400 with a descriptive JSON error body.
- [ ] No existing handler paths or response structures change.
- [ ] Unit tests cover at least one valid and one invalid input per parameter type.

## Testing Instructions
Run `go test ./internal/api/...` from `cmd/labctl`. Manually try `curl` with path parameters containing `../`, `;`, and shell metacharacters — all should return 400.

## Dependencies
None
