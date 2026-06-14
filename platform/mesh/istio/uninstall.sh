#!/usr/bin/env bash
set -euo pipefail

NAMESPACE=istio-system
MESH_APP_NAMESPACE="${MESH_APP_NAMESPACE:-go-api}"

echo "Uninstalling Istio..."

# Remove sidecar injection label from app namespace before uninstalling,
# then restart pods so they shed their sidecars.
if kubectl get namespace "${MESH_APP_NAMESPACE}" >/dev/null 2>&1; then
  echo "Removing sidecar injection label from namespace ${MESH_APP_NAMESPACE}..."
  kubectl label namespace "${MESH_APP_NAMESPACE}" istio-injection- 2>/dev/null || true
  echo "Rolling pods in ${MESH_APP_NAMESPACE} to remove sidecars..."
  kubectl rollout restart deployment -n "${MESH_APP_NAMESPACE}" 2>/dev/null || true
fi

if helm status istiod -n "${NAMESPACE}" >/dev/null 2>&1; then
  helm uninstall istiod -n "${NAMESPACE}"
  echo "istiod removed."
fi

if helm status istio-base -n "${NAMESPACE}" >/dev/null 2>&1; then
  helm uninstall istio-base -n "${NAMESPACE}"
  echo "istio-base removed."
fi

# Remove CRDs installed by istio-base (they persist after Helm uninstall)
kubectl get crd -o name 2>/dev/null | grep '\.istio\.io' | xargs -r kubectl delete || true

if kubectl get namespace "${NAMESPACE}" >/dev/null 2>&1; then
  kubectl delete namespace "${NAMESPACE}" --timeout=60s || true
  echo "Namespace '${NAMESPACE}' deleted."
fi

echo "Istio uninstalled."
