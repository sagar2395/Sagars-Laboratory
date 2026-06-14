#!/usr/bin/env bash
# Installs Kafka via Bitnami Helm chart (KRaft mode, ephemeral storage for k3d).
set -euo pipefail

NAMESPACE=kafka
CHART_VERSION="${KAFKA_CHART_VERSION:-26.8.5}"

echo "Installing Kafka ${CHART_VERSION}..."

helm repo add bitnami https://charts.bitnami.com/bitnami --force-update
helm repo update

helm upgrade --install kafka bitnami/kafka \
  --namespace "${NAMESPACE}" \
  --create-namespace \
  --version "${CHART_VERSION}" \
  -f "$(dirname "$0")/values.yaml" \
  --wait --timeout 8m

echo ""
echo "Kafka ${CHART_VERSION} installed in namespace ${NAMESPACE}."
echo "  Bootstrap: kafka.kafka.svc.cluster.local:9092"
echo "  Smoke test:"
echo "    kubectl run kafka-test -it --rm --restart=Never --image=bitnami/kafka:latest -- \\"
echo "      kafka-console-producer.sh --bootstrap-server kafka.${NAMESPACE}.svc.cluster.local:9092 --topic test"
