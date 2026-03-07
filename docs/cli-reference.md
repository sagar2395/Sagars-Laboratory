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
