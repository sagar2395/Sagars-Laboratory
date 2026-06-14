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
- [x] Canonical module/vanity path chosen and applied; `go build ./...` green
- [x] All imports, docs, and help text reference the new identity consistently
- [x] Vanity path resolves (meta tags) if adopted

## Testing Instructions
`go build ./... && go test ./...` after the rename; `go get` via the vanity path
resolves (if configured).

## Dependencies
Maintainer sub-decision A0 — the product name (manual-actions doc)

## Progress
- Decision: module path **`go.flightdeck.dev/labctl`** (brand Flightdeck +
  vanity domain `go.flightdeck.dev` were already resolved A3; A0 product name =
  Flightdeck — only the trademark search/registration remains, which does not
  block the code rename per the manual-actions doc).
- Done: renamed `module github.com/sagars-lab/labctl` → `go.flightdeck.dev/labctl`
  in `cmd/labctl/go.mod` and rewrote the import path in all 49 Go files; `go
  build ./...` / `go test ./...` green; imports re-gofmt'd (the new path sorts
  after `github.com/...`). CLI help (`labctl` Short/Long) rebranded to Flightdeck.
  README title + identity note, CLAUDE.md product line updated. Vanity meta-tag
  page added at `docs/vanity/labctl/index.html`; maintainer hosting step recorded
  in MAINTAINER-MANUAL-ACTIONS (A0).
- Deliberately left unchanged: concrete GitHub repo URLs / clone dir still point
  at the current origin (the `snowops/flightdeck` transfer is maintainer action
  C1 — rewriting links now would break them); the `/etc/hosts` marker comments in
  `cmd/hosts.go` (renaming would orphan existing host entries on users' machines).
- Caveat: full `go get go.flightdeck.dev/labctl` resolution also needs the vanity
  host live (A0) and the module's subdir layout addressed at the org transfer;
  internal builds are from source and unaffected.
