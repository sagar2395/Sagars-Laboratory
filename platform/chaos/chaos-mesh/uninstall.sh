#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="chaos-mesh"

echo "Uninstalling Chaos Mesh..."

# Remove all chaos experiments first
echo "Cleaning up active chaos experiments..."
for kind in PodChaos NetworkChaos StressChaos IOChaos TimeChaos DNSChaos HTTPChaos; do
  kubectl delete "$kind" --all --all-namespaces >/dev/null 2>&1 || true
done

# Uninstall Helm release
echo "Removing Chaos Mesh Helm release..."
helm uninstall chaos-mesh -n $NAMESPACE >/dev/null 2>&1 || true

# Clean up CRDs
echo "Removing Chaos Mesh CRDs..."
kubectl delete crd -l app.kubernetes.io/part-of=chaos-mesh >/dev/null 2>&1 || true
kubectl get crd -o name | grep chaos-mesh.org | xargs -r kubectl delete >/dev/null 2>&1 || true

# Delete namespace
echo "Removing namespace..."
kubectl delete namespace $NAMESPACE --ignore-not-found >/dev/null 2>&1 || true

echo ""
echo "Chaos Mesh uninstalled successfully"
