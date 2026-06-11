# Task 042: Traffic generator — k6 load profiles as a platform service

## Phase
M1 — Scenario Engine v2

## Type
feature

## Priority
P1

## Description
Simulations are only realistic under load. Add a k6-based traffic generator
that runs in-cluster with selectable profiles (steady, spike, soak) targeting
any app URL, controlled via `labctl traffic start|stop|status`.

## Files to Modify
- `services/traffic/` (new: k6 scripts per profile, install/uninstall/status scripts)
- `cmd/labctl/` (traffic command group wrapping the scripts)
- `docs/cli-reference.md`, `docs/SIMULATOR.md` cross-reference

## Implementation Notes
- Run k6 as a Kubernetes Job (steady/spike) or long-running Deployment
  (soak) in a `traffic` namespace. Profiles are k6 JS files in
  `services/traffic/profiles/` — declarative, not Go.
- Target URL and rate via env (`TRAFFIC_TARGET`, `TRAFFIC_RPS`), defaults
  to `http://go-api.${DOMAIN_SUFFIX:-k3d.local}/health` resolved in-cluster
  via the ingress service.
- k6 emits Prometheus metrics (experimental-prometheus-rw) so load shows up
  in Grafana next to app metrics.
- Idempotent: `traffic start` while running replaces the job.

## Acceptance Criteria
- [ ] `labctl traffic start --profile steady` generates visible RPS in Grafana
- [ ] `spike` and `soak` profiles work; `traffic stop` cleans up fully
- [ ] Scenario components can reference traffic profiles (used by chaos scenario)
- [ ] Scripts are portable (golden rule 1) and idempotent

## Testing Instructions
Manual flow in `docs/runbooks/07-scenario-engine-v2.md`: start traffic,
watch the Grafana RPS panel, inject chaos, observe impact, stop traffic.

## Dependencies
None (pairs well with 040/041)
