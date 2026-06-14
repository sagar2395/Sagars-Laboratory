# Provider variables — override in .env or on the command line.
# These must match a subdirectory under platform/<category>/<provider>/.
INGRESS_PROVIDER  ?= traefik
METRICS_PROVIDER  ?= prometheus
# LOGGING_PROVIDER ?= loki          # set to enable logging stack
# TRACING_PROVIDER ?= tempo         # set to enable tracing stack
# MESH_PROVIDER    ?= istio         # set to enable service mesh (istio|linkerd)

platform-up: platform-ingress-up platform-monitoring-up
ifdef LOGGING_PROVIDER
platform-up: platform-logging-up
endif
ifdef TRACING_PROVIDER
platform-up: platform-tracing-up
endif
ifdef MESH_PROVIDER
platform-up: platform-mesh-up
endif

platform-down: platform-monitoring-down platform-ingress-down
ifdef LOGGING_PROVIDER
platform-down: platform-logging-down
endif
ifdef TRACING_PROVIDER
platform-down: platform-tracing-down
endif
ifdef MESH_PROVIDER
platform-down: platform-mesh-down
endif

platform-status: platform-ingress-status platform-monitoring-status
ifdef LOGGING_PROVIDER
platform-status: platform-logging-status
endif
ifdef TRACING_PROVIDER
platform-status: platform-tracing-status
endif
ifdef MESH_PROVIDER
platform-status: platform-mesh-status
endif

# Ingress (INGRESS_PROVIDER=traefik|nginx)
platform-ingress-up:
	@echo "[platform] ingress provider: $(INGRESS_PROVIDER)"
	bash platform/ingress/$(INGRESS_PROVIDER)/install.sh

platform-ingress-down:
	@echo "[platform] removing ingress provider: $(INGRESS_PROVIDER)"
	@bash platform/ingress/$(INGRESS_PROVIDER)/uninstall.sh

platform-ingress-status:
	@echo "=== Ingress ($(INGRESS_PROVIDER)) Status ==="
	@bash platform/ingress/$(INGRESS_PROVIDER)/status.sh

# Monitoring: metrics backend + Grafana (always paired)
platform-monitoring-up:
	@echo "[platform] metrics provider: $(METRICS_PROVIDER)"
	bash platform/monitoring/metrics/$(METRICS_PROVIDER)/install.sh
	bash platform/monitoring/grafana/install.sh

platform-monitoring-down:
	@echo "[platform] removing monitoring stack ($(METRICS_PROVIDER) + grafana)"
	@bash platform/monitoring/grafana/uninstall.sh
	@bash platform/monitoring/metrics/$(METRICS_PROVIDER)/uninstall.sh

platform-monitoring-status:
	@echo "=== Metrics ($(METRICS_PROVIDER)) Status ==="
	@bash platform/monitoring/metrics/$(METRICS_PROVIDER)/status.sh
	@echo ""
	@echo "=== Grafana Status ==="
	@bash platform/monitoring/grafana/status.sh

# Logging (LOGGING_PROVIDER=loki|...) — only when LOGGING_PROVIDER is set
platform-logging-up:
	@echo "[platform] logging provider: $(LOGGING_PROVIDER)"
	bash platform/logging/$(LOGGING_PROVIDER)/install.sh

platform-logging-down:
	@echo "[platform] removing logging provider: $(LOGGING_PROVIDER)"
	@bash platform/logging/$(LOGGING_PROVIDER)/uninstall.sh

platform-logging-status:
	@echo "=== Logging ($(LOGGING_PROVIDER)) Status ==="
	@bash platform/logging/$(LOGGING_PROVIDER)/status.sh

# Tracing (TRACING_PROVIDER=tempo|...) — only when TRACING_PROVIDER is set
platform-tracing-up:
	@echo "[platform] tracing provider: $(TRACING_PROVIDER)"
	bash platform/tracing/$(TRACING_PROVIDER)/install.sh

platform-tracing-down:
	@echo "[platform] removing tracing provider: $(TRACING_PROVIDER)"
	@bash platform/tracing/$(TRACING_PROVIDER)/uninstall.sh

platform-tracing-status:
	@echo "=== Tracing ($(TRACING_PROVIDER)) Status ==="
	@bash platform/tracing/$(TRACING_PROVIDER)/status.sh

# Mesh (MESH_PROVIDER=istio|linkerd) — only when MESH_PROVIDER is set
platform-mesh-up:
	@echo "[platform] mesh provider: $(MESH_PROVIDER)"
	bash platform/mesh/$(MESH_PROVIDER)/install.sh

platform-mesh-down:
	@echo "[platform] removing mesh provider: $(MESH_PROVIDER)"
	@bash platform/mesh/$(MESH_PROVIDER)/uninstall.sh

platform-mesh-status:
	@echo "=== Mesh ($(MESH_PROVIDER)) Status ==="
	@bash platform/mesh/$(MESH_PROVIDER)/status.sh

.PHONY: platform-up platform-down platform-status \
        platform-ingress-up platform-ingress-down platform-ingress-status \
        platform-monitoring-up platform-monitoring-down platform-monitoring-status \
        platform-logging-up platform-logging-down platform-logging-status \
        platform-tracing-up platform-tracing-down platform-tracing-status \
        platform-mesh-up platform-mesh-down platform-mesh-status
