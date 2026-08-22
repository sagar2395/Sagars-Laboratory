# Runbook 00 — Cross-Platform Setup (macOS / Linux)

> Goal: from a clean machine to a healthy local k3d cluster, on macOS (Apple
> Silicon or Intel) or any modern Linux. Verifies Phase 0.

## Prereqs

- Git, Make.
- A container runtime:
  - **macOS:** Docker Desktop, **or** Colima (`brew install colima && colima start`),
    **or** OrbStack.
  - **Linux:** Docker Engine running (`sudo systemctl start docker`), your user in
    the `docker` group.
- Go 1.24+ (to build `labctl`).

## Steps

```bash
# 1. Configure
cp .env.example .env        # edit if you want a different cluster name / domain

# 2. Install pinned tools for the k3d profile (OS/arch auto-detected)
make setup-tools PROFILE=k3d

# 3. Build the CLI for your machine (no committed binary)
make cli-build
bin/labctl --help

# 4. Bring up the cluster + platform
make init                   # = setup-tools + runtime-up + platform-up

# 5. Map ingress hostnames (pick one)
bin/labctl hosts add        # preferred (Task 032); will sudo if needed
#   or manually add to /etc/hosts:
#   127.0.0.1 go-api.k3d.local grafana.k3d.local prometheus.k3d.local argocd.k3d.local
```

## Expected

- `make setup-tools` installs the **correct** binaries for your OS/arch:
  ```bash
  kubectl version --client     # matches versions.env
  k3d version                  # matches versions.env
  helm version                 # matches versions.env
  file bin/labctl              # reports YOUR architecture (arm64 on Apple Silicon)
  ```
- `bin/labctl status` shows a running cluster with the expected node count and
  Kubernetes version.
- `curl http://grafana.k3d.local` resolves (after `hosts add`).

## Cleanup

```bash
make teardown                # remove apps + platform + cluster
bin/labctl hosts remove      # remove /etc/hosts entries
```

## Cross-compilation (for releases)

```bash
make cli-build-all    # builds dist/labctl-{darwin,linux}-{arm64,amd64}
```

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| `Cannot connect to the Docker daemon` on macOS | Start Docker Desktop / `colima start` / OrbStack. |
| `bin/labctl: cannot execute binary file` | You're running a stale/foreign binary — `make cli-build` again. |
| Ingress host won't resolve | `/etc/hosts` entry missing — `bin/labctl hosts add`. |
| `labctl hosts add` says "permission denied" | Re-exec with sudo is automatic; approve the sudo prompt. |

## Notes for AI sessions

Phase 0 tasks 030–034 are complete. This runbook is the acceptance check — if
any step requires a manual edit to work on macOS, investigate and reopen the
relevant task.
