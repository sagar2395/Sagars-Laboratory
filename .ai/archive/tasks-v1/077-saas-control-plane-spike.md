# Task 077: SaaS / hosted control-plane spike

## Phase
M9 — Commercial & Hosted (deferred)

## Type
spike

## Priority
P2

## Description
Design spike for a managed/hosted edition: browser-based labs, multi-tenant
clusters, auth/SSO, billing, and a control-plane that wraps the SAME OSS engine
(no engine fork). Output is an architecture RFC, not production code.

## Implementation Notes
- The control-plane lives in the private repo and orchestrates the OSS engine +
  entitlement service; the engine stays unchanged.
- Evaluate per-tenant isolation (namespace vs vcluster vs cluster-per-tenant),
  ephemeral lab lifecycle, and cost controls.

## Acceptance Criteria
- [ ] RFC covering tenancy, lab lifecycle, auth, billing, and the engine seam
- [ ] Confirms no OSS engine changes are required (validates the §6 invariant)

## Dependencies
070, 074
