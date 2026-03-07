#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="chaos-mesh"
SCRIPT_DIR="$(dirname "$0")"

echo "Installing Chaos Mesh..."

# Create namespace if it doesn't exist
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# Add Helm repo and update
echo "Adding Chaos Mesh Helm repository..."
helm repo add chaos-mesh https://charts.chaos-mesh.org >/dev/null 2>&1 || true
helm repo update

# Install or upgrade Chaos Mesh
echo "Installing Chaos Mesh chart..."
helm upgrade --install chaos-mesh chaos-mesh/chaos-mesh \
  --namespace $NAMESPACE \
  --create-namespace \
  -f "$SCRIPT_DIR/values.yaml"

# Wait for controller manager to be ready
echo "Waiting for Chaos Mesh controller manager to be ready..."
kubectl rollout status deployment/chaos-controller-manager -n $NAMESPACE --timeout=180s || true

# Wait for dashboard
echo "Waiting for Chaos Mesh dashboard to be ready..."
kubectl rollout status deployment/chaos-dashboard -n $NAMESPACE --timeout=120s || true

echo ""
echo "Chaos Mesh installed successfully"
echo "Namespace: $NAMESPACE"
echo "Status: kubectl get pods -n $NAMESPACE"
echo ""
echo "Dashboard: kubectl port-forward -n $NAMESPACE svc/chaos-dashboard 2333:2333"
echo "  Then open http://localhost:2333"
echo ""
echo "Create experiments via CRDs or the Chaos Dashboard UI."
