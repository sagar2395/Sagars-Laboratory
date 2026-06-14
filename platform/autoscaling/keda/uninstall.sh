#!/usr/bin/env bash
set -euo pipefail

NAMESPACE=keda

echo "Uninstalling KEDA..."

# Remove ScaledObjects in all namespaces before removing the operator,
# otherwise the finalizers will block deletion.
for so in $(kubectl get scaledobject --all-namespaces -o name 2>/dev/null); do
  kubectl delete "${so}" --ignore-not-found 2>/dev/null || true
done

if helm status keda -n "${NAMESPACE}" >/dev/null 2>&1; then
  helm uninstall keda -n "${NAMESPACE}"
  echo "Helm release removed."
fi

# Remove CRDs installed by KEDA
for crd in $(kubectl get crd -o name 2>/dev/null | grep '\.keda\.sh'); do
  kubectl delete "${crd}" 2>/dev/null || true
done

if kubectl get namespace "${NAMESPACE}" >/dev/null 2>&1; then
  kubectl delete namespace "${NAMESPACE}" --timeout=60s || true
  echo "Namespace '${NAMESPACE}' deleted."
fi

echo "KEDA uninstalled."
