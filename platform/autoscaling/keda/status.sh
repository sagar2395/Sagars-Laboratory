#!/usr/bin/env bash
set -euo pipefail

NAMESPACE=keda

echo "=== KEDA Status ==="
echo ""

if ! kubectl get namespace "${NAMESPACE}" >/dev/null 2>&1; then
  echo "KEDA is not installed (namespace ${NAMESPACE} not found)"
  exit 0
fi

echo "Pods:"
kubectl get pods -n "${NAMESPACE}" 2>/dev/null || echo "  No pods found"
echo ""

echo "ScaledObjects across all namespaces:"
kubectl get scaledobject --all-namespaces 2>/dev/null || echo "  No ScaledObjects found"
echo ""

echo "ScaledJobs across all namespaces:"
kubectl get scaledjob --all-namespaces 2>/dev/null || echo "  No ScaledJobs found"
echo ""

echo "TriggerAuthentications:"
kubectl get triggerauthentication --all-namespaces 2>/dev/null || echo "  None found"
