# Task 066: Public SDK boundary (`pkg/`) + explicit schema apiVersion

## Phase
M7 — OSS & Ecosystem Foundation

## Type
foundation

## Priority
P0

## Description
Carve a clean, semver-stable public API that pack authors and third-party tools
depend on, separated from internal engine wiring. This decoupling is what lets
content and engine evolve on independent clocks (the core sustainability lever).

## Files to Modify
- `pkg/scenario/`, `pkg/checks/`, `pkg/pack/`, `pkg/results/` (move stable
  schema types out of `cmd/labctl/internal/...`)
- `sdk/schemas/` — JSON Schema for `scenario.yaml` (editor validation)
- `docs/authoring/` — SDK stability policy + compatibility matrix
- scenario.yaml: introduce explicit `apiVersion: scenario.lab.dev/v2`
- `go.work` if splitting into a separate module

## Implementation Notes
- Keep `internal/` for everything non-public (command glue, executor, api server).
- The engine must support the current schema apiVersion AND N-1 going forward —
  add a version negotiation/validation path now even if only v2 exists.
- Write the stability policy: what's covered, deprecation window, breaking-change
  process (RFC-gated).

## Acceptance Criteria
- [ ] Stable schema types live under `pkg/`; `internal/` no longer exports schema
- [ ] `scenario.yaml` carries an explicit apiVersion; loader validates it
- [ ] JSON Schemas published; documented stability policy
- [ ] All existing scenarios + tests still pass unchanged

## Testing Instructions
`go test ./...`; load every repo scenario and assert apiVersion validation;
validate a scenario.yaml against the published JSON Schema.

## Progress
- Phase 1 (landed): explicit `apiVersion: scenario.flightdeck.dev/v2` accepted +
  validated with a back-compat default; published JSON Schema
  (`sdk/schemas/scenario.schema.json`); stability policy
  (`docs/authoring/sdk-stability-policy.md`); RFC 0001 accepted
  (`docs/rfcs/0001-public-sdk-boundary.md`).
- Phase 2 (next): move schema types `internal/scenario`,`internal/checks` →
  `pkg/scenario`,`pkg/checks` (pure refactor, no behavior change).
- Phase 3: add `pkg/pack` (with 067), `pkg/results`, wire `pkg/entitlement` +
  `pkg/extension` (with 070).

## Dependencies
None (enables 067, 070)
