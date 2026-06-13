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

## Acceptance check (task 054)

- [x] `MESH_PROVIDER=istio labctl platform up mesh` meshes go-api (sidecar visible)
- [x] Swapping to linkerd on the same cluster works after uninstall
- [x] `status.sh` reports mesh health accurately for both providers
- [x] Scripts portable + idempotent; versions pinned in `versions.env`

> Data, secrets, and autoscaling sections (tasks 055–057) and the new scenarios
> (task 058) are appended here as those tasks ship.
