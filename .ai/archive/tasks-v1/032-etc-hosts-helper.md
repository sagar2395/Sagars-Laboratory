# Task 032: /etc/hosts Helper for *.k3d.local

## Phase
0

## Type
feature

## Priority
P1

## Description
Ingress routes use hostnames like `go-api.k3d.local`, `grafana.k3d.local`. These
only resolve if mapped to `127.0.0.1` in `/etc/hosts`. Today this is an
undocumented manual step that trips up first-time users on both macOS and Linux.
Provide a helper that adds/removes the entries.

## Files to Modify
- `cmd/labctl/cmd/` (new `hosts.go` command) OR `make/` + a script — pick one,
  prefer a `labctl hosts add|remove` command for consistency with the CLI.
- `bin`/script as needed; `docs/runbooks/00-cross-platform-setup.md`

## Implementation Notes
- Hostnames to manage are derived from `DOMAIN_SUFFIX` and the known ingress hosts
  (apps + grafana, prometheus, argocd, traefik, kubernetes-dashboard). Keep a
  single list; build entries as `127.0.0.1  <host>.<DOMAIN_SUFFIX>`.
- Edit `/etc/hosts` with a clearly marked managed block:
  ```
  # BEGIN sagars-laboratory
  127.0.0.1 go-api.k3d.local grafana.k3d.local ...
  # END sagars-laboratory
  ```
  `add` replaces the block; `remove` deletes it. Idempotent.
- Needs root: detect and re-exec with `sudo`, or print the exact block for the
  user to paste if not root. Works the same on macOS and Linux (`/etc/hosts` path
  is identical).
- Do NOT clobber unrelated lines. Only touch the managed block.

## Acceptance Criteria
- [ ] `labctl hosts add` inserts/updates the managed block; `remove` deletes it.
- [ ] Running `add` twice yields one block, not duplicates.
- [ ] Works on macOS and Linux; clear message if elevation is required.
- [ ] After `add`, `curl http://grafana.k3d.local` resolves to the local cluster.

## Testing Instructions
Run `labctl hosts add`, inspect `/etc/hosts`, curl an ingress host. Then
`labctl hosts remove` and confirm the block is gone. Follow
`docs/runbooks/00-cross-platform-setup.md`.

## Dependencies
None
