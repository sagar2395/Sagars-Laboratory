# Task 020: Fix kubectl Pod Output Parsing to Use JSON Format

## Priority
P1

## Assigned To
Bug

## Description
`cmd/labctl/internal/k8s/client.go` runs `kubectl get pods` with a custom `--output=custom-columns` format and then parses the output by splitting on whitespace. This approach breaks when pod names, namespaces, or node names contain spaces (unlikely but possible), and more critically, breaks whenever the column widths change or `kubectl` adds alignment padding. The correct approach is to request JSON output and unmarshal it, which is stable across all kubectl versions.

## Files to Modify
- `cmd/labctl/internal/k8s/client.go`

## Implementation Notes
Replace the `--output=custom-columns` approach in `GetNamespacePods` with:

1. Run `kubectl get pods -n <namespace> -o json`.
2. Unmarshal the response into a small local struct mirroring the Kubernetes PodList JSON:
   ```go
   type podList struct {
       Items []struct {
           Metadata struct {
               Name string `json:"name"`
           } `json:"metadata"`
           Status struct {
               Phase             string `json:"phase"`
               ContainerStatuses []struct {
                   Ready        bool  `json:"ready"`
                   RestartCount int32 `json:"restartCount"`
               } `json:"containerStatuses"`
           } `json:"status"`
       } `json:"items"`
   }
   ```
3. Populate `PodInfo` from the unmarshaled data. Compute `Ready` by counting ready containers. Keep `Status` as `phase`. Keep `Restarts` as the sum of `restartCount` across all containers.
4. Drop the `Age` field calculation (or set it to `""`) — computing human-readable age from `creationTimestamp` is a follow-up; do not block this fix on it.

Do not change the `PodInfo` struct fields or any callers.

## Acceptance Criteria
- [ ] `GetNamespacePods` returns correct pod info for all pods in a namespace including pods with multi-word phase strings.
- [ ] Multiple containers per pod: `Ready` reflects the count of ready containers, `Restarts` is the total across all containers.
- [ ] Unit test using a JSON fixture covers at least: running pod (all containers ready), pending pod (no containers ready), pod with restarts.
- [ ] No whitespace-split parsing remains in the function.
- [ ] `go test ./internal/k8s/...` passes.

## Testing Instructions
Run `labctl status` against a live k3d cluster. Confirm pod statuses listed under each app match what `kubectl get pods -A` reports. Run unit tests with fixture data.

## Dependencies
None
