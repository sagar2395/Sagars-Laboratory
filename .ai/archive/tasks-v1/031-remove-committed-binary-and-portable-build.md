# Task 031: Remove Committed Binary; Make Build Host-Native

## Phase
0

## Type
infra

## Priority
P0

## Description
A prebuilt Linux x86-64 `bin/labctl` was committed to git (built on WSL). It will
not run on macOS and must never be in version control. `bin/` is now gitignored
and the binary removed. Ensure `make cli-build` reliably produces a host-native
binary on any platform, and document cross-compilation for releases.

## Files to Modify
- `make/cli.mk`
- `README.md` (build instructions)

## Implementation Notes
- `make cli-build` should run `go build` for the host (Go picks host GOOS/GOARCH
  by default) and output to `bin/labctl`. Keep the existing step that copies
  `ui/dist/` into `cmd/labctl/ui/dist/` before building (for `go:embed`).
- Add a `cli-build-all` (or documented one-liner) for release cross-compilation:
  ```bash
  for t in darwin/arm64 darwin/amd64 linux/amd64 linux/arm64; do
    GOOS=${t%/*} GOARCH=${t#*/} go build -o dist/labctl-${t%/*}-${t#*/} ./cmd/labctl
  done
  ```
- Verify `bin/labctl` is gitignored (it is) and not tracked (`git ls-files bin/`
  returns nothing).

## Acceptance Criteria
- [ ] `git ls-files bin/` is empty.
- [ ] `make cli-build` produces a runnable `bin/labctl` on macOS and on Linux.
- [ ] A documented command cross-builds for darwin/linux × arm64/amd64.

## Testing Instructions
`make cli-build && file bin/labctl` → reports the host architecture. `bin/labctl
--help` runs. Follow `docs/runbooks/00-cross-platform-setup.md`.

## Dependencies
None
