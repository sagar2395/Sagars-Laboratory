#!/usr/bin/env bash
# Installs HashiCorp Vault in dev mode for the lab simulator.
# Dev mode: single pod, in-memory, root token = "root", auto-unsealed.
# WARNING: dev mode is NOT suitable for production — use the non-dev values
# profile (values-prod-like.yaml) as a starting point for real deployments.
set -euo pipefail

NAMESPACE=vault
CHART_VERSION="${VAULT_CHART_VERSION:-0.28.1}"
DOMAIN_SUFFIX="${DOMAIN_SUFFIX:-k3d.local}"

echo "Installing Vault ${CHART_VERSION} (dev mode)..."

helm repo add hashicorp https://helm.releases.hashicorp.com --force-update
helm repo update

helm upgrade --install vault hashicorp/vault \
  --namespace "${NAMESPACE}" \
  --create-namespace \
  --version "${CHART_VERSION}" \
  -f "$(dirname "$0")/values.yaml" \
  --wait --timeout 5m

echo "Waiting for Vault to be ready..."
kubectl rollout status deployment/vault -n "${NAMESPACE}" --timeout=90s

echo "Seeding demo KV secret (secret/go-api)..."
kubectl exec -n "${NAMESPACE}" deploy/vault -- \
  vault kv put secret/go-api \
    db_password="lab-demo-changeme" \
    api_key="lab-demo-api-key-001"

echo ""
echo "Vault ${CHART_VERSION} installed."
echo "  UI: http://vault.${DOMAIN_SUFFIX}"
echo "  Root token: root  (dev mode — DO NOT use in production)"
echo "  CLI: vault login -address=http://vault.${DOMAIN_SUFFIX} root"
echo "  Demo secret: vault kv get -address=http://vault.${DOMAIN_SUFFIX} secret/go-api"
echo ""
echo "  For External Secrets Operator integration:"
echo "    SECRETS_ESO=1 make platform-secrets-eso-up"
