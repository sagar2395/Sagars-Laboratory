# Runbook 11 — Multi-Env Promotion (Dev → Staging → Prod)

**Scenario:** `env-promotion`  
**Commands:** `labctl env list`, `labctl env promote`  
**Task:** 059

---

## What this simulates

A release pipeline with three environments (dev, staging, prod) each running
`go-api` in a dedicated Kubernetes namespace. Promotion between environments is
done by updating a `env-metadata` ConfigMap — a lightweight stand-in for the
GitOps pattern of committing updated image-tag values to a manifest repository.

**Default mode** (namespace): one cluster, three namespaces (`env-dev`,
`env-staging`, `env-prod`). Laptop-friendly — no extra clusters needed.

---

## Prerequisites

- Cluster running (`labctl runtime up`)
- Ingress installed (`labctl platform up ingress`)
- `go-api` built and available in the cluster's image store:
  ```bash
  cd apps/go-api && make docker-build   # or: docker build -t go-api .
  k3d image import go-api:latest -c sagars-cluster
  ```

---

## Activate the scenario

```bash
labctl scenario up env-promotion
```

This creates three namespaces and deploys go-api in each with staggered
declared tags: dev=v1.2.0, staging=v1.1.0, prod=v1.0.0.

---

## Verify the environments are healthy

```bash
labctl scenario verify env-promotion
# Checks: namespaces exist, deployments ready, env-metadata ConfigMaps present
```

---

## Inspect the release train

```bash
labctl env list
```

Expected output (initial state):

```
ENV        APP          TAG          STATUS     PROMOTED_AT
dev        go-api       v1.2.0       running    never
staging    go-api       v1.1.0       running    never
prod       go-api       v1.0.0       running    never
```

---

## Promote dev → staging

```bash
labctl env promote dev staging
```

Output:

```
Promoting go-api: dev (v1.2.0) → staging (v1.1.0 → v1.2.0)
  dev:      v1.2.0  (unchanged)
  staging:  v1.2.0  (was v1.1.0)

Run 'labctl env list' to confirm...
```

Then:

```bash
labctl env list
# staging should now show v1.2.0
```

---

## Promote staging → prod (full pipeline)

```bash
labctl env promote staging prod
labctl env list
# prod should now show v1.2.0
```

---

## Inspect the promotion record

```bash
kubectl -n env-staging get cm env-metadata -o yaml
# .data.promoted_from = "dev"
# .data.promoted_at   = "<timestamp>"
# .data.image_tag     = "v1.2.0"
```

---

## Re-run promotion (idempotency check)

```bash
labctl env promote dev staging
# Output: "staging is already on v1.2.0 — nothing to promote."
```

---

## Tear down

```bash
labctl scenario down env-promotion
# Deletes namespaces env-dev, env-staging, env-prod and all resources within
```

---

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| `could not read image_tag from env-dev/env-metadata` | Run `labctl scenario up env-promotion` first |
| `unknown environment 'xyz'` | Valid values: `dev`, `staging`, `prod` |
| Deployment stuck in `not-ready` | Check image pull: `kubectl -n env-dev describe pod -l app=go-api` |
| `ERROR: source and destination must be different` | Use two different env names |

---

## How this maps to a real GitOps pipeline

| Namespace-mode step | Real GitOps equivalent |
|---------------------|------------------------|
| `kubectl patch cm env-metadata` | `git commit values/staging.yaml` with new `image.tag` |
| Deployment label update | ArgoCD detects drift and reconciles the Deployment |
| `labctl env list` | Checking Argo CD UI / `argocd app list` |
