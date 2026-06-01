# Runbook 04 — Observability Scenario

> Goal: activate the observability scenario and see real logs, traces, and
> dashboards for the apps. Verifies Phase 4.

## Prereqs

- Runbook 00 (cluster up) + Runbook 01 (go-api deployed).
- Platform monitoring installed (`make platform-up`).

## Steps

```bash
# 1. Inspect the scenario before running it
bin/labctl scenario info observability-sre
bin/labctl scenario list

# 2. Activate it (Loki + Promtail + Tempo + alerting + SLO dashboards)
bin/labctl scenario up observability-sre

# 3. Generate traffic so there is something to observe
for i in $(seq 1 200); do curl -s http://go-api.k3d.local/health >/dev/null; done

# 4. Explore in Grafana (admin/admin)
open http://grafana.k3d.local        # macOS;  xdg-open on Linux
#   Explore -> Loki  -> {app="go-api"}                 (logs)
#   Explore -> Tempo -> search recent traces           (traces)
#   Dashboards -> SLO / Log Explorer

# 5. Check status + version of the running app
curl http://go-api.k3d.local/version  # Task 013

# 6. Tear the scenario down
bin/labctl scenario down observability-sre
```

## Expected

- Scenario activation streams progress and ends marked **active** (no race — the
  status reflects reality immediately after the command returns).
- Loki returns log lines for `{app="go-api"}`; Tempo shows traces.
- Re-running `scenario up` is a no-op (Task 019, idempotency).
- The observability namespace is configurable (Task 018) — not hardcoded.
- Loki enforces a retention policy (Task 028).
- `/version` reflects the built image (Task 013); `--verbose` raises log detail
  (Task 014).

## Cleanup

```bash
bin/labctl scenario down observability-sre
```

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| No logs in Loki | Promtail not scraping; check `kubectl get pods -n <obs-ns>`. |
| Scenario shows inactive right after `up` | Success-broadcast race (already fixed) — re-check `scenario status`. |
| Grafana has no Loki/Tempo datasource | Datasources are set in Grafana values; re-run `platform-up`. |
| Empty traces | App must emit OTel spans; generate traffic and confirm the collector endpoint. |
