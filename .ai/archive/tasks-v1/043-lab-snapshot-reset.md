# Task 043: Lab snapshot & reset — `labctl lab snapshot/restore/reset`

## Phase
M1 — Scenario Engine v2

## Type
feature

## Priority
P1

## Description
Fast iteration is the point of a simulator. Add `labctl lab snapshot` (record
the desired state: active platform components, deployed apps, active
scenarios), `labctl lab restore <snapshot>` (re-converge to it), and
`labctl lab reset` (tear down everything back to post-init: cluster +
ingress only).

## Files to Modify
- `cmd/labctl/internal/lab/` (new: snapshot model + orchestration)
- `cmd/labctl/` (command group + API routes)
- `docs/cli-reference.md`

## Implementation Notes
- Snapshot = a small YAML manifest of *intent* (which providers, apps,
  scenarios were active + their key env vars), NOT an etcd/volume backup.
  Restore replays the existing idempotent up/install paths — golden rule 5
  makes this nearly free.
- Store snapshots in `.labctl/snapshots/<name>.yaml` (gitignored runtime
  state, golden rule 6).
- `reset` = scenario down for all active, app undeploy, platform uninstall
  except ingress; confirm with `--yes` flag for non-interactive use.
- Surface progress through the existing action-event stream.

## Acceptance Criteria
- [ ] snapshot → reset → restore returns the lab to a state where
      `labctl scenario verify` passes for previously active scenarios
- [ ] `reset` completes on k3d in < 5 minutes
- [ ] Snapshots survive labctl restarts; never committed to git
- [ ] Unit tests for snapshot serialization + restore planning

## Testing Instructions
`cd cmd/labctl && go test ./...`; manual loop in
`docs/runbooks/07-scenario-engine-v2.md`.

## Dependencies
041 (verify used to prove restore correctness)
