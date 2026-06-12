#!/usr/bin/env bash
set -euo pipefail

NS="${TARGET_NAMESPACE:-go-api}"
DEPLOY="${TARGET_WORKLOAD:-go-api}"
MARK="labfault-crashloop-bad-config"

if [ "$(kubectl -n "$NS" get deploy "$DEPLOY" -o "jsonpath={.metadata.annotations.$MARK}" 2>/dev/null)" = "injected" ]; then
  echo "Fault already injected — nothing to do."
  exit 0
fi

echo "Injecting: replacing $DEPLOY's container command with one that exits immediately..."
kubectl -n "$NS" patch deploy "$DEPLOY" --type=json \
  -p '[{"op":"add","path":"/spec/template/spec/containers/0/command","value":["/bin/false"]}]'
kubectl -n "$NS" annotate deploy "$DEPLOY" "$MARK=injected" --overwrite

echo "Injected. New pods will crash-loop; the rollout will never complete."
