# The Pack Registry & Discovery

The **registry index** is how users find packs without knowing an exact ref. It
is a static, optionally signed JSON catalog — the Krew / Artifact-Hub pattern:
no servers, PR-moderated, fully auditable. A hosted catalog API (task 073) is a
future *superset*, never a replacement for this portable artifact.

## How the CLI uses it

```bash
labctl pack search [term]      # find packs (no term = list all)
labctl pack info  <name>       # details — installed pack, else registry entry
labctl pack add   <name>       # resolve <name> via the index, then install
```

`pack add` resolves its argument in order:

1. an **OCI ref** (`oci://…`),
2. a **git source** (anything with `://` or a `git@` prefix),
3. otherwise a **registry name** — looked up in the index and resolved to the
   **latest** version's `ref` (a bare `kafka-drills` matches `acme/kafka-drills`).

### Configuration

| Env var | Default | Meaning |
|---|---|---|
| `PACK_REGISTRY_INDEX` | `https://snowops.github.io/registry/index.json` | index URL (supports `https://` and `file://`) |
| `PACK_REGISTRY_KEY` | _(empty)_ | cosign public key; when set, the index signature (`<url>.sig`) is verified before the index is trusted |

The index is cached at `.labctl/cache/registry-index.json` with a 1-hour TTL.
Within the TTL the cache is used directly; afterwards it is refetched. If a
refetch fails, the last good cache is used (with a warning) so discovery keeps
working offline — **fail-closed only applies to signature verification**, never
to availability.

## The index format

`apiVersion: packs.flightdeck.dev/v1`, `kind: PackIndex`, plus `entries[]`. Each
entry mirrors `pack.yaml` metadata and adds the resolvable `ref` and a `verified`
flag. Schema: [`sdk/schemas/index.schema.json`](../../sdk/schemas/index.schema.json).

```json
{
  "apiVersion": "packs.flightdeck.dev/v1",
  "kind": "PackIndex",
  "generated": "2026-06-14T00:00:00Z",
  "entries": [
    {
      "name": "acme/kafka-drills",
      "version": "1.4.2",
      "ref": "oci://ghcr.io/acme/kafka-drills:1.4.2",
      "publisher": "acme",
      "license": "Apache-2.0",
      "tier": "community",
      "keywords": ["kafka", "incident"],
      "verified": true
    }
  ]
}
```

- `ref` is `oci://…` or `git+https://…@ver` (the `git+` prefix is stripped before
  the git installer runs).
- Multiple versions of a pack are separate entries; `search` shows the newest and
  `add <name>` installs the newest.
- `verified` is set by maintainers after confirming publisher identity.

## Publishing to the registry

The index source of truth lives in the [`registry/`](../../registry/) seed (and,
in the OSS layout, a separate public `registry` repo served via Pages). To add or
update a pack, open a PR that adds an entry and passes validation:

```bash
labctl pack validate-index registry/index.json
```

The [`validate-index`](../../.github/workflows/validate-index.yml) workflow runs
the same check on every PR. See [`registry/README.md`](../../registry/README.md)
for the full contributor flow.
