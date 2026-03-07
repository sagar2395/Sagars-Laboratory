# Next Steps: What To Work On

> This file tracks what to do next when starting a new session.
> Last updated: 2026-03-07

---

## Current Phase: ALL PHASES COMPLETE (1-6) + Hardening + Polishing + Documentation + Operational Testing DONE.

### Post-session Follow-ups (Remaining)

- [ ] Add an explicit dependency note in echo-server deploy docs: install Redis (`labctl service up redis`) before enabling `REDIS_URL`.
- [ ] Add a preflight warning in deploy flow when `REDIS_URL` is set but Redis service DNS is not resolvable.
- [ ] Consider adding optional Helm value flag in echo-server chart (e.g., `redis.enabled`) to make intent explicit.
- [ ] Add a quick smoke-check command/target to verify `/health` and `/ready` immediately after app deploy.
- [ ] Replace single-file HTML UI with proper React/Svelte build system.
- [ ] Add /etc/hosts management automation for k3d.local domains (or document the manual step prominently).
- [ ] Test AKS/EKS cloud runtimes with real cloud accounts.
- [ ] Verify Chaos Mesh containerd socket path works on cloud runtimes.

### Pre-work: Quick Fixes — DONE

- [x] Add `.gitignore` to project root
- [x] Fix `.env.example` line 1 (has stray `make build` text)
- [x] Add `set -euo pipefail` to scripts that lack it (all 12 scripts fixed)
- [x] Parameterize hardcoded `k3d.local` in platform install scripts
- [x] Add `INGRESS_PROVIDER`, `METRICS_PROVIDER` etc. to `.env.example`
- [x] Create `runtimes/k3d/runtime.env` extracting k3d-specific defaults

---

### Phase 1A: CLI — DONE

All items complete. `bin/labctl` compiles and runs.

1. [x] Go module with cobra, viper, gorilla/mux, gorilla/websocket, yaml.v3
2. [x] Config package (loads .env, app.env, runtime.env)
3. [x] Executor package (RunScript, RunCommand, RunHelm, RunKubectl)
4. [x] Platform registry (discovers providers, routes install/uninstall/status)
5. [x] K8s client (cluster info, app status, namespace checks via kubectl)
6. [x] 9 CLI command files (root, init_cmd, runtime, app, platform, check, status, scenario, ui)
7. [x] API server + handlers (REST + WebSocket)
8. [x] `make/cli.mk` (cli-build, cli-install, cli-tidy, cli-clean)

### Phase 1B: Web UI — FUNCTIONAL (embedded via go:embed)

1. [x] HTML dashboard with cluster, platform, apps, and scenarios cards
2. [x] REST API integration + WebSocket real-time updates
3. [x] Scenario activate/deactivate buttons in UI
4. [ ] Replace with proper React/Svelte build system
5. [x] Embed static assets via `go:embed` for single-binary distribution

---

### Phase 2: Scenario Framework + Observability — DONE

1. [x] **Scenario engine** — `cmd/labctl/internal/scenario/engine.go`
2. [x] **Observability SRE scenario** — `scenarios/observability-sre/`
3. [x] **CLI wired** — `labctl scenario list/up/down/status/info` all functional
4. [x] **API wired** — full REST endpoints for scenarios
5. [x] **go-api enhanced** — OTel, slog, labeled metrics, /toggle-failure
6. [x] **UI updated** — Scenario section with activate/deactivate

---

### Phase 3: GitOps + CI/CD — DONE

1. [x] **ArgoCD platform provider** — `platform/gitops/argocd/`
2. [x] **Redis shared service** — `services/redis/`
3. [x] **Echo-server app** — `apps/echo-server/`
4. [x] **GitOps scenario** — `scenarios/gitops-cicd/`
5. [x] **GitHub Actions workflow templates** — `delivery/github-actions/`
6. [x] **Services support** — CLI + API + Makefile integration
7. [x] **Build verified** — all compiles

---

### Phase 4: Security & Compliance — DONE

1. [x] **Kyverno policy provider** — `platform/security/policy/kyverno/`
   - `install.sh` — Adds kyverno Helm repo, installs kyverno chart
   - `uninstall.sh` — Helm uninstall + CRD cleanup
   - `status.sh` — Pods, cluster policies, policy reports summary
   - `values.yaml` — Admission/background/cleanup/reports controllers, minimal resources
2. [x] **cert-manager TLS provider** — `platform/security/tls/cert-manager/`
   - `install.sh` — Installs cert-manager with CRDs, creates self-signed + CA ClusterIssuers
   - `uninstall.sh` — Helm uninstall + CRD cleanup
   - `status.sh` — Pods, ClusterIssuers, Certificates, CertificateRequests
   - `values.yaml` — Minimal resources, Prometheus ServiceMonitor enabled
   - `cluster-issuer.yaml` — Self-signed issuer + lab CA chain
3. [x] **Sealed Secrets provider** — `platform/security/secrets/sealed-secrets/`
   - `install.sh` — Installs sealed-secrets controller with usage instructions
   - `uninstall.sh` — Helm uninstall + CRD cleanup
   - `status.sh` — Pods, services, sealed secrets, public key
   - `values.yaml` — Minimal resources, Prometheus ServiceMonitor
4. [x] **Network policies** — `platform/security/network-policies/`
   - `install.sh` — Applies all network policy manifests
   - `uninstall.sh` — Removes all network policy manifests
   - `status.sh` — Lists network policies and opted-in namespaces
   - `default-deny.yaml` — Deny all ingress+egress in app namespaces
   - `allow-dns.yaml` — Allow DNS resolution (port 53)
   - `allow-monitoring.yaml` — Allow Prometheus scraping from monitoring namespace
   - `allow-ingress.yaml` — Allow Traefik ingress + echo-server->Redis egress
5. [x] **Security scenario** — `scenarios/security-compliance/`
   - `scenario.yaml` — 5 components: Kyverno, cert-manager, kyverno-policies, network-policies, dashboards
   - `values/kyverno.yaml` — Scenario-specific Kyverno values
   - `values/cert-manager.yaml` — Scenario-specific cert-manager values
   - `manifests/kyverno-policies.yaml` — 6 ClusterPolicies:
     - disallow-privileged-containers (Enforce)
     - require-labels (Audit)
     - disallow-root-user (Audit)
     - disallow-host-path (Enforce)
     - require-resource-limits (Audit)
     - disallow-latest-tag (Audit)
   - `manifests/network-policies.yaml` — Full set for go-api + echo-server namespaces
   - `dashboards/security-dashboard.json` — Grafana dashboard: policy violations (enforce/audit), passing count, cert-manager certificates, policy results over time, violations by policy table, cert expiry, network policy count by namespace, admission latency, controller resources
6. [x] **Platform auto-discovery** — All 4 security providers show in `labctl platform status`
7. [x] **Build verified** — `bin/labctl` compiles, 3 scenarios discovered, security info works

---

### Phase 5: Chaos Engineering — DONE

1. [x] **Chaos Mesh platform provider** — `platform/chaos/chaos-mesh/`
   - `install.sh` — Adds chaos-mesh Helm repo, installs chart, waits for controller + dashboard
   - `uninstall.sh` — Cleans up experiments, Helm uninstall + CRD removal
   - `status.sh` — Pods, active experiments (PodChaos/NetworkChaos/StressChaos), schedules, workflows
   - `values.yaml` — Controller manager, chaos daemon (containerd/k3s socket), dashboard, ServiceMonitor
2. [x] **PodDisruptionBudgets** — Added PDB templates + values to both Helm charts
   - `apps/go-api/deploy/helm/templates/pdb.yaml` — Conditional PDB (minAvailable/maxUnavailable)
   - `apps/echo-server/deploy/helm/templates/pdb.yaml` — Same pattern
   - Default: disabled. Enabled via `podDisruptionBudget.enabled: true`
3. [x] **Chaos experiments** — `scenarios/chaos-engineering/manifests/chaos-experiments.yaml`
   - 8 experiments across 3 Chaos Mesh CRD types:
     - PodChaos: pod-kill (go-api), pod-kill (echo-server), pod-failure (go-api)
     - NetworkChaos: delay (echo-server→Redis 500ms), partition (go-api↔traefik), loss (echo-server→Redis 50%)
     - StressChaos: CPU stress (go-api, 2 workers 80% load), memory stress (echo-server, 200Mi)
   - `manifests/pod-disruption-budgets.yaml` — PDB manifests applied as scenario component
4. [x] **Chaos Grafana dashboard** — `dashboards/chaos-dashboard.json`
   - 14 panels: active experiments stat, total experiments stat, pod restarts stat, unavailable replicas stat, HTTP request rate (go-api + echo-server), HTTP latency p50/p95/p99 (go-api + echo-server), pod restart timeline, ready vs desired replicas, CPU usage + limits, memory usage + limits, network errors, OOMKill events
5. [x] **Chaos scenario** — `scenarios/chaos-engineering/scenario.yaml`
   - 4 components: chaos-mesh (helm), pod-disruption-budgets (manifest), chaos-experiments (manifest), chaos-dashboards (grafana-dashboard)
   - 7 explore commands: port-forward, list experiments, run pod-kill, watch pods, generate traffic, cleanup, check PDB
   - 8 explore tips for chaos experimentation
6. [x] **Platform auto-discovery** — Chaos Mesh shows in `labctl platform status`
7. [x] **Build verified** — `bin/labctl` compiles, 4 scenarios discovered

---

### Phase 6: Cloud Abstraction — DONE

1. [x] **AKS runtime** — `runtimes/aks/`
   - `up.sh` — az login check, resource group creation, terraform init/apply, az aks get-credentials
   - `down.sh` — terraform destroy (or direct az aks delete fallback), kubeconfig cleanup
   - `runtime.env` — INGRESS_CLASS=nginx, STORAGE_CLASS=managed-csi, DOMAIN_SUFFIX=sagarslab.io, REGISTRY_TYPE=acr
2. [x] **EKS runtime** — `runtimes/eks/`
   - `up.sh` — aws sts check, terraform init/apply, aws eks update-kubeconfig
   - `down.sh` — terraform destroy (or eksctl delete fallback), kubeconfig cleanup
   - `runtime.env` — INGRESS_CLASS=nginx, STORAGE_CLASS=gp3, DOMAIN_SUFFIX=sagarslab.io, REGISTRY_TYPE=ecr
3. [x] **Terraform AKS module** — `foundation/terraform/modules/aks/`
   - main.tf: AKS cluster (Calico, SystemAssigned), Log Analytics, ACR, AcrPull role
   - variables.tf: 11 variables (cluster, resource group, VM size, autoscaling, ACR)
   - outputs.tf: 9 outputs (cluster_name, kube_config_raw, acr_login_server, etc.)
4. [x] **Terraform EKS module** — `foundation/terraform/modules/eks/`
   - main.tf: VPC (2 public + 2 private subnets), IGW, NAT, route tables, IAM roles (cluster + node), EKS cluster, managed node group, ECR repos + lifecycle
   - variables.tf: 12 variables (cluster, aws_region, vpc_cidr, instance_type, ECR)
   - outputs.tf: 9 outputs (cluster_name, endpoint, vpc_id, ecr_repository_urls, etc.)
5. [x] **Terraform environments** — `foundation/terraform/environments/`
   - `dev/main.tf` — Wires AKS or EKS module via `var.runtime` toggle, B2s/t3.medium, 2 nodes
   - `staging/main.tf` — B4ms/t3.large, autoscaling 3-8 nodes
6. [x] **Cloud build strategies** — `engine/build/`
   - `acr.sh` — az acr login, docker build, push to Azure Container Registry
   - `ecr.sh` — aws ecr login, docker build, push to AWS ECR
7. [x] **Runtime dispatch** — `make/runtime.mk` dispatches to `runtimes/$(PROFILE)/`
8. [x] **Bootstrap updated** — `setup-tools.sh` adds eks profile (aws-cli + terraform + helm)
9. [x] **versions.env updated** — Added TERRAFORM_VERSION=1.7.0, AWS_CLI_VERSION=2
10. [x] **.env.example updated** — Added cloud runtime config section (Azure + AWS variables)
11. [x] **Build verified** — `bin/labctl` compiles, 4 scenarios, 8 platform providers

---

## All Phases Summary

| Phase | Name | Status |
|-------|------|--------|
| Pre-work | Quick Fixes | DONE |
| Phase 1A | CLI (`labctl`) | DONE |
| Phase 1B | Web UI | FUNCTIONAL (embedded via go:embed) |
| Phase 2 | Scenarios + Observability | DONE |
| Phase 3 | GitOps + CI/CD + Services | DONE |
| Phase 4 | Security & Compliance | DONE |
| Phase 5 | Chaos Engineering | DONE |
| Phase 6 | Cloud Abstraction | DONE |
| Hardening | Contracts, Nginx, Security, Tests | DONE |
| Polishing | Tests, go:embed, CI/CD, Cloud Values | DONE |
| Documentation | README, Architecture, CLI, Scenarios, Cloud, CI/CD, Apps, Platform docs | DONE |

---

### Hardening Round — DONE

1. [x] **_interface.yaml contracts** — 9 files documenting provider swappability
   - `platform/ingress/_interface.yaml` (traefik, nginx)
   - `platform/monitoring/metrics/_interface.yaml` (prometheus, victoria-metrics)
   - `platform/monitoring/grafana/_interface.yaml` (grafana)
   - `platform/chaos/_interface.yaml` (chaos-mesh)
   - `platform/gitops/_interface.yaml` (argocd)
   - `platform/security/policy/_interface.yaml` (kyverno)
   - `platform/security/tls/_interface.yaml` (cert-manager)
   - `platform/security/secrets/_interface.yaml` (sealed-secrets)
   - `platform/security/network-policies/_interface.yaml` (kubernetes-native)
2. [x] **Nginx ingress provider** — `platform/ingress/nginx/`
   - install.sh, uninstall.sh, status.sh, values.yaml
   - ServiceMonitor for Prometheus, admission webhooks enabled
3. [x] **Traefik backfill** — Added missing uninstall.sh and status.sh to `platform/ingress/traefik/`
4. [x] **Terraform make targets** — `make/terraform.mk`
   - terraform-init, terraform-plan, terraform-apply, terraform-destroy, terraform-output, terraform-status
   - TF_ENV variable for environment selection (dev/staging)
5. [x] **Non-root security fix** — Both apps now run as non-root
   - Dockerfiles: added appuser/appgroup, USER directive, /app workdir
   - Helm values: runAsNonRoot=true, runAsUser=65534, readOnlyRootFilesystem=true
6. [x] **Helm chart tests** — Test connection templates for both charts
   - `apps/go-api/deploy/helm/templates/tests/test-connection.yaml`
   - `apps/echo-server/deploy/helm/templates/tests/test-connection.yaml`
7. [x] **Build verified** — CLI compiles, all Helm templates render, make help shows terraform targets

---

### Polishing Round — DONE

1. [x] **Go tests** — 35 tests across 5 packages
   - `internal/config/config_test.go` — 6 tests: ListApps, LoadAppConfig, Load defaults
   - `internal/executor/executor_test.go` — 9 tests: New, SetEnv, RunCommand, CaptureOutput, RunScript, buildEnv
   - `internal/platform/registry_test.go` — 6 tests: Discovery, GetProvider, HasScript, edge cases
   - `internal/scenario/engine_test.go` — 8 tests: Discovery, Get, ResolveTemplate, Status, active state, invalid YAML
   - `internal/services/registry_test.go` — 7 tests: Discovery, Get, HasScript, edge cases
2. [x] **go:embed for UI** — Single-binary now includes embedded UI assets
   - `cmd/labctl/ui/embed.go` — `//go:embed all:dist`
   - `cmd/labctl/ui/dist/.gitkeep` — Placeholder for build-time copy
   - `make/cli.mk` updated — Copies `ui/dist/` before Go build
   - `internal/api/server.go` — Accepts `io/fs.FS`, auto-detects embedded content with filesystem fallback
3. [x] **GitHub Actions workflows** — Active CI/CD in `.github/workflows/`
   - `ci.yaml` — Multi-app CI: lint apps + CLI, test apps + CLI, build images, helm lint (matrix strategy)
   - `cd.yaml` — Multi-app CD: detect changed apps, build+push to GHCR, update Helm values
   - `helm-validation.yaml` — PR validation: dynamic chart discovery, lint + template validation
4. [x] **app.env documentation** — Clarified runtime vars are reference-only (injected via Helm)
5. [x] **Terraform backend configs** — Added commented-out azurerm + S3 backends to staging environment
6. [x] **Cloud Helm values** — `values-cloud.yaml` for both go-api and echo-server
   - nginx ingress class, pullPolicy: Always, probes, anti-affinity, cloud-appropriate resources
7. [x] **Build verified** — CLI compiles (15MB with embedded UI), all 35 tests pass with race detector, Helm lint passes for all profiles

### Documentation Round — DONE

1. [x] **Root README.md** — Project overview, quickstart (k3d), project structure tree, make targets, configuration guide, scenarios table, key URLs
2. [x] **docs/architecture.md** — Design principles, full directory layout, strategy dispatch pattern, provider swappability, CLI architecture, web UI, config loading, Helm value profiles, naming conventions
3. [x] **docs/cli-reference.md** — Full labctl command tree (lifecycle, runtime, app, platform, scenario, service, check, UI), global flags, CLI vs Make comparison
4. [x] **docs/scenarios.md** — How scenarios work, using scenarios, all 4 scenarios documented with explore hints, YAML format reference, template variables, component types, creating new scenarios
5. [x] **docs/cloud-runtimes.md** — Runtime profiles, Azure/AWS prerequisites, Terraform modules (AKS/EKS), environment sizes, provisioning commands, remote state backends, building for cloud, cost estimates, troubleshooting
6. [x] **docs/ci-cd.md** — 3 GitHub Actions workflows (CI, CD, Helm validation), job descriptions, secrets configuration, customization guide
7. [x] **apps/README.md** — App directory structure, app.env contract, build/deploy commands, Helm value profiles, adding new apps guide, go-api + echo-server endpoint tables, security practices
8. [x] **platform/README.md** — Components by category (ingress, monitoring, gitops, security, chaos), interface contracts, adding providers guide, directory structure

---

## How To Start a New Session

When resuming work on this project, read these files in order:

1. `.claude/PROJECT_STATE.md` — what exists today
2. `.claude/PLAN.md` — the full implementation roadmap
3. `.claude/NOTES.md` — conventions and architecture decisions
4. `.claude/NEXT_STEPS.md` — this file, what to do right now

Then check:
- Which phase are we in?
- What items are checked off above?
- Pick up from the first unchecked item.

---

## Session Log

| Date | Session | What Was Done |
|------|---------|--------------|
| 2026-03-05 | #1 | Full project review. Created .claude/ documentation. Designed 6-phase roadmap with CLI, Web UI, scenario framework, tool swappability. Plan approved. |
| 2026-03-05 | #1 (cont.) | **Pre-work**: Created .gitignore, fixed .env.example, created runtimes/k3d/runtime.env, fixed all 12 shell scripts (set -euo pipefail + parameterized k3d.local). **Phase 1A**: Built entire `labctl` CLI — Go module, 5 internal packages (config, executor, platform, k8s, api), 9 command files, main.go. Binary compiles and runs. Created make/cli.mk. **Phase 1B started**: Created functional HTML dashboard. |
| 2026-03-05 | #2 | **Phase 2 complete**: Built scenario engine (YAML-based, 4 component types). Created observability-sre scenario (Loki, Promtail, Tempo, alerting rules, 2 Grafana dashboards). Wired CLI commands (list/up/down/status/info) and API handlers. Enhanced go-api with OpenTelemetry, slog, labeled metrics, /toggle-failure. Updated UI with scenario activate/deactivate. |
| 2026-03-05 | #3 | **Phase 3 complete**: Created ArgoCD provider, Redis shared service, echo-server app (Go HTTP echo + Redis caching, full Helm chart), GitOps scenario, GitHub Actions workflows. Built services registry + CLI commands + API endpoints + Makefile targets. |
| 2026-03-05 | #4 | **Phase 4 complete**: Created 4 security platform providers — Kyverno (policy enforcement), cert-manager (TLS with self-signed CA chain), Sealed Secrets (encrypted secrets in Git), Network Policies (namespace isolation with default-deny + explicit allows). Created security-compliance scenario with 6 Kyverno ClusterPolicies, network policy manifests, and Grafana security dashboard. All auto-discovered by platform registry. |
| 2026-03-05 | #5 | **Phase 5 complete**: Created Chaos Mesh platform provider (install/uninstall/status/values with containerd socket for k3d). Added PDB templates to both go-api and echo-server Helm charts. Created 8 chaos experiments (pod-kill, pod-failure, network-delay, network-partition, network-loss, cpu-stress, memory-stress) using Chaos Mesh CRDs. Created 14-panel Grafana chaos dashboard correlating experiments with pod restarts, HTTP rates, latency, CPU/memory, network errors, OOMKills. Created chaos-engineering scenario with 4 components, 7 explore commands. All 4 scenarios discovered, build verified. |
| 2026-03-06 | #6 | **Phase 6 complete**: Created AKS + EKS cloud runtimes (up.sh, down.sh, runtime.env each). Created Terraform modules for AKS (cluster + Log Analytics + ACR) and EKS (VPC + NAT + IAM + cluster + node group + ECR). Created dev + staging environment configs with runtime toggle. Added ACR + ECR cloud build strategies. Updated runtime.mk for profile-based dispatch, bootstrap/setup-tools.sh with eks profile (aws-cli + terraform), versions.env with Terraform + AWS CLI versions. Updated .env.example with cloud config section. **ALL 6 PHASES COMPLETE.** |
| 2026-03-06 | #7 | **Hardening round**: Created 9 _interface.yaml provider contracts for all platform categories. Created Nginx ingress provider (install/uninstall/status/values). Backfilled missing Traefik uninstall.sh + status.sh. Created make/terraform.mk with 6 targets. Fixed non-root security in both Dockerfiles + Helm values (runAsNonRoot, readOnlyRootFilesystem, drop all capabilities). Added Helm chart test templates for both apps. All builds verified. |
| 2026-03-06 | #8 | **Polishing round**: Created 35 Go tests across 5 CLI packages (config, executor, platform, scenario, services). Implemented go:embed for single-binary UI (embed.go, .gitkeep, Makefile copy step, server FS fallback). Created 3 active GitHub Actions workflows (.github/workflows/ ci.yaml, cd.yaml, helm-validation.yaml) with multi-app matrix strategy. Clarified app.env runtime vars as reference-only. Added Terraform backend configs to staging. Created values-cloud.yaml for both apps (nginx ingress, cloud-appropriate settings). All builds and tests verified. |
| 2026-03-06 | #9 | **Documentation round**: Created 8 documentation files — root README.md (quickstart, structure, make targets, configuration, URLs), docs/architecture.md (strategy pattern, provider swappability, CLI architecture, naming conventions), docs/cli-reference.md (full command tree, CLI vs Make comparison), docs/scenarios.md (framework guide, all 4 scenarios, YAML format reference), docs/cloud-runtimes.md (AKS/EKS setup, Terraform modules, remote state, costs), docs/ci-cd.md (3 GitHub Actions workflows), apps/README.md (app conventions, adding new apps), platform/README.md (provider categories, swapping, interface contracts). Updated .claude/PROJECT_STATE.md with documentation section and file inventory. |
| 2026-03-07 | #10 | **Operational testing round (Part 1)**: Fixed 9 real-world issues found during UI testing. Scenarios: switched all engine operations to streamed execution for WebSocket output. Runtime: created runtime manager with cluster name support. Platform: created Loki+Promtail and Tempo components, added Grafana datasources. UI: conditional Install/Remove buttons, full-width platform card, Grafana deep-links for logs. Config: added exec.SetEnv() propagation from root.go. Fixed ArgoCD URL, added ServiceExists helper. |
| 2026-03-07 | #10 (cont.) | **Operational testing round (Part 2)**: Fixed persistent 404s — root cause was dual Traefik (k3d-bundled + custom) and config not reaching scripts. Disabled bundled Traefik via `--k3s-arg "--disable=traefik@server:*"` in k3d up.sh. Added HelmChart CRD cleanup. Replaced Prometheus static ingress.yaml with dynamic heredoc. Fixed Grafana dynamic ingress. Fixed K8s Dashboard: disabled Kong TLS, HTTP ingress, direct tarball install (both Helm repo and OCI broken). Fixed Kyverno policy exclusions for ArgoCD. Fixed scenario status race condition with explicit success broadcasts. Added stuck release cleanup for Loki/Tempo. Made cluster name consistent via .env config. Updated all .claude/ documentation files. |
