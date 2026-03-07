# Terraform management — plan, apply, destroy for cloud infrastructure
# Modules: foundation/terraform/modules/{aks,eks}
# Environments: foundation/terraform/environments/{dev,staging}

TF_ENV ?= dev
TF_DIR := foundation/terraform/environments/$(TF_ENV)

terraform-init:
	@echo "Initializing Terraform in $(TF_DIR)..."
	@terraform -chdir=$(TF_DIR) init -input=false

terraform-plan:
	@echo "Planning Terraform changes (env=$(TF_ENV), runtime=$(PROFILE))..."
	@terraform -chdir=$(TF_DIR) plan \
		-var="runtime=$(PROFILE)" \
		-var="cluster_name=$(CLUSTER_NAME)"

terraform-apply:
	@echo "Applying Terraform changes (env=$(TF_ENV), runtime=$(PROFILE))..."
	@terraform -chdir=$(TF_DIR) apply -auto-approve -input=false \
		-var="runtime=$(PROFILE)" \
		-var="cluster_name=$(CLUSTER_NAME)"

terraform-destroy:
	@echo "Destroying Terraform resources (env=$(TF_ENV), runtime=$(PROFILE))..."
	@terraform -chdir=$(TF_DIR) destroy -auto-approve -input=false \
		-var="runtime=$(PROFILE)" \
		-var="cluster_name=$(CLUSTER_NAME)"

terraform-output:
	@terraform -chdir=$(TF_DIR) output

terraform-status:
	@echo "Terraform state (env=$(TF_ENV)):"
	@echo ""
	@terraform -chdir=$(TF_DIR) show -no-color 2>/dev/null || echo "  No state found. Run 'make terraform-init' first."
