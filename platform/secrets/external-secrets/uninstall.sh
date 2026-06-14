#!/usr/bin/env bash
set -euo pipefail

OPERATOR_NS=external-secrets
APP_NAMESPACE="${VAULT_APP_NAMESPACE:-go-api}"

echo "Uninstalling External Secrets Operator..."

# Remove ExternalSecret and SecretStore from app namespace first
if kubectl get namespace "${APP_NAMESPACE}" >/dev/null 2>&1; then
  kubectl delete externalsecret go-api-secrets -n "${APP_NAMESPACE}" \
    --ignore-not-found 2>/dev/null || true
  kubectl delete secretstore vault-backend -n "${APP_NAMESPACE}" \
    --ignore-not-found 2>/dev/null || true
  kubectl delete secret vault-token -n "${APP_NAMESPACE}" \
    --ignore-not-found 2>/dev/null || true
  kubectl delete secret go-api-external-secret -n "${APP_NAMESPACE}" \
    --ignore-not-found 2>/dev/null || true
  echo "Cleaned up resources in namespace ${APP_NAMESPACE}."
fi

if helm status external-secrets -n "${OPERATOR_NS}" >/dev/null 2>&1; then
  helm uninstall external-secrets -n "${OPERATOR_NS}"
  echo "Helm release removed."
fi

# Remove CRDs
for crd in $(kubectl get crd -o name 2>/dev/null | grep 'external-secrets.io'); do
  kubectl delete "${crd}" 2>/dev/null || true
done

if kubectl get namespace "${OPERATOR_NS}" >/dev/null 2>&1; then
  kubectl delete namespace "${OPERATOR_NS}" --timeout=60s || true
  echo "Namespace '${OPERATOR_NS}' deleted."
fi

echo "External Secrets Operator uninstalled."
