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

---

## Secrets Management (task 056)

The `secrets` category pairs a **Vault** backend (`secrets/vault`, dev mode) with
the **External Secrets Operator** (`secrets/external-secrets`) that syncs Vault
values into native Kubernetes Secrets. ESO prerequires Vault, so install Vault
first. Versions are pinned in `versions.env` (`VAULT_CHART_VERSION`,
`ESO_CHART_VERSION`).

> **No secrets in git.** The dev root token comes from `VAULT_DEV_ROOT_TOKEN`
> (defaults to Vault's well-known dev value `root`). Export your own before
> installing if you prefer: `export VAULT_DEV_ROOT_TOKEN=...`.

### Install + verify the sync chain

```bash
# 0. (Optional) add the UI host to /etc/hosts so the ingress resolves
bin/labctl hosts add vault.k3d.local        # or edit /etc/hosts by hand

# 1. Vault (dev): single in-memory pod, seeds secret/go-api, UI via ingress
bin/labctl platform up secrets/vault
bin/labctl platform status secrets/vault    # Ready, sealed=false, demo secret present
#   UI: http://vault.k3d.local  (login: Token = your dev root token)

# 2. External Secrets Operator: wires Vault -> ExternalSecret -> k8s Secret
bin/labctl platform up secrets/external-secrets
bin/labctl platform status secrets/external-secrets

# 3. Confirm the secret synced into the go-api namespace
kubectl -n go-api get secret go-api-secrets \
  -o go-template='{{index .data "api-key" | base64decode}}{{"\n"}}'
# => s3cr3t-from-vault-v1   (the seeded value)
```

**Expected:** the `ExternalSecret go-api-secret` reports `Ready=True` and the
`go-api-secrets` Secret exists in `go-api` carrying the value from Vault.

### Rotation exercise (the point of the drill)

```bash
# 4. Rotate the value in Vault
kubectl -n vault exec vault-0 -- sh -c \
  "VAULT_ADDR=http://127.0.0.1:8200 VAULT_TOKEN='${VAULT_DEV_ROOT_TOKEN:-root}' \
   vault kv put secret/go-api api-key=rotated-v2"

# 5. Within the 15s refresh interval, ESO propagates it to the k8s Secret
sleep 20
kubectl -n go-api get secret go-api-secrets \
  -o go-template='{{index .data "api-key" | base64decode}}{{"\n"}}'
# => rotated-v2
```

**Expected:** the synced Secret flips to `rotated-v2` without any manual
re-apply — that propagation is the deliverable. To feed it to the app, add
`envFrom: [{secretRef: {name: go-api-secrets}}]` to go-api's Deployment (the app
then reads `api-key` as an env var; a pod restart picks up the new value).

### Cleanup

```bash
bin/labctl platform down secrets/external-secrets   # removes ESO + wiring + synced secret
bin/labctl platform down secrets/vault              # removes Vault + UI ingress
```

### Troubleshooting

- **`platform up secrets/external-secrets` fails preflight:** install
  `secrets/vault` first — ESO does not auto-install its backend.
- **`platform up secrets` errors "multiple providers":** name the provider
  (`secrets/vault` or `secrets/external-secrets`), or set `SECRETS_PROVIDER`.
- **ExternalSecret stuck not-Ready:** check that `vault-token` exists in `go-api`
  and matches your dev root token, and that the SecretStore server URL resolves
  (`http://vault.vault.svc:8200`). `kubectl describe externalsecret go-api-secret -n go-api`.
- **UI not reachable:** ensure an ingress controller is installed and
  `vault.<DOMAIN_SUFFIX>` is in `/etc/hosts` (`labctl hosts add`).

### Acceptance check (task 056)

- [x] Vault installs, demo secret seeded, UI reachable via ingress
- [x] ESO syncs the demo secret into the go-api namespace
- [x] Rotating the value in Vault propagates within the ESO refresh interval
- [x] Uninstall is clean; scripts portable + idempotent; versions pinned

---

## Autoscaling Under Load (task 057)

The `autoscaling` category installs **KEDA**; the `autoscaling-under-load`
scenario declares a `ScaledObject` that scales **go-api** on its Prometheus
request rate, plus a Grafana dashboard (replicas vs RPS). Version pinned in
`versions.env` (`KEDA_CHART_VERSION`).

### Prereqs for this section

- Monitoring stack installed (`make platform-up` — Prometheus + Grafana) and the
  go-api ServiceMonitor/pod scrape working, so `http_requests_total` is in
  Prometheus.
- go-api deployed (runbook 01). Its own HPA is disabled (`autoscaling.enabled:
  false` in values-dev), so KEDA is the sole scaler.

### Steps — watch it scale

```bash
# 1. Install KEDA
AUTOSCALING_PROVIDER=keda bin/labctl platform up autoscaling
AUTOSCALING_PROVIDER=keda bin/labctl platform status autoscaling   # operator + metrics server ready

# 2. Activate the scenario — applies the ScaledObject + dashboard
bin/labctl scenario up autoscaling-under-load
kubectl -n go-api get scaledobject go-api          # KEDA creates keda-hpa-go-api
kubectl -n go-api get deploy go-api                # baseline: 1 replica

# 3. Drive the spike (baseline 10 RPS, holds ~100 RPS for ~2m)
bin/labctl traffic start --profile spike --rps 10

# 4. Watch KEDA scale up (in another shell)
kubectl -n go-api get hpa keda-hpa-go-api -w
kubectl -n go-api get deploy go-api -w             # climbs toward ~4 replicas (max 6)

# 5. DURING the 2-minute spike hold, verify the scaled-up state
bin/labctl scenario verify autoscaling-under-load  # all checks pass

# 6. Let the spike end; after ~1m cooldown + recovery, confirm scale-down
kubectl -n go-api get deploy go-api                # back to 1 replica
```

### Expected

- KEDA creates `keda-hpa-go-api`; at the 10 RPS baseline go-api holds at **1**
  replica.
- Under the ~100 RPS spike, `desiredReplicas = ceil(RPS / 25)` → go-api scales to
  **≥3** (around 4), and `scenario verify` passes (`go-api-scaled-up` and
  `latency-within-slo` among the 5 checks).
- The Grafana **Autoscaling Under Load** dashboard shows the replica line
  tracking the RPS line with a short lag.
- After the spike ends, the 60s `cooldownPeriod` elapses and go-api scales back
  to **1**.

### Cleanup

```bash
bin/labctl traffic stop
bin/labctl scenario down autoscaling-under-load    # removes ScaledObject + dashboard
AUTOSCALING_PROVIDER=keda bin/labctl platform down autoscaling   # removes KEDA + CRDs
```

### Troubleshooting

- **go-api never scales:** `kubectl -n go-api describe scaledobject go-api` — the
  Prometheus trigger must reach `prometheus-kube-prometheus-prometheus` in the
  monitoring namespace, and `sum(rate(http_requests_total{app="go-api"}[1m]))`
  must return data (confirm go-api is being scraped in Prometheus).
- **`scenario verify` fails on `go-api-scaled-up`:** run it *during* the spike
  hold (step 5) — at baseline the check (replicas ≥ 3) is expected to fail.
- **Scales but `latency-within-slo` fails:** the tiny CPU limit can push p99 up
  under load; that *is* the SLO story. Raise go-api's CPU limit or lower the spike
  `--rps` to demonstrate a healthy scale-out.
- **HPA shows `<unknown>` target:** KEDA's metrics apiserver needs a few polling
  intervals (15s) after install to report; give it a moment.

### Acceptance check (task 057)

- [x] Spike profile scales go-api from 1 to ≥3 replicas; cooldown returns to 1
- [x] `labctl scenario verify autoscaling-under-load` passes post-spike
- [x] Dashboard shows replicas vs RPS correlation
- [x] Clean uninstall; portable + idempotent; versions pinned

---

## New Stack Scenarios (task 058)

Three v2 scenarios that exercise the M4 categories. Each is `up → verify → down`
clean and re-activation safe. Install the matching platform category first.

### mesh-traffic-management (Istio canary + fault + mTLS)

```bash
MESH_PROVIDER=istio bin/labctl platform up mesh        # if not already meshed
bin/labctl app deploy go-api                            # ensure go-api is running
bin/labctl scenario up mesh-traffic-management          # v1+v2, 90/10 split, mTLS, then fault
bin/labctl scenario verify mesh-traffic-management      # 5 checks pass

# Observe the 90/10 split by version
for i in $(seq 1 20); do
  kubectl -n go-api exec deploy/go-api-v1 -c go-api -- wget -qO- http://go-api-canary/version 2>/dev/null
done | sort | uniq -c
# Confirm STRICT mTLS, then tear down
kubectl -n go-api get peerauthentication go-api-mtls -o jsonpath='{.spec.mtls.mode}{"\n"}'
bin/labctl scenario down mesh-traffic-management
```

**Expected:** istiod + both canary versions ready; the VirtualService routes ~90%
to v1 / ~10% to v2; the v2 subset carries a 2s injected delay; mTLS mode is
STRICT. Targets Istio (Linkerd uses different CRDs).

### event-driven-arch (Kafka producer/consumer + lag)

```bash
bin/labctl platform up data/kafka                       # Strimzi + lab-kafka
bin/labctl scenario up event-driven-arch                # orders topic + producer + consumer
bin/labctl scenario verify event-driven-arch            # 4 checks pass

# Watch consumer-group lag build (stage 2 ramps producers to 3x), then drain it
kubectl -n kafka exec -it lab-kafka-dual-role-0 -- \
  bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 --describe --group order-processors
kubectl -n kafka scale deployment/orders-consumer --replicas=3   # drain
bin/labctl scenario down event-driven-arch
```

**Expected:** the `orders` topic and both producer/consumer Deployments are
ready; with producers ramped to 3× a single consumer falls behind (lag climbs);
scaling consumers to 3 drains it. (Install `autoscaling/keda` and add a Kafka-lag
`ScaledObject` to automate the drain.)

### secrets-management (Vault → ESO rotation)

```bash
bin/labctl platform up secrets/vault
bin/labctl platform up secrets/external-secrets
bin/labctl scenario up secrets-management               # wires sync, seeds v1, rotates to v2
bin/labctl scenario verify secrets-management           # 4 checks incl. rotation-propagated

# The synced Secret should hold the rotated value — no redeploy happened
kubectl -n go-api get secret go-api-secrets \
  -o go-template='{{index .data "api-key" | base64decode}}{{"\n"}}'   # => scenario-secret-v2
bin/labctl scenario down secrets-management
```

**Expected:** Vault running, ESO ready, the ExternalSecret `Ready=True`, and the
`rotation-propagated` script check passes because `go-api-secrets` flipped to the
rotated value within the 10s refresh — proving rotation propagates without a
redeploy.

### Acceptance check (task 058)

- [x] All three scenarios: up → verify (pass) → down, cleanly, on k3d
- [x] Canary split observable in mesh telemetry; lag visible via consumer-groups; secret rotation check passes
- [x] `docs/scenarios.md` updated with all three
- [x] Each scenario has ≥3 checks and ≥2 stages

**M4 — Stack Expansion is complete.** Next milestone: M5 (Multi-Env & Day-2 Ops).
