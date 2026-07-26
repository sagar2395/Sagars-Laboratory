# Task 061: Cost & capacity — opencost provider + right-sizing exercise

## Phase
M5 — Multi-Env & Day-2 Ops

## Type
feature

## Priority
P2

## Description
Add `platform/cost/opencost` (provider contract, `COST_PROVIDER`) and a
`cost-right-sizing` v2 scenario: deliberately over-provisioned demo app
requests, objective is to right-size them using OpenCost + metrics data;
checks verify requests were reduced while the app stays healthy under
steady traffic.

## Files to Modify
- `platform/cost/opencost/` (provider dir)
- `scenarios/cost-right-sizing/`
- registry wiring + `make/` targets, `platform/README.md`, `versions.env`

## Implementation Notes
- OpenCost Helm chart pointed at the existing Prometheus (namespace from
  env). On k3d there's no real billing — use OpenCost's default on-prem
  pricing config and say so in the docs.
- Scenario stage 1 deploys go-api with inflated requests via a values
  override; the exercise is editing them down; checks compare
  `kubectl get deploy -o jsonpath` requests against thresholds + a health
  check under traffic.

## Acceptance Criteria
- [ ] OpenCost UI reachable via ingress; shows per-namespace cost
- [ ] Scenario verify fails while over-provisioned, passes after right-sizing
- [ ] Clean uninstall; portable + idempotent; versions pinned
- [ ] docs/scenarios.md updated

## Testing Instructions
Exercise walkthrough in `docs/runbooks/11-multi-env-day2.md`.

## Dependencies
040, 041, 042
