# Task 034: Shell Portability Sweep + Lint

## Phase
0

## Type
infra

## Priority
P1

## Description
Audit every shell script in the repo for GNU/Linux-only constructs that break on
macOS (BSD userland), and add an automated check so regressions are caught in CI.

## Files to Modify
- Any `*.sh` containing non-portable constructs (sweep `bootstrap/`, `runtimes/`,
  `platform/`, `engine/`, `services/`, `scenarios/`)
- `.github/workflows/ci.yaml` (add a portability lint job — coordinate with task 006)

## Implementation Notes
Hunt for and fix these macOS-incompatible patterns:
- `grep -oP` / `\K` / `grep -P`  → use `sed`/`awk` or `grep -oE`.
- `sed -i` with no suffix         → use `sed -i.bak ... && rm -f *.bak` or a temp file.
- `readlink -f`                   → portable dir resolution via
  `cd "$(dirname "$0")" && pwd`.
- `date -d` / `date --date`       → avoid, or branch on `uname`.
- `stat -c`                       → `stat -f` differs on macOS; avoid or branch.
- `mktemp` flag differences, `xargs -r`, `sort -V` on old BSD.
- Prefer `#!/usr/bin/env bash` (already the convention) and `set -euo pipefail`.
- Add `shellcheck` to CI; optionally add a grep-based denylist step that fails the
  build if any forbidden construct reappears.

## Acceptance Criteria
- [ ] A repo-wide grep for the forbidden patterns above returns nothing.
- [ ] `shellcheck` passes (or documented, justified suppressions) on all scripts.
- [ ] CI fails if a forbidden construct is reintroduced.

## Testing Instructions
`grep -rnE 'grep -oP|sed -i |readlink -f|date -d|stat -c' --include='*.sh' .`
returns nothing. Run the new CI job locally if possible. Spot-check a few scripts
on macOS.

## Dependencies
030 (bootstrap is the biggest offender)
