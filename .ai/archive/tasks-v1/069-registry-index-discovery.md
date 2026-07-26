# Task 069: Registry index + pack discovery

## Phase
M7 — OSS & Ecosystem Foundation

## Type
feature

## Priority
P1

## Description
Stand up the marketplace catalog as a static, signed index (zero infra) and wire
discovery into the CLI. One file per pack/version in a separate public `registry`
repo, served via Pages/CDN. This is the Krew/Artifact-Hub pattern — cheap,
auditable, PR-moderated.

## Files to Modify
- `pkg/pack/index.go` — fetch/cache/verify the index; resolve name → ref
- `cmd/labctl/cmd/pack.go` — `labctl pack search <term>`, `labctl pack info <name>`
- `docs/authoring/registry.md`
- (separate repo `registry`) — index schema, seed entries, validate-index CI

## Implementation Notes
- Index entry schema mirrors `pack.yaml` metadata + the resolvable ref
  (oci://… or git+https://…@ver) + publisher verification status.
- CLI resolves: index name first, else a direct ref; cache with TTL; verify
  index signature.
- Keep the index a portable static artifact; a hosted API (task 073) is a future
  superset, not a replacement.

## Acceptance Criteria
- [x] `labctl pack search/info` read the static index and resolve installs
- [x] Index entry JSON Schema + a validate-index PR gate (in the registry repo)
- [x] Seed index lists the first-party community packs
- [x] Signed index; CLI verifies before trusting

## Testing Instructions
Point the CLI at a local index fixture; search, info, and add by index name;
confirm signature verification and TTL caching. See runbook 13 (section 5).

## Dependencies
067, 068

## Progress
- Done: `pkg/pack/index.go` — `Index`/`IndexEntry` types, `ParseIndex`/`Validate`
  (schema + semantics: name/semver/ref/tier/duplicates), `Find` (bare-name +
  highest-version resolution), `Search` (substring over name/desc/keywords,
  deduped to latest), `ResolveRef` (strips `git+`), and `IndexClient`
  (`Getter`-injectable fetch with `file://` support, TTL file cache, stale
  fallback on network error, optional cosign `verify-blob` that fails closed)
  with full hermetic tests.
- CLI: `labctl pack search [term]`, `pack info <name>` now falls back to the
  index, `pack add <name>` resolves a bare name → latest ref → install (OCI or
  git), `pack validate-index <file>` (the PR gate). Config: `PACK_REGISTRY_INDEX`
  + `PACK_REGISTRY_KEY`; cache at `.labctl/cache/registry-index.json` (1h TTL).
- Registry artifacts: `sdk/schemas/index.schema.json`, `registry/index.json`
  seed (lists first-party hello-pack) + `registry/README.md` contributor flow,
  `.github/workflows/validate-index.yml` PR gate.
- Docs: `docs/authoring/registry.md`, runbook 13 §5, updated packs.md / authoring
  README / cli-reference.
- Notes: the index lives here in seed form; the OSS layout publishes it to a
  separate public `registry` repo over Pages (maintainer action). Signature
  verification is opt-in (`PACK_REGISTRY_KEY`) — community indexes are unsigned;
  availability never fails closed (stale cache fallback), only verification does.
