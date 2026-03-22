# Task 013: Add /version Endpoint to go-api and echo-server

## Priority
P2

## Assigned To
Feature

## Description
Neither sample application exposes a `/version` endpoint. This is a standard operational primitive: it allows operators, CI pipelines, and the Kubernetes Dashboard to confirm which build is running without exec-ing into pods or reading pod metadata labels. Both apps should expose build-time version information through a consistent endpoint.

## Files to Modify
- `apps/go-api/main.go`
- `apps/echo-server/main.go`

## Implementation Notes
1. Inject version information at build time using `go build -ldflags` variables. Define package-level variables for `version`, `commit`, and `buildDate`, defaulting to `"dev"`, `"unknown"`, and `"unknown"` respectively.
2. Add a `GET /version` route that returns JSON:
   ```json
   {
     "version": "v1.2.3",
     "commit":  "abc1234",
     "buildDate": "2026-03-21T00:00:00Z",
     "app": "go-api"
   }
   ```
3. Update each app's `Dockerfile` to pass `--build-arg` values for these variables so real version info is embedded during image builds. Use `git describe --tags --always` and `git rev-parse --short HEAD` in the build scripts or engine.
4. The endpoint requires no authentication and must return HTTP 200.

Keep the implementation idiomatic Go with no new dependencies. Follow the route registration pattern already used in each `main.go`.

## Acceptance Criteria
- [ ] `GET /version` on `go-api` returns HTTP 200 with valid JSON containing `version`, `commit`, `buildDate`, and `app` fields.
- [ ] `GET /version` on `echo-server` returns HTTP 200 with the same structure.
- [ ] When built with default ldflags (local dev), fields show `"dev"` / `"unknown"`.
- [ ] When built with injected ldflags, fields show the injected values.
- [ ] No existing routes or behavior are changed.

## Testing Instructions
Run `go build ./...` from each app directory and `curl http://localhost:<PORT>/version`. Build with `-ldflags "-X main.version=v0.0.1 -X main.commit=abc123 -X main.buildDate=2026-01-01"` and confirm the values are reflected.

## Dependencies
None
