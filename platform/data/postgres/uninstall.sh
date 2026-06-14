#!/usr/bin/env bash
set -euo pipefail

OPERATOR_NS=cnpg-system
DB_NS=postgres

echo "Uninstalling CloudNativePG..."

# Delete the Cluster CR first (triggers PVC cleanup via finalizers)
if kubectl get cluster lab-postgres -n "${DB_NS}" >/dev/null 2>&1; then
  kubectl delete cluster lab-postgres -n "${DB_NS}" --timeout=60s || true
  echo "Cluster CR deleted."
fi

# Delete any leftover PVCs (CNPG creates them with retain policy)
if kubectl get namespace "${DB_NS}" >/dev/null 2>&1; then
  kubectl delete pvc --all -n "${DB_NS}" --timeout=60s 2>/dev/null || true
  kubectl delete namespace "${DB_NS}" --timeout=60s || true
  echo "Namespace '${DB_NS}' deleted."
fi

# Uninstall operator
if helm status cnpg -n "${OPERATOR_NS}" >/dev/null 2>&1; then
  helm uninstall cnpg -n "${OPERATOR_NS}"
  echo "CNPG operator Helm release removed."
fi

# Remove CRDs installed by CNPG
for crd in $(kubectl get crd -o name 2>/dev/null | grep 'postgresql.cnpg.io'); do
  kubectl delete "${crd}" 2>/dev/null || true
done

if kubectl get namespace "${OPERATOR_NS}" >/dev/null 2>&1; then
  kubectl delete namespace "${OPERATOR_NS}" --timeout=60s || true
  echo "Namespace '${OPERATOR_NS}' deleted."
fi

echo "CloudNativePG uninstalled."
