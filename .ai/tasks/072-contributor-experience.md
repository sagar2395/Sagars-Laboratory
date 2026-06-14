# Task 072: Contributor experience — scaffolds, authoring guide, docs

## Phase
M7 — OSS & Ecosystem Foundation

## Type
feature

## Priority
P1

## Description
Lower the barrier to a first contribution so the community can produce content at
scale. Provide scaffolds, an authoring guide, editor validation, and a starter
set of issues.

## Files to Modify
- `cmd/labctl/cmd/` — `labctl pack init`, `labctl scenario new` (scaffold from
  `sdk/pack-template/`, `sdk/scenario-template/`)
- `sdk/pack-template/`, `sdk/scenario-template/`
- `docs/authoring/` — "write a scenario pack" end-to-end guide
- `sdk/schemas/` — wire JSON Schemas for editor (VS Code) validation
- `.github/` — "good first issue"/"help wanted" labels + a few seed issues (manual)

## Implementation Notes
- Scaffolds produce a valid, lints-clean, verify-ready pack/scenario out of the box.
- Authoring guide walks: init → edit scenario.yaml → add checks → `scenario verify`
  → `pack publish`.
- JSON Schema references in templates give inline validation in editors.

## Acceptance Criteria
- [ ] `labctl pack init` / `labctl scenario new` produce valid, verifiable output
- [ ] Authoring guide covers the full create→verify→publish loop
- [ ] JSON Schema validation works in a standard editor
- [ ] A handful of good-first-issues exist (maintainer seeds these)

## Testing Instructions
Scaffold a new pack, run lint + `scenario verify` on it unchanged, publish to a
local registry; confirm a newcomer can follow the guide end to end.

## Dependencies
066, 067
