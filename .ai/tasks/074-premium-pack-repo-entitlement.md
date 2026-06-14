# Task 074: Premium pack repo + entitlement service (private)

## Phase
M8 — Marketplace (deferred)

## Type
feature

## Priority
P2

## Description
Stand up the private `lab-premium` repo and a license-key entitlement service
that implements the `pkg/entitlement` interface (task 070). Premium/enterprise
packs are distributed via an authenticated OCI registry; the license token is the
access credential. The OSS engine runs them through the same public interfaces —
no fork, no special build.

## Implementation Notes
- Lives entirely OUTSIDE the OSS tree (separate private repo, proprietary EULA).
- Entitlement service issues + verifies license tokens; integrates with a billing
  provider (Stripe/Paddle) later.
- Premium packs use the same `pack.yaml` format with `tier: premium|enterprise`.

## Acceptance Criteria
- [ ] Private repo with premium packs under a proprietary license
- [ ] Entitlement service implements pkg/entitlement; OSS engine unmodified
- [ ] Authenticated OCI pull gated by license token; fails closed without one

## Dependencies
068, 070
