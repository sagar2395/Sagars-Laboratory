#!/usr/bin/env bash
set -euo pipefail

OPERATOR_NS=external-secrets
APP_NAMESPACE="${VAULT_APP_NAMESPACE:-go-api}"

echo "=== External Secrets Operator Status ==="
echo ""

if ! kubectl get namespace "${OPERATOR_NS}" >/dev/null 2>&1; then
  echo "External Secrets Operator is not installed (namespace ${OPERATOR_NS} not found)"
  exit 0
fi

echo "Operator pods:"
kubectl get pods -n "${OPERATOR_NS}" 2>/dev/null || echo "  No pods found"
echo ""

echo "SecretStores in ${APP_NAMESPACE}:"
kubectl get secretstore -n "${APP_NAMESPACE}" 2>/dev/null || echo "  No SecretStores found"
echo ""

echo "ExternalSecrets in ${APP_NAMESPACE}:"
kubectl get externalsecret -n "${APP_NAMESPACE}" 2>/dev/null || echo "  No ExternalSecrets found"
echo ""

# Check sync status
ES=$(kubectl get externalsecret go-api-secrets -n "${APP_NAMESPACE}" \
  -o jsonpath='{.status.conditions[0].message}' 2>/dev/null || echo "")
if [ -n "${ES}" ]; then
  echo "Last sync status: ${ES}"
fi
echo ""

echo "Synced secret:"
if kubectl get secret go-api-external-secret -n "${APP_NAMESPACE}" >/dev/null 2>&1; then
  echo "  go-api-external-secret exists in namespace ${APP_NAMESPACE}"
  echo "  Keys: $(kubectl get secret go-api-external-secret -n "${APP_NAMESPACE}" \
    -o jsonpath='{.data}' 2>/dev/null | tr ',' '\n' | grep -o '"[^"]*":' | tr -d '":' | tr '\n' ' ')"
else
  echo "  Secret 'go-api-external-secret' not yet synced"
fi
