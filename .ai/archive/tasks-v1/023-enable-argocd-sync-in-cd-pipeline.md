# Task 023: Enable ArgoCD Sync Step in CD Pipeline

## Priority
P1

## Assigned To
DevOps

## Description
The `notify-argocd` job in `delivery/github-actions/cd.yaml` has been permanently disabled with `if: false`. The entire GitOps automation loop depends on this step: ArgoCD is installed by the `gitops-cicd` scenario, but CD never triggers a sync, so image updates committed by the `update-manifests` job are only applied when ArgoCD polls (every 3 minutes by default) or when a user manually syncs. The step must be re-enabled with proper secrets-based configuration so it works without hardcoded credentials.

## Files to Modify
- `delivery/github-actions/cd.yaml`
- `docs/ci-cd.md`

## Implementation Notes
1. Remove the `if: false` gate from the `notify-argocd` job.
2. Replace any hardcoded ArgoCD server URL with a GitHub Actions secret: `${{ secrets.ARGOCD_SERVER }}`.
3. Use `${{ secrets.ARGOCD_AUTH_TOKEN }}` for authentication (not a username/password). ArgoCD supports API tokens via `argocd account generate-token`.
4. The sync command must target the correct app name — use a job input or derived value from the changed app path:
   ```yaml
   run: argocd app sync ${{ env.APP_NAME }} --grpc-web --server ${{ secrets.ARGOCD_SERVER }} --auth-token ${{ secrets.ARGOCD_AUTH_TOKEN }}
   ```
5. Add a `needs: [update-manifests]` dependency so the sync only runs after the manifest update is committed.
6. Add a timeout of 5 minutes for the sync step so CI doesn't hang indefinitely.
7. Update `docs/ci-cd.md` to document the two required secrets (`ARGOCD_SERVER`, `ARGOCD_AUTH_TOKEN`) and how to generate an ArgoCD token.

Do not add new Actions; use the existing `argocd` CLI installed in the runner image or add a one-line setup step using the ArgoCD download URL from `versions.env`.

## Acceptance Criteria
- [ ] The `notify-argocd` job runs on push to `main` (no `if: false`).
- [ ] ArgoCD server URL and auth token are read from GitHub Actions repository secrets, not hardcoded.
- [ ] The job depends on `update-manifests` completing successfully.
- [ ] A 5-minute timeout prevents CI from hanging.
- [ ] `docs/ci-cd.md` documents the required secrets and how to configure them.

## Testing Instructions
Configure `ARGOCD_SERVER` and `ARGOCD_AUTH_TOKEN` secrets in a test fork. Push a trivial change to `apps/go-api/` and confirm the `notify-argocd` job runs and the ArgoCD app status transitions to `Synced`.

## Dependencies
None (can be implemented without task 006)
