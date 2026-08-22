# Task 052: Score & progress persistence + REST API

## Phase
M3 — Learning & Assessment

## Type
feature

## Priority
P1

## Description
Unify the run records produced by incidents (048), learning paths (050) and
challenges (051) into one results store in `.labctl/history/`, and expose
read endpoints: `GET /api/results`, `GET /api/results/{kind}`,
`GET /api/progress`. This is the data layer the UI (053) and team
leaderboard (063) read from.

## Files to Modify
- `cmd/labctl/internal/results/` (new package: shared record schema, JSONL store)
- `cmd/labctl/internal/{incident,learn,challenge}/` (write through the shared store)
- `cmd/labctl/` (API routes)
- `docs/cli-reference.md` (API section)

## Implementation Notes
- One record schema with a `kind` field (incident|module|challenge) and
  kind-specific payload — avoid three divergent formats now, it is the
  leaderboard's input later.
- JSONL append-only files, no DB. Include an optional `user` field
  (defaults to `$USER`) so team mode (062/063) needs no migration.
- Structured error responses per the existing API conventions (task 022).

## Acceptance Criteria
- [ ] Incident, learn, and challenge runs all land in the unified store
- [ ] API endpoints return filtered, time-sorted results
- [ ] Existing 048 history command reads from the shared store (no dual writes)
- [ ] Unit tests for the store; docs updated

## Testing Instructions
`cd cmd/labctl && go test ./...`; curl the endpoints after one run of each
kind. Runbook: `docs/runbooks/09-learning-and-challenges.md`.

## Dependencies
048, 050, 051
