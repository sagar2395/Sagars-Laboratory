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
- [ ] `labctl pack search/info` read the static index and resolve installs
- [ ] Index entry JSON Schema + a validate-index PR gate (in the registry repo)
- [ ] Seed index lists the first-party community packs
- [ ] Signed index; CLI verifies before trusting

## Testing Instructions
Point the CLI at a local index fixture; search, info, and add by index name;
confirm signature verification and TTL caching.

## Dependencies
067, 068
