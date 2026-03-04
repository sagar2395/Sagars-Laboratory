#!/bin/bash

NAMESPACE="monitoring"

echo "Prometheus Stack Status:"
echo "========================"
echo ""

if ! kubectl get namespace $NAMESPACE &>/dev/null; then
  echo "✗ Namespace '$NAMESPACE' not found"
  exit 1
fi

echo "Pods in '$NAMESPACE' namespace:"
kubectl get pods -n $NAMESPACE

echo ""
echo "Services in '$NAMESPACE' namespace:"
kubectl get svc -n $NAMESPACE

echo ""
echo "Prometheus Ingress:"
kubectl get ingress -n $NAMESPACE

echo ""
echo "Prometheus metrics endpoint:"
echo "  - External: http://prometheus.k3d.local"
echo "  - Internal: http://prometheus-kube-prometheus-prometheus.$NAMESPACE.svc.cluster.local:9090"
