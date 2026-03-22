# Task 011: Add Container Image Security Scanning to CI

## Priority
P1

## Assigned To
DevOps

## Description
The CI pipeline builds container images for `go-api` and `echo-server` but does not scan them for known CVEs or misconfigurations before publishing. This is a supply chain risk: a vulnerable base image or dependency could be silently promoted. A Trivy scan step must be added to the CI workflow to gate image promotion on a clean security report.

## Files to Modify
- `.github/workflows/ci.yaml`
- `docs/ci-cd.md`

## Implementation Notes
1. Add a new job `image-scan` that depends on the build job (or runs after images are built but before they are pushed/tagged as latest).
2. Use `aquasecurity/trivy-action@master` (or pin to a stable version from `versions.env` if one is added) to scan the built image tarball or image reference.
3. Configure Trivy to:
   - Fail on `CRITICAL` severity issues only (to avoid blocking on informational findings initially).
   - Ignore unfixed vulnerabilities (`--ignore-unfixed`) so the build is not blocked by issues without available patches.
   - Output results in SARIF format and upload to GitHub Security tab via `github/codeql-action/upload-sarif`.
4. Update `docs/ci-cd.md` to document the scan step, its thresholds, and how to triage findings.

Do not add a separate workflow file; integrate into the existing `ci.yaml` as a new job.

## Acceptance Criteria
- [ ] CI scans built images on every PR and push to `main`.
- [ ] Builds fail when CRITICAL severity fixable CVEs are detected.
- [ ] SARIF output is uploaded to the GitHub Security tab.
- [ ] `docs/ci-cd.md` documents the scan configuration and failure thresholds.
- [ ] The scan job does not significantly increase CI wall-clock time (Trivy should complete in under 2 minutes for these images).

## Testing Instructions
Open a PR that intentionally uses an older Alpine base image version (e.g., `alpine:3.14`) and confirm the scan job triggers and potentially reports findings. Confirm the SARIF upload step appears in the Actions run.

## Dependencies
None (independent of task 006 which covers shell/Terraform checks)
