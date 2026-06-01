# Task 033: Support macOS Container Runtimes

## Phase
0

## Type
infra

## Priority
P0

## Description
k3d needs a Docker-compatible daemon. On Linux that's usually native Docker; on
macOS it's Docker Desktop, Colima, or OrbStack. The preflight check and k3d up
flow assume Linux Docker. Detect an available runtime and give a clear, correct
message when none is running, on both OSes.

## Files to Modify
- `cmd/labctl/cmd/check.go` (or the `check` implementation)
- `runtimes/k3d/up.sh`
- `docs/runbooks/00-cross-platform-setup.md`

## Implementation Notes
- Preflight: `docker info` (or `docker version --format ...`) is the portable
  liveness probe regardless of which backend provides the socket. If it fails,
  print OS-specific guidance:
  - macOS: "Start Docker Desktop, or `colima start`, or OrbStack."
  - Linux: "Start the docker service (`sudo systemctl start docker`) and ensure
    your user is in the `docker` group."
- Do not hardcode a socket path for the liveness probe. (Note: Chaos Mesh DOES
  need a containerd socket — that is a separate, cloud/runtime concern handled in
  task 026, not here.)
- k3d itself is cross-platform; verify `k3d cluster create` works under Colima/
  OrbStack (they expose a Docker-compatible socket). Document any required
  `DOCKER_HOST` for Colima.

## Acceptance Criteria
- [ ] `labctl check cluster` (preflight) correctly detects a running runtime on
      macOS (Docker Desktop / Colima / OrbStack) and Linux.
- [ ] Clear, OS-appropriate message when no runtime is running.
- [ ] `labctl runtime up` (k3d) succeeds on macOS with at least one of these backends.

## Testing Instructions
On macOS with Docker Desktop (and again with Colima if available): stop the
runtime → `labctl check` reports it clearly; start it → `labctl runtime up`
creates a cluster. Follow `docs/runbooks/00-cross-platform-setup.md`.

## Dependencies
None
