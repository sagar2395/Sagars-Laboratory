#!/usr/bin/env bash
set -euo pipefail

NAMESPACE=monitoring
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "Installing Loki (single-binary mode)..."

helm repo add grafana https://grafana.github.io/helm-charts --force-update
helm repo update

# Clean up stuck pending releases if any
if helm status loki -n "$NAMESPACE" 2>/dev/null | grep -q "pending-"; then
  echo "Cleaning up stuck Loki release..."
  helm delete loki -n "$NAMESPACE" --wait 2>/dev/null || true
fi

helm upgrade --install loki grafana/loki \
  --namespace "$NAMESPACE" \
  --create-namespace \
  -f "$SCRIPT_DIR/values.yaml" \
  --wait --timeout 5m

echo "Waiting for Loki to be ready..."
kubectl rollout status statefulset/loki -n "$NAMESPACE" --timeout=120s || true

echo "Installing Promtail..."

# Clean up stuck pending releases if any
if helm status promtail -n "$NAMESPACE" 2>/dev/null | grep -q "pending-"; then
  echo "Cleaning up stuck Promtail release..."
  helm delete promtail -n "$NAMESPACE" --wait 2>/dev/null || true
fi

helm upgrade --install promtail grafana/promtail \
  --namespace "$NAMESPACE" \
  -f "$SCRIPT_DIR/promtail-values.yaml" \
  --wait --timeout 5m

echo "Waiting for Promtail to be ready..."
kubectl rollout status daemonset/promtail -n "$NAMESPACE" --timeout=120s || true

echo "Loki + Promtail installed successfully."
