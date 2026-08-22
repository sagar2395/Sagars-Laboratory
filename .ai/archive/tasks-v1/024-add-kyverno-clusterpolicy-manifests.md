# Task 024: Add Kyverno ClusterPolicy Manifests

## Priority
P1

## Assigned To
DevOps

## Description
The `platform/security/policy/kyverno/` component installs the Kyverno admission controller via Helm, but ships zero policy definitions. Without ClusterPolicy resources, Kyverno runs but enforces nothing. The `security-compliance` scenario references `kyverno-policies` as a manifest component — those policies must be defined in `scenarios/security-compliance/manifests/`. Currently, the scenario YAML declares this component but the actual policy manifests are either empty or missing, meaning the scenario silently provides no security enforcement.

## Files to Modify
- `scenarios/security-compliance/manifests/` *(add new manifest files)*
- `scenarios/security-compliance/scenario.yaml` *(verify component path is correct)*

## Implementation Notes
Create the following ClusterPolicy manifests in `scenarios/security-compliance/manifests/kyverno-policies.yaml`:

1. **Deny privilege escalation** — Block containers with `allowPrivilegeEscalation: true`:
   ```yaml
   rules:
   - name: deny-privilege-escalation
     match: { resources: { kinds: [Pod] } }
     validate:
       message: "Privilege escalation is not allowed."
       pattern:
         spec:
           containers:
           - securityContext:
               allowPrivilegeEscalation: false
   ```

2. **Require non-root user** — Require `runAsNonRoot: true` for all containers.

3. **Require resource limits** — Require CPU and memory limits on all containers (prevents noisy-neighbor events).

4. **Block latest tag** — Reject images tagged `:latest` or with no tag (require explicit version pinning).

For each policy, set `validationFailureAction: Audit` (not `Enforce`) so that pre-existing deployments are not immediately disrupted. The scenario's `explore` commands already include a test for privilege escalation blocking.

All policies must include namespace exclusions for `kube-system`, `kyverno`, `cert-manager`, `argocd`, and `chaos-mesh` to avoid blocking system components.

## Acceptance Criteria
- [ ] Four ClusterPolicy resources are created in the manifests file.
- [ ] Each policy excludes system namespaces.
- [ ] `validationFailureAction` is set to `Audit` for all policies.
- [ ] Running `labctl scenario up security-compliance` applies the policies without errors.
- [ ] `kubectl get clusterpolicies` shows all four policies in `Ready` state.
- [ ] The existing scenario `explore` command for policy testing returns a policy violation response (not 404).

## Testing Instructions
Activate the security-compliance scenario: `labctl scenario up security-compliance`. Run: `kubectl run test --image=nginx --restart=Never --privileged` and confirm the Kyverno audit log shows a violation. Run `kubectl get clusterpolicies` to verify all four policies appear.

## Dependencies
None — independently deployable.
