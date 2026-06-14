#!/usr/bin/env bash
# Installs Linkerd service mesh via Helm.
# Generates trust-anchor and issuer certs with openssl (portable, no step CLI needed).
set -euo pipefail

NAMESPACE=linkerd
CRDS_CHART_VERSION="${LINKERD_CRDS_CHART_VERSION:-1.9.0}"
CP_CHART_VERSION="${LINKERD_CP_CHART_VERSION:-1.16.0}"
MESH_APP_NAMESPACE="${MESH_APP_NAMESPACE:-go-api}"

echo "Installing Linkerd (crds ${CRDS_CHART_VERSION}, control-plane ${CP_CHART_VERSION})..."

# ── Certificate generation ────────────────────────────────────────────────────
TMPDIR_CERTS=$(mktemp -d)
# shellcheck disable=SC2064
trap "rm -rf ${TMPDIR_CERTS}" EXIT

echo "Generating trust anchor and issuer certificates..."

# Trust anchor CA config
cat > "${TMPDIR_CERTS}/ca.cnf" <<'OPENSSL_CNF'
[req]
distinguished_name = dn
x509_extensions    = v3_ca
prompt             = no

[dn]
CN = root.linkerd.cluster.local

[v3_ca]
basicConstraints       = critical,CA:TRUE
keyUsage               = critical,cRLSign,keyCertSign
subjectKeyIdentifier   = hash
OPENSSL_CNF

# Issuer cert config (subordinate CA signed by trust anchor)
cat > "${TMPDIR_CERTS}/issuer.cnf" <<'OPENSSL_CNF'
[req]
distinguished_name = dn
prompt             = no

[dn]
CN = identity.linkerd.cluster.local

[v3_issuer]
basicConstraints     = critical,CA:TRUE
keyUsage             = critical,cRLSign,keyCertSign
subjectKeyIdentifier = hash
OPENSSL_CNF

# Generate trust anchor key + self-signed cert
openssl ecparam -name prime256v1 -genkey -noout \
  -out "${TMPDIR_CERTS}/ca.key" 2>/dev/null
openssl req -new -x509 -days 8760 \
  -key "${TMPDIR_CERTS}/ca.key" \
  -out "${TMPDIR_CERTS}/ca.crt" \
  -config "${TMPDIR_CERTS}/ca.cnf" \
  -extensions v3_ca 2>/dev/null

# Generate issuer key + CSR
openssl ecparam -name prime256v1 -genkey -noout \
  -out "${TMPDIR_CERTS}/issuer.key" 2>/dev/null
openssl req -new \
  -key "${TMPDIR_CERTS}/issuer.key" \
  -out "${TMPDIR_CERTS}/issuer.csr" \
  -config "${TMPDIR_CERTS}/issuer.cnf" 2>/dev/null

# Sign issuer cert with trust anchor
openssl x509 -req -days 8760 \
  -in  "${TMPDIR_CERTS}/issuer.csr" \
  -CA  "${TMPDIR_CERTS}/ca.crt" \
  -CAkey "${TMPDIR_CERTS}/ca.key" \
  -CAcreateserial \
  -out "${TMPDIR_CERTS}/issuer.crt" \
  -extfile "${TMPDIR_CERTS}/issuer.cnf" \
  -extensions v3_issuer 2>/dev/null

CA_CRT=$(cat "${TMPDIR_CERTS}/ca.crt")
ISSUER_CRT=$(cat "${TMPDIR_CERTS}/issuer.crt")
ISSUER_KEY=$(cat "${TMPDIR_CERTS}/issuer.key")

# ── Helm install ──────────────────────────────────────────────────────────────
helm repo add linkerd https://helm.linkerd.io/stable --force-update
helm repo update

echo "Installing linkerd-crds..."
helm upgrade --install linkerd-crds linkerd/linkerd-crds \
  --namespace "${NAMESPACE}" \
  --create-namespace \
  --version "${CRDS_CHART_VERSION}" \
  --wait --timeout 3m

echo "Installing linkerd-control-plane..."
helm upgrade --install linkerd-control-plane linkerd/linkerd-control-plane \
  --namespace "${NAMESPACE}" \
  --version "${CP_CHART_VERSION}" \
  -f "$(dirname "$0")/values.yaml" \
  --set-string identityTrustAnchorsPEM="${CA_CRT}" \
  --set-string identity.issuer.tls.crtPEM="${ISSUER_CRT}" \
  --set-string identity.issuer.tls.keyPEM="${ISSUER_KEY}" \
  --wait --timeout 5m

# Enable proxy injection in the app namespace
if kubectl get namespace "${MESH_APP_NAMESPACE}" >/dev/null 2>&1; then
  echo "Enabling Linkerd proxy injection in namespace ${MESH_APP_NAMESPACE}..."
  kubectl annotate namespace "${MESH_APP_NAMESPACE}" \
    linkerd.io/inject=enabled --overwrite
  echo "Rolling pods in ${MESH_APP_NAMESPACE} to acquire proxies..."
  kubectl rollout restart deployment -n "${MESH_APP_NAMESPACE}" 2>/dev/null || true
fi

echo ""
echo "Linkerd installed."
echo "  Control plane: kubectl get pods -n ${NAMESPACE}"
echo "  Linkerd check: linkerd check (if CLI installed)"
echo "  Proxy injection active on namespace: ${MESH_APP_NAMESPACE}"
