# CI/CD

This project includes GitHub Actions workflows for continuous integration and deployment.

## Workflows

Three active workflows live in `.github/workflows/`:

### CI Pipeline (`ci.yaml`)

**Triggers:** Push to `main` or `feature/*` branches, PRs to `main` (when app, CLI, shell, Terraform, scenario, or platform files change).

**Jobs:**

| Job | What It Does | Runs On |
|-----|-------------|---------|
| `lint-apps` | `go vet` + `staticcheck` for each app | Matrix: go-api, echo-server |
| `lint-cli` | `go vet` + `staticcheck` for labctl CLI | Single |
| `lint-shell` | `shellcheck --severity=warning` + GNU-only construct check on all `*.sh` in `engine/`, `platform/`, `runtimes/`, `services/`, `bootstrap/` | Single |
| `lint-terraform` | `terraform fmt -check -recursive` + `terraform validate` on all modules in `foundation/terraform/modules/` | Single |
| `lint-yaml` | `yamllint -d relaxed` on `scenarios/` and `platform/` | Single |
| `test-apps` | `go test -v -race` with coverage for each app | Matrix: go-api, echo-server |
| `test-cli` | `go test -v -race` for CLI internal packages + build check | Single |
| `build-images` | `docker build` for each app (after lint + test); exports image tar as artifact | Matrix: go-api, echo-server |
| `image-scan` | Trivy scan on built image tarball; fails on CRITICAL fixable CVEs; uploads SARIF to GitHub Security | Matrix: go-api, echo-server |
| `helm-lint` | `helm lint` + `helm template` for each app's chart | Matrix: go-api, echo-server |

### CD Pipeline (`cd.yaml`)

**Triggers:** Push to `main` branch (when app code changes).

**Jobs:**

| Job | What It Does |
|-----|-------------|
| `detect-changes` | Diffs HEAD~1 to find which apps changed |
| `build-and-push` | Builds Docker images and pushes to GitHub Container Registry (ghcr.io) |
| `update-manifests` | Updates image tag in `values-dev.yaml` and commits the change |
| `notify-argocd` | Installs ArgoCD CLI from `versions.env` and triggers an immediate sync for each changed app |

The CD pipeline uses GHCR for image hosting. Images are tagged with both the commit SHA and `latest`.

The `notify-argocd` job only runs when the `ARGOCD_SERVER` secret is set (skipped if the GitOps scenario is not active).

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
  -> lint-shell (shellcheck + portability check) [all shell scripts]
  -> lint-terraform (fmt check + validate per module)
  -> lint-yaml (yamllint scenarios/ platform/)
  -> test-apps (go test -race) [parallel per app]
  -> test-cli (go test -race + build check)
  -> build-images (docker build + export tar) [depends on lint-apps + test-apps]
  -> image-scan (trivy CRITICAL only, SARIF upload) [depends on build-images]
  -> helm-lint (helm lint + template) [parallel per app]
```

### CD Flow

```
Push to main (app changes only)
  -> detect-changes (which apps changed?)
  -> build-and-push (docker build + push to ghcr.io) [per changed app]
  -> update-manifests (sed image tag in values-dev.yaml + git commit) [per changed app]
  -> notify-argocd (argocd app sync) [per changed app, only if ARGOCD_SERVER secret is set]
```

## Configuration

### Secrets

| Secret | Required By | Description |
|--------|-------------|-------------|
| `GITHUB_TOKEN` | CD (auto-provided) | Push images to GHCR; commit manifest updates |
| `ARGOCD_SERVER` | CD `notify-argocd` | ArgoCD server hostname (e.g., `argocd.k3d.local`) |
| `ARGOCD_AUTH_TOKEN` | CD `notify-argocd` | ArgoCD API token |

To generate an ArgoCD token:
```bash
argocd account generate-token --account <username> \
  --server $ARGOCD_SERVER --grpc-web
```
Then add the output as the `ARGOCD_AUTH_TOKEN` repository secret.

`ARGOCD_SERVER` and `ARGOCD_AUTH_TOKEN` are only needed when the `gitops-cicd` scenario is active. The `notify-argocd` job is skipped if `ARGOCD_SERVER` is not set.

### Image Scan Configuration

The `image-scan` job uses [aquasecurity/trivy-action@0.28.0](https://github.com/aquasecurity/trivy-action):

- **Severity threshold:** `CRITICAL` only (informational and HIGH findings do not block the build).
- **Unfixed CVEs:** ignored — only vulnerabilities with an available fix gate the build.
- **Output:** SARIF uploaded to the repository's GitHub Security → Code scanning tab.

To triage a scan failure:
1. Navigate to the **Security → Code scanning alerts** tab in GitHub.
2. Filter by `trivy-<app>` category.
3. For each CRITICAL finding, check whether a base-image update resolves it:
   ```bash
   # Update the FROM line in apps/<app>/Dockerfile to the latest patch version
   # then re-push the PR to trigger a fresh scan.
   ```
4. If a CVE has no available fix, open a task to track it and consider accepting the risk temporarily via a `.trivyignore` file in the relevant `apps/<app>/` directory.

### Shell Portability Check

The `lint-shell` job rejects any `*.sh` file in `engine/`, `platform/`, `runtimes/`, `services/`, or `bootstrap/` that contains:

| Forbidden | Reason |
|-----------|--------|
| `grep -oP` | PCRE — not available on macOS `grep` |
| `sed -i ` (no backup ext) | GNU sed only — macOS requires `sed -i ''` |
| `readlink -f` | GNU coreutils — not available on macOS |
| `date -d` | GNU date — macOS uses `date -j -f` |
| `stat -c` | GNU stat format — macOS uses `stat -f` |

### Terraform Validation

`lint-terraform` checks all modules under `foundation/terraform/modules/`. It runs:

1. `terraform fmt -check -recursive` — fails if any `.tf` file is not formatted.
2. `terraform init -backend=false` + `terraform validate` per module.

To auto-fix formatting before pushing:
```bash
terraform fmt -recursive foundation/terraform
```

### Customization

#### Adding a new app to CI

The CI workflow uses a matrix strategy. To add a new app, edit `.github/workflows/ci.yaml`:

```yaml
strategy:
  matrix:
    app: [go-api, echo-server, my-new-app]  # add here
```

Update this in the `lint-apps`, `test-apps`, `build-images`, `image-scan`, and `helm-lint` jobs.

#### Changing the container registry

Edit `.github/workflows/cd.yaml`:

```yaml
env:
  REGISTRY: ghcr.io    # change to your registry
```

For Azure ACR or AWS ECR, update the login step to use the appropriate action.

## Templates

The `delivery/github-actions/` directory contains single-app workflow templates:

| Template | Description |
|----------|-------------|
| `ci.yaml` | Single-app CI template |
| `cd.yaml` | Single-app CD template with ArgoCD sync |
| `helm-release.yaml` | Helm chart validation template |

These are reference files. The active workflows in `.github/workflows/` are the ones that run.

## Local Validation

Before pushing, run the same checks locally:

```bash
# Go lint + test
cd apps/go-api && go vet ./... && go test -race ./... && cd ../..
cd cmd/labctl && go vet ./... && go test -race ./internal/... && cd ../..

# Shell portability
grep -rnE 'grep -oP|sed -i |readlink -f|date -d|stat -c' --include='*.sh' \
  engine platform runtimes services bootstrap && echo "FORBIDDEN" || echo "clean"

# shellcheck (requires shellcheck: brew install shellcheck)
find engine platform runtimes services bootstrap -name '*.sh' \
  -exec shellcheck --severity=warning {} +

# Terraform format
terraform fmt -check -recursive foundation/terraform

# Helm lint
helm lint apps/go-api/deploy/helm -f apps/go-api/deploy/helm/values-dev.yaml

# YAML lint (requires yamllint: pip install yamllint)
yamllint -d relaxed scenarios/ platform/
```
