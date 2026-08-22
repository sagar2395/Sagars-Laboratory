# Task 001: Fix Platform Category Resolution and Status

## Priority
P0

## Assigned To
Feature

## Description
Align the `labctl` platform lifecycle with the actual platform registry category model. The current CLI and API attempt to install metrics providers from `monitoring`, but provider discovery registers Prometheus under `monitoring/metrics`. As a result, platform installs and status checks are inconsistent with the repository layout and with the documented provider model.

## Files to Modify
- `cmd/labctl/cmd/platform.go`
- `cmd/labctl/cmd/status.go`
- `cmd/labctl/internal/api/handlers.go`
- `cmd/labctl/internal/platform/registry.go`
- `cmd/labctl/internal/platform/registry_test.go`

## Implementation Notes
Update platform orchestration to use the correct category keys for discovered providers, especially `monitoring/metrics`. Keep provider selection driven by config rather than hard-coded assumptions. Replace namespace-name heuristics in status responses with provider-aware checks so Prometheus, Grafana, ingress, logging, and tracing report consistently. Preserve existing output shape unless a small additive change is necessary for correctness.

## Acceptance Criteria
- [ ] `labctl platform up` installs the configured ingress provider, metrics provider, and Grafana using category keys that match registry discovery.
- [ ] `labctl platform down` and `labctl platform status` operate on the same category model used by discovery.
- [ ] API status endpoints no longer report provider health by assuming the namespace matches the provider name.
- [ ] Unit tests cover nested category discovery and lifecycle lookups for `monitoring/metrics`.

## Testing Instructions
Run `go test ./internal/platform/... ./internal/api/... ./cmd/...` from `cmd/labctl`. Manually verify that `labctl platform status` and `GET /api/platform` behave sensibly for a config that uses Prometheus and Grafana.

## Dependencies
None
