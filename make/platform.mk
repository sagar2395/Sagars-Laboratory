platform-up: platform-ingress-up platform-monitoring-up

platform-down: platform-monitoring-down platform-ingress-down

platform-status: platform-ingress-status platform-monitoring-status

# Traefik Ingress
platform-ingress-up:
	bash platform/ingress/traefik/install.sh

platform-ingress-down:
	@bash platform/ingress/traefik/uninstall.sh

platform-ingress-status:
	@bash platform/ingress/traefik/status.sh

# Monitoring Stack (Prometheus + Grafana)
platform-monitoring-up:
	@echo "Installing monitoring stack..."
	bash platform/monitoring/prometheus/install.sh
	bash platform/monitoring/grafana/install.sh
	@echo "✓ Monitoring stack installed"

platform-monitoring-down:
	@echo "Uninstalling monitoring stack..."
	bash platform/monitoring/grafana/uninstall.sh
	bash platform/monitoring/prometheus/uninstall.sh
	@echo "✓ Monitoring stack uninstalled"

platform-monitoring-status:
	@echo "=== Prometheus Status ==="
	@bash platform/monitoring/prometheus/status.sh
	@echo ""
	@echo "=== Grafana Status ==="
	@bash platform/monitoring/grafana/status.sh