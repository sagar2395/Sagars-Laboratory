#!/usr/bin/env bash
set -euo pipefail

NAMESPACE=istio-system
MESH_APP_NAMESPACE="${MESH_APP_NAMESPACE:-go-api}"

echo "=== Istio Service Mesh Status ==="
echo ""

if ! kubectl get namespace "${NAMESPACE}" >/dev/null 2>&1; then
  echo "Istio is not installed (namespace ${NAMESPACE} not found)"
  exit 0
fi

echo "Control Plane (${NAMESPACE}):"
kubectl get pods -n "${NAMESPACE}" 2>/dev/null || echo "  No pods found"
echo ""

echo "Services:"
kubectl get svc -n "${NAMESPACE}" 2>/dev/null || echo "  No services found"
echo ""

echo "Istiod deployment:"
kubectl get deployment istiod -n "${NAMESPACE}" 2>/dev/null || echo "  istiod not found"
echo ""

echo "Sidecar injection namespaces:"
kubectl get namespace --show-labels 2>/dev/null | grep 'istio-injection=enabled' || echo "  No namespaces with sidecar injection enabled"
echo ""

echo "Meshed pods in ${MESH_APP_NAMESPACE}:"
if kubectl get namespace "${MESH_APP_NAMESPACE}" >/dev/null 2>&1; then
  kubectl get pods -n "${MESH_APP_NAMESPACE}" \
    -o jsonpath='{range .items[*]}{.metadata.name}{"  containers: "}{range .spec.containers[*]}{.name}{" "}{end}{"\n"}{end}' \
    2>/dev/null | head -20 || echo "  No pods found"
else
  echo "  Namespace ${MESH_APP_NAMESPACE} not found"
fi
echo ""

echo "Istio CRDs installed:"
kubectl get crd -o name 2>/dev/null | grep -c '\.istio\.io' | xargs -I{} echo "  {} Istio CRDs"
