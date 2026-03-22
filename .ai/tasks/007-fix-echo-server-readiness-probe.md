# Task 007: Fix Echo-Server Readiness Probe

## Priority
P0

## Assigned To
Bug

## Description
The echo-server `/ready` endpoint returns unhealthy when Redis is unavailable, even though Redis is declared as an optional dependency. This misrepresents the app's actual readiness to serve traffic and causes unnecessary pod restarts under Kubernetes liveness/readiness probes. The readiness check must reflect whether the application itself is functional, not whether an optional cache backend is reachable.

## Files to Modify
- `apps/echo-server/main.go`

## Implementation Notes
The `/ready` handler currently gates on `redisClient.Ping()` success. Change it so that:
1. The app always reports ready as long as its core HTTP listener is up.
2. Redis connectivity is reported separately in `/health` (or a dedicated `/health/redis` sub-check) as a degraded-but-functional status — not a hard failure.
3. Keep the existing `/health` and `/ready` route paths and response shapes; only change the readiness logic.
4. If Redis was never configured (`REDIS_ADDR` is empty), readiness must still return 200.

Do not add new dependencies. Keep the change minimal and focused on the readiness contract.

## Acceptance Criteria
- [ ] `GET /ready` returns HTTP 200 when the HTTP server is running, regardless of Redis state.
- [ ] `GET /health` or a sub-path clearly indicates Redis connectivity status as a non-fatal detail.
- [ ] Pod does not enter CrashLoopBackOff when `REDIS_ADDR` is unset or Redis is temporarily down.
- [ ] Existing integration behavior (caching) continues to work when Redis is available.

## Testing Instructions
1. Run `go build ./...` from `apps/echo-server/`.
2. Start the server with `REDIS_ADDR=` (empty) and `curl /ready` — expect 200.
3. Start with a valid `REDIS_ADDR` pointing to a stopped Redis and `curl /ready` — expect 200.
4. Confirm `curl /health` reports Redis status distinctly.

## Dependencies
None
