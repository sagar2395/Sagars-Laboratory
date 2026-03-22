# Task 025: Fix ArgoCD Values Hardcoded k3d.local Domain

## Priority
P1

## Assigned To
DevOps

## Description
`platform/gitops/argocd/values.yaml` hardcodes `argocd.k3d.local` as the ingress hostname. When the ArgoCD component is installed on AKS or EKS (where `DOMAIN_SUFFIX` is `sagarslab.io`), the hostname remains `argocd.k3d.local` and the ingress never resolves. The hostname must be templated from `DOMAIN_SUFFIX`, following the same pattern used by Prometheus, Grafana, and Traefik.

## Files to Modify
- `platform/gitops/argocd/install.sh`
- `platform/gitops/argocd/values.yaml`

## Implementation Notes
The pattern used by other components (e.g., `platform/monitoring/prometheus/install.sh`) is:
1. Set a placeholder in `values.yaml`: Replace the hardcoded `argocd.k3d.local` with a clearly named placeholder such as `argocd.DOMAIN_SUFFIX_PLACEHOLDER`.
2. In `install.sh`, before calling `helm upgrade --install`, substitute the placeholder with the actual `DOMAIN_SUFFIX` environment variable using `sed`:
   ```bash
   TEMP_VALUES=$(mktemp)
   sed "s/DOMAIN_SUFFIX_PLACEHOLDER/${DOMAIN_SUFFIX}/g" "${SCRIPT_DIR}/values.yaml" > "${TEMP_VALUES}"
   helm upgrade --install argocd argo/argo-cd \
     --namespace argocd --create-namespace \
     -f "${TEMP_VALUES}" \
     --wait --timeout 5m
   rm -f "${TEMP_VALUES}"
   ```
3. Alternatively, pass `--set server.ingress.hostname=argocd.${DOMAIN_SUFFIX}` directly on the `helm` command line if the chart supports per-value overrides at that path — check the Argo CD chart's values structure first and use whichever approach is less fragile.

Do not break existing k3d setups where `DOMAIN_SUFFIX=k3d.local` — the substituted value will be `argocd.k3d.local` which is identical to the current hardcoded value.

## Acceptance Criteria
- [ ] After `platform/gitops/argocd/install.sh` completes on k3d, the ArgoCD ingress hostname is `argocd.k3d.local`.
- [ ] After the same script runs with `DOMAIN_SUFFIX=sagarslab.io`, the ArgoCD ingress hostname is `argocd.sagarslab.io`.
- [ ] `values.yaml` no longer contains any hardcoded domain strings.
- [ ] The install script is idempotent (running it twice produces the same result).

## Testing Instructions
On k3d, run `platform/gitops/argocd/install.sh` and confirm `kubectl get ingress -n argocd` shows host `argocd.k3d.local`. Override `DOMAIN_SUFFIX=test.example.com` and run again — confirm the ingress hostname changes to `argocd.test.example.com`.

## Dependencies
None
