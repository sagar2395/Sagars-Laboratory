# Runbook 05 — CI/CD

> Goal: understand and verify the automated checks on PRs and the CD path.
> Verifies Phase 5.

## What runs where

| Workflow | File | Trigger | Does |
|----------|------|---------|------|
| CI | `.github/workflows/ci.yaml` | PR / push | lint + unit tests (apps + CLI), image build, Helm lint, **shell portability lint** (Task 006/034) |
| CD | `.github/workflows/cd.yaml` | merge to main | build + push images to GHCR, bump Helm values, **ArgoCD sync** (Task 023) |
| Helm validation | `.github/workflows/helm-validation.yaml` | PR | discover charts, lint + template render |
| Image scan | (in CI) | PR | container vulnerability scan (Task 011) |

## Verify locally before pushing

```bash
# CLI
cd cmd/labctl && go vet ./... && go test ./... && cd ../..

# Apps
cd apps/go-api && go test ./... && cd ../..

# Helm
helm lint apps/go-api/deploy/helm/go-api
helm template apps/go-api/deploy/helm/go-api -f apps/go-api/deploy/helm/go-api/values-dev.yaml >/dev/null

# Shell portability (the denylist from Task 034)
grep -rnE 'grep -oP|sed -i |readlink -f|date -d|stat -c' --include='*.sh' . && echo "FORBIDDEN FOUND" || echo "clean"
```

## Verify in GitHub

1. Open a PR; confirm CI + Helm validation + image scan run and pass.
2. Merge; confirm CD builds/pushes images and (with the GitOps scenario active)
   ArgoCD shows the apps **Synced/Healthy**.

## Expected

- A PR that introduces a GNU-only shell construct **fails** CI (Task 034).
- A PR with a vulnerable base image is flagged by the scan (Task 011).
- After merge, ArgoCD reconciles the new image (Task 023).

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| CD can't push to GHCR | Repo secrets / `GITHUB_TOKEN` package write permission. |
| ArgoCD never syncs | GitOps scenario not active, or sync step disabled (Task 023). |
| Image scan blocks everything | Tune severity threshold in the scan step. |
