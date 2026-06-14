#!/usr/bin/env bash
# Checks that the lab-events topic exists in the Kafka cluster.
set -euo pipefail

BOOTSTRAP="kafka.kafka.svc.cluster.local:9092"

TOPICS=$(kubectl exec -n kafka deploy/kafka-controller -- \
  kafka-topics.sh --bootstrap-server "${BOOTSTRAP}" --list 2>/dev/null || echo "")

if echo "${TOPICS}" | grep -q '^lab-events$'; then
  echo "Topic 'lab-events' exists."
  exit 0
else
  echo "Topic 'lab-events' not found. Topics: ${TOPICS}"
  exit 1
fi
