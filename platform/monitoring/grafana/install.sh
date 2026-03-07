#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="monitoring"
SCRIPT_DIR="$(dirname "$0")"

echo "Installing Grafana..."

# Create namespace if it doesn't exist
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# Add Helm repo and update
echo "Adding Grafana Helm repository..."
helm repo add grafana https://grafana.github.io/helm-charts >/dev/null 2>&1 || true
helm repo update

# Get admin password from env or use default
GRAFANA_ADMIN_PASSWORD="${GRAFANA_ADMIN_PASSWORD:-admin}"

# Install or upgrade Grafana
echo "Installing Grafana chart..."
helm upgrade --install grafana grafana/grafana \
  --namespace $NAMESPACE \
  --create-namespace \
  -f "$SCRIPT_DIR/values.yaml" \
  --set adminPassword="$GRAFANA_ADMIN_PASSWORD"

# Wait for Grafana deployment to be ready
echo "Waiting for Grafana to be ready..."
kubectl rollout status deployment/grafana -n $NAMESPACE --timeout=120s || true

echo "✓ Grafana installed successfully"
echo ""
echo "Access Grafana at: http://grafana.${DOMAIN_SUFFIX:-k3d.local}"
echo "Default credentials: admin / $GRAFANA_ADMIN_PASSWORD"
echo "Namespace: $NAMESPACE"
echo "Status: kubectl get pods -n $NAMESPACE"
