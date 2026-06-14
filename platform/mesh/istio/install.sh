#!/usr/bin/env bash
# Installs Istio service mesh (sidecar mode) via Helm.
# Selects sidecar over ambient for broad demo app compatibility.
set -euo pipefail

NAMESPACE=istio-system
CHART_VERSION="${ISTIO_CHART_VERSION:-1.24.2}"
DOMAIN_SUFFIX="${DOMAIN_SUFFIX:-k3d.local}"
MESH_APP_NAMESPACE="${MESH_APP_NAMESPACE:-go-api}"

echo "Installing Istio ${CHART_VERSION} (sidecar mode)..."

helm repo add istio https://istio-release.storage.googleapis.com/charts --force-update
helm repo update

echo "Installing istio/base..."
helm upgrade --install istio-base istio/base \
  --namespace "${NAMESPACE}" \
  --create-namespace \
  --version "${CHART_VERSION}" \
  --wait --timeout 3m

echo "Installing istiod (control plane)..."
helm upgrade --install istiod istio/istiod \
  --namespace "${NAMESPACE}" \
  --version "${CHART_VERSION}" \
  -f "$(dirname "$0")/values.yaml" \
  --wait --timeout 5m

# Enable sidecar injection in the app namespace
if kubectl get namespace "${MESH_APP_NAMESPACE}" >/dev/null 2>&1; then
  echo "Enabling sidecar injection in namespace ${MESH_APP_NAMESPACE}..."
  kubectl label namespace "${MESH_APP_NAMESPACE}" istio-injection=enabled --overwrite
  echo "Rolling pods in ${MESH_APP_NAMESPACE} to acquire sidecars..."
  kubectl rollout restart deployment -n "${MESH_APP_NAMESPACE}" 2>/dev/null || true
fi

echo ""
echo "Istio ${CHART_VERSION} installed."
echo "  Control plane: kubectl get pods -n ${NAMESPACE}"
echo "  Kiali (if installed): http://kiali.${DOMAIN_SUFFIX}"
echo "  Sidecar injection active on namespace: ${MESH_APP_NAMESPACE}"
