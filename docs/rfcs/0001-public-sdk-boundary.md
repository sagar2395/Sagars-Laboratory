# RFC 0001: Public SDK boundary (`pkg/`)

- **Status:** accepted
- **Author(s):** @sagar2395
- **Created:** 2026-06-14
- **Tracking task:** .ai/tasks/066-public-sdk-boundary.md

## Summary

Carve a semver-stable public Go SDK under `pkg/` that pack authors and
third-party tools depend on, separated from internal engine wiring under
`internal/`. Introduce an explicit, versioned scenario schema (`apiVersion`) and
publish JSON Schemas. This is the decoupling that lets content and engine evolve
on independent clocks.

## Motivation

Today the stable schema types (Scenario, Stage, Component, Check, Pack, Results)
live in `cmd/labctl/internal/...`. `internal/` cannot be imported by external
code, so the moment a third party wants to generate or validate Flightdeck
content programmatically, they have no stable surface — and we have no clear
boundary protecting us from accidental API lock-in. An open ecosystem
(scenario packs, editor tooling, CI validators) needs a public, versioned SDK.

## Design

Target layout (core monorepo):

```
pkg/
  scenario/   # Scenario/Stage/Component/Prerequisites types + loader + apiVersion
  checks/     # Check + Result types + runner contract
  pack/       # pack.yaml schema + validate + resolve (task 067)
  results/    # results/score schema
  entitlement/  extension/   # task 070 interfaces (no-op defaults)
sdk/schemas/  # JSON Schema for scenario.yaml and pack.yaml (editor validation)
```

Phasing (each independently shippable; existing content keeps working):

1. **(this RFC, phase 1 — landed)** Add an explicit `apiVersion`
   (`scenario.flightdeck.dev/v2`) to the scenario schema, accepted-and-validated
   with a back-compat default; publish `sdk/schemas/scenario.schema.json`; write
   the [stability policy](../authoring/sdk-stability-policy.md).
2. **(phase 2)** Move the schema types from `cmd/labctl/internal/scenario` and
   `internal/checks` into `pkg/scenario` and `pkg/checks`; keep thin aliases or
   update imports across `internal/`. No behavior change.
3. **(phase 3)** Add `pkg/pack` (task 067) and `pkg/results`; wire
   `pkg/entitlement` + `pkg/extension` (task 070).

The engine supports the current and previous schema versions (N and N-1);
unknown `apiVersion`s are rejected with an actionable error.

## Compatibility & migration

- Backward compatible: scenarios without `apiVersion` default to the current
  version; v1 (flat `components`) remains valid.
- The physical type move (phase 2) is internal refactoring with no user-visible
  change; imports inside the repo update in one PR.
- From the first 1.0 release, `pkg/` follows SemVer per the stability policy.

## Alternatives considered

- **Keep types in `internal/` and expose a codegen'd schema only.** Rejected:
  third-party Go tooling still couldn't import the types; weaker ecosystem.
- **Expose everything as one big `pkg/api`.** Rejected: poor cohesion; smaller
  focused packages version and document better.
- **Compiled Go plugins for extensibility.** Rejected (project-wide stance):
  portability and security tar pit; we stay declarative-data + sandboxed-script.

## Risks

- *Churn during the move* — mitigated by doing phase 2 as a pure, test-covered
  refactor with no behavior change.
- *Premature API lock-in* — mitigated by the stability policy + small packages +
  pre-1.0 latitude.

## Unresolved questions

- Whether `pkg/` becomes a separate Go module (own `go.mod` + `go.work`) now or
  at the 1.0 cut. Leaning: separate module at 1.0.
