# Project State: Sagars-Laboratory

> Last updated: 2026-03-07
> Branch: `feature/experiment-2`
> Status: All 6 phases complete + hardening + polishing + documentation + operational testing round. Active testing and bug-fixing of UI-driven workflows.

## Session Update (2026-03-07) — Operational Testing & Bug Fixes

Two sessions of intensive testing and fixing real-world issues discovered while using the labctl UI.

### Round 1: 9 Issues Fixed

1. **Scenarios not activating (Issue 1 & 5)** — `engine.go` was using non-streamed execution (`RunHelm`, `RunKubectl`, `RunScript`). Switched ALL component operations to streamed equivalents (`RunCommandStreamed`, `RunScriptStreamed`) so events reach the UI via WebSocket. Added `Description` and `Runtimes` to `ScenarioStatus` struct.

2. **Platform components Install/Remove buttons (Issue 2)** — Made buttons conditional on component `active` state in both UI files.

3. **Traefik & ArgoCD dashboard URLs (Issue 3)** — ArgoCD URL fixed from `https://` to `http://`. Traefik IngressRoute created for dashboard access at `traefik.<domain>/dashboard/`.

4. **Logging & Tracing platform components (Issue 4)** — Created `platform/logging/loki/` (install, uninstall, status, values, promtail-values) and `platform/tracing/tempo/` (install, uninstall, status, values). Added Loki + Tempo datasources to Grafana values.yaml.

5. **K8s Dashboard install (Issue 6)** — Kubernetes Dashboard project archived (`kubernetes-retired`). Both Helm repo (404) and OCI registry (403) broken. Switched to direct tarball install from GitHub release: `kubernetes-dashboard-7.14.0.tgz`. Disabled Kong proxy TLS for HTTP ingress compatibility with Traefik.

6. **Runtimes dropdown (Issue 7)** — Created `cmd/labctl/internal/runtime/manager.go` with `ClusterName` field, `List()`, `Activate()`, `Deactivate()`. Fixed `expectedContext` to use `"k3d-" + clusterName`. Replaced slow `kubectl cluster-info` with local `config get-contexts` check.

7. **Platform card width (Issue 8)** — Made platform card full-width in UI.

8. **Log Aggregation in UI (Issue 9)** — Added `openAppLogs()` function with Grafana Explore deep-links for Loki/Tempo. Added "Logs" button to app rows.

9. **Config propagation** — Added `exec.SetEnv()` in `root.go` for `CLUSTER_NAME`, `DOMAIN_SUFFIX`, `HTTP_PORT`, `HTTPS_PORT`, `INGRESS_CLASS`, `STORAGE_CLASS`, `PROFILE`. All child scripts now reliably receive config values.

### Round 2: Post-Testing Fixes

1. **K8s Dashboard 404** — Disabled Kong proxy TLS (`kong.proxy.tls.enabled: false` in values.yaml). Fixed ingress: replaced deprecated `kubernetes.io/ingress.class` annotation with `spec.ingressClassName: traefik`, removed TLS section, switched to port 80.

2. **ArgoCD blocked by Kyverno** — Added `argocd` and `kubernetes-dashboard` to ALL 6 Kyverno ClusterPolicy exclusion lists in `scenarios/security-compliance/manifests/kyverno-policies.yaml`.

3. **Scenario status race condition** — Added explicit success/failure `action_end` broadcasts in `handleScenarioUp`/`handleScenarioDown` (handlers.go). Root cause: `Up()` calls `markActive()` at the end, but the last component's streamed event fires BEFORE `markActive()`, so the UI refresh still saw inactive.

4. **Grafana persistent 404** — Root cause: `DOMAIN_SUFFIX` not reaching install scripts. Fixed by removing dead `.active-runtime.env` sourcing, using executor-provided env vars, and adding dynamic `--set "ingress.hosts[0]=grafana.${DOMAIN_SUFFIX}"`.

5. **Prometheus persistent 404** — Root cause: static `ingress.yaml` with hardcoded `prometheus.k3d.local`. Replaced with inline heredoc in install.sh using `$DOMAIN_SUFFIX` variable.

6. **Traefik IngressClass conflict** — k3d bundles Traefik in `kube-system` with `HelmChart` CRD that re-creates it after deletion. Fix: (a) Disable bundled Traefik at cluster creation with `--k3s-arg "--disable=traefik@server:*"`, (b) Added cleanup of `HelmChart`/`HelmChartConfig` CRDs in install script for existing clusters.

7. **Tempo/Loki race condition** — Added stuck release cleanup (detect `pending-` state, delete before install) to both install scripts.

8. **Cluster name consistency** — `up.sh`/`down.sh` now use `${1:-${CLUSTER_NAME:-sagars-cluster}}`. `Manager.Activate()`/`Deactivate()` pass `m.ClusterName` as argument.

### Files Changed

**Go source (cmd/labctl/):**
- `cmd/root.go` — Config propagation via `exec.SetEnv()`
- `cmd/ui.go` — (previously updated)
- `internal/api/handlers.go` — DomainSuffix in response, logging/tracing detection, success broadcasts, URL fixes
- `internal/api/server.go` — (previously updated)
- `internal/executor/executor.go` — (previously updated)
- `internal/executor/broadcast.go` — NEW: WebSocket event broadcaster
- `internal/k8s/client.go` — Added `ServiceExists()` helper
- `internal/platform/registry.go` — Added `InstallStreamed`/`UninstallStreamed` methods
- `internal/runtime/manager.go` — NEW: Runtime manager with cluster name support
- `internal/scenario/engine.go` — Streamed execution, Description/Runtimes fields

**Platform scripts:**
- `platform/ingress/traefik/install.sh` — HelmChart CRD cleanup, `--force-update`, `--wait`, IngressRoute
- `platform/monitoring/prometheus/install.sh` — Dynamic ingress via heredoc, `--force-update`, `--wait`
- `platform/monitoring/grafana/install.sh` — Removed `.active-runtime.env` dependency, dynamic ingress host
- `platform/monitoring/grafana/values.yaml` — Added Loki + Tempo datasources
- `platform/dashboard/kubernetes-dashboard/install.sh` — Direct tarball install, HTTP ingress
- `platform/dashboard/kubernetes-dashboard/values.yaml` — `kong.proxy.tls.enabled: false`
- `platform/logging/loki/` — NEW: install.sh, uninstall.sh, status.sh, values.yaml, promtail-values.yaml
- `platform/tracing/tempo/` — NEW: install.sh, uninstall.sh, status.sh, values.yaml

**Runtime scripts:**
- `runtimes/k3d/up.sh` — `--disable=traefik`, cluster name from env, skip-if-exists
- `runtimes/k3d/down.sh` — Target specific cluster, cluster name from env

**Scenario manifests:**
- `scenarios/security-compliance/manifests/kyverno-policies.yaml` — Added argocd + kubernetes-dashboard exclusions

**UI:**
- `ui/dist/index.html` — Full-width platform card, conditional buttons, Grafana deep-links, Logs button
- `cmd/labctl/ui/dist/index.html` — Same changes (embedded copy)

**Config:**
- `.env` — Added `LOGGING_PROVIDER=loki`, `TRACING_PROVIDER=tempo`

---

## What This Project Is

A Kubernetes-based homelab for testing Platform Engineering and DevOps scenarios. Designed to be a reproducible, containerized lab environment where you can spin up infrastructure, deploy apps, activate testing scenarios (chaos, security, GitOps, observability), and tear everything down in minutes.

---

## What Is Already Built

### 1. Cluster Runtimes (`runtimes/`)
- **k3d** (`runtimes/k3d/`) — Local Kubernetes with 2 agent nodes
  - Port mapping: host 80 -> LB 80, host 443 -> LB 443
  - Bundled Traefik disabled (`--disable=traefik`) — managed separately via platform install
  - `up.sh` (creates cluster with env-driven name/ports, skips if exists), `down.sh` (deletes specific cluster), `runtime.env`
- **AKS** (`runtimes/aks/`) — Azure Kubernetes Service via Terraform
- **EKS** (`runtimes/eks/`) — AWS Elastic Kubernetes Service via Terraform

### 2. Tool Bootstrap (`bootstrap/`)
- `setup-tools.sh` — installs tools with version pinning per profile

### 3. Applications (`apps/`)
- **go-api** — Go 1.24, Prometheus + OpenTelemetry, multi-stage Docker build
- **echo-server** — Go 1.24, Redis (optional), Prometheus metrics, multi-stage Docker build

### 4. Helm Charts (`apps/*/deploy/helm/`)
- Both charts: v0.1.0, PDB templates, multi-profile values (dev, prod-like, cloud, test)

### 5. Engine Layer (`engine/`)
- Strategy pattern for build (docker, acr, ecr) and deploy (helm)

### 6. Platform Components (`platform/`)
- **Ingress**: Traefik (with IngressRoute dashboard, k3d cleanup), Nginx
- **Monitoring**: Prometheus (dynamic ingress), Grafana (dynamic ingress, Loki+Tempo datasources)
- **Logging**: Loki + Promtail (stuck release cleanup)
- **Tracing**: Tempo (stuck release cleanup)
- **Dashboard**: Kubernetes Dashboard (direct tarball install from kubernetes-retired, HTTP via Kong)
- **GitOps**: ArgoCD
- **Security**: Kyverno, cert-manager, Sealed Secrets, Network Policies
- **Chaos**: Chaos Mesh

### 7. Lab Controller CLI — `labctl` (`cmd/labctl/`)
- **Language**: Go 1.24, Cobra + Viper + Gorilla + yaml.v3
- **Binary**: `bin/labctl` (~15MB with embedded UI)
- **Commands**: init, teardown, reset, status, runtime (up/down/status), app (build/deploy/destroy/list), platform (up/down/status), check (tools/cluster/ingress), scenario (list/up/down/status/info), service (list/up/down/status), ui
- **Internal packages**:
  - `internal/config/` — loads .env, app.env, runtime.env; propagates to executor env
  - `internal/executor/` — wraps os/exec, supports streamed output + WebSocket broadcasting
  - `internal/executor/broadcast.go` — ActionEvent broadcaster for real-time UI updates
  - `internal/platform/` — provider registry with streamed install/uninstall
  - `internal/k8s/` — cluster info, app status, namespace/service existence checks
  - `internal/api/` — HTTP API server with REST + WebSocket, DomainSuffix in status response
  - `internal/runtime/` — runtime manager with cluster name, activate/deactivate with args
  - `internal/scenario/` — YAML-based engine with streamed execution, success/failure broadcasts
  - `internal/services/` — service registry

### 8. Web UI Dashboard (`ui/dist/`)
- Dark theme, responsive grid layout, served by `labctl ui` on port 3939
- Cards: Cluster info (with runtime selector), Platform components (full-width, conditional Install/Remove), Applications (with Logs deep-link), Scenarios (with status updates)
- REST API integration + WebSocket for real-time streamed output
- Grafana Explore deep-links for Loki (logs) and Tempo (traces)

### 9. Scenario Framework (`scenarios/`)
- 4 scenarios: observability-sre, gitops-cicd, security-compliance, chaos-engineering
- Kyverno policies exclude argocd + kubernetes-dashboard namespaces
- Engine supports: helm, manifest, grafana-dashboard, script component types
- Status includes Description and Runtimes fields

### 10. Config Flow
- `.env` -> `runtimes/<profile>/runtime.env` -> `config.Load()` -> `exec.SetEnv()` -> child scripts
- Key propagated vars: CLUSTER_NAME, DOMAIN_SUFFIX, HTTP_PORT, HTTPS_PORT, INGRESS_CLASS, STORAGE_CLASS, PROFILE
- Scripts use `${DOMAIN_SUFFIX:-k3d.local}` pattern for fallback

---

## Architecture Patterns To Preserve

1. **Strategy dispatch**: `app.env` -> `engine/build.sh` -> `engine/build/<strategy>.sh`
2. **Platform components**: Each gets `install.sh`, `uninstall.sh`, `status.sh`, `values.yaml`
3. **Per-app config**: `apps/<name>/app.env` is the contract between apps and engine
4. **Env-based profiles**: `PROFILE=k3d|aks|eks` drives tool installation and runtime selection
5. **Version pinning**: `versions.env` controls all tool versions
6. **Idempotent operations**: All installs use `helm upgrade --install`
7. **Provider swappability**: Env var selects provider per category, registry routes to scripts
8. **CLI wraps scripts**: `labctl` calls existing shell scripts, doesn't replace them
9. **Services pattern**: `services/<name>/` follows same convention as platform providers
10. **Scenario YAML**: Declarative `scenario.yaml` with components, prerequisites, explore hints
11. **Config propagation**: `root.go` sets env vars on executor; scripts inherit them automatically
12. **Streamed execution**: All UI-visible operations use `RunCommandStreamed`/`RunScriptStreamed` for real-time WebSocket output
13. **Success broadcasts**: Handler goroutines send explicit `action_end` events after both success AND failure, ensuring UI status refresh works reliably
14. **Dynamic ingress hosts**: All install scripts use `$DOMAIN_SUFFIX` for ingress host configuration, never hardcoded domain names in static YAML files

---

## File Inventory

```
Total: ~260+ source files
  Shell scripts:   55+ (+loki, tempo, updated traefik/prometheus/grafana/k8s-dashboard, updated k3d up/down)
  YAML configs:    65+ (+loki values, promtail-values, tempo values, updated grafana values, kyverno policies)
  Terraform:       10
  Make includes:    9
  Go source (CLI): 28+ (+broadcast.go, manager.go, updated engine/handlers/root/client)
  Go source (apps): 6
  Markdown docs:   12+
  Env files:        8
  Dockerfiles:      2
  Helm charts:      2
  HTML:             1 (+ embedded copy)
  JSON dashboards:  4
  Scenario YAMLs:   4
  CI/CD workflows:  6
  Go test files:    5 (35 tests total)
```
