#!/bin/bash

set -e

NAMESPACE="monitoring"

echo "Installing Prometheus Stack (prometheus-operator, kube-prometheus-stack, node-exporter, kube-state-metrics)..."

# Add Helm repo and update
echo "Adding Prometheus Helm repository..."
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts >/dev/null 2>&1 || true
helm repo update

# Install Prometheus Stack
echo "Installing kube-prometheus-stack chart..."
helm upgrade --install prometheus prometheus-community/kube-prometheus-stack \
  --namespace $NAMESPACE \
  --create-namespace \
  -f "$(dirname "$0")/values.yaml"

# Wait for Prometheus to be ready (StatefulSet, not Deployment)
echo "Waiting for Prometheus to be ready..."
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=prometheus -n $NAMESPACE --timeout=120s 2>/dev/null || true

# Create Prometheus Ingress
echo "Creating Prometheus ingress..."
kubectl apply -f "$(dirname "$0")/ingress.yaml"

echo "✓ Prometheus Stack installed successfully"
echo ""
echo "Access Prometheus at: http://prometheus.k3d.local (via Traefik ingress)"
echo "Namespace: $NAMESPACE"
echo "Status: kubectl get pods -n $NAMESPACE"
