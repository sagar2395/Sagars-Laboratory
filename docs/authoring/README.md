# Authoring Guide

How to create content for Flightdeck — scenarios, packs, platform modules — and
the stability guarantees you can rely on.

> New here? Start with [Your First Scenario Pack](first-pack.md) — `labctl
> scenario new` / `labctl pack init` scaffold valid, verify-ready content in
> seconds.

## Contents

- [Your First Scenario Pack](first-pack.md) — the fast path: scaffold → edit →
  verify → publish, for a first contribution.
- [Authoring Scenario Packs](packs.md) — `pack.yaml`, the `labctl pack` commands,
  and how content is bundled and distributed.
- [Publishing Packs (OCI)](publishing.md) — `labctl pack publish`, OCI artifact
  format, cosign signing, and fail-closed verification on install.
- [The Pack Registry & Discovery](registry.md) — `labctl pack search`, the static
  index format, name resolution, TTL caching, and the validate-index PR gate.
- [Extension Seams](extensions.md) — the entitlement, resolver, and hook
  interfaces that let premium/hosted builds plug in without forking the engine.
- [SDK & Schema Stability Policy](sdk-stability-policy.md) — what we keep stable
  and how the schema versions.
- JSON Schemas (for editor validation):
  - [`sdk/schemas/scenario.schema.json`](../../sdk/schemas/scenario.schema.json)
  - [`sdk/schemas/pack.schema.json`](../../sdk/schemas/pack.schema.json)
- Scenario format reference: [../scenarios.md](../scenarios.md)
- Platform module contract: [../../platform/README.md](../../platform/README.md)

## Quick orientation

- A **scenario** is declarative YAML (`scenario.yaml`) + assets (manifests,
  values, dashboards, scripts). It declares `objectives`, `stages` of
  `components`, and machine-verifiable `checks`. Set
  `apiVersion: scenario.flightdeck.dev/v2`.
- A **scenario pack** bundles one or more scenarios (and optionally platform
  modules / incidents / learning content) for independent install — formalised
  in task 067 (`pack.yaml`).
- Content is **cross-platform, idempotent, and declarative** — never hardcode
  logic in Go (see [CONTRIBUTING.md](../../CONTRIBUTING.md) golden rules).

## Validate before you PR

```bash
bin/labctl scenario info <name>      # parses + renders stages/checks
bin/labctl scenario verify <name>    # runs the checks against a live cluster
yamllint -d relaxed scenarios/<name>/
```
