# Task 057: `platform/autoscaling` category — KEDA + scale-on-load scenario

## Phase
M4 — Stack Expansion

## Type
feature

## Priority
P2

## Description
Add an autoscaling category with a `keda` provider (`AUTOSCALING_PROVIDER`),
plus a small `autoscaling-under-load` scenario: ScaledObject on go-api
driven by Prometheus RPS (and optionally Kafka consumer lag when
data/kafka is installed), verified under traffic from the 042 generator.

## Files to Modify
- `platform/autoscaling/keda/` (provider dir)
- `scenarios/autoscaling-under-load/` (scenario.yaml v2 + manifests)
- registry wiring + `make/` targets, `platform/README.md`, `versions.env`

## Implementation Notes
- KEDA via its Helm chart. ScaledObject with the prometheus scaler
  querying the existing monitoring stack (namespace from env, rule 3/4).
- Scenario stages: baseline (1 replica) → spike traffic (042 profile) →
  checks assert replicas scaled up and p99 stayed under threshold →
  cooldown check asserts scale-down.
- This is the flagship "watch autoscaling actually work" demo — the
  Grafana dashboard panel (replicas vs RPS) is part of the scenario.

## Acceptance Criteria
- [ ] Spike profile scales go-api from 1 to ≥3 replicas; cooldown returns to 1
- [ ] `labctl scenario verify autoscaling-under-load` passes post-spike
- [ ] Dashboard shows replicas vs RPS correlation
- [ ] Clean uninstall; portable + idempotent; versions pinned

## Testing Instructions
Full spike walkthrough in `docs/runbooks/10-stack-expansion.md`.

## Dependencies
040, 041, 042
