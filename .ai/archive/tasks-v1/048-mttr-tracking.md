# Task 048: MTTR tracking — time-to-detect / time-to-resolve

## Phase
M2 — Incident Engine

## Type
feature

## Priority
P1

## Description
Record timing for every incident run: injected-at, first-status-checked-at
(proxy for detection), resolved-at, hints used, resolved-by (manual fix vs
`incident resolve`). Persist run history in `.labctl/` and expose
`labctl incident history` + REST endpoint. This is the raw data for
challenge scoring (051) and team leaderboards (063).

## Files to Modify
- `cmd/labctl/internal/incident/` (run record model + persistence)
- `cmd/labctl/` (history subcommand + REST route)
- `docs/cli-reference.md`

## Implementation Notes
- Append-only JSONL history file in `.labctl/history/incidents.jsonl` —
  simple, greppable, no DB. One record per run.
- `incident status --watch` (uses 041's watch mode) timestamps the resolve
  the moment the detection check first passes — more accurate than manual
  status polls.
- Using `incident resolve` (escape hatch) marks the run `resolved_by:
  "auto"` — it still resolves, but scores as a non-completion later.

## Acceptance Criteria
- [ ] Every inject→resolve loop appends one complete run record
- [ ] `incident history` shows name, MTTR, hints used, resolved-by
- [ ] Records survive labctl restarts; never committed to git
- [ ] Unit tests for record lifecycle

## Testing Instructions
Run two faults end-to-end (one manual fix, one auto-resolve), check
`incident history`. Runbook: `docs/runbooks/08-incident-engine.md`.

## Dependencies
046
