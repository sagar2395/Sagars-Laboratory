#!/usr/bin/env bash
# Installs KEDA (Kubernetes Event-Driven Autoscaling) via Helm.
# KEDA enables ScaledObjects powered by Prometheus, Kafka, and many other
# triggers — it works alongside the native HPA by managing it under the hood.
set -euo pipefail

NAMESPACE=keda
CHART_VERSION="${KEDA_CHART_VERSION:-2.15.1}"

echo "Installing KEDA ${CHART_VERSION}..."

helm repo add kedacore https://kedacore.github.io/charts --force-update
helm repo update

helm upgrade --install keda kedacore/keda \
  --namespace "${NAMESPACE}" \
  --create-namespace \
  --version "${CHART_VERSION}" \
  -f "$(dirname "$0")/values.yaml" \
  --wait --timeout 5m

echo ""
echo "KEDA ${CHART_VERSION} installed in namespace ${NAMESPACE}."
echo "  Verify: kubectl get pods -n ${NAMESPACE}"
echo "  Docs:   https://keda.sh/docs/"
echo ""
echo "  To use the Prometheus scaler:"
echo "    kubectl apply -f scenarios/autoscaling-under-load/manifests/scaledobject.yaml"
