#!/usr/bin/env bash
# Installs External Secrets Operator + wires it to the lab Vault instance.
# Requires Vault to be running (SECRETS_VAULT=1 make platform-secrets-vault-up).
set -euo pipefail

OPERATOR_NS=external-secrets
APP_NAMESPACE="${VAULT_APP_NAMESPACE:-go-api}"
ESO_CHART_VERSION="${ESO_CHART_VERSION:-0.10.5}"
VAULT_ADDR="http://vault.vault.svc.cluster.local:8200"

# Preflight: Vault must be installed and reachable
if ! kubectl get namespace vault >/dev/null 2>&1; then
  echo "ERROR: Vault is not installed. Install it first:"
  echo "  SECRETS_VAULT=1 make platform-secrets-vault-up"
  exit 1
fi
if ! kubectl get deployment vault -n vault >/dev/null 2>&1; then
  echo "ERROR: Vault deployment not found. Run: SECRETS_VAULT=1 make platform-secrets-vault-up"
  exit 1
fi

echo "Installing External Secrets Operator ${ESO_CHART_VERSION}..."

helm repo add external-secrets https://charts.external-secrets.io --force-update
helm repo update

helm upgrade --install external-secrets external-secrets/external-secrets \
  --namespace "${OPERATOR_NS}" \
  --create-namespace \
  --version "${ESO_CHART_VERSION}" \
  --set installCRDs=true \
  --wait --timeout 5m

echo "Waiting for ESO webhooks to be ready..."
kubectl rollout status deployment/external-secrets -n "${OPERATOR_NS}" --timeout=90s
kubectl rollout status deployment/external-secrets-webhook -n "${OPERATOR_NS}" --timeout=90s

# Create the Vault root token secret in the app namespace
echo "Creating Vault token secret in namespace ${APP_NAMESPACE}..."
if ! kubectl get namespace "${APP_NAMESPACE}" >/dev/null 2>&1; then
  echo "Warning: namespace ${APP_NAMESPACE} not found — skipping SecretStore setup"
  echo "ESO is installed. Configure SecretStore manually when the namespace exists."
  exit 0
fi

kubectl create secret generic vault-token \
  --from-literal=token=root \
  -n "${APP_NAMESPACE}" \
  --dry-run=client -o yaml | kubectl apply -f -

# Create SecretStore pointing to Vault
cat <<EOF | kubectl apply -f -
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: vault-backend
  namespace: ${APP_NAMESPACE}
spec:
  provider:
    vault:
      server: "${VAULT_ADDR}"
      path: "secret"
      version: "v2"
      auth:
        tokenSecretRef:
          name: vault-token
          key: token
EOF

# Create ExternalSecret to sync go-api credentials
cat <<EOF | kubectl apply -f -
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: go-api-secrets
  namespace: ${APP_NAMESPACE}
spec:
  refreshInterval: 30s
  secretStoreRef:
    name: vault-backend
    kind: SecretStore
  target:
    name: go-api-external-secret
    creationPolicy: Owner
  data:
    - secretKey: db_password
      remoteRef:
        key: go-api
        property: db_password
    - secretKey: api_key
      remoteRef:
        key: go-api
        property: api_key
EOF

echo ""
echo "External Secrets Operator ${ESO_CHART_VERSION} installed."
echo "  Syncing: Vault secret/go-api → k8s secret 'go-api-external-secret' in ${APP_NAMESPACE}"
echo ""
echo "Rotation exercise:"
echo "  1. Update in Vault: vault kv put secret/go-api db_password=new-password"
echo "  2. Wait up to 30 s (ESO refresh interval)"
echo "  3. Verify:  kubectl get secret go-api-external-secret -n ${APP_NAMESPACE} -o jsonpath='{.data.db_password}' | base64 -d"
