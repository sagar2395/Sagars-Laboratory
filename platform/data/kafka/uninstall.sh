#!/usr/bin/env bash
set -euo pipefail

NAMESPACE=kafka

echo "Uninstalling Kafka..."

if helm status kafka -n "${NAMESPACE}" >/dev/null 2>&1; then
  helm uninstall kafka -n "${NAMESPACE}"
  echo "Helm release removed."
fi

if kubectl get namespace "${NAMESPACE}" >/dev/null 2>&1; then
  kubectl delete namespace "${NAMESPACE}" --timeout=60s || true
  echo "Namespace '${NAMESPACE}' deleted."
fi

echo "Kafka uninstalled."
