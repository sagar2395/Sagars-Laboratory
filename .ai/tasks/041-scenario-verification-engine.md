# Task 041: Verification engine — `labctl scenario verify`

## Phase
M1 — Scenario Engine v2

## Type
feature

## Priority
P0

## Description
Implement the check runners for the v2 `checks` block and a `labctl scenario
verify <name>` command (plus REST endpoint) that executes them and reports
pass/fail per check with a summary. This is the grading primitive reused by
challenges (051) and incident resolution detection (045/046).

## Files to Modify
- `cmd/labctl/internal/scenario/` (verify orchestration)
- `cmd/labctl/internal/checks/` (new package: one runner per check type)
- `cmd/labctl/` (cobra command + API route)
- `docs/cli-reference.md`

## Implementation Notes
- Runner contract: `Run(ctx, check, templateVars) CheckResult{Name, Pass,
  Got, Want, Err}`. Respect a per-check timeout (default 30s) and an
  overall `--timeout`.
- `http`: plain GET via net/http. `kubectl`: shell out to kubectl with
  `-o jsonpath` (CLI wraps scripts — don't import client-go logic for this).
  `promql`: query Prometheus HTTP API at the monitoring service (namespace
  from env, golden rule 3/4). `script`: run with the standard script env.
- `--watch` flag: re-run checks every N seconds until all pass or timeout —
  needed later for incident MTTR (048).
- Exit code 0 only when all checks pass (CI-friendly).

## Acceptance Criteria
- [ ] `labctl scenario verify observability-sre` passes on an active scenario
- [ ] Deleting a scenario component makes verify fail with a useful message
- [ ] All four check types have unit tests (fake HTTP server / stub exec)
- [ ] REST endpoint returns structured results; docs updated

## Testing Instructions
`cd cmd/labctl && go test ./...`; manual flow in
`docs/runbooks/07-scenario-engine-v2.md`.

## Dependencies
040
