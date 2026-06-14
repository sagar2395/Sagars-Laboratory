# Task 071: Module path & brand alignment

## Phase
M7 — OSS & Ecosystem Foundation

## Type
foundation

## Priority
P1

## Description
Resolve the identity inconsistency (module path `github.com/sagars-lab/labctl` vs
repo `sagar2395/Sagars-Laboratory`) and adopt a stable, brandable identity before
an ecosystem forms around it. Pick a vanity import path so the module path never
has to change again even if the GitHub home moves.

Decided: GitHub org **`snowops`**, monorepo core. Remaining: the exact
**trademark-able product name** (and the vanity import path derived from it,
e.g. `go.<product>.dev/...`) — the one sub-decision A0 in the manual-actions doc.

## Files to Modify
- `cmd/labctl/go.mod` (+ any `pkg/` go.mod), all import paths
- vanity import-path config (e.g. `go.<brand>.dev/lab`) — meta tag / proxy
- `README`, docs, install instructions, CLI help strings
- repo/org references throughout docs

## Implementation Notes
- Prefer a vanity path decoupled from GitHub (`go.<brand>.dev/...`) so a future
  org move is a redirect, not a code change.
- This is a mechanical but wide change — do it before the contributor base grows.
- Coordinate with decision A3; keep the current repo as the project origin.

## Acceptance Criteria
- [ ] Canonical module/vanity path chosen and applied; `go build ./...` green
- [ ] All imports, docs, and help text reference the new identity consistently
- [ ] Vanity path resolves (meta tags) if adopted

## Testing Instructions
`go build ./... && go test ./...` after the rename; `go get` via the vanity path
resolves (if configured).

## Dependencies
Maintainer sub-decision A0 — the product name (manual-actions doc)
