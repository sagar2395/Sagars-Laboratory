# Task 040: Scenario format v2 — stages, objectives, checks

## Phase
M1 — Scenario Engine v2 (see docs/ROADMAP.md Part II)

## Type
feature

## Priority
P0

## Description
Extend `scenario.yaml` with three optional blocks that turn scenarios into
verifiable simulations: `stages` (ordered groups of components that can be
activated independently), `objectives` (human-readable goals), and `checks`
(machine-verifiable assertions). This is the foundational primitive for the
whole simulator era — verification (041), challenge grading (051), and
incident detection (045) all consume `checks`.

Backward compatibility is mandatory: a v1 scenario with only `components`
must keep working unchanged (treated as a single implicit stage, no checks).

## Files to Modify
- `cmd/labctl/internal/scenario/` (schema structs, parser, validation)
- `docs/scenarios.md` (document the v2 format with examples)
- `scenarios/observability-sre/scenario.yaml` (first scenario migrated as reference)

## Implementation Notes
- Check types: `http` (url, expectStatus, optional bodyContains), `kubectl`
  (resource, namespace, jsonpath, operator, value), `promql` (query,
  operator, value), `script` (path, exit 0 = pass). Define the schema only
  here; runners are task 041.
- Checks may reference template vars (`{{.DomainSuffix}}` etc.) like
  explore commands already do.
- `stages[].components` reuses the existing component schema verbatim —
  no Go install logic changes beyond grouping/ordering.
- Validation errors must be structured and name the offending field.

## Acceptance Criteria
- [ ] Parser accepts v1 scenarios unchanged (all four existing scenarios still load)
- [ ] Parser accepts v2 blocks and rejects malformed checks with clear errors
- [ ] observability-sre migrated to v2 with ≥1 stage, ≥2 objectives, ≥3 checks
- [ ] `docs/scenarios.md` documents the v2 format
- [ ] Unit tests for schema parsing/validation pass

## Testing Instructions
`cd cmd/labctl && go test ./...`; then `bin/labctl scenario info
observability-sre` shows stages/objectives. Runbook:
`docs/runbooks/07-scenario-engine-v2.md`.

## Dependencies
None
