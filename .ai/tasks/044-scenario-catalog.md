# Task 044: Scenario catalog — install external scenario packs

## Phase
M1 — Scenario Engine v2

## Type
feature

## Priority
P2

## Description
Make scenarios shareable: `labctl scenario install <git-url>[@ref]` clones a
scenario pack into `.labctl/catalog/<name>/`, validates it against the v2
schema, and makes it available to `scenario list/up/verify` alongside
in-repo scenarios. `labctl scenario uninstall <name>` removes it.

## Files to Modify
- `cmd/labctl/internal/scenario/` (multi-root scenario discovery)
- `cmd/labctl/` (install/uninstall subcommands)
- `docs/scenarios.md` (authoring + publishing a pack)

## Implementation Notes
- A pack is just a git repo (or subdir) containing one or more scenario
  dirs with `scenario.yaml`. Reuse the v2 parser for validation; reject
  packs that fail schema validation or contain absolute paths.
- Installed packs live under `.labctl/` (runtime state, gitignored).
- Name collisions: installed pack loses; warn and require `--name` override.
- Security note in docs: packs run scripts/manifests — only install trusted
  sources. No auto-update.

## Acceptance Criteria
- [ ] Installing a pack from a public git URL makes it appear in `scenario list` (marked as external)
- [ ] Invalid packs are rejected with a clear validation error
- [ ] `scenario up/verify/down` work identically for external scenarios
- [ ] Authoring/publishing guide added to docs/scenarios.md

## Testing Instructions
Create a minimal pack in a scratch repo, install, activate, verify, remove.
Steps in `docs/runbooks/07-scenario-engine-v2.md`.

## Dependencies
040, 041
