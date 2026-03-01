platform-up:
	bash platform/ingress/traefik/install.sh

platform-down:
	@bash platform/ingress/traefik/uninstall.sh

platform-status:
	@bash platform/ingress/traefik/status.sh