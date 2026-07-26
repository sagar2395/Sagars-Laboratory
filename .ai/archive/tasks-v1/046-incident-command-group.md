# Task 046: `labctl incident` command group

## Phase
M2 — Incident Engine

## Type
feature

## Priority
P0

## Description
CLI + API surface for the fault library: `labctl incident list`, `incident
inject <name>` (or `--random`, `--category <cat>`), `incident status`
(runs the active fault's detection check), `incident resolve` (escape
hatch — runs `resolve.sh`). Active incident state lives in `.labctl/`.

## Files to Modify
- `cmd/labctl/internal/incident/` (new package: discovery, state, orchestration)
- `cmd/labctl/` (cobra command group + REST routes)
- `docs/cli-reference.md`

## Implementation Notes
- The Go side only discovers fault dirs, parses `fault.yaml`, and shells
  out to inject/resolve scripts with the standard env — golden rule 2.
- `--random` picks from faults whose prerequisites (apps deployed) are met;
  supports `--seed` for reproducible team exercises and `--silent` to
  suppress printing which fault was injected (game-day mode).
- Only one active incident at a time in v1; `inject` while active errors
  with the active incident's name unless `--force`.
- `incident status` runs the detection check via the 041 runners; when it
  passes, mark resolved and clear active state.

## Acceptance Criteria
- [ ] Full loop works: list → inject → status (failing) → fix manually → status (resolved)
- [ ] `--random --silent` hides the fault name; `incident resolve` always recovers
- [ ] REST endpoints expose the same operations with structured responses
- [ ] Unit tests for discovery/state; docs updated

## Testing Instructions
`cd cmd/labctl && go test ./...`; manual loop in
`docs/runbooks/08-incident-engine.md`.

## Dependencies
041, 045
