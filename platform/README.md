# Platform Components

The `platform/` directory contains infrastructure components organized by category. Each component is a provider that can be swapped independently.

## How It Works

Components follow a standard layout:

```
platform/<category>/<provider>/
  install.sh       # Install via Helm (idempotent)
  uninstall.sh     # Remove via Helm + cleanup CRDs
  status.sh        # Health check and status output
  values.yaml      # Helm chart configuration
```

The active provider for each category is selected via environment variables in `.env`:

```bash
INGRESS_PROVIDER=traefik       # or nginx
METRICS_PROVIDER=prometheus
```

Each category also has an `_interface.yaml` file documenting the contract that all providers in that category must satisfy.

## Installation

```bash
# Install all platform components
make platform-up
# or
labctl platform up

# Check status
make platform-status
# or
labctl platform status

# Remove all
make platform-down
# or
labctl platform down
```

## Components by Category

### Ingress

Controls external traffic routing into the cluster.

| Provider | Chart | Description |
|----------|-------|-------------|
| **traefik** | `traefik/traefik` | Default for k3d. LoadBalancer service, API dashboard |
| **nginx** | `ingress-nginx/ingress-nginx` | Default for cloud. Admission webhooks, ServiceMonitor |

Switch: `INGRESS_PROVIDER=nginx` in `.env`, then `make platform-up`.

### Monitoring

#### Metrics (`monitoring/metrics/`)

| Provider | Chart | Description |
|----------|-------|-------------|
| **prometheus** | `prometheus-community/kube-prometheus-stack` | Prometheus Operator, Node Exporter, Kube-State-Metrics, Alertmanager |

See [monitoring/README.md](monitoring/README.md) for detailed setup and verification.

#### Visualization (`monitoring/grafana/`)

| Provider | Chart | Description |
|----------|-------|-------------|
| **grafana** | `grafana/grafana` | Auto-provisioned Prometheus datasource, dashboard sidecar, 5Gi PVC |

Access: `http://grafana.k3d.local` (admin/admin)

### GitOps (`gitops/`)

| Provider | Chart | Description |
|----------|-------|-------------|
| **argocd** | `argo/argo-cd` | GitOps continuous delivery. Traefik ingress at `argocd.k3d.local` |

Activated via the `gitops-cicd` scenario or manually.

### Security (`security/`)

| Subcategory | Provider | Chart | Description |
|-------------|----------|-------|-------------|
| Policy | **kyverno** | `kyverno/kyverno` | Policy enforcement (admission control) |
| TLS | **cert-manager** | `jetstack/cert-manager` | Certificate management with self-signed CA |
| Secrets | **sealed-secrets** | `sealed-secrets/sealed-secrets` | Encrypted secrets in Git |
| Network | **kubernetes-native** | N/A | NetworkPolicy manifests (default-deny + explicit allows) |

These are typically activated via the `security-compliance` scenario.

### Chaos (`chaos/`)

| Provider | Chart | Description |
|----------|-------|-------------|
| **chaos-mesh** | `chaos-mesh/chaos-mesh` | Failure injection (pod kill, network delay, stress) |

Activated via the `chaos-engineering` scenario. Includes a web dashboard (port-forward to 2333).

## Provider Interface Contracts

Each category has an `_interface.yaml` documenting:

```yaml
category: ingress
description: Routes external HTTP/HTTPS traffic to cluster services
provides:
  - IngressClass resource
  - LoadBalancer or NodePort service
requires:
  - Kubernetes cluster
env_vars:
  INGRESS_CLASS: traefik | nginx
implementations:
  - name: traefik
    chart: traefik/traefik
  - name: nginx
    chart: ingress-nginx/ingress-nginx
```

## Adding a New Provider

1. Create directory: `platform/<category>/<provider-name>/`
2. Create the four required files: `install.sh`, `uninstall.sh`, `status.sh`, `values.yaml`
3. Follow the `_interface.yaml` contract for the category
4. Update `_interface.yaml` to list the new implementation
5. The CLI's platform registry will auto-discover it

### Script Template

```bash
#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${NAMESPACE:-my-namespace}"
RELEASE_NAME="my-provider"
CHART="repo/chart-name"

echo "Installing $RELEASE_NAME..."
helm repo add myrepo https://charts.example.com
helm repo update

kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

helm upgrade --install "$RELEASE_NAME" "$CHART" \
  --namespace "$NAMESPACE" \
  --values "$(dirname "$0")/values.yaml" \
  --wait --timeout 5m

echo "$RELEASE_NAME installed."
```

## Directory Structure

```
platform/
  ingress/
    _interface.yaml
    traefik/              install.sh, uninstall.sh, status.sh, values.yaml
    nginx/                install.sh, uninstall.sh, status.sh, values.yaml
  monitoring/
    README.md             Detailed monitoring setup guide
    metrics/
      _interface.yaml
      prometheus/         install.sh, uninstall.sh, status.sh, values.yaml
    grafana/
      _interface.yaml
      install.sh, uninstall.sh, status.sh, values.yaml
  gitops/
    _interface.yaml
    argocd/               install.sh, uninstall.sh, status.sh, values.yaml
  security/
    policy/
      _interface.yaml
      kyverno/            install.sh, uninstall.sh, status.sh, values.yaml
    tls/
      _interface.yaml
      cert-manager/       install.sh, uninstall.sh, status.sh, values.yaml, cluster-issuer.yaml
    secrets/
      _interface.yaml
      sealed-secrets/     install.sh, uninstall.sh, status.sh, values.yaml
    network-policies/
      _interface.yaml
      install.sh, uninstall.sh, status.sh
      default-deny.yaml, allow-dns.yaml, allow-monitoring.yaml, allow-ingress.yaml
  chaos/
    _interface.yaml
    chaos-mesh/           install.sh, uninstall.sh, status.sh, values.yaml
```
