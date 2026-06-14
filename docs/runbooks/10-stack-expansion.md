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

### Troubleshooting

- **Pods stuck at 0/2** — the sidecar injector webhook may not be ready. Wait
  30 s and re-run `kubectl rollout restart deployment -n go-api`.
- **Linkerd cert errors** — re-run `make platform-mesh-down MESH_PROVIDER=linkerd`
  and `make platform-mesh-up MESH_PROVIDER=linkerd`. Fresh certs are generated
  each install.
- **k3d resource limits** — if nodes OOM, reduce limits in `values.yaml`.
  For Istio, set `pilot.resources.requests.memory: 128Mi`.
