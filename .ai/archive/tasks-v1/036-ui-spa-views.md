# Task 036: Build the Four SPA Views

## Phase
3

## Type
feature

## Priority
P1

## Description
Implement the four dashboard views on top of the scaffold from task 035, each
backed by the existing API and updated live over WebSocket.

## Files to Modify
- `ui/src/` (views/components)
- `docs/runbooks/03-web-ui.md`

## Implementation Notes
Views (mirror the original dashboard's intent, but as real components):
1. **Dashboard** — cluster info (runtime, nodes, k8s version, context) + platform
   component health grid (per-category, active provider, green/yellow/red) + app
   list with pod counts and endpoints.
2. **Scenarios** — list with active/inactive state + category badges; detail panel
   (description, prerequisites, components, explore tips); Start/Stop buttons that
   stream progress live.
3. **Platform** — providers per category, active highlighted, "Switch provider"
   (down+up) action, per-component status + logs link.
4. **Apps** — deploy/destroy actions, pod status, recent logs / Grafana deep-link.
- All long-running actions (scenario up/down, platform install) must reflect the
  streamed `ActionEvent`s from the WebSocket (depends on task 003 having streamed
  events for all operations).
- Handle the job-id pattern from task 017 if implemented (poll/subscribe by job id).

## Acceptance Criteria
- [ ] All four views render real data from the API.
- [ ] Activating a scenario from the UI streams live progress, then shows it active.
- [ ] Switching an ingress provider from the UI works end-to-end.
- [ ] Deploying/destroying an app from the UI works and reflects pod status.

## Testing Instructions
With a live cluster and `bin/labctl ui` running, exercise each view: deploy an
app, activate `observability-sre`, switch ingress provider. Follow
`docs/runbooks/03-web-ui.md`.

## Dependencies
035; benefits from 003 and 017
