# Task 037: Embed the SPA Build into the Binary

## Phase
3

## Type
infra

## Priority
P1

## Description
Wire the SPA production build into the single `labctl` binary via `go:embed`, so
`bin/labctl ui` serves the dashboard with no external files — while keeping a
filesystem fallback for frontend dev.

## Files to Modify
- `make/cli.mk` (build the SPA, copy `ui/dist` → `cmd/labctl/ui/dist`, then `go build`)
- `cmd/labctl/ui/embed.go` (already `//go:embed all:dist`)
- `cmd/labctl/internal/api/server.go` (serve embedded FS; fall back to filesystem when empty)
- `docs/runbooks/03-web-ui.md`

## Implementation Notes
- `make cli-build` order: `cd ui && npm ci && npm run build` → copy `ui/dist/*`
  into `cmd/labctl/ui/dist/` → `go build`. Guard the npm step so a missing Node
  toolchain gives a clear error (or skip if `ui/dist` already built).
- The server already accepts an `io/fs.FS` and auto-detects embedded content;
  ensure it serves `index.html` for client-side routes (SPA fallback) and static
  assets correctly.
- Keep `cmd/labctl/ui/dist/.gitkeep`; never commit built assets.

## Acceptance Criteria
- [ ] `make cli-build` builds the SPA and embeds it; resulting binary needs no
      external UI files.
- [ ] `bin/labctl ui` serves the SPA at `localhost:3939`, deep links work (SPA fallback).
- [ ] Frontend dev still works via Vite proxy (filesystem fallback path).
- [ ] No built UI assets are tracked in git.

## Testing Instructions
`make cli-build`, move/rename `ui/dist`, run `bin/labctl ui` from a different
directory — UI still loads (proves embedding). Follow `docs/runbooks/03-web-ui.md`.

## Dependencies
035, 036
