# Task 009: Add Helm Pre-Deploy Lint and Post-Deploy Readiness Wait

## Priority
P1

## Assigned To
DevOps

## Description
The engine's Helm deploy strategy (`engine/deploy/helm.sh`) applies charts directly without first linting templates or waiting for rollout completion. A malformed values file or chart error results in a broken release with no clear signal to the caller. Additionally, after `helm upgrade`, the script does not wait for all workloads to become ready, so downstream steps may proceed against a partially deployed application.

## Files to Modify
- `engine/deploy/helm.sh`

## Implementation Notes
1. **Pre-deploy lint**: Before `helm upgrade --install`, run `helm lint` against the chart path and values file. If lint fails, print the lint output and exit non-zero without attempting the upgrade.
2. **Pre-deploy dry-run**: After lint, run `helm upgrade --install --dry-run` to catch rendering errors. Exit non-zero on failure.
3. **Post-deploy readiness wait**: After a successful upgrade, run `kubectl rollout status deployment/<RELEASE_NAME> -n <NAMESPACE> --timeout=<HELM_WAIT_TIMEOUT>` where `HELM_WAIT_TIMEOUT` defaults to `5m` but can be overridden in `app.env`.
4. Keep the existing function/argument signatures intact. Do not add new required env vars — only optional ones with documented defaults.

All new steps must produce clear, identifiable output prefixes (`[lint]`, `[dry-run]`, `[rollout]`) for log readability.

## Acceptance Criteria
- [ ] A chart with an invalid template fails the deploy before any `helm upgrade` is attempted.
- [ ] A successfully deployed chart waits for rollout completion before the script exits 0.
- [ ] `HELM_WAIT_TIMEOUT` in `app.env` overrides the default 5-minute rollout wait.
- [ ] Output from each step is prefixed and clearly distinguishable in logs.
- [ ] No existing env var names or function signatures change.

## Testing Instructions
Run `make app-deploy APP=go-api` against a running k3d cluster. Introduce a deliberate template error in the chart values and confirm the script exits non-zero before upgrading. Confirm the rollout wait fires after a successful deploy.

## Dependencies
None
