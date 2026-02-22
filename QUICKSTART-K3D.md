# Go API on k3d - Quick Start Guide

This guide helps you quickly set up and deploy the Go API application on a k3d (Kubernetes in Docker) cluster.

## Prerequisites

- Docker installed and running
- Make installed
- This repository cloned

## Quick Start (5 minutes)

### 1. Set Up Tools and Cluster

```bash
# Install required tools (kubectl, k3d, docker)
make setup-tools PROFILE=k3d

# Create k3d cluster
make cluster-up CLUSTER_NAME=sagar-lab
```

### 2. Build and Deploy Go API

```bash
# Build Docker image
make go-docker-build

# Import image into k3d
make go-docker-import CLUSTER_NAME=sagar-lab

# Deploy to k3d using Helm
make deploy-go-api

# (Optional) View Helm chart manifests before deploying
make helm-validate
```

### 3. Access the Application

Add to `/etc/hosts`:
```
127.0.0.1 go-api.k3d.local
```

Then access:
- **Application**: http://go-api.k3d.local/
- **Health Status**: http://go-api.k3d.local/health
- **Readiness Status**: http://go-api.k3d.local/ready
- **Prometheus Metrics**: http://go-api.k3d.local/metrics

### 4. Monitor the Deployment

```bash
# Watch pods
kubectl get pods -n go-api -w

# View logs
kubectl logs -n go-api -l app=go-api -f

# Check deployment status
kubectl describe deployment -n go-api go-api

# Check ingress
kubectl get ingress -n go-api
```

## Available Make Targets

### Setup & Environment

```bash
# Install tools for specific profile
make setup-tools PROFILE=k3d           # Kubernetes tools + Docker
make setup-tools PROFILE=aks           # Kubernetes tools + Azure CLI
make setup-tools PROFILE=common        # Kubernetes tools only
```

### Cluster Management

```bash
# Cluster lifecycle
make cluster-up CLUSTER_NAME=sagar-lab
make cluster-down

# Verify cluster
kubectl cluster-info
k3d cluster list
```

### Build & Test Locally

```bash
# Build Go binary
make go-build

# Run locally (port 8080)
make go-run

# Build Docker image
make go-docker-build
```

### Kubernetes Deployment

```bash
# Validate Helm chart
make helm-lint
make helm-validate

# Deploy with different configurations
make deploy-go-api                                          # Default config
make deploy-go-api HELM_VALUES=values-dev.yaml             # Development
make deploy-go-api HELM_VALUES=values-prod-like.yaml       # Production-like
make deploy-go-api HELM_VALUES=values-test.yaml            # Testing

# Manage deployment
make undeploy-go-api
```

## Configuration Profiles

### Development Profile (`values-dev.yaml`)
- Single replica for lighter resource usage
- No autoscaling
- Minimal resource requests

```bash
make deploy-go-api HELM_VALUES=values-dev.yaml
```

### Production-Like Profile (`values-prod-like.yaml`)
- 3 replicas
- Autoscaling enabled (3-10 pods)
- Higher resource requests/limits
- Pod anti-affinity for distribution

```bash
make deploy-go-api HELM_VALUES=values-prod-like.yaml
```

### Test Profile (`values-test.yaml`)
- Tests readiness failure scenario
- Useful for testing pod replacement and healing

```bash
make deploy-go-api HELM_VALUES=values-test.yaml
```

## Port Forwarding (Alternative Access)

If you don't want to modify `/etc/hosts`:

```bash
# Forward local port to service
kubectl port-forward -n go-api svc/go-api 8080:8080

# Access at http://localhost:8080
```

## Checking Logs

```bash
# View logs from all pods
kubectl logs -n go-api -l app=go-api

# Stream logs
kubectl logs -n go-api -l app=go-api -f

# View logs from specific pod
kubectl logs -n go-api <pod-name>

# View previous pod logs (if crashed)
kubectl logs -n go-api <pod-name> --previous
```

## Scaling Pods Manually

```bash
# Scale to specific number
kubectl scale deployment -n go-api go-api --replicas=5

# View current replicas
kubectl get deployment -n go-api go-api
```

## Testing Readiness Failure

```bash
# Use test profile to simulate readiness failure
make deploy-go-api HELM_VALUES=values-test.yaml

# Watch pods being replaced
kubectl get pods -n go-api -w

# Undeploy when done
make undeploy-go-api
```

## Prometheus Metrics

The application exposes Prometheus metrics at `/metrics`:

```bash
# Get metrics from running pod
kubectl port-forward -n go-api svc/go-api 8080:8080 &
curl http://localhost:8080/metrics | grep http_requests

# Available metrics:
# - http_requests_total: Counter of total HTTP requests
# - http_request_duration_seconds: Histogram of request duration
```

## Helm Commands

```bash
# List releases
helm list -n go-api

# Show release details
helm status go-api -n go-api

# Get values used
helm get values go-api -n go-api

# View manifest
helm get manifest go-api -n go-api

# Update release
helm upgrade go-api apps/go-api/deploy/helm/go-api -f <values.yaml>

# Rollback release
helm rollback go-api 1

# Uninstall release
helm uninstall go-api -n go-api
```

## Troubleshooting

### Pods not starting

```bash
# Check pod events
kubectl describe pod -n go-api <pod-name>

# View crash logs
kubectl logs -n go-api <pod-name> --previous

# Check resource availability
kubectl describe nodes
```

### Image not found error

```bash
# Ensure image is built
docker images | grep go-api

# Import into k3d if missing
k3d image import go-api:latest -c sagar-lab

# Verify in cluster
k3d image list -c sagar-lab
```

### Ingress not working

```bash
# Check ingress resource
kubectl describe ingress -n go-api

# Check traefik logs (k3d's ingress controller)
kubectl logs -n kube-system -l app=traefik

# Verify host entry in /etc/hosts
cat /etc/hosts | grep go-api

# Manual port-forward as alternative
kubectl port-forward -n go-api svc/go-api 8080:8080
```

### Slow startup

```bash
# Check resource limits
kubectl describe pod -n go-api <pod-name> | grep -A 5 "Limits\|Requests"

# Check node resources
kubectl top nodes
kubectl top pods -n go-api

# Increase limits in values.yaml if needed
```

## Cleanup

```bash
# Remove Go API deployment
make undeploy-go-api

# Remove k3d cluster
make cluster-down

# Remove Docker image
docker rmi go-api:latest
```

## Common Workflows

### Full Clean Restart

```bash
make cluster-down
make cluster-up CLUSTER_NAME=sagar-lab
make go-docker-build
make go-docker-import CLUSTER_NAME=sagar-lab
make deploy-go-api
```

### Rebuild and Redeploy

```bash
make go-docker-build
make go-docker-import CLUSTER_NAME=sagar-lab
helm upgrade go-api apps/go-api/deploy/helm/go-api -n go-api
kubectl rollout status deployment/go-api -n go-api
```

### Test Different Configurations

```bash
# Deploy with dev config
make deploy-go-api HELM_VALUES=values-dev.yaml
sleep 10

# Undeploy
make undeploy-go-api

# Deploy with production config
make deploy-go-api HELM_VALUES=values-prod-like.yaml
```

## Next Steps

1. **Add persistent volume** - For data storage needs
2. **Configure monitoring** - Set up Prometheus scraping with ServiceMonitor
3. **Add authentication** - Implement API authentication
4. **Enable HTTPS** - Configure TLS certificates for ingress
5. **Add more services** - Deploy additional microservices
6. **Setup CI/CD** - Automate builds and deployments

## Resources

- [Helm Documentation](https://helm.sh/docs/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [k3d Documentation](https://k3d.io/)
- [Traefik Documentation](https://doc.traefik.io/)

## Support

For issues or questions:
1. Check application logs: `kubectl logs -n go-api -l app=go-api`
2. Check pod status: `kubectl get pods -n go-api`
3. Check events: `kubectl describe events -n go-api`
4. Review Helm chart: [Go API Helm Chart README](./helm/README.md)
