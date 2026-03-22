# Task 010: Add kubernetes-dashboard uninstall.sh

## Priority
P1

## Assigned To
DevOps

## Description
The `platform/dashboard/kubernetes-dashboard/` component has an `install.sh` but no `uninstall.sh`. This breaks the platform interface contract (every component must have `install.sh`, `uninstall.sh`, and `status.sh`) and means the Kubernetes Dashboard cannot be torn down via the standard `labctl platform component destroy` path or the Make `platform-down` target.

## Files to Modify
- `platform/dashboard/kubernetes-dashboard/uninstall.sh` *(create)*

## Implementation Notes
Follow the exact same pattern as the sibling uninstall scripts in the repo (e.g., `platform/monitoring/grafana/uninstall.sh` or `services/redis/uninstall.sh`):
1. `#!/usr/bin/env bash` with `set -euo pipefail`.
2. Source `versions.env` from the repo root if it provides relevant version variables.
3. Determine the release name and namespace used by `install.sh` (check `install.sh` for the values — likely `kubernetes-dashboard` namespace and release name).
4. Run `helm uninstall <release> -n <namespace> --ignore-not-found`.
5. Optionally delete the namespace if it was created exclusively for this component.
6. Print a clear success message.

Do not introduce new tooling or patterns beyond what other uninstall scripts already use.

## Acceptance Criteria
- [ ] `platform/dashboard/kubernetes-dashboard/uninstall.sh` exists and is executable.
- [ ] Running the script against a cluster with the dashboard installed removes the Helm release cleanly.
- [ ] Running the script when the dashboard is not installed exits 0 (idempotent, `--ignore-not-found`).
- [ ] Script follows `set -euo pipefail` and uses the same namespace/release names as `install.sh`.

## Testing Instructions
Install the dashboard via `install.sh`, then run `uninstall.sh` and confirm `kubectl get all -n kubernetes-dashboard` returns no resources. Run `uninstall.sh` a second time and confirm it exits 0.

## Dependencies
None
