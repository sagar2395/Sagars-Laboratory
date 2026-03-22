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
- **Config values are provided by the executor environment** — scripts should NOT source `.active-runtime.env` or other file-based configs. Instead, use `${DOMAIN_SUFFIX:-k3d.local}` to read from env with a sensible default.
- **Dynamic ingress hosts**: Never use hardcoded domain names in static YAML files. Use inline heredocs with `$DOMAIN_SUFFIX` interpolation or `--set` overrides in Helm.
- **Helm repo add**: Always use `--force-update` flag to prevent stale cache issues.
- **Helm install**: Use `--wait --timeout 5m` for reliable readiness detection.
- **Stuck release cleanup**: Before `helm upgrade --install`, check for `pending-` state releases and delete them to avoid "another operation is in progress" errors.

---

## Key Config Files

| File | Purpose | Read By |
|------|---------|---------|
| `.env` | Global project config (PROFILE, CLUSTER_NAME, ports, providers) | Makefile, CLI |
| `.env.example` | Template for `.env` | Developers |
| `versions.env` | Tool version pins | bootstrap/setup-tools.sh |
| `apps/<name>/app.env` | Per-app config (strategies, helm settings) | engine/*.sh |
| `runtimes/<profile>/runtime.env` | Runtime-specific overrides | CLI config loader |

---

## Config Propagation Flow

```
.env (global)
  + runtimes/<profile>/runtime.env (runtime-specific)
  -> config.Load() [Go: os.Setenv for each key]
  -> exec.SetEnv() [Go: explicit on executor for key vars]
  -> buildEnv() [Go: os.Environ() + executor overrides]
  -> child process environment [shell scripts receive all vars]
```

Key propagated variables:
- `CLUSTER_NAME` — used by runtime up/down scripts
- `DOMAIN_SUFFIX` — used by ALL platform install scripts for ingress hosts
- `HTTP_PORT`, `HTTPS_PORT` — used by k3d cluster creation
- `INGRESS_CLASS`, `STORAGE_CLASS` — used by Helm values
- `PROFILE` — identifies active runtime (k3d, aks, eks)

**Important**: Scripts should NEVER source config files directly. They should read from environment variables with fallback defaults: `${DOMAIN_SUFFIX:-k3d.local}`.

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
2. ~~`go-api` runs as root (runAsUser: 0)~~ **FIXED** (Session #7)
3. ~~Grafana dashboards are placeholders~~ **PARTIALLY FIXED** — SLO and Log Explorer dashboards added via scenario
4. ~~No `.gitignore` file in the project root~~ **FIXED** (Session #1)
5. ~~`.env.example` has a stray `make build` on line 1~~ **FIXED** (Session #1)
6. ~~Hardcoded `k3d.local` domain in multiple places~~ **FIXED** — All scripts now use `$DOMAIN_SUFFIX` from executor env (Session #1 + #10)
7. ~~Platform install scripts don't accept provider variables~~ **FIXED** — Registry pattern + env vars (Session #1)
8. go-api needs to be rebuilt and redeployed after OpenTelemetry enhancement (new image required)
9. Web UI is a single HTML file — should be replaced with a proper frontend build
10. ~~No tests for the CLI or API server~~ **FIXED** — 35 tests across 5 packages (Session #8)
11. ~~echo-server app.env has commented-out REDIS_URL~~ **FIXED** (Session #8)
12. ~~GitHub Actions workflows are templates~~ **FIXED** (Session #8)
13. Chaos Mesh uses containerd socket at `/run/k3s/containerd/containerd.sock` — only works on k3d/k3s
14. PDB templates default to disabled — enable via values or scenario manifests
15. Terraform modules not tested against real cloud accounts
16. ~~Cloud runtime scripts assume Terraform state is local~~ **FIXED** (Session #8)
17. EKS module creates NAT Gateway (costs ~$32/month)
18. ~~Platform install scripts may need Nginx adjustment for cloud~~ **PARTIALLY FIXED** (Session #7)
19. ~~echo-server Helm defaults force Redis URL~~ **FIXED** (Session #9)
20. Deploy UX gap: no preflight validation for dependency URLs
21. ~~Traefik conflict: k3d-bundled Traefik in kube-system conflicts with custom install~~ **FIXED** — Disabled at cluster creation + HelmChart CRD cleanup (Session #10)
22. ~~DOMAIN_SUFFIX not reaching platform install scripts~~ **FIXED** — exec.SetEnv() propagation + removed dead .active-runtime.env sourcing (Session #10)
23. ~~Prometheus ingress hardcoded in static YAML~~ **FIXED** — Replaced with inline heredoc (Session #10)
24. ~~Kubernetes Dashboard Helm repo and OCI registry both broken~~ **FIXED** — Direct tarball install from GitHub release (Session #10)
25. ~~Kubernetes Dashboard Kong TLS causing Traefik proxy failure~~ **FIXED** — Disabled Kong TLS, HTTP ingress (Session #10)
26. ~~Scenario status race condition~~ **FIXED** — Explicit success broadcasts after markActive (Session #10)
27. ~~ArgoCD blocked by Kyverno policies~~ **FIXED** — Added namespace exclusions (Session #10)

