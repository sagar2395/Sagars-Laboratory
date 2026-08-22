# Task 021: Add Redis Cache TTL Configuration to echo-server

## Priority
P2

## Assigned To
Feature

## Description
The echo-server `/cache` (POST) handler stores keys in Redis with a TTL of `0`, which means keys never expire. In a long-lived lab environment, this causes unbounded Redis memory growth. The TTL should default to a sensible value (e.g., 1 hour) and be configurable via an environment variable.

## Files to Modify
- `apps/echo-server/main.go`

## Implementation Notes
1. Read a `CACHE_TTL` environment variable at startup. Parse it as a Go `time.Duration` string (e.g., `"1h"`, `"30m"`, `"0"` for no expiry). Default to `"1h"` if not set or if the value is empty.

2. Pass the parsed duration to `redisClient.Set(ctx, key, value, ttl)` in the POST branch of the `/cache` handler. Replace the current hardcoded `0`.

3. Document the variable in the server startup log line (if the server logs environment configuration at startup).

4. If parsing `CACHE_TTL` fails, log a warning and fall back to the 1-hour default — do not exit.

Do not add new dependencies. Use `time.ParseDuration` from the standard library.

## Acceptance Criteria
- [ ] By default, POST `/cache` stores keys with a 1-hour TTL.
- [ ] Setting `CACHE_TTL=30m` causes keys to expire after 30 minutes.
- [ ] Setting `CACHE_TTL=0` disables TTL (keys never expire) — preserving old behavior when explicitly requested.
- [ ] An invalid `CACHE_TTL` value logs a warning and uses the 1-hour default instead of crashing.
- [ ] `go build ./...` from `apps/echo-server/` succeeds.

## Testing Instructions
Run the server with `CACHE_TTL=5s`. POST a key to `/cache`, then `GET /cache?key=<key>` after 6 seconds — it should return 404 / empty. Confirm the default 1-hour TTL is used when `CACHE_TTL` is unset.

## Dependencies
Task 007 (echo-server readiness fix) — independently implementable, but should be done in the same app release.
