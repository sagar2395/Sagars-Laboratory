# CI/CD

This project includes GitHub Actions workflows for continuous integration and deployment.

## Workflows

Three active workflows live in `.github/workflows/`:

### CI Pipeline (`ci.yaml`)

**Triggers:** Push to `main` or `feature/*` branches, PRs to `main` (when app or CLI code changes).

**Jobs:**

| Job | What It Does | Runs On |
|-----|-------------|---------|
| `lint-apps` | `go vet` + `staticcheck` for each app | Matrix: go-api, echo-server |
| `lint-cli` | `go vet` + `staticcheck` for labctl CLI | Single |
| `test-apps` | `go test -v -race` with coverage for each app | Matrix: go-api, echo-server |
| `test-cli` | `go test -v -race` for CLI internal packages + build check | Single |
| `build-images` | `docker build` for each app (after lint + test) | Matrix: go-api, echo-server |
| `helm-lint` | `helm lint` + `helm template` for each app's chart | Matrix: go-api, echo-server |

### CD Pipeline (`cd.yaml`)

**Triggers:** Push to `main` branch (when app code changes).

**Jobs:**

| Job | What It Does |
|-----|-------------|
| `detect-changes` | Diffs HEAD~1 to find which apps changed |
| `build-and-push` | Builds Docker images and pushes to GitHub Container Registry (ghcr.io) |
| `update-manifests` | Updates image tag in `values-dev.yaml` and commits the change |

The CD pipeline uses GHCR for image hosting. Images are tagged with both the commit SHA and `latest`.

### Helm Validation (`helm-validation.yaml`)

**Triggers:** PRs that touch Helm chart files, platform configs, or scenario values.

**Jobs:**

| Job | What It Does |
|-----|-------------|
| `discover-charts` | Finds all `Chart.yaml` files in the repo |
| `validate` | Runs `helm lint` + `helm template` for each discovered chart |

## How It Works

### CI Flow

```
Push to feature/* or PR to main
  -> lint-apps (go vet, staticcheck) [parallel per app]
  -> lint-cli (go vet, staticcheck)
  -> test-apps (go test -race) [parallel per app]
  -> test-cli (go test -race + build check)
  -> build-images (docker build) [depends on lint + test]
  -> helm-lint (helm lint + template) [parallel per app]
```

### CD Flow

```
Push to main (app changes only)
  -> detect-changes (which apps changed?)
  -> build-and-push (docker build + push to ghcr.io) [per changed app]
  -> update-manifests (sed image tag in values-dev.yaml + git commit)
  -> ArgoCD auto-syncs on manifest change (if gitops-cicd scenario is active)
```

## Configuration

### Secrets

The CD pipeline uses `GITHUB_TOKEN` (automatically provided) for:
- Pushing images to GitHub Container Registry
- Committing updated manifests back to the repo

No additional secrets are needed for the basic workflow. If ArgoCD sync triggering is enabled, set `ARGOCD_AUTH_TOKEN`.

### Customization

#### Adding a new app to CI

The CI workflow uses a matrix strategy. To add a new app, edit `.github/workflows/ci.yaml`:

```yaml
strategy:
  matrix:
    app: [go-api, echo-server, my-new-app]  # add here
```

#### Changing the container registry

Edit `.github/workflows/cd.yaml`:

```yaml
env:
  REGISTRY: ghcr.io    # change to your registry
```

For Azure ACR or AWS ECR, update the login step to use the appropriate action.

## Templates

The `delivery/github-actions/` directory contains the original workflow templates that were customized for this project:

| Template | Description |
|----------|-------------|
| `ci.yaml` | Single-app CI template |
| `cd.yaml` | Single-app CD template with ArgoCD sync |
| `helm-release.yaml` | Helm chart validation template |

These are reference files. The active workflows in `.github/workflows/` are the ones that run.

## Local Validation

Before pushing, you can run the same checks locally:

```bash
# Lint
cd apps/go-api && go vet ./...
cd cmd/labctl && go vet ./...

# Test
cd apps/go-api && go test -race ./...
cd cmd/labctl && go test -race ./internal/...

# Build
make build APP_NAME=go-api
make cli-build

# Helm lint
helm lint apps/go-api/deploy/helm -f apps/go-api/deploy/helm/values-dev.yaml
```
