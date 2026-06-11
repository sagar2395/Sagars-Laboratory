# Task 063: Team sessions — shared remote deploy + shared leaderboard

## Phase
M6 — Team Mode & New Runtimes

## Type
feature

## Priority
P2

## Description
Let a team run one simulator for a game day: package the labctl server as a
Helm chart (`delivery/charts/labctl-server/` or similar) deployable into
any cluster it manages (in-cluster mode), and add a team leaderboard view
aggregating 052 results per user. Includes a "game day" flow: operator
injects `--random --silent` incidents, participants race, leaderboard ranks
by MTTR and hints used.

## Files to Modify
- Helm chart for the server (+ image build via the existing engine)
- `cmd/labctl/internal/server/` (in-cluster kubeconfig mode)
- `ui/src/` (leaderboard aggregation view)
- `docs/cloud-runtimes.md` or new section; runbook 12

## Implementation Notes
- In-cluster mode: detect serviceaccount kubeconfig; scripts the server
  shells out to must be in the image (the image is the repo + tools —
  document size honestly).
- Persist `.labctl/` on a PVC so history/scores survive pod restarts.
- RBAC for the server's ServiceAccount: scope to what scripts need; do
  not ship cluster-admin without a documented warning.

## Acceptance Criteria
- [ ] Chart deploys the server into k3d; UI reachable; can run a scenario from inside
- [ ] Two authenticated users' challenge results rank on the leaderboard
- [ ] History survives pod restart (PVC)
- [ ] Game-day flow documented end-to-end

## Testing Instructions
Full game-day dry run in `docs/runbooks/12-team-mode.md`.

## Dependencies
051, 052, 062
