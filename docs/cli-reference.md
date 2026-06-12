# CLI Reference

`labctl` is the command-line interface for managing the Sagars-Laboratory homelab.

## Installation

```bash
make cli-build        # builds bin/labctl
make cli-install      # builds + copies to PATH
```

## Global Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--project-dir` | string | auto-detected | Project root directory |
| `-v, --verbose` | bool | false | Verbose output |

## Commands

### Lifecycle

#### `labctl init`

Initialize the lab: setup tools, create cluster, install platform components.

```bash
labctl init
```

Equivalent to running `make setup-tools && make runtime-up && make platform-up`.

#### `labctl teardown`

Tear down the lab: destroy apps, remove platform, delete cluster.

```bash
labctl teardown
```

#### `labctl reset`

Full reset: teardown followed by init.

```bash
labctl reset
```

#### `labctl status`

Show overall lab status including cluster info, platform health, and deployed apps.

```bash
labctl status
```

---

### Runtime

Manage the underlying Kubernetes cluster.

#### `labctl runtime up`

Create the cluster using the configured runtime profile (k3d, aks, or eks).

```bash
labctl runtime up
```

#### `labctl runtime down`

Destroy the cluster.

```bash
labctl runtime down
```

#### `labctl runtime status`

Show cluster connectivity and node info.

```bash
labctl runtime status
```

---

### Applications

Manage application build and deployment lifecycle.

#### `labctl app list`

List all discovered applications with their build/deploy strategies.

```bash
labctl app list
```

#### `labctl app build <name>`

Build an application container image using its configured build strategy.

```bash
labctl app build go-api
labctl app build echo-server
```

#### `labctl app deploy <name>`

Deploy an application to the cluster using its configured deploy strategy.

```bash
labctl app deploy go-api
```

#### `labctl app destroy <name>`

Remove an application from the cluster.

```bash
labctl app destroy go-api
```

---

### Platform

Manage platform infrastructure components (ingress, monitoring, etc.).

#### `labctl platform up`

Install all platform components based on the configured providers.

```bash
labctl platform up
```

Installs components selected by `INGRESS_PROVIDER`, `METRICS_PROVIDER`, etc. in `.env`.

#### `labctl platform down`

Uninstall all platform components.

```bash
labctl platform down
```

#### `labctl platform status`

Show the status of all discovered platform components by category.

```bash
labctl platform status
```

---

### Scenarios

Manage declarative lab scenarios (observability, security, chaos, etc.).

#### `labctl scenario list`

List all available scenarios with their display names, categories, and activation status.

```bash
labctl scenario list
```

#### `labctl scenario info <name>`

Show detailed information about a scenario: description, prerequisites, components, and exploration hints.

```bash
labctl scenario info observability-sre
```

#### `labctl scenario up <name>`

Activate a scenario. Installs all declared components (Helm charts, manifests, dashboards).

```bash
labctl scenario up observability-sre
```

#### `labctl scenario down <name>`

Deactivate a scenario. Removes installed components.

```bash
labctl scenario down observability-sre
```

#### `labctl scenario status`

Show which scenarios are currently active.

```bash
labctl scenario status
```

#### `labctl scenario verify <name>`

Run the scenario's machine-verifiable `checks` (scenario format v2) and
report pass/fail per check. Exits non-zero if any check fails, so it is
safe in CI and scripts. See `docs/scenarios.md` for the checks schema.

```bash
labctl scenario verify observability-sre

# Re-run until everything passes (useful right after `scenario up`)
labctl scenario verify observability-sre --watch --interval 10s --timeout 5m
```

Flags:

| Flag | Default | Meaning |
|------|---------|---------|
| `--watch` | off | re-run checks until all pass or `--timeout` elapses |
| `--interval` | `10s` | delay between re-runs in watch mode |
| `--timeout` | `5m` | overall watch deadline |
| `--check-timeout` | `30s` | per-check timeout |

`promql` checks query Prometheus at `http://prometheus.<DOMAIN_SUFFIX>` by
default; override with the `PROMETHEUS_URL` environment variable.

The REST equivalent is `POST /api/scenarios/{name}/verify` (synchronous,
bounded to ~12s — use the CLI's `--watch` for long convergence).

---

#### `labctl scenario install <git-url>[@ref]`

Install a scenario pack from a git repository into `.labctl/catalog/`.
Every scenario in the pack is schema-validated before it becomes visible.
Flags: `--name` (default: repo basename), `--force` (replace an installed
pack). Companion commands: `labctl scenario packs` (list installed packs)
and `labctl scenario uninstall <pack-name>`. See "Scenario Packs" in
`docs/scenarios.md` — including the security note: packs run scripts on
your cluster, install only trusted sources.

---

### `labctl traffic` — Synthetic load generation

Run a k6 load generator in-cluster so scenarios, incidents, and autoscaling
play out under realistic traffic. The generator lives in its own `traffic`
namespace; scripts are in `services/traffic/`.

#### `labctl traffic start`

Start (or restart — running it again replaces the active run) the generator.

```bash
labctl traffic start                                   # steady 10 rps for 10m at go-api
labctl traffic start --profile spike --rps 20          # 20 rps baseline, 200 rps spike
labctl traffic start --profile soak --duration 4h
labctl traffic start --target http://echo-server.k3d.local/ --rps 50
```

| Flag | Default | Meaning |
|------|---------|---------|
| `--profile` | `steady` | `steady` (constant), `spike` (10x burst, fixed ~6m shape), `soak` (long sustained) |
| `--rps` | `10` | requests/sec (baseline for spike) |
| `--duration` | profile default | run length (`30s`, `10m`, `1h30m`); steady 10m, soak 2h, spike fixed |
| `--target` | go-api `/health` (in-cluster) | URL to load |

Set `K6_PROMETHEUS_RW_SERVER_URL` to also push k6's own metrics into
Prometheus via remote write (requires the receiver enabled); by default,
watch the load through the target app's request metrics in Grafana.

#### `labctl traffic stop`

Stop the generator and remove everything it created (job, configmap, namespace).

#### `labctl traffic status`

Show whether a run is active, its profile, pods, and recent k6 output.

#### `labctl traffic profiles`

List the available profiles (discovered from `services/traffic/profiles/`).

---

### `labctl incident` — Fault injection & game days

Inject realistic, reversible production faults from `incidents/` and
practice diagnosing them; the fault's detection check confirms when you've
actually fixed it. See `incidents/README.md` for the fault contract.

```bash
labctl incident list                          # browse the fault library
labctl incident inject oom-kill               # break the lab on purpose
labctl incident inject --random --silent      # game-day mode: surprise fault, name hidden
labctl incident inject --random --seed 42     # reproducible pick (same fault for the whole team)
labctl incident inject --random --category network
labctl incident status                        # runs the detection check; clears state when it passes
labctl incident resolve                       # escape hatch: undo the active fault
labctl incident resolve <name>                # works even if active state was lost
```

Rules: one active incident at a time (`--force` to override); injection is
gated on the fault's prerequisite apps being present; `resolve.sh` always
restores the lab regardless of partial manual fixes.

REST: `GET /api/incidents`, `POST /api/incidents/{name}/inject[?silent&force]`,
`POST /api/incidents/inject-random[?seed&category]`, `GET /api/incidents/status`,
`POST /api/incidents/resolve[?name=]`. Silent mode hides the fault's identity
in API responses until it is resolved.

---

### `labctl lab` — Snapshot, restore, reset

Lab-level state operations for fast iteration. A snapshot records **intent**
(which platform components, apps, and scenarios are active) as a small YAML
file in `.labctl/snapshots/` — not cluster bytes. Restore replays the normal
idempotent install paths; reset tears everything down to post-init.

```bash
labctl lab snapshot before-gameday    # record current state
labctl lab snapshots                  # list saved snapshots
labctl lab reset                      # interactive teardown (keeps cluster + ingress)
labctl lab reset --yes                # non-interactive
labctl lab restore before-gameday     # converge back: platform → apps → scenarios
labctl lab delete before-gameday
```

Details:

- **Snapshot sources** — platform components from labctl's install markers
  (`.labctl/platform/`, written by `platform up`/`down`; installs done
  outside labctl are not tracked), scenarios from the scenario engine's
  state, apps by live kubectl probe.
- **Restore order** — ingress first, then monitoring, then remaining
  platform components, then apps, then scenarios. Already-active pieces are
  skipped (idempotent), so restoring over a half-converged lab is safe.
- **Reset** — stops traffic, deactivates all scenarios, destroys deployed
  apps, uninstalls platform components **except the ingress category**, and
  keeps going past individual failures, reporting what stuck at the end.
  The cluster itself stays up.

REST equivalents: `GET/POST/DELETE /api/lab/snapshots[/{name}]`,
`POST /api/lab/snapshots/{name}/restore` (async, returns a job id),
`POST /api/lab/reset?confirm=true` (async; refuses without `confirm`).

---

### Services

Manage shared services (Redis, etc.) that apps depend on.

#### `labctl service list`

List all available shared services.

```bash
labctl service list
```

#### `labctl service up <name>`

Install a shared service.

```bash
labctl service up redis
```

#### `labctl service down <name>`

Uninstall a shared service.

```bash
labctl service down redis
```

#### `labctl service status [name]`

Show service status. If no name is given, shows all services.

```bash
labctl service status
labctl service status redis
```

---

### Checks

Run validation checks against the environment.

#### `labctl check tools`

Verify that all required CLI tools (kubectl, helm, docker, etc.) are installed and accessible.

```bash
labctl check tools
```

#### `labctl check cluster`

Check cluster connectivity via `kubectl cluster-info`.

```bash
labctl check cluster
```

#### `labctl check ingress`

Check that the ingress controller is running and responding.

```bash
labctl check ingress
```

---

### Web UI

#### `labctl ui`

Launch the web UI dashboard. Opens a browser automatically.

```bash
labctl ui                # default port 3939
labctl ui --port 8080    # custom port
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--port` | string | 3939 | Port to serve the UI on |

The dashboard shows:
- Cluster status and connection info
- Platform component health
- Applications with deploy/destroy actions
- Scenarios with activate/deactivate controls
- Real-time updates via WebSocket

## Comparison: CLI vs Make

Both interfaces work. Use whichever you prefer:

| Operation | CLI | Make |
|-----------|-----|------|
| Full setup | `labctl init` | `make init` |
| Build app | `labctl app build go-api` | `make build APP_NAME=go-api` |
| Deploy app | `labctl app deploy go-api` | `make deploy APP_NAME=go-api` |
| Platform status | `labctl platform status` | `make platform-status` |
| Activate scenario | `labctl scenario up observability-sre` | N/A (CLI only) |
| Web dashboard | `labctl ui` | N/A (CLI only) |
| Deploy all apps | N/A | `make deploy-all` |
| Terraform | N/A | `make terraform-apply TF_ENV=dev` |

The CLI adds scenario management, the web UI, and a unified status view. Make targets are more granular and support Terraform operations directly.
