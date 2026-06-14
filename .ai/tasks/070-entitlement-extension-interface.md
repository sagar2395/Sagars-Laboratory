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
- [ ] `pkg/entitlement` + `pkg/extension` interfaces with open/no-op defaults
- [ ] Pack install + scenario run route through them with zero behavior change in OSS
- [ ] A documented example of injecting a custom resolver/entitlement (test double)
- [ ] CODEOWNERS locks these packages to the lead maintainer

## Testing Instructions
Unit-test the no-op defaults (allow-all); inject a test-double entitlement that
denies a pack and assert the install path refuses it cleanly.

## Dependencies
066
