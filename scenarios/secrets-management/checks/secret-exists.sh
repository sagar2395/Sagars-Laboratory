#!/usr/bin/env bash
# Checks that the ESO-managed secret exists in the go-api namespace.
set -euo pipefail

if kubectl get secret go-api-external-secret -n go-api >/dev/null 2>&1; then
  KEYS=$(kubectl get secret go-api-external-secret -n go-api \
    -o jsonpath='{.data}' 2>/dev/null | grep -o '"[^"]*":' | tr -d '":'  | tr '\n' ' ')
  echo "Secret 'go-api-external-secret' exists with keys: ${KEYS}"
  exit 0
else
  echo "Secret 'go-api-external-secret' not found in namespace go-api."
  echo "ESO may not have synced yet — check: kubectl describe externalsecret go-api-secrets -n go-api"
  exit 1
fi
