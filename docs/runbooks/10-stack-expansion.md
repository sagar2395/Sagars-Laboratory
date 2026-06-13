# Runbook 10 — Stack Expansion (M4)

> Goal: install, swap, and remove the new **swappable platform categories** on a
> single cluster, on identical workloads. This runbook grows one section per M4
> task. Today it covers the **service mesh** category (task 054): istio and
> linkerd, selected by `MESH_PROVIDER`.

## Prereqs

- Runbook 00 completed (tools + cluster up). `openssl` on the host (needed for
  Linkerd's mTLS identity certs — it ships with macOS and every Linux distro).
- A workload to mesh. Runbook 01 deploys **go-api** into the `go-api` namespace,
  which is the default `MESH_NAMESPACE`.
- **k3d memory:** give the cluster at least **4Gi** (6Gi is comfortable with a
  meshed app running). A mesh control plane plus sidecars is not free.

---

## Service Mesh (task 054)

The mesh category lives at `platform/mesh/<provider>/` with the standard
four-file contract (`install.sh`, `uninstall.sh`, `status.sh`, `values.yaml`)
and a `_interface.yaml`. Chart versions are pinned in `versions.env`
(`ISTIO_VERSION`, `LINKERD_CRDS_CHART_VERSION`,
`LINKERD_CONTROL_PLANE_CHART_VERSION`) and overridable per-install.

### Steps

```bash
# 0. Make sure a workload exists in the namespace you'll mesh
bin/labctl init && bin/labctl app deploy go-api      # if not already running
kubectl get pods -n go-api

# 1. Install Istio (sidecar mode) and enrol the go-api namespace
MESH_PROVIDER=istio bin/labctl platform up mesh
#   ...or via make:  make platform-mesh-up MESH_PROVIDER=istio

# 2. Confirm the sidecar was injected (two containers: app + istio-proxy)
kubectl get pod -n go-api \
  -o jsonpath='{range .items[*]}{.metadata.name}{": "}{.spec.containers[*].name}{"\n"}{end}'

# 3. Health + enrolment at a glance
MESH_PROVIDER=istio bin/labctl platform status mesh

# 4. Swap to Linkerd on the SAME cluster (uninstall istio first)
MESH_PROVIDER=istio   bin/labctl platform down mesh
MESH_PROVIDER=linkerd bin/labctl platform up mesh

# 5. Confirm the linkerd-proxy sidecar is present now
kubectl get pod -n go-api \
  -o jsonpath='{range .items[*]}{.metadata.name}{": "}{.spec.containers[*].name}{"\n"}{end}'
MESH_PROVIDER=linkerd bin/labctl platform status mesh

# 6. Re-running install is a safe no-op (idempotency)
MESH_PROVIDER=linkerd bin/labctl platform up mesh    # no errors, no duplicates
```

### Expected

- After step 1, `platform status mesh` shows `istiod` ready, `go-api` listed
  under *Meshed namespaces*, and each go-api pod with an `istio-proxy` sidecar.
- After step 4, the istio control plane and the `istio-injection` label are gone;
  linkerd's control plane is ready and pods carry a `linkerd-proxy` sidecar.
- `platform up mesh` without `MESH_PROVIDER` set prints the available providers
  and the env var to choose one (it does not guess).
- Re-running install changes nothing (idempotent); re-running uninstall on an
  already-removed mesh is a clean no-op.

### Cleanup

```bash
MESH_PROVIDER=linkerd bin/labctl platform down mesh   # or istio, whichever is active
# The go-api namespace keeps running, now mesh-free (sidecars dropped on restart).
```

### Troubleshooting

- **Pods stuck `Init` / no sidecar:** the namespace must be enrolled *before* the
  pod starts. `install.sh` restarts existing deployments for you; if you deployed
  the app afterwards, just `kubectl rollout restart deployment -n go-api`.
- **istiod / linkerd pods `Pending` (Insufficient memory):** raise k3d memory (see
  Prereqs). The `values.yaml` files already request minimal resources.
- **Linkerd install fails on certs:** ensure `openssl` is on PATH. The installer
  generates an ECDSA P-256 trust anchor + issuer; on re-install it reuses the
  certs already stored in the cluster so identity stays stable.
- **Chart version not found:** pin a different release via the matching env var,
  e.g. `ISTIO_VERSION=1.24.1 MESH_PROVIDER=istio bin/labctl platform up mesh`.

---

### Acceptance check (task 054)

- [x] `MESH_PROVIDER=istio labctl platform up mesh` meshes go-api (sidecar visible)
- [x] Swapping to linkerd on the same cluster works after uninstall
- [x] `status.sh` reports mesh health accurately for both providers
- [x] Scripts portable + idempotent; versions pinned in `versions.env`

---

## Data Infrastructure (task 055)

The `data` category holds **additive sub-components** — `data/kafka` (Strimzi)
and `data/postgres` (CloudNativePG) coexist on one cluster. Each owns its own
namespace, so you install/remove them independently. Versions are pinned in
`versions.env` (`STRIMZI_VERSION`, `KAFKA_VERSION`, `CNPG_CHART_VERSION`).

### Kafka (Strimzi) — install + produce/consume smoke test

```bash
# 1. Install the operator + a 1-broker KRaft Kafka cluster (ephemeral storage)
bin/labctl platform up data/kafka
#   ...or:  make platform-data-kafka-up

# 2. Confirm readiness (operator ready + Kafka CR Ready=True)
bin/labctl platform status data/kafka

# 3. Produce a couple of messages with kcat, then consume them back
kubectl -n kafka run kcat-prod --rm -i --image=edenhill/kcat:1.7.1 --restart=Never -- \
  -b lab-kafka-kafka-bootstrap:9092 -t demo -P <<'MSGS'
hello
world
MSGS

kubectl -n kafka run kcat-cons --rm -i --image=edenhill/kcat:1.7.1 --restart=Never -- \
  -b lab-kafka-kafka-bootstrap:9092 -t demo -C -e -o beginning
# => prints: hello / world
```

**Expected:** `status data/kafka` shows `Ready: True`; the consumer prints the
two messages the producer sent. Re-running `platform up data/kafka` is a no-op.

### Postgres (CloudNativePG) — install + failover drill

```bash
# 1. Install the operator + a 2-instance HA Postgres cluster
bin/labctl platform up data/postgres
#   ...or:  make platform-data-postgres-up

# 2. Confirm 2/2 ready and note the current primary + per-pod roles
bin/labctl platform status data/postgres

# 3. Connect with psql (from inside the cluster)
kubectl -n postgres exec -it lab-postgres-1 -- psql -U postgres -c '\l'

# 4. Failover drill: delete the primary, watch a replica get promoted
kubectl -n postgres delete pod \
  "$(kubectl -n postgres get pod -l cnpg.io/instanceRole=primary -o name)"
kubectl -n postgres get pods -l cnpg.io/cluster=lab-postgres -w   # Ctrl-C when settled
bin/labctl platform status data/postgres   # currentPrimary now points at the other pod
```

**Expected:** after step 4 the deleted primary is replaced and a former replica
is promoted (`currentPrimary` changes); the cluster returns to `2/2` ready.

### Cleanup

```bash
bin/labctl platform down data/kafka       # removes operator + CR + PVCs + namespace
bin/labctl platform down data/postgres    # removes operator + Cluster + PVCs + namespaces
```

### Troubleshooting

- **`platform up data` errors "multiple providers":** that's intentional — name
  the sub-component (`data/kafka` or `data/postgres`), or set `DATA_PROVIDER`.
- **Kafka CR stuck not-Ready:** check the operator logs
  (`kubectl logs -n kafka deploy/strimzi-cluster-operator`) and that
  `KAFKA_VERSION` is supported by the installed `STRIMZI_VERSION`.
- **Postgres pods `Pending` (storage/memory):** k3d's `local-path` provisioner
  must be healthy; lower `POSTGRES_INSTANCES` or raise k3d memory if needed.
- **PVCs left behind:** uninstall deletes PVCs labelled with the cluster; if you
  changed `*_CLUSTER`, delete leftovers with the matching label selector.

### Acceptance check (task 055)

- [x] `labctl platform up data/kafka` yields a ready Kafka cluster (kcat smoke test)
- [x] `labctl platform up data/postgres` yields a ready 2-instance cluster; deleting the primary triggers failover
- [x] Uninstall removes operators + CRs + PVCs cleanly
- [x] Versions pinned; scripts portable + idempotent

> Secrets and autoscaling sections (tasks 056–057) and the new scenarios
> (task 058) are appended here as those tasks ship.
