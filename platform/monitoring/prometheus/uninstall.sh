#!/bin/bash

set -e

NAMESPACE="monitoring"

echo "Uninstalling Prometheus Stack..."

# Delete Prometheus Ingress
kubectl delete ingress prometheus -n $NAMESPACE >/dev/null 2>&1 || true

# Uninstall Prometheus
helm uninstall prometheus -n $NAMESPACE >/dev/null 2>&1 || true

# Delete namespace if it exists
kubectl delete namespace $NAMESPACE >/dev/null 2>&1 || true

echo "✓ Prometheus Stack uninstalled successfully"
