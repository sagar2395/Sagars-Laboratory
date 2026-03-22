# Task 015: Fix Kubernetes Version Parsing in k8s Client

## Priority
P0

## Assigned To
Bug

## Description
`cmd/labctl/internal/k8s/client.go` parses the Kubernetes server version by doing a raw string search for `"gitVersion"` in the `kubectl version --short` output. This approach is brittle: the `--short` flag is deprecated since Kubernetes 1.24, the output format changed in 1.28, and string-based parsing will silently produce garbage or panic on any new version format. The correct approach is to parse the structured JSON from `kubectl version -o json`.

## Files to Modify
- `cmd/labctl/internal/k8s/client.go`

## Implementation Notes
Replace the `--short` invocation and subsequent string index hacking with:

1. Run `kubectl version -o json` (no `--short` flag).
2. Unmarshal the response into a small local struct:
   ```go
   type versionOutput struct {
       ServerVersion struct {
           GitVersion string `json:"gitVersion"`
       } `json:"serverVersion"`
   }
   ```
3. Assign `info.K8sVersion = v.ServerVersion.GitVersion`.
4. Handle the case where the server is unreachable (the JSON field will be empty or the command will fail): set `K8sVersion` to `"unknown"` and do not propagate the error — surface it via `info.Connected = false` instead.

Do not change the `ClusterInfo` struct fields or any callers. The fix is purely internal to the parsing logic.

## Acceptance Criteria
- [ ] `GetClusterInfo()` returns the Kubernetes version string on any cluster version from 1.24 onward.
- [ ] When the cluster is unreachable, `K8sVersion` is `"unknown"` and `Connected` is `false` — no panic or garbled string.
- [ ] Existing callers (`handleStatus`, `status.go`) continue to compile and display the version correctly.
- [ ] Unit test using a JSON fixture covers both the happy path and the offline/empty case.

## Testing Instructions
Run `go test ./internal/k8s/...` from `cmd/labctl`. Manually run `labctl status` against a live k3d cluster and confirm version string is well-formed (e.g., `v1.31.0+k3s1`).

## Dependencies
None
