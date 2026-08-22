# Task 050: Learning paths — format + `labctl learn`

## Phase
M3 — Learning & Assessment (see docs/ROADMAP.md Part II)

## Type
feature

## Priority
P1

## Description
Guided, ordered learning tracks. A path is `learn/<path-name>/path.yaml`
declaring ordered modules; each module references a scenario stage or an
incident, plus prose (`intro.md`), objectives, and a completion check.
CLI: `labctl learn list`, `learn start <path>`, `learn next`, `learn
progress`. Ship the first path: `kubernetes-foundations` (init → deploy app
→ observability scenario → first incident).

## Files to Modify
- `learn/README.md` + `learn/kubernetes-foundations/` (format + first path)
- `cmd/labctl/internal/learn/` (new package)
- `cmd/labctl/` (command group + REST routes)
- `docs/cli-reference.md`, `docs/scenarios.md` cross-link

## Implementation Notes
- A module = `{name, intro, action: {scenario|incident|command}, check}`.
  Reuse 040 checks and 041 runners for completion verification — no new
  verification logic.
- `learn next` shows the next incomplete module's intro + objective and,
  after the user acts, verifies its check before advancing.
- Progress per path persisted in `.labctl/` (gitignored).
- Paths are declarative content — authoring a path must require zero Go.

## Acceptance Criteria
- [ ] `kubernetes-foundations` is completable end-to-end on a fresh k3d lab
- [ ] Progress survives restarts; `learn progress` shows N/M modules done
- [ ] Module completion is check-verified, not self-reported
- [ ] learn/README.md documents path authoring

## Testing Instructions
Walk the full path on a fresh lab. Runbook:
`docs/runbooks/09-learning-and-challenges.md`.

## Dependencies
041 (checks), 046 (incident modules)
