# Task 070: Entitlement & extension interface (no-op default)

## Phase
M7 — OSS & Ecosystem Foundation

## Type
foundation

## Priority
P0

## Description
Define the seams that let premium content and hosted/SaaS plug in WITHOUT forking
the engine. Ship them in the OSS core with default no-op/open implementations, so
the open engine behaves identically while leaving a clean extension point.

This is the key anti-lock-in task: building these interfaces early is cheap;
retrofitting them after premium logic has leaked into the engine is a rewrite.

## Files to Modify
- `pkg/entitlement/` — `Entitlement` interface (default: everything allowed) +
  license-token verifier seam
- `pkg/extension/` — content-source resolver interface (git/OCI/registry/…),
  and optional lifecycle hooks (pre/post stage, pre/post check)
- `cmd/labctl/internal/...` — wire the no-op defaults into pack install + scenario run
- `docs/architecture.md`, `docs/authoring/extensions.md`

## Implementation Notes
- OSS default `Entitlement.Allow(pack) == true` always; premium impls live in the
  private repo and are injected at construction — never compiled into OSS.
- Resolver interface abstracts WHERE a pack comes from so premium/private/hosted
  sources slot in without engine changes.
- No premium/business logic in the engine — only interfaces + open defaults.
- Keep hooks data/# script-driven; do NOT introduce compiled Go plugins.

## Acceptance Criteria
- [x] `pkg/entitlement` + `pkg/extension` interfaces with open/no-op defaults
- [x] Pack install + scenario run route through them with zero behavior change in OSS
- [x] A documented example of injecting a custom resolver/entitlement (test double)
- [x] CODEOWNERS locks these packages to the lead maintainer

## Testing Instructions
Unit-test the no-op defaults (allow-all); inject a test-double entitlement that
denies a pack and assert the install path refuses it cleanly. See
`pkg/{entitlement,extension}/*_test.go` and
`internal/scenario/entitlement_test.go`.

## Dependencies
066

## Progress
- Done: `pkg/entitlement` — `Entitlement` interface + `AllowAll` open default
  (`Default()`), `DeniedError`/`Denied`, and a `TokenVerifier` license-token seam
  (interface only — no concrete verifier in OSS). `pkg/extension` — `Resolver`
  interface + `Chain` + open built-ins (`OCIResolver` via pkg/pack,
  `GitResolver` via the git CLI, `LocalResolver` for dirs/file://) +
  `DefaultResolver`; `Hooks` interface + `NoopHooks` default
  (`DefaultHooks()`).
- Wiring (zero behavior change): `scenario.Engine` gains `Entitlement` + `Hooks`
  fields defaulted to the open impls in `NewEngine`; `installFetched` authorizes
  through entitlement; `Up` fires Pre/PostStage; `Verify` fires Pre/PostCheck;
  `InstallVia(resolver,…)` installs through the resolver seam (the OCI add path
  now flows through `extension.OCIResolver`).
- Example + tests: a `denyTier` entitlement double (docs + test) and a
  `denyAll` integration test proving the install path refuses cleanly with a
  `*DeniedError`; resolver `Chain`/built-in tests; `NoopHooks` test; a
  `Verify`-routes-through-hooks test. CODEOWNERS locks
  `/cmd/labctl/pkg/{entitlement,extension}/` to the lead maintainer.
- Docs: `docs/authoring/extensions.md`, architecture principle 8 (open core /
  injected extensions), authoring README index.
- Note: premium/business logic and concrete token verifiers stay in a private
  repo and are injected at construction — never compiled into the OSS binary.
  Hooks are data/script-driven; no compiled Go plugins.
