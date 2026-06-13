# Task 064: New runtimes — kind (CI-friendly) and GKE

## Phase
M6 — Team Mode & New Runtimes

## Type
feature

## Priority
P2

## Description
Broaden runtime support with two new profiles following the existing
contract (`runtimes/<profile>/up.sh|down.sh|runtime.env`):

1. **kind** — headless, CI-friendly local runtime so scenario/check suites
   can run in GitHub Actions (enables true e2e CI for the simulator).
2. **gke** — third cloud, Terraform module + environment mirroring the
   aks/eks structure.

## Files to Modify
- `runtimes/kind/`, `runtimes/gke/`
- `foundation/terraform/modules/gke/` + `environments/`
- CI workflow (kind-based e2e job, can be nightly not per-PR)
- `docs/cloud-runtimes.md`

## Implementation Notes
- kind: ingress needs the extraPortMappings pattern; mirror what k3d's
  up.sh exposes so platform scripts don't care which local runtime is
  underneath. DOMAIN_SUFFIX still env-driven.
- GKE follows the aks/eks Terraform module shape exactly (remote state per
  task 027). Same caveat as 038/039: verify against a real account once,
  record cost, tear down.
- CI e2e: kind up → platform up (ingress+monitoring) → one scenario up →
  `scenario verify` → down. Keep it under the runner's resource limits.

## Acceptance Criteria
- [ ] `labctl runtime up --profile kind` produces a working lab headlessly
- [ ] Nightly CI job runs one scenario verify on kind and is green
- [ ] GKE provision/deploy/teardown verified once against a real project
- [ ] docs/cloud-runtimes.md covers both

## Testing Instructions
kind: run the CI job locally. GKE: follow
`docs/runbooks/06-cloud-runtimes.md` (extended) + runbook 12 notes.

## Dependencies
041 (verify in CI); 027 pattern for GKE state
