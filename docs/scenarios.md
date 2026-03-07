# Scenarios Guide

Scenarios are declarative playgrounds that install a collection of related tools, configurations, and dashboards to explore specific DevOps concepts.

## How Scenarios Work

Each scenario is a directory in `scenarios/` containing a `scenario.yaml` that declares:

- **Prerequisites** - platform components and apps that must be running
- **Components** - what to install (Helm charts, manifests, dashboards, scripts)
- **Explore hints** - URLs, commands, and tips for experimenting

The scenario engine handles installation order, template resolution, and state tracking.

## Using Scenarios

### List available scenarios

```bash
labctl scenario list
```

### Get details before activating

```bash
labctl scenario info observability-sre
```

This shows the full description, prerequisites, components that will be installed, and exploration hints.

### Activate a scenario

```bash
labctl scenario up observability-sre
```

The engine will:
1. Validate prerequisites (platform components and apps)
2. Install each component in order (Helm charts, kubectl manifests, Grafana dashboards)
3. Mark the scenario as active
4. Print exploration tips

### Deactivate a scenario

```bash
labctl scenario down observability-sre
```

Removes all components installed by the scenario.

### Check status

```bash
labctl scenario status
```

You can also use the web UI (`labctl ui`) to activate/deactivate scenarios with a single click.

---

## Available Scenarios

### Observability & SRE (`observability-sre`)

**Category:** observability

**What it deploys:**
- Loki (log aggregation via `grafana/loki` Helm chart)
- Promtail (log shipping agent via `grafana/promtail`)
- Tempo (distributed tracing via `grafana/tempo`)
- Alerting rules (PrometheusRule CRDs for high error rate, latency, pod restarts)
- SLO dashboards (Grafana JSON dashboards for availability, latency, error budget)

**Prerequisites:**
- Platform: ingress, monitoring/metrics, monitoring/grafana
- Apps: go-api

**Explore after activation:**
- Open Grafana at `http://grafana.k3d.local`
- Explore > Select Loki datasource > Query `{namespace="go-api"}`
- Explore > Select Tempo datasource > Search by service name
- Generate traffic: `for i in $(seq 1 100); do curl -s http://go-api.k3d.local/health; done`
- Trigger failures: `curl http://go-api.k3d.local/toggle-failure` then hit `/ready`
- Check alerts: `kubectl -n monitoring get prometheusrules`

---

### GitOps & CI/CD (`gitops-cicd`)

**Category:** gitops

**What it deploys:**
- ArgoCD (via Helm chart with Traefik ingress)
- ArgoCD Application CRDs pointing at `apps/go-api/deploy/helm/` and `apps/echo-server/deploy/helm/`
- Multi-environment setup (dev/staging namespaces with different values files)

**Prerequisites:**
- Platform: ingress, monitoring/metrics, monitoring/grafana
- Apps: go-api

**Explore after activation:**
- Open ArgoCD dashboard at `http://argocd.k3d.local`
- Login: admin / (password printed during install)
- Watch both apps synced in the ArgoCD UI
- Change a values file, observe ArgoCD detect drift and sync
- Perform a rollback via ArgoCD UI

---

### Security & Compliance (`security-compliance`)

**Category:** security

**What it deploys:**
- Kyverno (policy enforcement engine via Helm)
- cert-manager (TLS certificate management via Helm)
- 6 Kyverno ClusterPolicies:
  - `disallow-privileged-containers` (Enforce)
  - `require-labels` (Audit)
  - `disallow-root-user` (Audit)
  - `disallow-host-path` (Enforce)
  - `require-resource-limits` (Audit)
  - `disallow-latest-tag` (Audit)
- Network Policies (namespace isolation for go-api and echo-server)
- Security Grafana dashboard (policy violations, certificates, admission latency)

**Prerequisites:**
- Platform: ingress, monitoring/metrics, monitoring/grafana
- Apps: go-api

**Explore after activation:**
- Try deploying a non-compliant pod: `kubectl run nginx --image=nginx` (Kyverno blocks it)
- Check policy violations: `kubectl get policyreports -A`
- View security dashboard in Grafana
- Test network isolation between namespaces

---

### Chaos Engineering (`chaos-engineering`)

**Category:** chaos

**What it deploys:**
- Chaos Mesh (failure injection engine via Helm)
- PodDisruptionBudgets for go-api and echo-server
- 8 pre-built chaos experiments:
  - **PodChaos:** pod-kill (go-api), pod-kill (echo-server), pod-failure (go-api)
  - **NetworkChaos:** delay (echo-server to Redis, 500ms), partition (go-api to Traefik), packet loss (echo-server to Redis, 50%)
  - **StressChaos:** CPU stress (go-api), memory stress (echo-server)
- Chaos Grafana dashboard (experiment timeline, pod restarts, HTTP metrics, resource usage)

**Prerequisites:**
- Platform: ingress, monitoring/metrics, monitoring/grafana
- Apps: go-api

**Explore after activation:**
- Port-forward Chaos Dashboard: `kubectl -n chaos-mesh port-forward svc/chaos-dashboard 2333:2333`
- Run an experiment: `kubectl apply -f scenarios/chaos-engineering/manifests/chaos-experiments.yaml`
- Watch pods recover: `kubectl get pods -n go-api -w`
- Generate traffic during experiments: `while true; do curl -s http://go-api.k3d.local/health; sleep 0.1; done`
- Monitor impact in Grafana chaos dashboard
- Check PDB status: `kubectl get pdb -A`

---

## Scenario YAML Format

```yaml
name: my-scenario                    # Must match directory name
displayName: "My Scenario"           # Shown in UI and CLI
description: "What this scenario teaches"
category: observability              # Grouping label

prerequisites:
  platform:                          # Required platform components
    - ingress
    - monitoring/metrics
  apps:                              # Required apps
    - go-api

runtimes:                            # Compatible runtimes (optional)
  - k3d
  - aks

components:                          # What to install (in order)
  - name: my-chart
    type: helm                       # helm | manifest | grafana-dashboard | script
    chart: repo/chart-name
    repo: https://charts.example.com # Helm repo URL
    version: "1.0.0"
    namespace: my-ns
    valuesFile: values/my-chart.yaml

  - name: my-manifests
    type: manifest
    path: manifests/my-resources.yaml
    namespace: my-ns

  - name: my-dashboards
    type: grafana-dashboard
    path: dashboards/
    namespace: monitoring

  - name: my-setup
    type: script
    path: scripts/setup.sh

explore:
  urls:
    - label: "My Dashboard"
      url: "http://my-app.{{.DomainSuffix}}"

  commands:
    - label: "Check status"
      command: "kubectl get pods -n my-ns"

  tips:
    - "First, generate some traffic to see data in dashboards"
    - "Try breaking things to see alerts fire"
```

### Template Variables

URLs and commands support Go template variables:

| Variable | Example Value | Description |
|----------|--------------|-------------|
| `{{.DomainSuffix}}` | `k3d.local` | Domain suffix from active runtime |
| `{{.ProjectRoot}}` | `/path/to/project` | Absolute path to project root |

### Component Types

| Type | What It Does |
|------|-------------|
| `helm` | Adds Helm repo, installs chart with values file |
| `manifest` | Applies Kubernetes YAML via `kubectl apply` |
| `grafana-dashboard` | Creates ConfigMap from dashboard JSON files (picked up by Grafana sidecar) |
| `script` | Runs a shell script from the scenario directory |

## Creating a New Scenario

1. Create directory: `scenarios/my-scenario/`
2. Write `scenario.yaml` following the format above
3. Add supporting files under `values/`, `manifests/`, `dashboards/` as needed
4. Test: `labctl scenario up my-scenario`

The scenario engine auto-discovers any directory under `scenarios/` that contains a valid `scenario.yaml`.
