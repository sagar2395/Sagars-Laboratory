#!/usr/bin/env bash
set -euo pipefail

NAMESPACE=vault
DOMAIN_SUFFIX="${DOMAIN_SUFFIX:-k3d.local}"

echo "=== Vault Status ==="
echo ""

if ! kubectl get namespace "${NAMESPACE}" >/dev/null 2>&1; then
  echo "Vault is not installed (namespace ${NAMESPACE} not found)"
  exit 0
fi

echo "Pods:"
kubectl get pods -n "${NAMESPACE}" 2>/dev/null || echo "  No pods found"
echo ""

echo "Services:"
kubectl get svc -n "${NAMESPACE}" 2>/dev/null || echo "  No services found"
echo ""

# Check Vault seal status (if available)
if kubectl get pod -n "${NAMESPACE}" -l app.kubernetes.io/name=vault \
     -o jsonpath='{.items[0].metadata.name}' >/dev/null 2>&1; then
  POD=$(kubectl get pod -n "${NAMESPACE}" -l app.kubernetes.io/name=vault \
    -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
  echo "Seal status:"
  kubectl exec -n "${NAMESPACE}" "${POD}" -- vault status 2>/dev/null \
    | grep -E "Sealed|Mode|Version" | sed 's/^/  /' || echo "  (vault not responding)"
  echo ""
fi

echo "Demo secret:"
echo "  vault kv get -address=http://vault.${DOMAIN_SUFFIX} secret/go-api"
echo "  (token: root)"
echo ""

echo "UI: http://vault.${DOMAIN_SUFFIX}"
