# Runbook 05 — CI/CD

> Goal: verify the automated checks on PRs and the CD path.
> Verifies Phase 5 (tasks 006, 011, 023).

## What runs where

| Workflow | File | Trigger | Does |
|----------|------|---------|------|
| CI | `.github/workflows/ci.yaml` | PR / push to main/feature/* | Go lint+test, shell portability lint, Terraform fmt+validate, YAML lint, image build+scan, Helm lint |
| CD | `.github/workflows/cd.yaml` | merge to main | build + push images to GHCR, bump Helm values, ArgoCD sync |
| Helm validation | `.github/workflows/helm-validation.yaml` | PR touching charts | discover charts, lint + template render |

## Verify locally before pushing

```bash
# Go
cd apps/go-api && go vet ./... && go test ./... && cd ../..
cd cmd/labctl && go vet ./... && go test ./... && cd ../..

# Helm
helm lint apps/go-api/deploy/helm -f apps/go-api/deploy/helm/values-dev.yaml
helm template go-api apps/go-api/deploy/helm -f apps/go-api/deploy/helm/values-dev.yaml >/dev/null

# Shell portability (should print "clean")
grep -rnE 'grep -oP|sed -i |readlink -f|date -d|stat -c' \
  --include='*.sh' engine platform runtimes services bootstrap && echo "FORBIDDEN" || echo "clean"

# shellcheck (requires: brew install shellcheck)
find engine platform runtimes services bootstrap -name '*.sh' \
  -exec shellcheck --severity=warning {} +

# Terraform format (requires terraform installed)
terraform fmt -check -recursive foundation/terraform

# YAML lint (requires: pip install yamllint)
yamllint -d relaxed scenarios/ platform/
```

## Verify in GitHub

### CI checks on a PR

1. Open a PR touching any file in `apps/`, `engine/`, `platform/`, `runtimes/`, `services/`, `bootstrap/`, `scenarios/`, or `foundation/terraform/`.
2. Confirm all of the following jobs appear and pass:
   - `lint-apps` (go-api, echo-server)
   - `lint-cli`
   - `lint-shell`
   - `lint-terraform`
   - `lint-yaml`
   - `test-apps` (go-api, echo-server)
   - `test-cli`
   - `build-images` (go-api, echo-server)
   - `image-scan` (go-api, echo-server)
   - `helm-lint` (go-api, echo-server)

### Image scan results

1. Navigate to the repo → **Security** → **Code scanning alerts**.
2. Filter by category `trivy-go-api` or `trivy-echo-server`.
3. CRITICAL fixable findings cause the `image-scan` job to fail; HIGH and below are reported only.

### CD after merge

1. Merge a PR that touches `apps/go-api/` or `apps/echo-server/`.
2. Confirm the CD workflow runs: `detect-changes` → `build-and-push` → `update-manifests`.
3. After `update-manifests` commits, check that `apps/<app>/deploy/helm/values-dev.yaml` has the new SHA.
4. If the `gitops-cicd` scenario is active and `ARGOCD_SERVER` / `ARGOCD_AUTH_TOKEN` secrets are set, confirm `notify-argocd` runs and the ArgoCD app transitions to **Synced/Healthy**.

### ArgoCD secrets setup (one-time)

```bash
# With the gitops-cicd scenario active and ArgoCD reachable:
ARGOCD_SERVER=argocd.k3d.local

# Generate a token for the admin account (or a dedicated CI account)
argocd account generate-token --account admin \
  --server $ARGOCD_SERVER --grpc-web

# Add the output as GitHub repository secrets:
# Settings → Secrets and variables → Actions → New repository secret
# ARGOCD_SERVER  = argocd.k3d.local
# ARGOCD_AUTH_TOKEN = <token from above>
```

## Expected failure modes

| Scenario | Which job fails | Fix |
|----------|----------------|-----|
| PR adds `grep -oP` to a shell script | `lint-shell` portability check | Replace with POSIX grep |
| `.tf` file is not formatted | `lint-terraform` | Run `terraform fmt -recursive foundation/terraform` |
| Invalid YAML in a scenario file | `lint-yaml` | Fix the YAML syntax error |
| App uses a vulnerable base image (CRITICAL fixable CVE) | `image-scan` | Update FROM in Dockerfile to a patched version |
| CD can't push to GHCR | `build-and-push` | Check `GITHUB_TOKEN` package write permission in repo settings |
| ArgoCD never syncs after merge | `notify-argocd` skipped or failed | Verify `ARGOCD_SERVER` secret is set; check ArgoCD is reachable from the runner |
