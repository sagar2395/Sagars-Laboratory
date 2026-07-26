# Task 054: `platform/mesh` category — istio + linkerd providers

## Phase
M4 — Stack Expansion (see docs/ROADMAP.md Part II)

## Type
feature

## Priority
P1

## Description
Add a service-mesh platform category with two swappable providers (istio,
linkerd), selected by `MESH_PROVIDER`. Same provider contract as every
other category: `install.sh`, `uninstall.sh`, `status.sh`, `values.yaml`.
This unlocks mesh traffic-management simulations (058) and side-by-side
PoC comparison — a core simulator use case.

## Files to Modify
- `platform/mesh/istio/` and `platform/mesh/linkerd/` (provider dirs)
- platform registry / category wiring in `cmd/labctl` (registration only)
- `make/` platform targets (provider-aware, per task 005 pattern)
- `platform/README.md`

## Implementation Notes
- Istio via the `istio/istiod` + `istio/base` Helm charts (ambient or
  sidecar — pick sidecar for broader demo compatibility); linkerd via its
  Helm charts (handle its cert generation in install.sh portably).
- Demo apps opt into the mesh via namespace labels set by install.sh —
  and removed by uninstall.sh (idempotent both ways).
- k3d resource reality: document minimum memory; keep values minimal.
- Pin chart versions in `versions.env`.

## Acceptance Criteria
- [ ] `MESH_PROVIDER=istio labctl platform up mesh` meshes go-api (sidecar visible)
- [ ] Swapping to linkerd on the same cluster works after uninstall
- [ ] `status.sh` reports mesh health accurately for both providers
- [ ] Scripts portable + idempotent; versions pinned

## Testing Instructions
Install/swap/uninstall loop on k3d. Runbook:
`docs/runbooks/10-stack-expansion.md`.

## Dependencies
None (uses existing platform machinery)
