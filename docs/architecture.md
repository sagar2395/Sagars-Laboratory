# Architecture

This document describes the project structure, design patterns, and conventions used in Sagars-Laboratory.

## Design Principles

1. **Script-first, CLI-wraps** - Shell scripts are the source of truth. The `labctl` CLI wraps them via `os/exec`, not replaces them. You can always fall back to `make` or run scripts directly.
2. **Strategy dispatch** - Build and deploy strategies are selected per-app via `app.env`, not hardcoded. Adding a new strategy means adding one script file.
3. **Provider swappability** - Platform components are interchangeable. Change `INGRESS_PROVIDER=nginx` and re-run `platform-up` to switch from Traefik to Nginx.
4. **Declarative scenarios** - Scenarios are YAML files that declare what to install. The engine handles the orchestration.
5. **Idempotent operations** - All installs use `helm upgrade --install`. All scripts are safe to run multiple times.

## Directory Layout

```
Sagars-Laboratory/
  .env.example                 # Global config template (PROFILE, CLUSTER_NAME, providers)
  .env                         # Local config (gitignored, created from .env.example)
  versions.env                 # Pinned tool versions

  apps/<name>/                 # Applications
    app.env                    # Per-app config (build + deploy strategy)
    main.go / Dockerfile       # Application source
    deploy/helm/               # Helm chart + value profiles

  bootstrap/
    setup-tools.sh             # Version-pinned tool installer

  cmd/labctl/                  # CLI source (Go)
    main.go                    # Entry point
    cmd/                       # Cobra command definitions
    internal/                  # Business logic packages
    ui/                        # go:embed assets for web UI

  delivery/
    github-actions/            # CI/CD workflow templates

  engine/
    build.sh / deploy.sh       # Strategy dispatchers
    build/<strategy>.sh        # docker.sh, acr.sh, ecr.sh
    deploy/<strategy>.sh       # helm.sh

  foundation/terraform/
    modules/aks/               # AKS Terraform module
    modules/eks/               # EKS Terraform module
    environments/dev/          # Dev environment config
    environments/staging/      # Staging environment config

  make/                        # Modular Makefile includes
    vars.mk                    # Shared variables
    bootstrap.mk               # Tool setup targets
    runtime.mk                 # Cluster lifecycle targets
    app.mk                     # App build/deploy targets
    platform.mk                # Platform install targets
    cli.mk                     # CLI build targets
    services.mk                # Service management targets
    terraform.mk               # Terraform targets
    check.mk                   # Validation targets

  platform/<category>/<provider>/
    install.sh                 # Install via Helm
    uninstall.sh               # Remove via Helm
    status.sh                  # Health check
    values.yaml                # Helm values
    _interface.yaml            # Provider contract (at category level)

  runtimes/<profile>/
    up.sh                      # Create cluster
    down.sh                    # Destroy cluster
    runtime.env                # Runtime-specific variables

  scenarios/<name>/
    scenario.yaml              # Scenario definition
    values/                    # Helm value overrides
    manifests/                 # Kubernetes manifests
    dashboards/                # Grafana dashboard JSON files

  services/<name>/
    install.sh / uninstall.sh / status.sh / values.yaml

  ui/dist/
    index.html                 # Web UI (single-page, embedded in CLI binary)

  .github/workflows/           # Active GitHub Actions workflows
```

## Strategy Pattern

The engine layer selects build and deploy implementations at runtime based on `app.env`:

```
User: make deploy APP_NAME=go-api
  -> engine/deploy.sh deploy go-api
    -> source apps/go-api/app.env      # reads DEPLOY_STRATEGY=helm
    -> engine/deploy/helm.sh deploy go-api
      -> helm upgrade --install ...
```

### Build Strategies

| Strategy | Script | What It Does |
|----------|--------|-------------|
| `docker` | `engine/build/docker.sh` | Build image + import to k3d |
| `acr` | `engine/build/acr.sh` | Build + push to Azure Container Registry |
| `ecr` | `engine/build/ecr.sh` | Build + push to AWS ECR |

### Deploy Strategies

| Strategy | Script | What It Does |
|----------|--------|-------------|
| `helm` | `engine/deploy/helm.sh` | `helm upgrade --install` with values file |

To add a new strategy, create a script in the appropriate directory. No other changes required.

## Provider Swappability

Platform components are organized as `platform/<category>/<provider>/` directories. Each provider must implement:

| File | Purpose |
|------|---------|
| `install.sh` | Install the component (usually via Helm) |
| `uninstall.sh` | Remove the component |
| `status.sh` | Check health and print status |
| `values.yaml` | Helm chart configuration |

The active provider for each category is selected via environment variable:

```bash
INGRESS_PROVIDER=traefik    # or nginx
METRICS_PROVIDER=prometheus  # or victoria-metrics
```

Each category has an `_interface.yaml` contract file documenting:
- What Kubernetes resources the provider creates
- What endpoints it exposes
- What env variables it sets
- What implementations exist

### Current Providers

| Category | Providers |
|----------|-----------|
| ingress | traefik, nginx |
| monitoring/metrics | prometheus |
| monitoring/grafana | grafana |
| gitops | argocd |
| security/policy | kyverno |
| security/tls | cert-manager |
| security/secrets | sealed-secrets |
| security/network-policies | kubernetes-native |
| chaos | chaos-mesh |

## CLI Architecture

The `labctl` CLI is built in Go with Cobra. It wraps existing shell scripts rather than reimplementing them.

```
cmd/labctl/
  main.go                      # Entry: calls cmd.Execute()
  cmd/
    root.go                    # Root command, global flags, config init
    init_cmd.go                # labctl init / teardown / reset
    runtime.go                 # labctl runtime up/down/status
    app.go                     # labctl app build/deploy/destroy/list
    platform.go                # labctl platform up/down/status
    scenario.go                # labctl scenario list/up/down/status/info
    service.go                 # labctl service list/up/down/status
    check.go                   # labctl check tools/cluster/ingress
    status.go                  # labctl status
    ui.go                      # labctl ui (web dashboard)
  internal/
    config/                    # Config loading (.env, app.env, runtime.env)
    executor/                  # Shell command + script execution
    platform/                  # Platform provider registry
    k8s/                       # Kubernetes operations (via kubectl)
    scenario/                  # Scenario engine (YAML-based)
    services/                  # Service registry
    api/                       # HTTP API server + WebSocket
  ui/
    embed.go                   # go:embed for static UI assets
    dist/                      # Embedded files (copied at build time)
```

### Web UI

The web UI is a single HTML file (`ui/dist/index.html`) embedded into the CLI binary via `go:embed`. When you run `labctl ui`, it starts an HTTP server on port 3939 serving:

- `/api/*` - REST API endpoints (status, apps, platform, scenarios, services)
- `/api/ws` - WebSocket for real-time updates
- `/` - Static UI files (embedded or filesystem fallback for dev)

## Configuration Loading

Configuration is layered:

1. **Global** (`.env`) - PROFILE, CLUSTER_NAME, ports, provider selection
2. **Runtime** (`runtimes/<profile>/runtime.env`) - INGRESS_CLASS, STORAGE_CLASS, DOMAIN_SUFFIX, REGISTRY_TYPE
3. **Per-app** (`apps/<name>/app.env`) - APP_NAME, BUILD_STRATEGY, DEPLOY_STRATEGY, HELM_VALUES

The CLI's `config.Load()` reads all three layers, with runtime.env overriding global defaults.

## Helm Value Profiles

Each app ships multiple Helm value files:

| File | Purpose |
|------|---------|
| `values.yaml` | Chart defaults |
| `values-dev.yaml` | Local k3d (traefik, low resources, 1 replica) |
| `values-prod-like.yaml` | Production simulation (HPA, anti-affinity, probes) |
| `values-cloud.yaml` | Cloud runtimes (nginx, pullPolicy: Always) |
| `values-test.yaml` | CI testing |

Select via `HELM_VALUES=values-cloud.yaml` in `app.env`.

## Naming Conventions

| Thing | Convention | Example |
|-------|-----------|---------|
| App directory | lowercase, hyphenated | `apps/go-api/` |
| App config | always `app.env` | `apps/go-api/app.env` |
| Helm chart | `deploy/helm/` inside app | `apps/go-api/deploy/helm/` |
| Platform component | `platform/<category>/<provider>/` | `platform/ingress/traefik/` |
| Runtime | `runtimes/<profile>/` | `runtimes/k3d/` |
| Scenario | `scenarios/<name>/scenario.yaml` | `scenarios/observability-sre/` |
| Service | `services/<name>/` | `services/redis/` |
| Make module | `make/<domain>.mk` | `make/platform.mk` |
| Env variables | UPPER_SNAKE_CASE | `BUILD_STRATEGY` |
| CLI commands | noun-verb | `labctl app deploy` |
