# Task 027: Configure Terraform Remote State Backend

## Priority
P2

## Assigned To
DevOps

## Description
Both `foundation/terraform/environments/dev/main.tf` and `foundation/terraform/environments/staging/main.tf` have remote state backend blocks commented out. Without remote state, every operator who runs `terraform apply` uses a local `terraform.tfstate` file. On a team (or across CI runs), this causes state divergence, accidental resource duplication, and potential for destructive concurrent applies. The backends must be uncommented and documented, with clear instructions for one-time backend initialization.

## Files to Modify
- `foundation/terraform/environments/dev/main.tf`
- `foundation/terraform/environments/staging/main.tf`
- `foundation/terraform/environments/dev/backend.tf` *(create)*
- `foundation/terraform/environments/staging/backend.tf` *(create)*
- `docs/cloud-runtimes.md`

## Implementation Notes
Move backend configuration into separate `backend.tf` files so they can be gitignored or overridden without touching `main.tf`:

**For AKS runtime (`backend.tf` in each environment):**
```hcl
terraform {
  backend "azurerm" {
    resource_group_name  = "sagars-lab-tfstate"
    storage_account_name = "sagarslabtfstate"
    container_name       = "tfstate"
    key                  = "dev/terraform.tfstate"  # or staging/
  }
}
```

**For EKS runtime:**
```hcl
terraform {
  backend "s3" {
    bucket         = "sagars-lab-tfstate"
    key            = "dev/terraform.tfstate"       # or staging/
    region         = "us-east-1"
    dynamodb_table = "sagars-lab-tfstate-lock"
    encrypt        = true
  }
}
```

Since the runtime (AKS vs EKS) is determined at `terraform apply` time, provide BOTH backend options as separate commented-out blocks with clear instructions to uncomment the one that matches the `runtime` variable. Add a `backend.tf.example` file next to each environment that shows both options.

Update `docs/cloud-runtimes.md` with a "State Management" section explaining how to initialize the backend (`terraform init -migrate-state`), the one-time setup steps for the storage resources, and why local state is only acceptable for local-only k3d experiments.

## Acceptance Criteria
- [ ] `backend.tf` files exist in both `dev/` and `staging/` environments.
- [ ] Both AWS S3 and Azure Blob backend options are documented in each file.
- [ ] `docs/cloud-runtimes.md` has a "State Management" section.
- [ ] Existing `main.tf` files do not have backend blocks (moved to `backend.tf`).
- [ ] A `backend.tf.example` shows how to configure each provider.

## Testing Instructions
Run `terraform init` in `foundation/terraform/environments/dev/` with one of the example backends configured — confirm no errors. Confirm `terraform validate` passes.

## Dependencies
None
