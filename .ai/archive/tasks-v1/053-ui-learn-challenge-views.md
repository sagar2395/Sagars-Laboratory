# Task 053: Web UI — Learn, Challenges, Leaderboard views

## Phase
M3 — Learning & Assessment

## Type
feature

## Priority
P2

## Description
Extend the SPA with three views: **Learn** (paths, module progress, start/
next actions), **Challenges** (list, live timer, hint button with penalty
warning, submit), **Results/Leaderboard** (history table from the results
API, per-user once team mode lands). Incident controls (inject/status/hint)
join the existing dashboard.

## Files to Modify
- `ui/src/` (new views + API client methods)
- `cmd/labctl/` (only if new WS event types are needed)
- `docs/runbooks/03-web-ui.md` (extend) and 09 runbook

## Implementation Notes
- Follow the SPA patterns from tasks 035–037 (same framework, same WS
  live-update mechanism). Challenge timer state comes from the server, not
  the browser, so refreshes don't reset it.
- Hint reveal in the UI must call the same REST endpoint as the CLI so the
  penalty is recorded identically.
- Production build embeds via the existing `make cli-build` flow (037) —
  no new build machinery.

## Acceptance Criteria
- [ ] All three views render real data from the results/learn/challenge APIs
- [ ] A full challenge can be driven entirely from the UI
- [ ] Live updates (timer, check results) stream over WebSocket
- [ ] Embedded build still produces a single binary

## Testing Instructions
`make cli-build && bin/labctl ui`; drive one challenge from the browser.
Runbook: `docs/runbooks/09-learning-and-challenges.md`.

## Dependencies
050, 051, 052
