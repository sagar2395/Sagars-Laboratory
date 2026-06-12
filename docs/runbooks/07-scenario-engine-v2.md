# Runbook 07 — Scenario Engine v2: Checks, Verify, Traffic & Snapshot/Reset

Covers milestone **M1** (tasks 040–043). The scenario catalog (task 044)
will extend this runbook when it ships.

## Prereqs

- A healthy local lab: `make init` (runbook 01) with ingress + monitoring up
  (runbook 02) and go-api deployed.
- CLI built: `make cli-build`.

## Steps

### 1. Inspect a v2 scenario

```bash
bin/labctl scenario info observability-sre
```

**Expected:** output now includes an `Objectives:` list, a `Stages (3):`
tree (log-aggregation, tracing, alerting-and-slo) with components nested
under each stage, and a `Checks (5)` list naming
`loki-ready, promtail-ready, tempo-ready, alerting-rules-installed,
grafana-reachable`.

### 2. Activate and verify

```bash
bin/labctl scenario up observability-sre
bin/labctl scenario verify observability-sre --watch --timeout 5m
```

**Expected:** `up` prints the objectives, installs components grouped by
stage (`=== Stage: log-aggregation ===` …), and ends with "This scenario
has 5 verifiable checks". `verify --watch` re-runs until every line reads
`PASS` and exits 0 with `All 5 checks passed.`

### 3. Prove verify detects breakage

```bash
kubectl -n monitoring delete statefulset loki
bin/labctl scenario verify observability-sre; echo "exit=$?"
```

**Expected:** the `loki-ready` check reports `FAIL`/error, the summary says
`1 of 5 checks failed`, and the exit code is non-zero.

Restore and re-verify:

```bash
bin/labctl scenario down observability-sre
bin/labctl scenario up observability-sre
bin/labctl scenario verify observability-sre --watch
```

### 4. Verify over the REST API

```bash
bin/labctl ui &   # serves the API on :3939
curl -s -X POST localhost:3939/api/scenarios/observability-sre/verify | jq .
```

**Expected:** JSON with `"passed": true` and a `results` array of 5 entries,
each with `name`, `type`, `pass`, `got`, `want`, `durationMs`.

### 5. Generate traffic and watch the impact

```bash
bin/labctl traffic profiles
bin/labctl traffic start --profile steady --rps 20 --duration 10m
bin/labctl traffic status
```

**Expected:** `start` creates the `traffic` namespace and a `traffic-k6`
job; `status` shows the job running and recent k6 output. In Grafana, the
go-api request-rate panel climbs to ~20 rps within a minute.

Now make it interesting — spike it and watch the dashboards:

```bash
bin/labctl traffic start --profile spike --rps 10
```

**Expected:** restarting while running replaces the run (no error). The
rate holds ~10 rps for 1 minute, ramps to ~100 rps for 2 minutes, then
recovers. Latency panels react during the spike.

Stop and confirm cleanup:

```bash
bin/labctl traffic stop
kubectl get namespace traffic   # → NotFound
bin/labctl traffic status        # → "Traffic generator: not running"
```

### 6. Snapshot, reset, restore — the fast-iteration loop

With the observability scenario active and go-api deployed:

```bash
bin/labctl lab snapshot before-reset
bin/labctl lab snapshots
```

**Expected:** the snapshot lists your installed platform components
(`ingress/...`, `monitoring/...`), `go-api` under apps, and
`observability-sre` under scenarios. Note: platform components only appear
if they were installed through labctl after task 043 (the markers live in
`.labctl/platform/`).

Reset the lab:

```bash
bin/labctl lab reset --yes
```

**Expected:** steps print in order — traffic-stop, scenario-down,
app-destroy, platform-uninstall (ingress is skipped) — and finish in
< 5 minutes on k3d. `labctl scenario status` shows nothing active;
`kubectl get ns` no longer shows the scenario/app namespaces.

Restore and prove convergence:

```bash
bin/labctl lab restore before-reset
bin/labctl scenario verify observability-sre --watch --timeout 5m
```

**Expected:** restore replays platform → apps → scenarios, and verify goes
green — the snapshot's checks are the proof that restore actually converged.

### 7. Schema validation gate (what CI runs)

```bash
cd cmd/labctl && go test ./internal/... && cd ../..
```

**Expected:** all tests pass. Break a scenario on purpose (e.g. change a
check's `type:` to `gopher` in any `scenarios/*/scenario.yaml`) and re-run:
the repo test must fail naming the file and field. Revert afterwards.

## Cleanup

```bash
bin/labctl traffic stop
bin/labctl scenario down observability-sre
kill %1 2>/dev/null  # stop labctl ui if started
```

## Troubleshooting

- **`promql` checks error with connection refused** — Prometheus isn't
  reachable at `http://prometheus.<DOMAIN_SUFFIX>`. Export
  `PROMETHEUS_URL` (e.g. a `kubectl port-forward` address) and re-run.
- **`grafana-reachable` fails with DNS errors** — your `/etc/hosts` lacks
  the ingress hostnames; see runbook 00 or `labctl hosts add`.
- **Scenario refuses to load after editing** — run
  `bin/labctl scenario list`; if it's missing, the schema validator
  rejected it. The CLI test `TestRepoScenarios_AllLoadAndValidate` prints
  the exact reason.
