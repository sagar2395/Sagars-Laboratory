# Task 067: Formal scenario-pack format (`pack.yaml`) + `labctl pack`

## Phase
M7 — OSS & Ecosystem Foundation

## Type
foundation

## Priority
P0

## Description
Formalise the pack format on top of the existing catalog machinery
(`internal/scenario/catalog.go`: InstallPack/ValidatePackDir/Packs). Add a
`pack.yaml` manifest (identity, semver, publisher, license, tier, engine
constraints, dependencies, provenance) and a first-class `labctl pack` command
group. Back-compatible with today's git-pack installs.

## Files to Modify
- `pkg/pack/` — pack.yaml schema, parse, validate, dependency resolve
- `sdk/schemas/pack.schema.json`
- `cmd/labctl/cmd/pack.go` — `labctl pack add|list|info|remove|search` (evolve
  the existing `scenario install|packs|uninstall`, keep aliases)
- `packs/community/` — repackage the built-in scenarios as the first packs
- `docs/authoring/packs.md`, `docs/scenarios.md`, `docs/cli-reference.md`

## Implementation Notes
- `pack.yaml` MUST include fields for `tier`, `license`, `checksum`,
  `signature`, and engine/scenario `apiVersion` constraints even if unused at
  first — adding them later is a breaking change.
- Packs install to `~/.labctl/packs/<publisher>/<name>@<version>/` (versioned;
  multiple versions coexist). Record a lockfile for reproducibility.
- Validate dependencies (platform categories reuse task 002 preflight; pack deps
  use semver ranges).

## Acceptance Criteria
- [ ] `pack.yaml` schema + validation; git packs without one still install (warn)
- [ ] `labctl pack add|list|info|remove|search` work; old `scenario install` aliases
- [ ] Built-in scenarios repackaged under `packs/community/` and installable
- [ ] Dependency + engine-compat resolution with actionable errors

## Testing Instructions
Install a sample pack from a local git dir; verify versioned install dir,
lockfile, dependency resolution, and that `scenario up` finds packaged scenarios.

## Dependencies
066
