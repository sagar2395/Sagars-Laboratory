runtime-up:
	@echo "Creating k3d cluster..."
	@bash runtimes/k3d/up.sh $(CLUSTER_NAME)

runtime-down:
	@echo "Shutting down k3d cluster..."
	@bash runtimes/k3d/down.sh