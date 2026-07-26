# Task 056: `platform/secrets` category — vault + external-secrets

## Phase
M4 — Stack Expansion

## Type
feature

## Priority
P2

## Description
Add a secrets-management category: `vault` (HashiCorp Vault in dev mode for
the lab, with a documented non-dev values profile) and `external-secrets`
(External Secrets Operator wired to Vault as its backend). Selected by
`SECRETS_PROVIDER`. Unlocks the secrets-management scenario (058) and a
future leaked-secret incident.

## Files to Modify
- `platform/secrets/vault/`, `platform/secrets/external-secrets/`
- registry wiring + `make/` targets
- `platform/README.md`, `versions.env`

## Implementation Notes
- Vault dev mode is fine for simulation (root token via env, NOT
  committed); install.sh seeds a demo KV secret for go-api.
- ESO provider depends on Vault being present — declare that as a
  prerequisite in its install.sh preflight rather than auto-installing.
- Demonstrate the full sync: Vault KV → ExternalSecret → k8s Secret →
  go-api env var.
- No secrets in git, ever. Document rotation as the exercise.

## Acceptance Criteria
- [ ] Vault installs, demo secret seeded, UI reachable via ingress
- [ ] ESO syncs the demo secret into the go-api namespace
- [ ] Rotating the value in Vault propagates within the ESO refresh interval
- [ ] Uninstall is clean; scripts portable + idempotent; versions pinned

## Testing Instructions
Rotation walkthrough in `docs/runbooks/10-stack-expansion.md`.

## Dependencies
None (058's secrets scenario depends on this)
