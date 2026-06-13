# Task 058: New scenarios — mesh-traffic-management, event-driven-arch, secrets-management

## Phase
M4 — Stack Expansion

## Type
feature

## Priority
P1

## Description
Three new v2 scenarios that exercise the M4 categories, each with stages,
objectives, and checks:

1. **mesh-traffic-management** — canary 90/10 split between two go-api
   versions, latency fault injection at mesh level, mTLS verification.
2. **event-driven-arch** — producer/consumer flow through Kafka; stage 2
   creates consumer lag, objective is to scale consumers (manually or via
   KEDA if installed).
3. **secrets-management** — app consumes a Vault-backed secret via ESO;
   stage 2 rotates it; checks verify propagation without redeploy.

## Files to Modify
- `scenarios/mesh-traffic-management/`, `scenarios/event-driven-arch/`,
  `scenarios/secrets-management/` (scenario.yaml + manifests/values/dashboards)
- `docs/scenarios.md` (document all three)
- `apps/` (only if echo-server needs a tiny consumer mode — prefer
  configuring existing apps over new ones)

## Implementation Notes
- Declare platform prerequisites properly (`mesh`, `data/kafka`,
  `secrets/*`) so preflight validation (task 002 machinery) gates
  activation.
- Each scenario must pass `labctl scenario verify` and be re-activation
  safe (golden rule 5).
- Keep explore blocks rich (urls/commands/tips) — they double as the
  learning content surface.

## Acceptance Criteria
- [ ] All three scenarios: up → verify (pass) → down, cleanly, on k3d
- [ ] Canary split observable in mesh telemetry; lag visible in Grafana;
      secret rotation check passes
- [ ] docs/scenarios.md updated with all three
- [ ] Each scenario has ≥3 checks and ≥2 stages

## Testing Instructions
Per-scenario walkthroughs in `docs/runbooks/10-stack-expansion.md`.

## Dependencies
040, 041, 054, 055, 056 (057 optional for the KEDA variant)
