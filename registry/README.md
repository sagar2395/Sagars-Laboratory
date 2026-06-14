# Flightdeck Pack Registry (seed)

This directory is the **source of truth** for the community pack registry index —
the static, PR-moderated catalog that `labctl pack search` / `labctl pack add
<name>` read. It is the Krew / Artifact-Hub pattern: no servers, fully
auditable, every change is a reviewable pull request.

> In the OSS layout this content is published to a separate public `registry`
> repo and served as `https://snowops.github.io/registry/index.json` via Pages.
> It lives here in seed form so the format, schema, and tooling ship with the
> engine. The CLI's index URL is `PACK_REGISTRY_INDEX` (see `docs/authoring/registry.md`).

## Files

- [`index.json`](index.json) — the published catalog: one entry per
  pack **version**, mirroring `pack.yaml` metadata plus a resolvable `ref`
  (`oci://…` or `git+https://…@ver`) and a `verified` flag.
- Schema: [`../sdk/schemas/index.schema.json`](../sdk/schemas/index.schema.json).

## Add or update a pack (PR flow)

1. Publish your pack first (`labctl pack publish … oci://…`, see
   `docs/authoring/publishing.md`).
2. Add an entry to `index.json`:

   ```json
   {
     "name": "your-org/your-pack",
     "version": "1.0.0",
     "ref": "oci://ghcr.io/your-org/your-pack:1.0.0",
     "publisher": "your-org",
     "license": "Apache-2.0",
     "tier": "community",
     "description": "what it does"
   }
   ```

3. Validate locally and open a PR:

   ```bash
   bin/labctl pack validate-index registry/index.json
   ```

The `validate-index` CI gate runs the same check on every PR. `verified: true`
is set by maintainers only, after confirming publisher identity.

## Signing

The published index may be cosign-signed; the CLI verifies it when
`PACK_REGISTRY_KEY` points at the registry's public key. Unsigned indexes are
accepted by default (community), so verification is opt-in for now.
