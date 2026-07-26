# Task 051: Challenge mode — timed, auto-graded runs

## Phase
M3 — Learning & Assessment

## Type
feature

## Priority
P1

## Description
Skills-assessment wrapper around scenarios and incidents: `labctl challenge
list`, `challenge start <name>` (starts timer, injects/activates, hides
hints behind score penalty), `challenge status`, `challenge submit` (runs
checks, computes grade, records result). Grade = checks passed / total,
elapsed time vs `parTime`, hints used.

## Files to Modify
- `challenges/README.md` + `challenges/<name>/challenge.yaml` (format + first 3 challenges)
- `cmd/labctl/internal/challenge/` (new package)
- `cmd/labctl/` (command group + REST routes)
- `docs/cli-reference.md`

## Implementation Notes
- `challenge.yaml` = `{name, description, parTime, setup: {scenario|incident},
  grading: {checks | useDetectionCheck}, hintPenalty}`. Declarative — reuse
  040/041/046 machinery entirely.
- First three challenges: "restore the broken deploy" (wraps
  bad-deploy-rollout), "find the memory leak" (wraps oom-kill), "make the
  SLO green" (observability checks under traffic from 042).
- Score formula documented in challenges/README.md; keep it simple and
  deterministic (no wall-clock randomness in grading).
- One active challenge at a time; `challenge abort` cleans up via the
  underlying scenario down / incident resolve.

## Acceptance Criteria
- [ ] Full loop: start → fix → submit produces a score record
- [ ] Hints taken during a challenge reduce the score per hintPenalty
- [ ] `challenge abort` always returns the lab to a clean state
- [ ] Three challenges shipped and completable

## Testing Instructions
Complete each challenge once. Runbook:
`docs/runbooks/09-learning-and-challenges.md`.

## Dependencies
041, 046, 047, 048
