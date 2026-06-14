# SDK & Schema Stability Policy

This policy defines what Flightdeck promises to keep stable so pack authors and
downstream tools can build on it with confidence. It is the contract that lets
**content and engine evolve on independent clocks**.

## What is covered (the public surface)

- **The scenario schema** — `scenario.yaml`, identified by `apiVersion`
  (`scenario.flightdeck.dev/v2`). Published as a JSON Schema in
  [`sdk/schemas/scenario.schema.json`](../../sdk/schemas/scenario.schema.json).
- **The check types** — `http`, `kubectl`, `promql`, `script` and their fields.
- **The pack format** — `pack.yaml` (`packs.flightdeck.dev/v1`), once task 067
  lands, with its own JSON Schema.
- **The public Go SDK** — packages under `pkg/` (once task 066's physical move
  lands). Today the stable types still live in `cmd/labctl/internal/...`; the
  move to `pkg/` is RFC-gated (see `docs/rfcs/0001-public-sdk-boundary.md`).

Anything under `internal/` is **not** public and may change at any time.

## Versioning rules

- **SemVer** for the engine/CLI and (separately) for each scenario pack.
- The **scenario schema** is versioned by `apiVersion`. The engine supports the
  **current and previous** schema versions (N and N-1). A breaking schema change
  ⇒ a new `apiVersion` (e.g. `…/v3`), not a silent change to `v2`.
- A missing `apiVersion` is treated as the engine's current default for
  backward compatibility (existing v1/v2 scenarios keep working).

## Compatibility promise

- Within an `apiVersion`, we only **add** optional fields; we never remove or
  repurpose a field, nor tighten validation in a way that breaks valid content.
- Deprecations are announced in the changelog with a **minimum one-minor-version
  window** before removal, and removal only happens at a new `apiVersion`.
- The CLI publishes a **compatibility matrix** (which CLI supports which scenario
  and pack `apiVersion`s) each release (see [RELEASING.md](../../RELEASING.md)).

## How breaking changes happen

Any change to a covered surface requires an **RFC** under
[`docs/rfcs/`](../rfcs/) approved by the lead maintainer
([GOVERNANCE.md](../../GOVERNANCE.md)). The RFC states the new `apiVersion`, the
migration path, and the deprecation window.

## For pack authors

- Pin the `apiVersion` in your `scenario.yaml` and the engine constraints in your
  `pack.yaml` so your pack fails fast on an incompatible engine rather than
  misbehaving.
- Validate locally against the published JSON Schema (your editor can do this).
