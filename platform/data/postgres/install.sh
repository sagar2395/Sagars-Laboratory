#!/usr/bin/env bash
# Installs CloudNativePG operator + a 2-instance PostgreSQL cluster.
# The 2-instance setup enables a built-in failover drill (delete primary → watch promotion).
set -euo pipefail

OPERATOR_NS=cnpg-system
DB_NS=postgres
CNPG_CHART_VERSION="${CNPG_CHART_VERSION:-0.22.0}"
PG_VERSION="${PG_VERSION:-16}"

echo "Installing CloudNativePG operator ${CNPG_CHART_VERSION}..."

helm repo add cnpg https://cloudnative-pg.github.io/charts --force-update
helm repo update

helm upgrade --install cnpg cnpg/cloudnative-pg \
  --namespace "${OPERATOR_NS}" \
  --create-namespace \
  --version "${CNPG_CHART_VERSION}" \
  -f "$(dirname "$0")/values.yaml" \
  --wait --timeout 5m

# Wait for operator webhook to be ready
echo "Waiting for CNPG operator webhook..."
kubectl rollout status deployment/cnpg-cloudnative-pg \
  -n "${OPERATOR_NS}" --timeout=90s

# Create database namespace + the Cluster CR
kubectl create namespace "${DB_NS}" --dry-run=client -o yaml | kubectl apply -f -

cat <<EOF | kubectl apply -f -
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: lab-postgres
  namespace: ${DB_NS}
spec:
  instances: 2
  imageName: ghcr.io/cloudnative-pg/postgresql:${PG_VERSION}
  primaryUpdateStrategy: unsupervised
  storage:
    size: 1Gi
  resources:
    requests:
      cpu: 50m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 512Mi
  postgresql:
    parameters:
      max_connections: "30"
      shared_buffers: 32MB
EOF

echo "Waiting for Cluster to become ready (may take 2–3 min)..."
# Poll until both instances are Running; timeout after 5 min
DEADLINE=$(( $(date +%s) + 300 ))
while true; do
  READY=$(kubectl get cluster lab-postgres -n "${DB_NS}" \
    -o jsonpath='{.status.readyInstances}' 2>/dev/null || echo "0")
  if [ "${READY}" = "2" ]; then
    break
  fi
  if [ "$(date +%s)" -ge "${DEADLINE}" ]; then
    echo "Warning: Cluster not fully ready within 5 min — check with:"
    echo "  kubectl describe cluster lab-postgres -n ${DB_NS}"
    break
  fi
  sleep 10
done

echo ""
echo "CloudNativePG ${CNPG_CHART_VERSION} + PostgreSQL ${PG_VERSION} installed."
echo "  Primary:  lab-postgres-rw.${DB_NS}.svc.cluster.local:5432"
echo "  Replicas: lab-postgres-ro.${DB_NS}.svc.cluster.local:5432"
echo "  Credentials: kubectl get secret lab-postgres-superuser -n ${DB_NS} -o jsonpath='{.data.password}' | base64 -d"
echo "  Failover drill: kubectl delete pod lab-postgres-1 -n ${DB_NS} && watch kubectl get pods -n ${DB_NS}"
