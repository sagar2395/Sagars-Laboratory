# Task 047: Progressive hints & solution walkthroughs

## Phase
M2 — Incident Engine

## Type
feature

## Priority
P1

## Description
`labctl incident hint` reveals the next unrevealed hint for the active
incident (from `hints.md`); `labctl incident solution` prints the full
walkthrough (with a confirmation, since it spoils the exercise). Hints
revealed are recorded in the run state so scoring (048/051) can penalize
them.

## Files to Modify
- `cmd/labctl/internal/incident/` (hint parsing + reveal state)
- `cmd/labctl/` (subcommands + REST routes)
- `docs/cli-reference.md`

## Implementation Notes
- `hints.md` format: hints separated by `## Hint N` headings — keep it
  plain markdown so fault authors need no special tooling.
- Reveal state is per-run (resets on new inject), stored with the active
  incident state in `.labctl/`.
- UI later consumes the same REST endpoints (053) — return hints as
  structured JSON `{index, total, text}`.

## Acceptance Criteria
- [ ] Repeated `incident hint` reveals hints in order, then says "no more hints"
- [ ] `incident solution` requires confirmation (or `--yes`)
- [ ] Revealed-hint count is persisted in the run record
- [ ] Works for all six 045 faults

## Testing Instructions
Manual loop in `docs/runbooks/08-incident-engine.md`.

## Dependencies
046
