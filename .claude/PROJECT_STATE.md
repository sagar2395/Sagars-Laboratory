# Project State: Sagars-Laboratory

> Last updated: 2026-03-07
> Branch: `feature/experiment`
> Status: All 6 phases complete + hardening round + polishing round. Tests, go:embed, GitHub Actions workflows, cloud Helm values, Terraform backends all done.

## Session Update (2026-03-07)

- Fixed `echo-server` readiness failures caused by unresolved Redis DNS in clusters where Redis service is not installed.
- Root cause: Helm defaults forced `REDIS_URL=redis://redis-master.services.svc.cluster.local:6379`, but no `services` namespace/service existed at runtime.
- Applied fix: `REDIS_URL` now defaults to empty in echo-server Helm values so Redis stays optional and `/ready` passes when Redis is absent.
- Updated files:
  - `apps/echo-server/deploy/helm/values.yaml`
  - `apps/echo-server/deploy/helm/values-dev.yaml`
  - `apps/echo-server/deploy/helm/values-cloud.yaml`
  - `apps/echo-server/app.env` (clarifying comment)
- Verified: redeployed `echo-server`; new pod started with `redis=false`, env showed empty `REDIS_URL`, pod reached `1/1` ready.

---

## What This Project Is

A Kubernetes-based homelab for testing Platform Engineering and DevOps scenarios. Designed to be a reproducible, containerized lab environment where you can spin up infrastructure, deploy apps, activate testing scenarios (chaos, security, GitOps, observability), and tear everything down in minutes.

---

## What Is Already Built

### 1. Cluster Runtimes (`runtimes/`)
- **k3d** (`runtimes/k3d/`) — Local Kubernetes with 2 agent nodes
  - Port mapping: host 80 -> LB 80, host 443 -> LB 443
  - `up.sh`, `down.sh`, `runtime.env` (INGRESS_CLASS=traefik, STORAGE_CLASS=local-path, DOMAIN_SUFFIX=k3d.local)
- **AKS** (`runtimes/aks/`) — Azure Kubernetes Service via Terraform
  - `up.sh` (az login check, resource group, terraform apply, get-credentials)
  - `down.sh` (terraform destroy or direct az aks delete)
  - `runtime.env` (INGRESS_CLASS=nginx, STORAGE_CLASS=managed-csi, DOMAIN_SUFFIX=sagarslab.io, REGISTRY_TYPE=acr)
- **EKS** (`runtimes/eks/`) — AWS Elastic Kubernetes Service via Terraform
  - `up.sh` (aws sts check, terraform apply, update-kubeconfig)
  - `down.sh` (terraform destroy or eksctl delete)
  - `runtime.env` (INGRESS_CLASS=nginx, STORAGE_CLASS=gp3, DOMAIN_SUFFIX=sagarslab.io, REGISTRY_TYPE=ecr)

### 2. Tool Bootstrap (`bootstrap/`)
- `setup-tools.sh` — installs tools with version pinning per profile
- Profiles: `k3d`, `aks`, `eks`, `common`, `all`
- All versions pinned in `versions.env`:
  - kubectl 1.29.3, k3d 5.8.3, docker 29.2.1, helm 3.14.0, az-cli 2.57.0, terraform 1.7.0, aws-cli 2

### 3. Application: go-api (`apps/go-api/`)
- **Language**: Go 1.24, standard library + Prometheus client + OpenTelemetry
- **Endpoints**: `/health`, `/ready`, `/metrics`, `/toggle-failure`, `/` (service info)
- **Features**: Graceful shutdown, structured JSON logging (slog), optional OTel tracing
- **Container**: Multi-stage Dockerfile (golang:1.24-alpine -> alpine:latest)
- **Config**: `app.env` — BUILD_STRATEGY=docker, DEPLOY_STRATEGY=helm

### 4. Application: echo-server (`apps/echo-server/`)
- **Language**: Go 1.24, standard library + Prometheus client + Redis client
- **Endpoints**: `/health`, `/ready` (Redis check), `/echo` (request details), `/cache` (GET/POST/DELETE with Redis), `/metrics`
- **Features**: Structured JSON logging (slog), Prometheus labeled metrics, graceful shutdown
- **Dependencies**: Redis (optional, via REDIS_URL env var)
- **Container**: Multi-stage Dockerfile (golang:1.24-alpine -> alpine:latest)
- **Config**: `app.env` — BUILD_STRATEGY=docker, DEPLOY_STRATEGY=helm
- **Helm chart**: 7 templates, Redis optional by default (`REDIS_URL: ""` in values files)

### 5. Helm Charts (`apps/*/deploy/helm/`)
- **go-api chart**: v0.1.0, 10 templates (added pdb.yaml)
- **echo-server chart**: v0.1.0, 8 templates (added pdb.yaml)
- **Value profiles**: values-dev.yaml, values-prod-like.yaml, values-cloud.yaml, values-test.yaml, values.yaml

### 6. Engine Layer (`engine/`)
- **Strategy pattern** for build and deploy
- `build.sh` -> loads `app.env` -> dispatches to `engine/build/<strategy>.sh`
- `deploy.sh` -> loads `app.env` -> dispatches to `engine/deploy/<strategy>.sh`
- **Build strategies**: `docker.sh` (build + k3d import), `acr.sh` (build + Azure ACR push), `ecr.sh` (build + AWS ECR push)
- **Deploy strategies**: `helm.sh` (install/upgrade/destroy/lint/validate)

### 7. Platform: Ingress (`platform/ingress/`)
- **Traefik** (`traefik/`) via Helm chart — LoadBalancer service type, API dashboard enabled
  - install.sh, uninstall.sh, status.sh, values.yaml
- **Nginx** (`nginx/`) via ingress-nginx Helm chart — Alternative for cloud environments
  - install.sh, uninstall.sh, status.sh, values.yaml
  - ServiceMonitor for Prometheus, admission webhooks enabled

### 8. Platform: Monitoring (`platform/monitoring/`)
- **Prometheus** (kube-prometheus-stack): Operator, Node Exporter, Kube-State-Metrics, Alertmanager
- **Grafana**: Auto-provisioned Prometheus datasource, dashboard sidecar, 5Gi PVC

### 9. Platform: GitOps (`platform/gitops/argocd/`)
- **ArgoCD** via Helm chart (argo/argo-cd)
- install.sh, uninstall.sh, status.sh, values.yaml
- Server config with Traefik ingress at argocd.k3d.local
- Reduced replicas for local dev, insecure mode for HTTP

### 10. Shared Services (`services/`)
- **Redis** (`services/redis/`) — Bitnami Redis via Helm
  - Standalone mode, no auth, 1Gi persistence
  - install.sh, uninstall.sh, status.sh, values.yaml
  - Namespace: services
  - Connection: redis://redis-master.services.svc.cluster.local:6379

### 11. Makefile System (`make/`)
- **Modular includes**: vars.mk, bootstrap.mk, runtime.mk, app.mk, platform.mk, check.mk, cli.mk, services.mk, terraform.mk
- **Lifecycle targets**: `init`, `teardown`, `reset`
- **App targets**: build, deploy, destroy-app, lint, validate, deploy-all, destroy-all-apps
- **Platform targets**: platform-up, platform-down, platform-status
- **Service targets**: service-list, service-up, service-down, service-status
- **CLI targets**: cli-build, cli-install, cli-tidy, cli-clean
- **Terraform targets**: terraform-init, terraform-plan, terraform-apply, terraform-destroy, terraform-output, terraform-status

---

### 12. Lab Controller CLI — `labctl` (`cmd/labctl/`)
- **Language**: Go 1.24, Cobra + Viper + Gorilla + yaml.v3
- **Binary**: `bin/labctl` (built via `make cli-build`, ~15MB with embedded UI)
- **Commands**: init, teardown, reset, status, runtime (up/down/status), app (build/deploy/destroy/list), platform (up/down/status), check (tools/cluster/ingress), scenario (list/up/down/status/info), service (list/up/down/status), ui
- **Internal packages**:
  - `internal/config/` — loads .env, app.env, runtime.env via viper
  - `internal/executor/` — wraps os/exec for script and command execution
  - `internal/platform/` — provider registry, discovers platform components
  - `internal/k8s/` — cluster info, app status, namespace checks via kubectl
  - `internal/api/` — HTTP API server with REST + WebSocket for web UI, embedded FS support
  - `internal/scenario/` — YAML-based scenario engine with helm/manifest/dashboard/script handlers
  - `internal/services/` — service registry, discovers and manages shared services
- **UI embedding**: `ui/embed.go` with `go:embed all:dist` — `make cli-build` copies `ui/dist/` before compile
- **Tests**: 35 tests across 5 packages (config, executor, platform, scenario, services)
- **Source files**: 25+ Go files (main.go, 10 cmd files, 6 internal packages, 5 test files, embed.go)
- **Module**: `github.com/sagars-lab/labctl`

### 13. Web UI Dashboard (`ui/dist/`)
- **Functional HTML dashboard** served by `labctl ui` on port 3939
- Dark theme, responsive grid layout
- Cards: Cluster info, Platform components, Applications, Scenarios
- REST API integration + WebSocket for real-time updates

### 14. Scenario Framework (`scenarios/`)
- **Engine**: `cmd/labctl/internal/scenario/engine.go`
  - Component types: helm, manifest, grafana-dashboard, script
  - Template resolution for `{{.DomainSuffix}}`, `{{.ProjectRoot}}`
  - State tracking via `.labctl/scenarios/<name>.active`
- **Observability SRE scenario**: `scenarios/observability-sre/`
  - Loki, Promtail, Tempo, 6 PrometheusRule alerts, 2 Grafana dashboards
- **GitOps CI/CD scenario**: `scenarios/gitops-cicd/`
  - ArgoCD helm component, ArgoCD Application CRDs for go-api + echo-server
- **Security Compliance scenario**: `scenarios/security-compliance/`
  - Kyverno + cert-manager (helm), 6 Kyverno policies + network policies (manifests), security Grafana dashboard
- **Chaos Engineering scenario**: `scenarios/chaos-engineering/`
  - Chaos Mesh (helm), PodDisruptionBudgets + 8 chaos experiments (manifests), chaos Grafana dashboard

### 15. CI/CD Templates & Workflows
- **Templates** (`delivery/github-actions/`):
  - `ci.yaml` — Build, test, lint, Helm lint workflow (template)
  - `cd.yaml` — Deploy-on-push workflow (template)
  - `helm-release.yaml` — Helm package and release workflow (template)
- **Active Workflows** (`.github/workflows/`):
  - `ci.yaml` — Multi-app CI: lint, test, build, Helm lint for go-api, echo-server, labctl CLI
  - `cd.yaml` — Multi-app CD: auto-detect changed apps, build+push to GHCR, update manifests
  - `helm-validation.yaml` — PR validation: discover and lint all Helm charts

### 16. Platform: Security (`platform/security/`)
- **Kyverno** (`policy/kyverno/`) — Policy enforcement engine
  - install.sh, uninstall.sh, status.sh, values.yaml
  - Admission, background, cleanup, reports controllers
- **cert-manager** (`tls/cert-manager/`) — Certificate management
  - install.sh, uninstall.sh, status.sh, values.yaml, cluster-issuer.yaml
  - Self-signed ClusterIssuer + CA chain for lab certificates
- **Sealed Secrets** (`secrets/sealed-secrets/`) — Encrypted secrets in Git
  - install.sh, uninstall.sh, status.sh, values.yaml
  - Prometheus ServiceMonitor enabled
- **Network Policies** (`network-policies/`) — Namespace isolation
  - install.sh, uninstall.sh, status.sh
  - default-deny.yaml, allow-dns.yaml, allow-monitoring.yaml, allow-ingress.yaml
  - Default deny all + explicit allows for DNS, Prometheus, Traefik, Redis

### 17. Platform: Chaos (`platform/chaos/`)
- **Chaos Mesh** (`chaos-mesh/`) — Failure injection engine
  - install.sh, uninstall.sh, status.sh, values.yaml
  - Controller manager, chaos daemon (containerd socket), dashboard
  - ServiceMonitor for Prometheus integration
  - Dashboard on port 2333 (via port-forward)

### 18. Foundation: Terraform (`foundation/terraform/`)
- **AKS module** (`modules/aks/`) — Azure Kubernetes Service
  - `main.tf` — AKS cluster, Log Analytics, ACR, ACR pull role assignment
  - `variables.tf` — cluster_name, resource_group, location, vm_size, node_count, autoscaling, ACR options
  - `outputs.tf` — cluster_name, cluster_id, kube_config_raw, acr_login_server
  - Uses azurerm provider ~> 3.80, Calico network policy, SystemAssigned identity
- **EKS module** (`modules/eks/`) — AWS Elastic Kubernetes Service
  - `main.tf` — VPC (2 public + 2 private subnets), IGW, NAT, route tables, IAM roles, EKS cluster, managed node group, ECR repos + lifecycle
  - `variables.tf` — cluster_name, aws_region, vpc_cidr, instance_type, node_count, ECR options
  - `outputs.tf` — cluster_name, endpoint, vpc_id, subnet_ids, ecr_repository_urls
  - Uses aws provider ~> 5.30, image scanning on push, keep-last-10 lifecycle
- **Environments** (`environments/`)
  - `dev/main.tf` — Wires AKS or EKS module via `var.runtime` toggle, shared variables, runtime-agnostic outputs
  - `staging/main.tf` — Larger nodes (B4ms/t3.large), autoscaling (3-8 nodes), separate ACR/ECR prefixes

### 19. Makefile: Runtime Dispatch (`make/runtime.mk`)
- `runtime-up` / `runtime-down` / `runtime-status` — dispatch to `runtimes/$(PROFILE)/` based on PROFILE env var
- Works for k3d, aks, eks without changes

### 20. Provider Interface Contracts (`_interface.yaml`)
- 9 interface contract files documenting provider swappability:
  - `platform/ingress/_interface.yaml` — ingress provider contract (traefik, nginx)
  - `platform/monitoring/metrics/_interface.yaml` — metrics provider contract (prometheus, victoria-metrics)
  - `platform/monitoring/grafana/_interface.yaml` — visualization provider contract
  - `platform/chaos/_interface.yaml` — chaos engineering provider contract (chaos-mesh)
  - `platform/gitops/_interface.yaml` — gitops provider contract (argocd)
  - `platform/security/policy/_interface.yaml` — policy engine contract (kyverno)
  - `platform/security/tls/_interface.yaml` — TLS certificate contract (cert-manager)
  - `platform/security/secrets/_interface.yaml` — secrets management contract (sealed-secrets)
  - `platform/security/network-policies/_interface.yaml` — network segmentation contract

### 21. Helm Chart Test Templates
- `apps/go-api/deploy/helm/templates/tests/test-connection.yaml` — HTTP /health endpoint test
- `apps/echo-server/deploy/helm/templates/tests/test-connection.yaml` — HTTP /health endpoint test
- Run with `helm test <release-name>`

### 22. Security Hardening
- Both Dockerfiles (go-api, echo-server) now run as non-root user (`appuser:appgroup`)
- Helm values enforce `runAsNonRoot: true`, `runAsUser: 65534`, `readOnlyRootFilesystem: true`
- All Linux capabilities dropped (`capabilities.drop: [ALL]`)

### 23. Documentation (`docs/` + in-folder READMEs)
- **Root** (`README.md`) — Project overview, quickstart guide, make targets, configuration, URLs
- **Architecture** (`docs/architecture.md`) — Directory layout, strategy pattern, provider swappability, CLI architecture, naming conventions
- **CLI Reference** (`docs/cli-reference.md`) — Full command tree with all subcommands, flags, and usage examples
- **Scenarios Guide** (`docs/scenarios.md`) — How scenarios work, YAML format reference, all 4 scenarios documented
- **Cloud Runtimes** (`docs/cloud-runtimes.md`) — AKS/EKS setup, Terraform modules, remote state, cost estimates
- **CI/CD** (`docs/ci-cd.md`) — GitHub Actions workflows (CI, CD, Helm validation)
- **Apps Guide** (`apps/README.md`) — App conventions, how to add new apps, endpoint reference
- **Platform Guide** (`platform/README.md`) — Provider categories, swapping, interface contracts, adding new providers

---

## What Is NOT Built Yet

| Directory | Purpose | Status |
|-----------|---------|--------|
| `ui/` (full build) | React/Svelte frontend project | Placeholder HTML only (embedded via go:embed) |

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
9. **Services pattern**: `services/<name>/` follows same convention as platform providers (install/uninstall/status/values)
10. **Scenario YAML**: Declarative `scenario.yaml` with components, prerequisites, explore hints

---

## Git History (Chronological)

```
7f3eaca  first commit
7e8dcbe  Adding installation scripts
7d91bca  Adding makefile for cluster up and down
83b9cca  Added golang application
eccfa51  Created runtime and bootstrap environment in wsl
66ebd81  Added helm charts for go-api
d1e924d  Arranging Makefiles
aa339b4  Updating make help
6ec1bfa  Updating make help
eae9973  Fixing app deployment and destroy in makefile
a54198a  Updated manifests and configurations
2b76e28  Setup go app on k3d cluster
9d403c6  Updated environment variables and setup defaults
42d7840  Taking specific version of tools
2c961fc  Added monitoring to cluster and application through prometheus and grafana
```

---

## File Inventory

```
Total: ~240+ source files
  Shell scripts:   49  (+nginx install/uninstall/status, traefik uninstall/status, acr, ecr)
  YAML configs:    60  (+9 _interface.yaml contracts, nginx values.yaml, 2 values-cloud.yaml)
  Terraform:       10  (modules/aks: 3, modules/eks: 3, environments/dev: 1, environments/staging: 1, + provider configs)
  Make includes:    9  (vars, bootstrap, runtime, app, platform, check, cli, services, terraform)
  Go source (CLI): 26+ (main.go, go.mod, go.sum, 10 cmd files, 6 internal packages, 5 test files, ui/embed.go)
  Go source (apps): 6  (go-api: main.go/go.mod/go.sum, echo-server: main.go/go.mod/go.sum)
  Markdown docs:   12  (README.md, apps/README.md, platform/README.md, platform/monitoring/README.md, apps/go-api/README.md, apps/go-api/deploy/helm/README.md, docs/architecture.md, docs/cli-reference.md, docs/scenarios.md, docs/cloud-runtimes.md, docs/ci-cd.md, + .claude docs)
  Env files:        8  (.env.example, versions.env, go-api/app.env, echo-server/app.env, k3d/runtime.env, aks/runtime.env, eks/runtime.env)
  Dockerfiles:      2  (go-api, echo-server) — both run as non-root
  Helm charts:      2  (go-api: 11 templates + test, echo-server: 9 templates + test)
  HTML:             1  (ui/dist/index.html)
  JSON dashboards:  4  (SLO, Log Explorer, Security & Compliance, Chaos Engineering)
  Scenario YAMLs:   4  (observability-sre, gitops-cicd, security-compliance, chaos-engineering)
  CI/CD workflows:  6  (3 templates in delivery/ + 3 active in .github/workflows/)
  Go test files:    5  (config, executor, platform, scenario, services — 35 tests total)
```
