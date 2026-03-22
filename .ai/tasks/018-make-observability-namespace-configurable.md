# Task 018: Make Observability Namespace Configurable

## Priority
P1

## Assigned To
Feature

## Description
The `monitoring` namespace is hardcoded in at least three distinct places: `platform/registry.go` (Namespace() function), `scenario/engine.go` (Grafana dashboard ConfigMap creation), and `api/handlers.go` (dashboard URL construction for Loki and Tempo). If an operator deploys observability components into a different namespace (e.g., `observability`, `ops`), all of these silently use the wrong namespace, producing broken dashboard links, failed ConfigMap creation, and incorrect status checks.

## Files to Modify
- `cmd/labctl/internal/platform/registry.go`
- `cmd/labctl/internal/scenario/engine.go`
- `cmd/labctl/internal/api/handlers.go`
- `cmd/labctl/internal/config/config.go`

## Implementation Notes
1. Add a `MonitoringNamespace` field to the `Config` struct in `config.go`, defaulting to `"monitoring"` if not set in config or environment.

2. In `platform/registry.go`, change the hardcoded `"monitoring"` string in the `Namespace()` method to use `cfg.MonitoringNamespace` (pass config into Registry or add a method to override the namespace).

3. In `scenario/engine.go`, where `installGrafanaDashboard` hardcodes `"monitoring"`, use the engine's config namespace instead.

4. In `api/handlers.go`, replace hardcoded `"monitoring"` namespace references in dashboard URL construction with the config value.

5. Expose the setting as an environment variable `MONITORING_NAMESPACE` (already loaded from `app.env` via the existing env loading pattern in `config.go`).

Keep all defaults as `"monitoring"` so existing deployments are unaffected.

## Acceptance Criteria
- [ ] Setting `MONITORING_NAMESPACE=observability` in `app.env` causes all namespace references in platform, scenarios, and API to use `observability`.
- [ ] The default value is `"monitoring"`, preserving all existing behavior.
- [ ] `go test ./internal/...` passes with no changes to test fixtures.
- [ ] `app.env.example` documents the new `MONITORING_NAMESPACE` variable.

## Testing Instructions
Set `MONITORING_NAMESPACE=test-ns` in `app.env` and run `labctl platform status`. Confirm the namespace used in status output is `test-ns`. Run `labctl scenario info observability-sre` and confirm dashboard URLs reference `test-ns`.

## Dependencies
None
