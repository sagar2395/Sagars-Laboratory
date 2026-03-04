#!/bin/bash

set -e

NAMESPACE="monitoring"

echo "Uninstalling Grafana..."

# Uninstall Grafana
helm uninstall grafana -n $NAMESPACE >/dev/null 2>&1 || true

echo "✓ Grafana uninstalled successfully"
