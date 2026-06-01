# Task 030: Make Tool Bootstrap OS/Arch-Aware

## Phase
0

## Type
infra

## Priority
P0

## Description
`bootstrap/setup-tools.sh` was written on WSL and assumes Linux x86-64. It
downloads `https://dl.k8s.io/.../bin/linux/amd64/kubectl` and parses versions with
`grep -oP` (PCRE, GNU-only). On macOS (BSD grep, arm64) it fails or installs the
wrong binaries. Make the bootstrap detect the host OS and CPU architecture and
install the correct tools for it.

## Files to Modify
- `bootstrap/setup-tools.sh`
- `README.md` (prerequisites section, if install steps change)

## Implementation Notes
- Detect once at the top:
  ```bash
  OS="$(uname -s | tr '[:upper:]' '[:lower:]')"      # linux | darwin
  ARCH="$(uname -m)"                                  # x86_64 | arm64 | aarch64
  case "$ARCH" in x86_64|amd64) ARCH=amd64;; arm64|aarch64) ARCH=arm64;; esac
  ```
- Use `${OS}/${ARCH}` in download URLs:
  - kubectl: `https://dl.k8s.io/release/v${KUBECTL_VERSION}/bin/${OS}/${ARCH}/kubectl`
  - k3d, helm: their release assets follow the same `${OS}-${ARCH}` convention.
- Replace every `grep -oP '...\K...'` with portable `sed`/`awk` (no `\K`, no `-P`).
- Install location: keep `/usr/local/bin` (writable with sudo on both macOS and
  Linux). On macOS, prefer `brew` when available but do NOT hard-require it.
- Docker: keep the manual-install guidance, but on macOS point to Docker
  Desktop / Colima / OrbStack (see task 033).
- Keep `set -euo pipefail` and idempotent "already installed at correct version" checks.

## Acceptance Criteria
- [ ] No `linux/amd64` or other hardcoded OS/arch strings remain in the script.
- [ ] No `grep -oP` / `\K` / other GNU-only constructs remain.
- [ ] `make setup-tools PROFILE=k3d` installs correct binaries on macOS (arm64) and Linux (amd64).
- [ ] Re-running is a no-op when versions already match.

## Testing Instructions
Run `bash -n bootstrap/setup-tools.sh` (syntax). On a Mac: `make setup-tools
PROFILE=k3d` then `kubectl version --client`, `k3d version`, `helm version` all
report the pinned versions. Follow `docs/runbooks/00-cross-platform-setup.md`.

## Dependencies
None
