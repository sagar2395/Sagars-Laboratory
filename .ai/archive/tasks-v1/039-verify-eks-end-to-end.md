# Task 039: Verify EKS End-to-End

## Phase
6

## Type
infra

## Priority
P2

## Description
Provision an EKS dev cluster via Terraform, deploy go-api with the cloud Helm
profile, run one scenario, and tear everything down. First real-world validation
of the EKS path. **Incurs spend (incl. NAT Gateway ~$32/mo if left running).**

## Files to Modify
- Likely fixes to `runtimes/eks/{up,down}.sh`, `foundation/terraform/modules/eks/`,
  `foundation/terraform/environments/dev/`, `engine/build/ecr.sh`
- `docs/runbooks/06-cloud-runtimes.md` (record timings/cost/manual steps)

## Implementation Notes
- Follow `docs/runbooks/06-cloud-runtimes.md` with `--profile eks`.
- Configure remote state first (Task 027).
- The EKS module creates a VPC + NAT Gateway — confirm teardown removes them.
- Capture manual workarounds and fold them back into the scripts.

## Acceptance Criteria
- [ ] `labctl runtime up --profile eks` provisions a working cluster unattended.
- [ ] go-api builds via ECR and deploys with `values-cloud.yaml`.
- [ ] `observability-sre` scenario runs on EKS.
- [ ] `runtime down` removes all resources (VPC, NAT, node group, ECR); cost recorded.

## Testing Instructions
Execute runbook 06 fully with `--profile eks`. Verify the AWS console shows no
leftover VPC/NAT/EKS/ECR after teardown.

## Dependencies
027; depends on Phase 2 stability
