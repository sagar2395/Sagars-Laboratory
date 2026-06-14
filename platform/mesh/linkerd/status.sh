#!/usr/bin/env bash
set -euo pipefail

NAMESPACE=linkerd
MESH_APP_NAMESPACE="${MESH_APP_NAMESPACE:-go-api}"

echo "=== Linkerd Service Mesh Status ==="
echo ""

if ! kubectl get namespace "${NAMESPACE}" >/dev/null 2>&1; then
  echo "Linkerd is not installed (namespace ${NAMESPACE} not found)"
  exit 0
fi

echo "Control Plane (${NAMESPACE}):"
kubectl get pods -n "${NAMESPACE}" 2>/dev/null || echo "  No pods found"
echo ""

echo "Services:"
kubectl get svc -n "${NAMESPACE}" 2>/dev/null || echo "  No services found"
echo ""

echo "Proxy-injection namespaces:"
kubectl get namespace --show-labels 2>/dev/null | grep 'linkerd.io/inject=enabled' || echo "  No namespaces with proxy injection enabled"
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

echo "Linkerd CRDs installed:"
kubectl get crd -o name 2>/dev/null | grep -c '\.linkerd\.io' | xargs -I{} echo "  {} Linkerd CRDs"
echo ""

# Linkerd CLI check if available
if command -v linkerd >/dev/null 2>&1; then
  echo "Linkerd CLI check:"
  linkerd check --proxy 2>&1 | tail -5 || true
else
  echo "Tip: install the linkerd CLI for a full health check (linkerd check)"
fi
