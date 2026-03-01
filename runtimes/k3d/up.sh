# Create a k3d cluster
# Expose HTTP/HTTPS ports on the host so that ingress rules using
# hostnames (e.g. go-api.k3d.local) resolve to localhost and traffic
# reaches the load‑balancer pod.  Environment variables allow overrides.

CLUSTER_NAME="${1:-two-node-cluster}"
HTTP_PORT="${HTTP_PORT:-80}"
HTTPS_PORT="${HTTPS_PORT:-443}"

k3d cluster create "$CLUSTER_NAME" --agents 2 \
  -p "${HTTP_PORT}:80@loadbalancer" \
  -p "${HTTPS_PORT}:443@loadbalancer"

kubectl config use-context "k3d-$CLUSTER_NAME"