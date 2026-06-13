# Task 062: Optional auth + per-user RBAC for API/UI

## Phase
M6 — Team Mode & New Runtimes (see docs/ROADMAP.md Part II)

## Type
feature

## Priority
P2

## Description
Make the labctl server safe to share: optional authentication (off by
default for local use) with two roles — `operator` (full control) and
`participant` (run challenges/incidents/learn, read status; cannot
uninstall platform or reset the lab). Identity flows into result records
(052's `user` field).

## Files to Modify
- `cmd/labctl/internal/server/` (auth middleware, role checks per route)
- `cmd/labctl/` (user management subcommands or static config)
- `ui/src/` (login flow when auth is enabled)
- `docs/cli-reference.md`

## Implementation Notes
- Keep it simple: static users file (`.labctl/users.yaml`, bcrypt hashes)
  + session tokens. No OIDC in v1 — document it as future work. Enabled
  via `LABCTL_AUTH=true`.
- Role checks live at the route layer; default-deny for participant on
  mutating platform/runtime/lab endpoints.
- When auth is off, behavior must be byte-identical to today (local UX
  unchanged).

## Acceptance Criteria
- [ ] Auth off: no behavioral change, all tests pass
- [ ] Auth on: participant can run a challenge but cannot `lab reset`
- [ ] Result records carry the authenticated user
- [ ] Session handling has unit tests; docs updated

## Testing Instructions
Two-browser manual test (operator + participant). Runbook:
`docs/runbooks/12-team-mode.md`.

## Dependencies
052
