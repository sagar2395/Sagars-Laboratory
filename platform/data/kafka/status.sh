#!/usr/bin/env bash
set -euo pipefail

NAMESPACE=kafka

echo "=== Kafka Status ==="
echo ""

if ! kubectl get namespace "${NAMESPACE}" >/dev/null 2>&1; then
  echo "Kafka is not installed (namespace ${NAMESPACE} not found)"
  exit 0
fi

echo "Pods:"
kubectl get pods -n "${NAMESPACE}" 2>/dev/null || echo "  No pods found"
echo ""

echo "Services:"
kubectl get svc -n "${NAMESPACE}" 2>/dev/null || echo "  No services found"
echo ""

# Check controller readiness
READY=$(kubectl get deployment -n "${NAMESPACE}" -o jsonpath='{range .items[*]}{.metadata.name}{": "}{.status.readyReplicas}{"/"}{.status.replicas}{"\n"}{end}' 2>/dev/null)
if [ -n "${READY}" ]; then
  echo "Deployments:"
  while IFS= read -r line; do echo "  ${line}"; done <<< "${READY}"
  echo ""
fi

echo "Bootstrap address:"
echo "  kafka.${NAMESPACE}.svc.cluster.local:9092"
echo ""
echo "Quick smoke test:"
echo "  kubectl run kafka-smoke -it --rm --restart=Never --image=bitnami/kafka:latest -- \\"
echo "    kafka-topics.sh --bootstrap-server kafka.${NAMESPACE}.svc.cluster.local:9092 --list"
