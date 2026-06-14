# Task 068: OCI pack distribution + signing

## Phase
M7 — OSS & Ecosystem Foundation

## Type
feature

## Priority
P1

## Description
Add OCI artifacts as a pack distribution encoding alongside git. OCI gives
versioning, content addressing, signing (cosign), and auth-gated access — the
same mechanism that later gates premium/enterprise packs (registry returns 401
without a license token). Git packs remain the simple community path.

## Files to Modify
- `pkg/pack/` — OCI push/pull (oras-go), checksum + cosign verify
- `cmd/labctl/cmd/pack.go` — `labctl pack add oci://...`, `labctl pack publish`
- `.github/workflows/` — publish first-party community packs to GHCR on release
- `docs/authoring/publishing.md`

## Implementation Notes
- Use ORAS for OCI artifact push/pull; do not hard-depend on a single registry
  vendor — `publisher/name@version` identity stays registry-neutral.
- Signature verification optional for community packs, required for "verified".
- Entitlement check (task 070 interface) runs before pulling premium packs.

## Acceptance Criteria
- [x] Push/pull a pack as an OCI artifact to/from GHCR
- [x] Checksum + optional cosign signature verification on install
- [x] First-party community packs published on release via CI
- [x] Portable + idempotent; no registry-vendor lock-in

## Testing Instructions
Publish a test pack to a local/registry, pull it by digest and tag, verify
signature, confirm tamper detection fails closed. See runbook 13.

## Dependencies
067, 070

## Progress
- Done: `pkg/pack/archive.go` (deterministic reproducible tar.gz of a pack,
  zip-slip-safe extraction, sha256 checksum + fail-closed `VerifyChecksum`) and
  `pkg/pack/oci.go` (`Publish`/`Pull` that shell out to `oras`/`cosign` via
  injectable `Runner` funcs — golden rule #2, mirroring the catalog's `GitFunc`;
  fail-closed verify-before-extract; keyless + key-pair cosign; oci:// helpers)
  with full offline tests (round-trip, determinism, path-traversal, tamper,
  publish/pull arg construction with stub runners, keyless-identity guard).
- CLI: `labctl pack publish <dir> oci://...` (`--sign`, `--cosign-key`) and
  `labctl pack add oci://...` (`--require-signature`, `--cosign-key`,
  `--certificate-identity`, `--certificate-oidc-issuer`); git path unchanged.
  Refactored `catalog.go` to share a validate-then-rename install tail
  (`installFetched`/`prepareDest`) + `InstallPackFromDir` for the OCI flow.
- CI: `.github/workflows/pack-publish.yml` publishes + keyless-signs the
  `packs/examples/*` community packs to GHCR on release (registry-neutral
  identity).
- Docs: `docs/authoring/publishing.md`, runbook `13-pack-distribution.md`,
  updated `packs.md` / authoring README / `cli-reference.md`.
- Note: 068 was started ahead of its listed dep 070 (entitlement interface) per
  the roadmap's `next` pointer; `--require-signature` is the manual seam that 070
  will later drive from policy. The heavy crypto/OCI clients are intentionally
  NOT vendored — the `oras`/`cosign` binaries are wrapped, keeping the binary
  small, cross-platform, and the SDK import-light.
