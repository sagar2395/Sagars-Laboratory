# Sagars-Laboratory

A Kubernetes-based homelab for Platform Engineering and DevOps. Spin up infrastructure, deploy apps, activate testing scenarios (observability, GitOps, security, chaos engineering), and tear everything down in minutes.

## What's Inside

| Layer | What It Does |
|-------|-------------|
| **Runtimes** | k3d (local), AKS (Azure), EKS (AWS) cluster lifecycle |
| **Apps** | go-api (HTTP + metrics + tracing), echo-server (HTTP + Redis caching) |
| **Platform** | Ingress (Traefik/Nginx), Monitoring (Prometheus + Grafana), GitOps (ArgoCD), Security (Kyverno, cert-manager, Sealed Secrets, Network Policies), Chaos (Chaos Mesh) |
| **Scenarios** | Declarative playgrounds: Observability SRE, GitOps CI/CD, Security Compliance, Chaos Engineering |
| **Services** | Shared dependencies: Redis |
| **CLI** | `labctl` - single binary with embedded web UI dashboard |
| **IaC** | Terraform modules for AKS and EKS with dev/staging environments |
| **CI/CD** | GitHub Actions workflows for lint, test, build, deploy |

## Prerequisites

- Docker (running)
- One of: k3d (local) / Azure CLI (AKS) / AWS CLI (EKS)
- kubectl, Helm 3

Install all tools automatically:

```bash
make setup-tools            # installs for current PROFILE (default: k3d)
make setup-tools PROFILE=all  # installs everything
```

Tool versions are pinned in [`versions.env`](versions.env):

| Tool | Version |
|------|---------|
| kubectl | 1.29.3 |
| k3d | 5.8.3 |
| Helm | 3.14.0 |
| Docker | 29.2.1 |
| Terraform | 1.7.0 |
| Azure CLI | 2.57.0 |
| AWS CLI | 2 |

## Quickstart (Local with k3d)

```bash
# 1. Clone and configure
git clone <repo-url> && cd Sagars-Laboratory
cp .env.example .env          # edit if needed

# 2. One-command setup: creates cluster + installs platform
make init

# 3. Build and deploy an app
make build APP_NAME=go-api
make deploy APP_NAME=go-api

# 4. Verify
curl http://go-api.k3d.local/health
curl http://grafana.k3d.local              # admin/admin

# 5. Activate a scenario
labctl scenario up observability-sre

# 6. Tear down when done
make teardown
```

Or use the CLI directly:

```bash
make cli-build                # builds bin/labctl
./bin/labctl init             # creates cluster + platform
./bin/labctl app deploy go-api
./bin/labctl scenario list
./bin/labctl ui               # opens web dashboard at http://localhost:3939
```

## Project Structure

```
Sagars-Laboratory/
  .env.example              # Global config template
  .github/workflows/        # Active CI/CD workflows
  apps/                     # Application source + Helm charts
    go-api/                 # Go HTTP server with metrics + tracing
    echo-server/            # Go echo server with Redis caching
  bootstrap/                # Tool installation scripts
  cmd/labctl/               # CLI source (Go, Cobra)
  delivery/                 # CI/CD pipeline templates
  docs/                     # Documentation
  engine/                   # Build/deploy strategy dispatchers
  foundation/terraform/     # IaC modules (AKS, EKS) + environments
  make/                     # Modular Makefile includes
  platform/                 # Platform components (ingress, monitoring, security, chaos, gitops)
  runtimes/                 # Cluster lifecycle scripts (k3d, aks, eks)
  scenarios/                # Declarative playground definitions
  services/                 # Shared service deployments (Redis)
  ui/                       # Web UI frontend source
```

## Documentation

| Document | Description |
|----------|-------------|
| [Architecture](docs/architecture.md) | Project structure, design patterns, and conventions |
| [CLI Reference](docs/cli-reference.md) | All `labctl` commands and flags |
| [Scenarios Guide](docs/scenarios.md) | How scenarios work + available playgrounds |
| [Cloud Runtimes](docs/cloud-runtimes.md) | AKS/EKS setup with Terraform |
| [CI/CD](docs/ci-cd.md) | GitHub Actions workflows |
| [Apps Guide](apps/README.md) | How to add and manage applications |
| [Platform Guide](platform/README.md) | Platform components and provider swapping |

## Make Targets

Run `make help` for the full list. Key targets:

```
Lifecycle:
  make init                 # Setup tools + create cluster + install platform
  make teardown             # Destroy apps + platform + cluster
  make reset                # Teardown + init

Apps (set APP_NAME=<name>):
  make build                # Build app container image
  make deploy               # Deploy app to cluster via Helm
  make destroy-app          # Remove app from cluster
  make deploy-all           # Deploy all discovered apps

Platform:
  make platform-up          # Install all platform components
  make platform-down        # Remove all platform components
  make platform-status      # Show platform health

Services:
  make service-up SVC=redis # Install a shared service
  make service-list         # List available services

CLI:
  make cli-build            # Build labctl binary (bin/labctl)
  make cli-install          # Build + install to PATH

Terraform (set TF_ENV=dev|staging):
  make terraform-plan       # Plan infrastructure changes
  make terraform-apply      # Apply infrastructure
  make terraform-destroy    # Destroy infrastructure
```

## Configuration

Global settings live in `.env` (created from [`.env.example`](.env.example)):

```bash
PROFILE=k3d                    # Runtime: k3d | aks | eks
CLUSTER_NAME=sagars-cluster
INGRESS_PROVIDER=traefik       # traefik | nginx
METRICS_PROVIDER=prometheus
```

Per-app settings live in `apps/<name>/app.env`:

```bash
APP_NAME=go-api
BUILD_STRATEGY=docker          # docker | acr | ecr
DEPLOY_STRATEGY=helm
HELM_VALUES=values-dev.yaml    # values-dev | values-prod-like | values-cloud
```

Helm values profiles per environment:

| Profile | Ingress Class | Use Case |
|---------|--------------|----------|
| `values-dev.yaml` | traefik | Local k3d development |
| `values-prod-like.yaml` | traefik | Production simulation on k3d |
| `values-cloud.yaml` | nginx | AKS/EKS cloud deployments |
| `values-test.yaml` | traefik | Automated testing |

## Available Scenarios

Activate with `labctl scenario up <name>` or via the web UI:

| Scenario | What It Deploys |
|----------|----------------|
| `observability-sre` | Loki + Promtail + Tempo + alerting rules + SLO dashboards |
| `gitops-cicd` | ArgoCD + Application CRDs for both apps |
| `security-compliance` | Kyverno policies + cert-manager + network policies + security dashboard |
| `chaos-engineering` | Chaos Mesh + PDBs + 8 chaos experiments + chaos dashboard |

See [Scenarios Guide](docs/scenarios.md) for details.

## Key URLs (k3d)

| Service | URL | Credentials |
|---------|-----|-------------|
| go-api | http://go-api.k3d.local | - |
| echo-server | http://echo-server.k3d.local | - |
| Grafana | http://grafana.k3d.local | admin / admin |
| Prometheus | http://prometheus.k3d.local | - |
| ArgoCD | http://argocd.k3d.local | admin / (see install output) |
| labctl UI | http://localhost:3939 | - |

## License

Private project. Not licensed for redistribution.
