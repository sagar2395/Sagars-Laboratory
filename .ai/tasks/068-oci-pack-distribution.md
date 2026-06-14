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
- [ ] Push/pull a pack as an OCI artifact to/from GHCR
- [ ] Checksum + optional cosign signature verification on install
- [ ] First-party community packs published on release via CI
- [ ] Portable + idempotent; no registry-vendor lock-in

## Testing Instructions
Publish a test pack to a local/registry, pull it by digest and tag, verify
signature, confirm tamper detection fails closed.

## Dependencies
067, 070
