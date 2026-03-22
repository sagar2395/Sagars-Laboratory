# Task 028: Add Loki Log Retention Policy

## Priority
P2

## Assigned To
DevOps

## Description
`platform/logging/loki/values.yaml` deploys Loki in single-binary mode with a 5Gi PVC but specifies no log retention policy. Without `limits_config.retention_period` and `compactor.retention_enabled`, logs accumulate indefinitely until the PVC fills and Loki crashes. In a lab environment running long-term, this causes silent data loss (Loki fills disk and ingestion fails). A default retention of 7 days should be configured, with an environment-variable override for longer retention.

## Files to Modify
- `platform/logging/loki/values.yaml`
- `platform/logging/loki/install.sh`

## Implementation Notes
Add the following to `platform/logging/loki/values.yaml` under the `loki:` key:

```yaml
loki:
  limits_config:
    retention_period: 168h  # 7 days (overridable via LOKI_RETENTION_HOURS)
  compactor:
    retention_enabled: true
    delete_request_store: filesystem
  storage:
    type: filesystem
```

In `install.sh`, read a `LOKI_RETENTION_HOURS` environment variable (default `168`) and substitute the retention period value before installing:

```bash
RETENTION="${LOKI_RETENTION_HOURS:-168}h"
# use sed on a temp copy of values.yaml to substitute the retention value
```

Follow the same temp-file + sed pattern used by other platform scripts (e.g., Prometheus install.sh creates a temp values file).

Document `LOKI_RETENTION_HOURS` in `apps/app.env.example`.

## Acceptance Criteria
- [ ] After `platform/logging/loki/install.sh`, Loki's compactor has `retention_enabled: true`.
- [ ] Default retention is 7 days (168h).
- [ ] Setting `LOKI_RETENTION_HOURS=336` before running install doubles retention to 14 days.
- [ ] `kubectl logs -n monitoring -l app=loki` shows no startup errors related to compactor config.
- [ ] `apps/app.env.example` documents `LOKI_RETENTION_HOURS`.

## Testing Instructions
Run `platform/logging/loki/install.sh`. Run `kubectl exec -n monitoring <loki-pod> -- cat /etc/loki/config/config.yaml | grep retention` — confirm `retention_period` is `168h`. Send test logs and confirm they appear in Grafana → Loki datasource.

## Dependencies
None — independent of other tasks.
