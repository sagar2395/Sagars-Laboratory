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

---

# Day-2 Operations Drills (task 060)

Three scenarios rehearse the operations on-call engineers fear, each with checks
that **grade availability** through the operation rather than guessing at it.

| Drill | Scenario | Grades |
|-------|----------|--------|
| Node drain under load | `node-drain-drill` | success rate ≥ 99.5%, PDB held, no cordon left |
| Rolling cluster upgrade | `cluster-upgrade-drill` | success rate ≥ 99%, all nodes Ready |
| Namespace backup & restore | `backup-restore-drill` | namespace round-trips, marker intact |

**Shared prerequisites (drains/upgrades):**

```bash
labctl runtime up                      # k3d defaults to 2 agents (multi-node)
labctl platform up ingress
labctl platform up monitoring/metrics  # promql availability checks need Prometheus
cd apps/go-api && make docker-build && k3d image import go-api:latest -c sagars-cluster && cd ../..
```

The availability checks read `http_requests_total` from Prometheus, so **start
traffic before** running a drill and keep it running in another terminal:

```bash
labctl traffic start --profile steady --rps 20
```

---

## Drill 1 — Node drain under load (`node-drain-drill`)

```bash
labctl scenario up node-drain-drill          # applies the PDB
bash scenarios/node-drain-drill/scripts/drain.sh   # cordon → drain → uncordon
labctl scenario verify node-drain-drill      # grades success rate + PDB + no cordon
labctl scenario down node-drain-drill
```

`drain.sh` scales go-api to 3, picks a worker node, cordons and drains it
(respecting the PDB), waits for pods to reschedule, then **always uncordons** on
exit. Watch pods move with `kubectl -n go-api get pods -o wide -w`.

**Expected:** the drain blocks briefly while a replacement go-api pod becomes
Ready on another node; the `availability-held-during-drain` check passes.

---

## Drill 2 — Rolling cluster upgrade (`cluster-upgrade-drill`)

```bash
# Optionally pin an older version at cluster creation to upgrade FROM:
#   K3S_VERSION=v1.28.8-k3s1 AGENTS=2 labctl runtime up
kubectl get nodes                            # note the current version
labctl scenario up cluster-upgrade-drill     # applies the PDB
TARGET_K3S_VERSION=v1.29.4-k3s1 \
  bash scenarios/cluster-upgrade-drill/scripts/upgrade.sh
labctl scenario verify cluster-upgrade-drill # grades success rate + node readiness
labctl scenario down cluster-upgrade-drill
```

`upgrade.sh` rolls each agent node one at a time: drain → `k3d node delete` →
`k3d node create` on the target image → wait for go-api. The PDB keeps the app
available across the roll.

**Honest scope:** k3d can't upgrade a node in place, so this replaces worker
nodes — a faithful rolling worker upgrade. The control-plane node is untouched
(managed clusters upgrade it first).

---

## Drill 3 — Namespace backup & restore (`backup-restore-drill`)

Requires `jq` (used to scrub server-managed fields from the archive).

```bash
labctl scenario up backup-restore-drill              # plants the restore marker
bash scenarios/backup-restore-drill/scripts/backup.sh go-api
kubectl -n go-api delete configmap restore-marker    # simulate accidental loss
labctl scenario verify backup-restore-drill          # marker check now FAILS
bash scenarios/backup-restore-drill/scripts/restore.sh go-api
labctl scenario verify backup-restore-drill          # round-trip — checks PASS
labctl scenario down backup-restore-drill
```

Archives land in `.labctl/backups/` (gitignored); `go-api-latest.json` points at
the most recent. Harder variant: `kubectl delete namespace go-api` then restore —
the archive recreates the namespace and its objects.

**Manifest-level backup:** round-trips Kubernetes objects, **not** PersistentVolume
data. For stateful data use a volume snapshot or Velero with restic.

---

## Day-2 drills — troubleshooting

| Symptom | Fix |
|---------|-----|
| Drain hangs forever | A pod has no PDB headroom or no other node can host it. `kubectl describe` the pending pod; ensure ≥2 nodes (`kubectl get nodes`). |
| `only 1 node(s) in the cluster` | Recreate multi-node: `AGENTS=2 labctl runtime up`. |
| Availability check fails with no traffic | Start `labctl traffic start` first — `rate(http_requests_total[…])` is 0/0 without requests. |
| `TARGET_K3S_VERSION is required` | Pass a version newer than the current one (see `kubectl get nodes`). |
| `'jq' is required` (backup) | `brew install jq` / `apt-get install jq`. |
| Node stuck `SchedulingDisabled` | `kubectl uncordon <node>` — the drill scripts do this automatically on exit. |
