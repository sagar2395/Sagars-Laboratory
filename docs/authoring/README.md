# Authoring Guide

How to create content for Flightdeck — scenarios, packs, platform modules — and
the stability guarantees you can rely on.

> The end-to-end "write a scenario pack" walkthrough and the `labctl pack init` /
> `labctl scenario new` scaffolds land with task 072. This index will grow as the
> SDK (tasks 066–067) lands.

## Contents

- [Authoring Scenario Packs](packs.md) — `pack.yaml`, the `labctl pack` commands,
  and how content is bundled and distributed.
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
