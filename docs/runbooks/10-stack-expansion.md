# Runbook 10 — Stack Expansion (M4)

Covers milestone **M4** (tasks 054–058): additional platform categories and
new scenarios that use them.

---

## Part A — Service Mesh (task 054)

Adds `platform/mesh/` with two swappable providers: **istio** and **linkerd**.

### Prerequisites

- A healthy cluster with ingress running (runbooks 00–01).
- Helm 3.x on PATH.
- `openssl` on PATH (built-in on macOS and most Linux distros).
- Minimum memory: Istio needs ~1 GB free in the cluster; Linkerd ~512 MB.

### 1. Install Istio

```bash
export MESH_PROVIDER=istio
make platform-mesh-up
# or
MESH_PROVIDER=istio labctl platform up mesh
```

**Expected:**
- `istio-system` namespace created.
- `istio-base` and `istiod` Helm releases installed.
- `go-api` namespace labelled `istio-injection=enabled`.
- Pods in `go-api` restarted with `istio-proxy` sidecar container.

Verify:

```bash
kubectl get pods -n istio-system
kubectl get pods -n go-api  # each pod should show 2/2 Ready
make platform-mesh-status MESH_PROVIDER=istio
```

### 2. Swap to Linkerd (provider swap test)

```bash
export MESH_PROVIDER=istio
make platform-mesh-down     # removes label + rolls pods, then removes Helm releases

export MESH_PROVIDER=linkerd
make platform-mesh-up
```

**Expected:**
- `go-api` namespace re-labelled with `linkerd.io/inject=enabled`.
- `linkerd` namespace created with control plane pods.
- Pods in `go-api` restarted with `linkerd-proxy` sidecar.

Verify:

```bash
kubectl get pods -n linkerd
kubectl get pods -n go-api   # each pod should show 2/2 Ready (app + proxy)
make platform-mesh-status MESH_PROVIDER=linkerd
```

Optional (if Linkerd CLI installed):

```bash
linkerd check
linkerd viz install | kubectl apply -f -  # optional dashboard
```

### 3. Uninstall

```bash
make platform-mesh-down MESH_PROVIDER=linkerd
# Namespace go-api annotation removed, pods rolled back to 1/1
```

### 4. `labctl platform` integration

The registry auto-discovers the new providers (scans for `install.sh`):

```bash
labctl platform status   # shows mesh category if MESH_PROVIDER set
labctl platform up       # installs mesh if MESH_PROVIDER is set in env
labctl platform down     # removes mesh if MESH_PROVIDER is set in env
```

### 5. Custom app namespace

By default the mesh is applied to the `go-api` namespace. Override:

```bash
MESH_APP_NAMESPACE=my-app MESH_PROVIDER=istio make platform-mesh-up
```

### Troubleshooting (Mesh)

- **Pods stuck at 0/2** — the sidecar injector webhook may not be ready. Wait
  30 s and re-run `kubectl rollout restart deployment -n go-api`.
- **Linkerd cert errors** — re-run `make platform-mesh-down MESH_PROVIDER=linkerd`
  and `make platform-mesh-up MESH_PROVIDER=linkerd`. Fresh certs are generated
  each install.
- **k3d resource limits** — if nodes OOM, reduce limits in `values.yaml`.
  For Istio, set `pilot.resources.requests.memory: 128Mi`.

---

## Part B — Data Infrastructure (task 055)

Adds `platform/data/kafka/` and `platform/data/postgres/`. Both are independent
and additive (like monitoring/metrics + monitoring/grafana).

### Prerequisites

- A healthy cluster (runbooks 00–01). 4 GB+ free RAM recommended when running
  both data providers alongside monitoring.

### 6. Install Kafka

```bash
DATA_KAFKA=1 make platform-data-kafka-up
# or directly:
bash platform/data/kafka/install.sh
```

**Expected:**
- `kafka` namespace created.
- Single-pod Kafka cluster (KRaft mode) in `Running` state within ~3 min.

Verify:

```bash
make platform-data-kafka-status DATA_KAFKA=1
kubectl get pods -n kafka
```

Smoke test (produce + consume):

```bash
# Produce a message
kubectl run kafka-smoke -it --rm --restart=Never \
  --image=bitnami/kafka:latest -- \
  kafka-console-producer.sh \
    --bootstrap-server kafka.kafka.svc.cluster.local:9092 \
    --topic test-topic
# Type a message, press Ctrl+D

# Consume
kubectl run kafka-consume -it --rm --restart=Never \
  --image=bitnami/kafka:latest -- \
  kafka-console-consumer.sh \
    --bootstrap-server kafka.kafka.svc.cluster.local:9092 \
    --topic test-topic --from-beginning --max-messages 1
```

**Expected:** The message you typed appears in the consumer output.

### 7. Install PostgreSQL

```bash
DATA_POSTGRES=1 make platform-data-postgres-up
# or directly:
bash platform/data/postgres/install.sh
```

**Expected:**
- `cnpg-system` namespace with the operator.
- `postgres` namespace with `lab-postgres-1` (primary) and `lab-postgres-2` (replica).

Verify:

```bash
make platform-data-postgres-status DATA_POSTGRES=1
kubectl get cluster -n postgres
```

Connect:

```bash
PG_PASS=$(kubectl get secret lab-postgres-superuser -n postgres \
  -o jsonpath='{.data.password}' | base64 -d)
kubectl run psql-test -it --rm --restart=Never \
  --image=postgres:16 \
  --env="PGPASSWORD=${PG_PASS}" -- \
  psql -h lab-postgres-rw.postgres.svc.cluster.local -U postgres -c "\l"
```

### 8. Failover drill (PostgreSQL)

```bash
# Find the primary
PRIMARY=$(kubectl get cluster lab-postgres -n postgres \
  -o jsonpath='{.status.currentPrimary}')
echo "Deleting primary: ${PRIMARY}"
kubectl delete pod "${PRIMARY}" -n postgres

# Watch the failover — replica promotes in ~30 s
watch kubectl get pods -n postgres
```

**Expected:** `lab-postgres-2` (or the surviving replica) transitions to primary.
`lab-postgres-rw` service follows the new primary automatically.

### 9. Uninstall data components

```bash
DATA_KAFKA=1     make platform-data-kafka-down
DATA_POSTGRES=1  make platform-data-postgres-down
```

### Troubleshooting (Data)

- **Kafka pods Pending** — insufficient cluster resources. Edit
  `platform/data/kafka/values.yaml`, reduce `controller.resources.requests.memory`
  to `128Mi`.
- **CNPG operator webhook timeout** — increase `--wait --timeout` in
  `platform/data/postgres/install.sh` or re-run install (idempotent).
- **PostgreSQL PVCs stuck on uninstall** — if the namespace hangs, manually
  remove the finalizers:
  `kubectl patch pvc <name> -n postgres -p '{"metadata":{"finalizers":null}}'`

---

## Part C — Secrets Management (task 056)

Adds `platform/secrets/vault/` and `platform/secrets/external-secrets/`.
Install in order: Vault first, then ESO.

### Prerequisites

- A healthy cluster (runbooks 00–01).
- Vault requires the ingress to be running for UI access (optional but useful).

### 10. Install Vault (dev mode)

```bash
SECRETS_VAULT=1 make platform-secrets-vault-up
# or directly:
bash platform/secrets/vault/install.sh
```

**Expected:**
- `vault` namespace with a single Vault pod (dev mode: auto-unsealed, in-memory).
- Demo secret seeded: `secret/go-api` with `db_password` and `api_key` keys.
- UI accessible at `http://vault.k3d.local` (token: `root`).

Verify:

```bash
make platform-secrets-vault-status SECRETS_VAULT=1

# Browse secrets via the CLI
kubectl exec -n vault deploy/vault -- vault kv get secret/go-api
```

### 11. Install External Secrets Operator

```bash
SECRETS_ESO=1 make platform-secrets-eso-up
# or directly:
bash platform/secrets/external-secrets/install.sh
```

**Expected:**
- `external-secrets` namespace with operator + webhook pods.
- `vault-backend` SecretStore created in the `go-api` namespace.
- `go-api-secrets` ExternalSecret created; syncs within 30 s.

Verify:

```bash
make platform-secrets-eso-status SECRETS_ESO=1
kubectl get externalsecret -n go-api
kubectl get secret go-api-external-secret -n go-api
```

### 12. Secret rotation exercise

```bash
# 1. Update the secret in Vault
kubectl exec -n vault deploy/vault -- \
  vault kv put secret/go-api db_password=rotated-password-v2

# 2. Wait up to 30 s (ESO refreshInterval)
sleep 35

# 3. Verify the k8s Secret is updated
kubectl get secret go-api-external-secret -n go-api \
  -o jsonpath='{.data.db_password}' | base64 -d
```

**Expected:** Output is `rotated-password-v2`.

### 13. Uninstall

```bash
SECRETS_ESO=1   make platform-secrets-eso-down    # remove ESO first
SECRETS_VAULT=1 make platform-secrets-vault-down  # then Vault
```

### Troubleshooting (Secrets)

- **Vault ingress 404** — check that Traefik is running and the `vault.k3d.local`
  entry is in `/etc/hosts` (runbook 00). The Vault UI renders at `/ui/`.
- **ESO SecretStore shows "Invalid"** — Vault may not be reachable. Check:
  `kubectl exec -n external-secrets deploy/external-secrets -- wget -qO- http://vault.vault.svc.cluster.local:8200/v1/sys/health`
- **ExternalSecret stuck "Syncing"** — the token secret may be missing or the
  KV path may differ. Run:
  `kubectl describe externalsecret go-api-secrets -n go-api`
