# Task 060: Day-2 ops drills — upgrade, drain, backup/restore

## Phase
M5 — Multi-Env & Day-2 Ops

## Type
feature

## Priority
P2

## Description
The operations people fear, as repeatable checked scenarios:

1. **cluster-upgrade-drill** — upgrade the k3d cluster's k8s version while
   traffic (042) runs; checks measure request success rate through the
   upgrade.
2. **node-drain-drill** — multi-node k3d; cordon+drain a node under load;
   checks assert PDBs held and zero failed requests.
3. **backup-restore-drill** — back up cluster state (velero or
   export-manifests approach — decide in implementation), destroy a
   namespace, restore, verify with checks.

## Files to Modify
- `scenarios/cluster-upgrade-drill/`, `scenarios/node-drain-drill/`,
  `scenarios/backup-restore-drill/`
- `runtimes/k3d/` (multi-node + version-pinned cluster options)
- `docs/scenarios.md`

## Implementation Notes
- k3d makes upgrade simulation possible via cluster recreate with a newer
  image while keeping the registry — be honest in docs about what's
  simulated vs real.
- Each drill's "grade" is the measured availability during the operation —
  surfaced via a promql check (e.g. success rate ≥ 99.5%).
- These drills assume PDBs from the chaos scenario exist — declare as
  prerequisites.

## Acceptance Criteria
- [ ] All three drills run end-to-end on k3d with verify passing
- [ ] Availability during drain/upgrade is measured and reported, not guessed
- [ ] Backup/restore round-trips the go-api namespace
- [ ] docs/scenarios.md updated

## Testing Instructions
Per-drill walkthroughs in `docs/runbooks/11-multi-env-day2.md`.

## Dependencies
040, 041, 042
