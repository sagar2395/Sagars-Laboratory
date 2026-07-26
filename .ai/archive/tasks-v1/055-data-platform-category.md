# Task 055: `platform/data` category — kafka (strimzi) + postgres (cnpg)

## Phase
M4 — Stack Expansion

## Type
feature

## Priority
P1

## Description
Add a data-infrastructure category with two providers: `kafka` (Strimzi
operator + a small cluster) and `postgres` (CloudNativePG operator + a
small HA cluster). Selected by `DATA_PROVIDER`; both can coexist (they are
additive, like monitoring components) — model them as sub-components
`data/kafka`, `data/postgres` following the `monitoring/*` pattern.
Unlocks the event-driven scenario (058) and DB-failover drills (M5).

## Files to Modify
- `platform/data/kafka/`, `platform/data/postgres/` (provider dirs)
- registry wiring + `make/` targets
- `platform/README.md`, `versions.env`

## Implementation Notes
- Strimzi: operator Helm chart + a `Kafka` CR (1 broker, ephemeral storage
  for k3d; values file keeps it tiny). CNPG: operator chart + a `Cluster`
  CR (2 instances) — its failover is a ready-made day-2 drill.
- `status.sh` should check operator + CR readiness conditions via kubectl
  jsonpath (portable).
- Keep resource requests minimal — this must fit a laptop k3d alongside
  monitoring.

## Acceptance Criteria
- [ ] `labctl platform up data/kafka` yields a ready Kafka cluster (produce/consume smoke test in status or runbook)
- [ ] `labctl platform up data/postgres` yields a ready 2-instance cluster; deleting the primary triggers failover
- [ ] Uninstall removes operators + CRs + PVCs cleanly
- [ ] Versions pinned; scripts portable + idempotent

## Testing Instructions
Smoke tests in `docs/runbooks/10-stack-expansion.md` (kcat produce/consume,
psql connect, failover watch).

## Dependencies
None
