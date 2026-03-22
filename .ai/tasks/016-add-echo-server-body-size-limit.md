# Task 016: Add Echo-Server Request Body Size Limit

## Priority
P0

## Assigned To
Bug

## Description
The `echo-server` `/echo` and `/cache` (POST) handlers read the full request body into memory with no size limit. An attacker can trivially send a multi-gigabyte request body, exhausting pod memory and causing an OOM kill or denial of service. This must be mitigated by wrapping each request body with `http.MaxBytesReader` before reading.

## Files to Modify
- `apps/echo-server/main.go`

## Implementation Notes
Apply `http.MaxBytesReader(w, r.Body, maxBodyBytes)` at the top of each handler that reads `r.Body` (the `/echo` handler and the POST branch of `/cache`). Define the limit as a package-level constant:

```go
const maxBodyBytes = 1 << 20 // 1 MiB
```

When the body exceeds the limit, `io.ReadAll` (or `io.ReadFull`) will return an error. Return HTTP 413 Request Entity Too Large with a JSON body:

```json
{"error": "request body too large"}
```

Do not change the behavior for requests within the limit. Do not add new dependencies.

## Acceptance Criteria
- [ ] `POST /echo` with a body larger than 1 MiB returns HTTP 413.
- [ ] `POST /cache` with a body larger than 1 MiB returns HTTP 413.
- [ ] Normal requests (≤1 MiB) continue to work exactly as before.
- [ ] The size limit constant is defined at the top of the file and not scattered across handlers.

## Testing Instructions
Run `go build ./...` from `apps/echo-server/`. Send an oversized body: `dd if=/dev/urandom bs=2M count=1 | curl -s -X POST -d @- http://localhost:8080/echo` — expect HTTP 413. Send a small body and confirm it echoes normally.

## Dependencies
None
