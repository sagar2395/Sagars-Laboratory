# Task 045: Fault library — incidents/ contract + first 6 faults

## Phase
M2 — Incident Engine (see docs/ROADMAP.md Part II)

## Type
feature

## Priority
P0

## Description
Create the fault contract and the first six realistic, **reversible**
production faults. Each fault is a directory `incidents/<name>/` containing:
`fault.yaml` (metadata, category, severity, target, detection `check` using
the 040 schema, alert expectation), `inject.sh`, `resolve.sh`, `hints.md`
(progressive hints), `solution.md` (full walkthrough).

First six faults: `crashloop-bad-config`, `oom-kill`, `network-blackhole`,
`bad-deploy-rollout`, `service-selector-broken`, `noisy-neighbor` (CPU hog pod).

> Implementation note: the originally planned `dns-blackhole` and `pvc-full`
> were swapped for `network-blackhole` (deny-all-ingress NetworkPolicy) and
> `service-selector-broken` (empty endpoints) — same teaching goals, but
> their detection works reliably on a default k3d cluster, whereas DNS exec
> probes and PVC behavior vary with the local CNI/storage setup.

## Files to Modify
- `incidents/README.md` (the contract, mirroring platform/README.md style)
- `incidents/<name>/...` (six fault dirs)
- `docs/SIMULATOR.md` cross-reference if contract details drift

## Implementation Notes
- Faults must target the demo apps (go-api, echo-server) or their
  namespaces — never platform components or kube-system.
- `inject.sh` must be idempotent and record what it changed (e.g. annotate
  the resource) so `resolve.sh` can always undo it — resolve is the escape
  hatch and must work even if the user half-fixed things.
- Detection check in `fault.yaml` = "the fault is RESOLVED when this check
  passes" (reuses 040 check schema; runners from 041).
- Portable shell, golden rule 1. No GNU-only flags.

## Acceptance Criteria
- [ ] All six faults inject, visibly break the app, and resolve cleanly
- [ ] Each fault.yaml has a working detection check
- [ ] hints.md has ≥3 progressive hints; solution.md reproduces the fix
- [ ] incidents/README.md documents the contract for new faults

## Testing Instructions
For each fault: inject → observe breakage (curl/Grafana) → run resolve.sh →
detection check passes. Runbook: `docs/runbooks/08-incident-engine.md`.

## Dependencies
040 (check schema)
