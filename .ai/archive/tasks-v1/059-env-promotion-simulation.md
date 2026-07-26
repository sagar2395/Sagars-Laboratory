# Task 059: Multi-cluster env promotion — dev → staging → prod via GitOps

## Phase
M5 — Multi-Env & Day-2 Ops (see docs/ROADMAP.md Part II)

## Type
feature

## Priority
P2

## Description
Simulate real release engineering locally: three k3d clusters (or, in a
constrained mode, three namespaces in one cluster) representing dev,
staging, prod; an image is promoted between them purely via Git commits
that ArgoCD reconciles. Shipped as the `env-promotion` v2 scenario plus
`labctl env` helpers (`env list`, `env promote <app> <from> <to>`).

## Files to Modify
- `runtimes/k3d/` (multi-cluster profile support, e.g. `K3D_CLUSTERS=dev,staging,prod`)
- `scenarios/env-promotion/` (scenario + ArgoCD app-of-apps per env)
- `cmd/labctl/` (env command group wrapping promotion script)
- `docs/cli-reference.md`

## Implementation Notes
- Promotion = a script that updates the image tag in the env's values file
  in the local Git repo serving ArgoCD (reuse the gitops scenario's repo
  mechanism) and commits — labctl never talks to ArgoCD's API directly.
- Constrained namespace mode keeps laptop usage viable; document the
  trade-off. Checks verify each env runs the expected image tag.
- Mind memory: three k3d clusters is heavy — default to namespace mode,
  cluster mode behind a flag.

## Acceptance Criteria
- [ ] `env promote go-api dev staging` results in staging running dev's tag, via Git only
- [ ] `scenario verify env-promotion` checks each env's running tag
- [ ] Both namespace and multi-cluster modes documented; namespace mode is default
- [ ] Teardown removes all envs cleanly

## Testing Instructions
Full promotion walkthrough in `docs/runbooks/11-multi-env-day2.md`.

## Dependencies
040, 041 (gitops platform category already exists)
