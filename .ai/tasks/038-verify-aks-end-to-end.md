# Task 038: Verify AKS End-to-End

## Phase
6

## Type
infra

## Priority
P2

## Description
Provision an AKS dev cluster via Terraform, deploy go-api with the cloud Helm
profile, run one scenario, and tear everything down. First real-world validation
of the AKS path (modules were never run against a live account). **Incurs spend.**

## Files to Modify
- Likely fixes to `runtimes/aks/{up,down}.sh`, `foundation/terraform/modules/aks/`,
  `foundation/terraform/environments/dev/`, `engine/build/acr.sh`
- `docs/runbooks/06-cloud-runtimes.md` (record timings/cost/manual steps)

## Implementation Notes
- Follow `docs/runbooks/06-cloud-runtimes.md` with `--profile aks`.
- Configure remote state first (Task 027) so a failed run doesn't strand state.
- Capture every manual workaround needed and fold it back into the scripts.
- Confirm `runtime down` leaves **zero** billable resources (check the portal).

## Acceptance Criteria
- [ ] `labctl runtime up --profile aks` provisions a working cluster unattended.
- [ ] go-api builds via ACR and deploys with `values-cloud.yaml`.
- [ ] `observability-sre` scenario runs on AKS.
- [ ] `runtime down` removes all resources; cost + timings recorded in the runbook.

## Testing Instructions
Execute runbook 06 fully (up → app → scenario → down). Verify the Azure portal
shows no leftover resource group / LB / ACR after teardown.

## Dependencies
027; depends on Phase 2 stability
