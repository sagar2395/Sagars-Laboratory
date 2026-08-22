# Task 049: On-call drill — alerts to a webhook receiver

## Phase
M2 — Incident Engine

## Type
feature

## Priority
P2

## Description
Close the loop from fault to page: faults that declare an `alert`
expectation in `fault.yaml` should fire a real Alertmanager alert, routed to
a configurable webhook receiver (Slack-compatible payload). Ship a tiny
in-cluster "pager" receiver that records received alerts so drills work
with zero external accounts, and document pointing it at real Slack/Teams.

## Files to Modify
- `platform/monitoring/` (Alertmanager route/receiver config, env-driven)
- `services/pager/` (new: minimal receiver deployment + status script)
- `incidents/*/fault.yaml` (alert expectations for applicable faults)
- `docs/cli-reference.md` / runbook

## Implementation Notes
- Webhook URL via `ALERT_WEBHOOK_URL` env (golden rule 3); default routes
  to the in-cluster pager service.
- `labctl incident status` should also report whether the expected alert
  fired (query Alertmanager API) — measures time-to-detect for real.
- Keep PrometheusRule additions inside the relevant fault dirs or the
  observability scenario — not hardcoded into the monitoring provider.

## Acceptance Criteria
- [ ] Injecting `oom-kill` fires an alert visible in the pager receiver within its `for:` window
- [ ] `ALERT_WEBHOOK_URL` override routes alerts externally (documented, manually verified once)
- [ ] `incident status` shows alert-fired state and timestamp
- [ ] All config is env-driven and idempotent

## Testing Instructions
Full drill in `docs/runbooks/08-incident-engine.md`: inject → wait for page
→ triage via Grafana/Loki → fix → verify resolution + alert cleared.

## Dependencies
045, 046, 048
