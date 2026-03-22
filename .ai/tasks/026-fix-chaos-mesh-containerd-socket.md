# Task 026: Fix Chaos-Mesh Containerd Socket Hardcoded Path

## Priority
P1

## Assigned To
Bug

## Description
`platform/chaos/chaos-mesh/values.yaml` hardcodes the containerd socket path as `/run/k3s/containerd/containerd.sock`. This path is specific to k3s/k3d. On AKS the containerd socket is at `/run/containerd/containerd.sock`, and on EKS it is at `/run/containerd/containerd.sock` as well. When the chaos-engineering scenario is activated on a cloud runtime, the Chaos Mesh daemon pods fail to start because the socket path does not exist, silently breaking all chaos experiments.

## Files to Modify
- `platform/chaos/chaos-mesh/values.yaml`
- `platform/chaos/chaos-mesh/install.sh`

## Implementation Notes
The Chaos Mesh `chaosDaemon.socketPath` Helm value controls this. The correct approach:

1. In `install.sh`, detect the active runtime from `PROFILE` (or by checking the kubeconfig context name) and set the socket path accordingly:
   ```bash
   case "${PROFILE:-k3d}" in
     k3d)    CONTAINERD_SOCKET="/run/k3s/containerd/containerd.sock" ;;
     aks|eks) CONTAINERD_SOCKET="/run/containerd/containerd.sock" ;;
     *)       CONTAINERD_SOCKET="/run/containerd/containerd.sock" ;;
   esac
   ```

2. Pass the socket path at helm install time:
   ```bash
   helm upgrade --install chaos-mesh chaos-mesh/chaos-mesh \
     --namespace chaos-mesh \
     --set chaosDaemon.socketPath="${CONTAINERD_SOCKET}" \
     -f "${SCRIPT_DIR}/values.yaml" \
     --wait --timeout 3m
   ```

3. Remove the hardcoded `socketPath` from `values.yaml` so the install script is the single source of truth.

4. Document the supported runtimes and their socket paths in a comment in `install.sh`.

## Acceptance Criteria
- [ ] `install.sh` selects the correct socket path based on `PROFILE`.
- [ ] On k3d, chaos-daemon pods start and the socket path matches `/run/k3s/containerd/containerd.sock`.
- [ ] `values.yaml` no longer contains the hardcoded k3s socket path.
- [ ] The script includes a comment explaining the platform-specific socket paths.

## Testing Instructions
On k3d, run `platform/chaos/chaos-mesh/install.sh` and verify `kubectl get pods -n chaos-mesh` shows all pods Running, then `kubectl exec -n chaos-mesh <daemon-pod> -- ls /run/k3s/containerd/containerd.sock` confirms the socket exists. Run `status.sh` and confirm no errors.

## Dependencies
None
