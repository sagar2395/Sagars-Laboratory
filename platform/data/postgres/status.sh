#!/usr/bin/env bash
set -euo pipefail

OPERATOR_NS=cnpg-system
DB_NS=postgres

echo "=== PostgreSQL (CloudNativePG) Status ==="
echo ""

if ! kubectl get namespace "${OPERATOR_NS}" >/dev/null 2>&1; then
  echo "CloudNativePG is not installed (namespace ${OPERATOR_NS} not found)"
  exit 0
fi

echo "Operator (${OPERATOR_NS}):"
kubectl get pods -n "${OPERATOR_NS}" 2>/dev/null || echo "  No pods found"
echo ""

if ! kubectl get namespace "${DB_NS}" >/dev/null 2>&1; then
  echo "Database namespace '${DB_NS}' not found."
  exit 0
fi

echo "Cluster status:"
kubectl get cluster -n "${DB_NS}" 2>/dev/null || echo "  No clusters found"
echo ""

echo "Database pods:"
kubectl get pods -n "${DB_NS}" 2>/dev/null || echo "  No pods found"
echo ""

echo "Services:"
kubectl get svc -n "${DB_NS}" 2>/dev/null | grep -v kubernetes || echo "  No services found"
echo ""

# Show primary instance
PRIMARY=$(kubectl get cluster lab-postgres -n "${DB_NS}" \
  -o jsonpath='{.status.currentPrimary}' 2>/dev/null || echo "unknown")
echo "Current primary: ${PRIMARY}"
echo ""

echo "Connection:"
echo "  Primary (RW): lab-postgres-rw.${DB_NS}.svc.cluster.local:5432"
echo "  Replicas (RO): lab-postgres-ro.${DB_NS}.svc.cluster.local:5432"
echo "  Credentials: kubectl get secret lab-postgres-superuser -n ${DB_NS} -o jsonpath='{.data.password}' | base64 -d"
echo ""
echo "Failover drill:"
echo "  kubectl delete pod ${PRIMARY} -n ${DB_NS}"
echo "  watch kubectl get pods -n ${DB_NS}"
