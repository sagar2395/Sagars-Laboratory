# Runbooks — Use & Verify Each Feature by Hand

These are step-by-step manual guides. Each one shows **how to use a feature** and
**how to confirm it works**, so you can learn the system by doing and so AI
sessions have an objective acceptance check beyond automated tests.

Each runbook maps to a phase in `docs/ROADMAP.md`.

| Runbook | Covers | Phase |
|---------|--------|-------|
| [00-cross-platform-setup.md](00-cross-platform-setup.md) | Install tools + cluster on macOS / Linux | 0 |
| [01-local-cluster-and-apps.md](01-local-cluster-and-apps.md) | init → deploy app → reach it → status | 1 |
| [02-platform-and-status.md](02-platform-and-status.md) | Install/swap platform components, accurate status | 2 |
| [03-web-ui.md](03-web-ui.md) | Build, run, and drive the web UI | 3 |
| [04-observability-scenario.md](04-observability-scenario.md) | Logs in Loki, traces in Tempo, dashboards | 4 |
| [05-ci-cd.md](05-ci-cd.md) | CI checks + CD + ArgoCD sync | 5 |
| [06-cloud-runtimes.md](06-cloud-runtimes.md) | Provision & tear down AKS/EKS | 6 |

## How to read a runbook

- **Prereqs** — what must already be true.
- **Steps** — copy-paste commands, in order.
- **Expected** — what success looks like (the assertion you verify).
- **Cleanup** — how to undo it.
- **Troubleshooting** — common failures and fixes.

## Conventions

- Commands assume you are in the repo root unless noted.
- `bin/labctl` is built locally (`make cli-build`) — there is no committed binary.
- Replace `k3d.local` with your `DOMAIN_SUFFIX` if you changed it in `.env`.
- Ingress hostnames need `/etc/hosts` entries — see runbook 00 (or `labctl hosts add`).
