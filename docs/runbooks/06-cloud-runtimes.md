# Runbook 06 — Cloud Runtimes (AKS / EKS)

> Goal: provision a real cloud cluster with Terraform, run an app + a scenario,
> and tear it all down. Verifies Phase 6. **This incurs cloud spend — be deliberate.**

## ⚠️ Before you start

- This creates billable resources (cluster, nodes, EKS NAT Gateway ~$32/mo, etc.).
- Always run the **Cleanup** section the same day.
- Configure remote Terraform state first (Task 027) so state isn't lost.

## Prereqs

- **AKS:** Azure CLI logged in (`az login`), a subscription, sufficient quota.
- **EKS:** AWS CLI configured (`aws sts get-caller-identity` works), quota.
- `make setup-tools PROFILE=eks` (installs terraform + aws-cli) or the AKS equivalent.

## Steps (AKS shown; EKS identical with `--profile eks`)

```bash
# 1. Select the cloud profile
#    edit .env: PROFILE=aks   (and set DOMAIN_SUFFIX, REGISTRY_TYPE=acr, etc.)

# 2. Provision
bin/labctl runtime up --profile aks      # terraform init/apply + get-credentials
kubectl get nodes                        # cloud nodes appear

# 3. Platform + app on cloud Helm profile
make platform-up
make build  APP_NAME=go-api BUILD_STRATEGY=acr     # push to ACR (ECR for EKS)
make deploy APP_NAME=go-api HELM_VALUES=values-cloud.yaml

# 4. Run one scenario to prove portability
bin/labctl scenario up observability-sre

# 5. TEAR DOWN (do not skip)
bin/labctl scenario down observability-sre
make teardown
bin/labctl runtime down --profile aks    # terraform destroy
```

## Expected

- `runtime up` provisions a cluster via the Terraform module
  (`foundation/terraform/modules/aks` or `eks`) using the dev environment sizing.
- The cloud build strategy (`engine/build/acr.sh` / `ecr.sh`) pushes the image.
- App deploys with `values-cloud.yaml` (nginx ingress class).
- The observability scenario runs the same as on k3d.
- Chaos Mesh, if used, finds the correct containerd socket for the cloud runtime
  (Task 026 — not the hardcoded k3s path).
- `runtime down` removes **all** billable resources (verify in the cloud console).

## Record after first successful run

- Time to provision / destroy.
- Approximate cost.
- Any manual steps still required (fold them back into the scripts / tasks).

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| Terraform state lost between runs | Configure remote backend (Task 027). |
| Image pull denied on cloud | Registry auth: `az acr login` / `aws ecr get-login-password`. |
| Chaos experiments fail on cloud | Containerd socket path is runtime-specific (Task 026). |
| Ingress has no external IP | Cloud LB still provisioning, or wrong ingress class for the provider. |
| Leftover charges after teardown | Check the cloud console for orphaned LBs, disks, NAT, registries. |
