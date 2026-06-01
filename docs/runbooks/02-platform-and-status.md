# Runbook 02 — Platform Components & Status

> Goal: install, inspect, and swap platform components, with accurate status and
> idempotent operations. Verifies Phase 2.

## Prereqs

- Runbook 00 completed (cluster up).

## Steps

```bash
# 1. Install the whole platform stack
make platform-up                 # or: bin/labctl platform up
bin/labctl platform status       # per-category health + active provider

# 2. Inspect a single component
bash platform/monitoring/metrics/prometheus/status.sh
bash platform/ingress/traefik/status.sh

# 3. Idempotency check — re-running must be a safe no-op
make platform-up                 # should not error or duplicate resources

# 4. Swap the ingress provider (Traefik -> Nginx)
#    edit .env: INGRESS_PROVIDER=nginx
bin/labctl platform down         # remove current ingress
bin/labctl platform up           # install the newly-selected provider
bin/labctl platform status       # active provider now shows nginx

# 5. Remove a component cleanly (Task 010 adds dashboard uninstall)
bash platform/dashboard/kubernetes-dashboard/uninstall.sh

# 6. Verify ArgoCD ingress uses DOMAIN_SUFFIX (Task 025)
bash platform/gitops/argocd/install.sh          # default: argocd.k3d.local
kubectl get ingress -n argocd -o jsonpath='{.items[0].spec.rules[0].host}'
# => argocd.k3d.local

DOMAIN_SUFFIX=test.example.com bash platform/gitops/argocd/install.sh
kubectl get ingress -n argocd -o jsonpath='{.items[0].spec.rules[0].host}'
# => argocd.test.example.com
```

## Expected

- `platform status` reports each category (ingress, metrics, logging, tracing,
  gitops, security, chaos) with the **active provider** and green/red health,
  derived from JSON not text scraping (Task 004, 020).
- Re-running `platform up` changes nothing (idempotent).
- Switching `INGRESS_PROVIDER` and re-running swaps the implementation; ingress
  hosts still resolve.
- Async actions return a **job id** (Task 017) and structured errors (Task 022).
- `make platform-*` targets respect the selected provider (Task 005).

## Verifying provider swappability

```bash
# Each provider directory must have all four scripts:
ls platform/ingress/nginx/        # install.sh uninstall.sh status.sh values.yaml
ls platform/ingress/traefik/
cat platform/ingress/_interface.yaml   # the contract both implement
```

## Cleanup

```bash
make platform-down               # remove all platform components
```

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| Grafana/Prometheus 404 | Ingress host vs `DOMAIN_SUFFIX` mismatch; confirm `/etc/hosts`. |
| "another operation in progress" on install | Stuck Helm release; install scripts clean `pending-` state — re-run. |
| Traefik reappears after delete on k3d | k3d bundles Traefik; cluster is created with `--disable=traefik`. Recreate cluster if it predates that. |
| Status shows installed but pods are down | Component health check (Task 004) — inspect with the component `status.sh`. |
