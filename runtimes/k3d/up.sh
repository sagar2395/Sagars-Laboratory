# Create a k3d cluster
CLUSTER_NAME="${1:-two-node-cluster}"
k3d cluster create "$CLUSTER_NAME" --agents 2
kubectl config use-context "k3d-$CLUSTER_NAME"