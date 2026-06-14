#!/usr/bin/env bash
set -euo pipefail

NAMESPACE=linkerd
MESH_APP_NAMESPACE="${MESH_APP_NAMESPACE:-go-api}"

echo "Uninstalling Linkerd..."

# Remove proxy injection annotation from app namespace before uninstalling
if kubectl get namespace "${MESH_APP_NAMESPACE}" >/dev/null 2>&1; then
  echo "Removing Linkerd injection annotation from namespace ${MESH_APP_NAMESPACE}..."
  kubectl annotate namespace "${MESH_APP_NAMESPACE}" linkerd.io/inject- 2>/dev/null || true
  echo "Rolling pods in ${MESH_APP_NAMESPACE} to remove proxies..."
  kubectl rollout restart deployment -n "${MESH_APP_NAMESPACE}" 2>/dev/null || true
fi

if helm status linkerd-control-plane -n "${NAMESPACE}" >/dev/null 2>&1; then
  helm uninstall linkerd-control-plane -n "${NAMESPACE}"
  echo "linkerd-control-plane removed."
fi

if helm status linkerd-crds -n "${NAMESPACE}" >/dev/null 2>&1; then
  helm uninstall linkerd-crds -n "${NAMESPACE}"
  echo "linkerd-crds removed."
fi

# Remove any lingering Linkerd CRDs
kubectl get crd -o name 2>/dev/null | grep '\.linkerd\.io' | xargs -r kubectl delete || true

if kubectl get namespace "${NAMESPACE}" >/dev/null 2>&1; then
  kubectl delete namespace "${NAMESPACE}" --timeout=60s || true
  echo "Namespace '${NAMESPACE}' deleted."
fi

echo "Linkerd uninstalled."
