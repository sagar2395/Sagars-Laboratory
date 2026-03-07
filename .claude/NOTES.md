# Architecture Notes & Conventions

> Reference document for maintaining consistency across sessions.
> Last updated: 2026-03-07

---

## Directory Convention

```
Sagars-Laboratory/
  .claude/                  # AI session context (this folder)
  apps/<name>/              # Each app: source code + app.env + deploy/
  bootstrap/                # Tool installation scripts
  cmd/labctl/               # CLI + Web UI source (Go) [Phase 1]
  delivery/                 # CI/CD pipeline definitions [Phase 3]
  engine/                   # Build/deploy strategy dispatchers
  foundation/terraform/     # IaC modules for cloud [Phase 6]
  make/                     # Modular makefile includes
  platform/                 # Platform components (ingress, monitoring, gitops, security)
  runtimes/<profile>/       # Cluster lifecycle scripts per runtime
  scenarios/<name>/         # Playground scenario definitions [Phase 2+]
  services/<name>/          # Shared services (Redis, etc.) [Phase 3]
  ui/                       # Web UI frontend source [Phase 1B]
```

---

## Naming Conventions

| Thing | Convention | Example |
|-------|-----------|---------|
| App directory | lowercase, hyphenated | `apps/go-api/` |
| App config | always `app.env` | `apps/go-api/app.env` |
| Helm chart | lives in `deploy/helm/` | `apps/go-api/deploy/helm/` |
| Helm values | `values-<profile>.yaml` | `values-dev.yaml`, `values-prod-like.yaml` |
| Platform component | `platform/<category>/<provider>/` | `platform/ingress/traefik/` |
| Platform scripts | `install.sh`, `uninstall.sh`, `status.sh` | Always these three |
| Runtime | `runtimes/<profile>/` | `runtimes/k3d/`, `runtimes/aks/`, `runtimes/eks/` |
| Runtime scripts | `up.sh`, `down.sh`, `runtime.env` | Always these three |
| Terraform module | `foundation/terraform/modules/<cloud>/` | `modules/aks/`, `modules/eks/` |
| Terraform env | `foundation/terraform/environments/<env>/` | `environments/dev/`, `environments/staging/` |
| Build strategy | `engine/build/<strategy>.sh` | `docker.sh`, `acr.sh`, `ecr.sh` |
| Scenario | `scenarios/<name>/scenario.yaml` | `scenarios/observability-sre/scenario.yaml` |
| Service | `services/<name>/` | `services/redis/` |
| Service scripts | `install.sh`, `uninstall.sh`, `status.sh` | Same as platform |
| Make module | `make/<domain>.mk` | `make/platform.mk`, `make/services.mk` |
| Env variables | UPPER_SNAKE_CASE | `BUILD_STRATEGY`, `DEPLOY_STRATEGY` |
| CLI commands | noun-verb | `labctl app deploy`, `labctl scenario up` |

---

## Strategy Pattern (Core Architecture)

The engine layer uses a dispatch pattern:

```
User runs: make deploy APP_NAME=go-api
  -> engine/deploy.sh deploy go-api
    -> source apps/go-api/app.env (reads DEPLOY_STRATEGY=helm)
    -> engine/deploy/helm.sh deploy go-api
      -> helm upgrade --install ...
```

Adding a new strategy = adding a single script in `engine/build/` or `engine/deploy/`.
Adding a new app = adding a directory in `apps/` with `app.env` + source + deploy config.

---

## Provider Swappability Pattern (New in This Plan)

Platform components are organized as:
```
platform/<category>/<provider>/
  install.sh      # Helm install with values.yaml
  uninstall.sh    # Helm uninstall + cleanup
  status.sh       # Health check
  values.yaml     # Helm chart values
```

Active provider selected via env variable (e.g., `INGRESS_PROVIDER=traefik`).

The `_interface.yaml` file documents what each category must provide:
- What Kubernetes resources it creates
- What endpoints it exposes
- What env variables it sets for other components
- What implementations exist (chart references + notes)

All 9 platform categories now have `_interface.yaml` contracts:
- `platform/ingress/_interface.yaml`
- `platform/monitoring/metrics/_interface.yaml`
- `platform/monitoring/grafana/_interface.yaml`
- `platform/chaos/_interface.yaml`
- `platform/gitops/_interface.yaml`
- `platform/security/policy/_interface.yaml`
- `platform/security/tls/_interface.yaml`
- `platform/security/secrets/_interface.yaml`
- `platform/security/network-policies/_interface.yaml`

---

## Helm Conventions

- All installs use `helm upgrade --install` (idempotent)
- Namespace created separately via `kubectl create namespace --dry-run=client -o yaml | kubectl apply -f -`
- Values files are per-environment: `values-dev.yaml`, `values-prod-like.yaml`, `values-cloud.yaml`, `values-test.yaml`
- Templates support conditional sections: `{{- if .Values.feature.enabled }}`

---

## Shell Script Conventions

- Shebang: `#!/usr/bin/env bash`
- Error handling: `set -euo pipefail` (should be added to scripts that lack it)
- Color output: using ANSI codes for status messages
- All scripts should be idempotent (safe to run multiple times)
- Source shared variables from env files, don't hardcode

---

## Key Config Files

| File | Purpose | Read By |
|------|---------|---------|
| `.env` | Global project config (PROFILE, CLUSTER_NAME, ports) | Makefile, CLI |
| `.env.example` | Template for `.env` | Developers |
| `versions.env` | Tool version pins | bootstrap/setup-tools.sh |
| `apps/<name>/app.env` | Per-app config (strategies, helm settings) | engine/*.sh |
| `runtimes/<profile>/runtime.env` | Runtime-specific overrides [new] | CLI, platform scripts |

---

## Decisions Made

### CLI Language: Go
- Project already uses Go for the app
- Cobra is the standard Go CLI framework
- Can embed web UI assets via `embed.FS`
- Single binary deployment

### Web UI: Embedded in CLI binary
- `labctl ui` starts a local HTTP server
- Frontend built as static assets, embedded via Go `embed`
- `cmd/labctl/ui/embed.go` embeds `ui/dist/` directory (copied at build time by `make cli-build`)
- `cmd/labctl/ui/dist/.gitkeep` committed; actual HTML copied from `ui/dist/` at build time
- API layer in `cmd/labctl/internal/api/`
- Server auto-detects embedded content; falls back to filesystem path for dev mode
- No separate process to manage

### Scenario Format: YAML-based
- Declarative scenario definitions in `scenario.yaml`
- Engine handles helm, kubectl, grafana-dashboard component types
- Prerequisites validated before activation
- Explore tips printed after activation

### Tool Swappability: Provider pattern
- Env variable selects active provider per category
- All providers follow same interface (install/uninstall/status/values)
- Switching = change variable + re-run platform-up

---

## Known Technical Debt

1. ~~Some shell scripts lack `set -euo pipefail`~~ **FIXED** (Session #1)
2. ~~`go-api` runs as root (runAsUser: 0)~~ **FIXED** — Both apps now use non-root user (appuser:appgroup, runAsUser: 65534, readOnlyRootFilesystem) (Session #7)
3. ~~Grafana dashboards are placeholders~~ **PARTIALLY FIXED** — SLO and Log Explorer dashboards added via scenario
4. ~~No `.gitignore` file in the project root~~ **FIXED** (Session #1)
5. ~~`.env.example` has a stray `make build` on line 1~~ **FIXED** (Session #1)
6. ~~Hardcoded `k3d.local` domain in multiple places~~ **FIXED** — Parameterized via `${DOMAIN_SUFFIX:-k3d.local}` (Session #1)
7. ~~Platform install scripts don't accept provider variables~~ **FIXED** — Registry pattern + env vars (Session #1)
8. go-api needs to be rebuilt and redeployed after OpenTelemetry enhancement (new image required)
9. Web UI is a single HTML file — should be replaced with a proper frontend build
10. ~~No tests for the CLI or API server~~ **FIXED** — 35 tests across 5 packages (config, executor, platform, scenario, services) (Session #8)
11. ~~echo-server app.env has commented-out REDIS_URL — actual default is set in values-dev.yaml~~ **FIXED** — Clarified with comments that runtime vars in app.env are documentation only; real values injected via Helm (Session #8)
12. ~~GitHub Actions workflows are templates — need to be placed in `.github/workflows/` and customized per repo~~ **FIXED** — Created 3 workflows in `.github/workflows/` (ci.yaml, cd.yaml, helm-validation.yaml) customized for multi-app project (Session #8)
13. Chaos Mesh uses containerd socket at `/run/k3s/containerd/containerd.sock` — only works on k3d/k3s, cloud runtimes need different socket paths
14. PDB templates exist in both Helm charts but default to disabled — enable via values or scenario manifests
15. Terraform modules not tested against real cloud accounts — need validation with actual AKS/EKS provisioning
16. ~~Cloud runtime scripts assume Terraform state is local — backend config blocks are commented out (need S3/AzureRM backend for teams)~~ **FIXED** — Both dev and staging environments now have commented-out azurerm + S3 backend config blocks (Session #8)
17. EKS module creates NAT Gateway (costs ~$32/month) — consider NAT instance for cost savings in dev
18. ~~Platform install scripts may need adjustment for cloud runtimes (e.g., Traefik -> Nginx ingress)~~ **PARTIALLY FIXED** — Nginx ingress provider created (Session #7)
19. ~~echo-server Helm defaults force Redis URL even when Redis service is absent, causing `/ready` failures~~ **FIXED** — `REDIS_URL` now defaults to empty in values files; Redis remains optional unless explicitly configured (Session #9, 2026-03-07)
20. Deploy UX gap: no preflight validation currently warns when dependency URLs (for example Redis) point to non-existent in-cluster services; add deploy-time warning/check.
21. Operational observation: `kube-system` Traefik helm-install job pods were seen in CrashLoop/Error during troubleshooting; verify if these are stale and clean up if needed.

