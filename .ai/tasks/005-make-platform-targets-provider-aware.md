# Task 005: Make Platform Targets Provider-Aware

## Priority
P1

## Assigned To
DevOps

## Description
Update the Make-based platform lifecycle so it respects provider selection from `.env` instead of hard-coding Traefik, Prometheus, and Grafana. The current Make targets diverge from both the documented provider model and the CLI orchestration.

## Files to Modify
- `Makefile`
- `make/platform.mk`
- `.env.example`
- `docs/ci-cd.md`

## Implementation Notes
Preserve the simple `make platform-up`, `make platform-down`, and `make platform-status` interface, but derive the invoked scripts from configured provider variables. Support at least ingress and metrics providers, and add optional handling for logging and tracing when configured. If some categories are intentionally not included in the default platform lifecycle, document that boundary clearly in the Make help text and docs.

## Acceptance Criteria
- [ ] `make platform-up` and `make platform-down` use configured provider variables instead of hard-coded provider paths.
- [ ] `make platform-status` reports status for the same categories handled by the lifecycle targets.
- [ ] Help text and `.env.example` accurately describe which provider variables affect Make-based platform orchestration.
- [ ] The updated targets work without changing the existing command surface.

## Testing Instructions
Run `make platform-status`, `make -n platform-up`, and `make -n platform-down` with different provider overrides such as `INGRESS_PROVIDER=nginx` and `METRICS_PROVIDER=prometheus`.

## Dependencies
None
