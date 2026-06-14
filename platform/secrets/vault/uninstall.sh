#!/usr/bin/env bash
set -euo pipefail

NAMESPACE=vault

echo "Uninstalling Vault..."

if helm status vault -n "${NAMESPACE}" >/dev/null 2>&1; then
  helm uninstall vault -n "${NAMESPACE}"
  echo "Helm release removed."
fi

if kubectl get namespace "${NAMESPACE}" >/dev/null 2>&1; then
  kubectl delete namespace "${NAMESPACE}" --timeout=60s || true
  echo "Namespace '${NAMESPACE}' deleted."
fi

echo "Vault uninstalled."
