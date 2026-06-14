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
- [x] `labctl pack init` / `labctl scenario new` produce valid, verifiable output
- [x] Authoring guide covers the full create→verify→publish loop
- [x] JSON Schema validation works in a standard editor
- [~] A handful of good-first-issues exist (labels + guidance shipped; maintainer seeds the issues)

## Testing Instructions
Scaffold a new pack, run lint + `scenario verify` on it unchanged, publish to a
local registry; confirm a newcomer can follow the guide end to end.

## Dependencies
066, 067

## Progress
- Done: `internal/scaffold` (single source of truth) generates valid,
  verify-ready content; `labctl scenario new <name>` → `scenarios/<name>/` and
  `labctl pack init <name>` → a pack dir with `pack.yaml` + README + one bundled
  scenario (both with a passing `checks/ready.sh`). E2E verified: scaffold →
  `scenario verify` green → `pack publish` validates the manifest.
- Reference templates `sdk/scenario-template/` + `sdk/pack-template/` (with
  `# yaml-language-server: $schema=…` modelines) for manual copiers.
- Editor validation: `.vscode/settings.json` maps `sdk/schemas/*` to
  scenario.yaml / pack.yaml / registry index globs (Red Hat YAML ext); the
  modelines also validate outside the repo.
- Guide: `docs/authoring/first-pack.md` (scaffold→edit→verify→publish), wired
  into the authoring README + cli-reference; `.github/labels.yml` defines
  `good first issue` / `help wanted` + area/kind labels.
- Tests: `internal/scaffold` asserts scaffolded scenario + pack parse, validate,
  preserve publisher/name, and keep the check script executable.
- Maintainer action (not code): seed a few starter issues under the labels.
