# CLAUDE.md — Project Context for AI Sessions

> Read this file first in every session. It is the lean entry point.
> Anything deeper lives in `docs/` — only open those when the task needs them.
> Last updated: 2026-06-01

---

## What this project is

**Sagars-Laboratory** is a Kubernetes homelab toolkit for practicing Platform
Engineering / DevOps. A single Go binary (`labctl`) plus a web UI wraps a set of
shell scripts to:

- spin up a cluster (k3d locally; AKS/EKS in the cloud),
- install **swappable** platform components (ingress, monitoring, logging,
  tracing, GitOps, security, chaos),
- activate declarative **scenarios** (observability, GitOps, security, chaos),
- and tear it all down in minutes.

The core idea: **the CLI orchestrates, shell scripts do the work.** `labctl`
never reimplements `helm`/`kubectl` logic — it calls scripts and streams their
output. Keep that boundary.

## Current status (read before planning work)

The old `.claude/` docs claimed "all 6 phases complete." **That was aspirational.**
Reality as of this rewrite:

- ✅ Go CLI builds and all tests pass on macOS (arm64) and Linux.
- ✅ All major directories exist (apps, platform, scenarios, runtimes, terraform).
- ⚠️ ~28 real bugs / hardening tasks are open — see `.ai/tasks/` and `docs/ROADMAP.md`.
- ❌ Not yet verified end-to-end on macOS (tool bootstrap was WSL/Linux-only).
- ❌ Web UI is a single hand-written HTML file (being rebuilt as a proper SPA).
- ❌ Cloud runtimes (AKS/EKS) have never been run against real accounts.

**The plan to reach the desired state is `docs/ROADMAP.md`.** Work the phases in order.

## The golden rules (do not violate)

1. **Cross-platform.** This must run on macOS (Apple Silicon + Intel) and any
   modern Linux. No GNU-only flags in shell scripts (`grep -oP`, `sed -i` w/o
   backup, `readlink -f`, `date -d`). No hardcoded `linux/amd64`. Detect OS/arch.
2. **CLI wraps scripts.** Don't move script logic into Go. Add a script, register it.
3. **Config flows through the environment.** Scripts read `${VAR:-default}` from
   the executor's env. They must NOT source `.env`/`.active-runtime.env` themselves.
4. **No hardcoded domains.** Always use `${DOMAIN_SUFFIX:-k3d.local}`.
5. **Idempotent everything.** `helm upgrade --install`, `kubectl apply`, "skip if
   exists" cluster creation, scenario re-activation safe.
6. **Never commit build artifacts.** No binaries in git (`bin/` is gitignored),
   no runtime state (`.labctl/`).
7. **Every feature ships with a runbook.** If you add/fix a feature, add or update
   the matching file in `docs/runbooks/` so a human can verify it by hand.

## Repository map

```
labctl CLI ............ cmd/labctl/              Go: Cobra CLI + REST/WS API + embedded UI
web UI source ......... ui/                      Frontend (being rebuilt as SPA)
apps .................. apps/<name>/             Source + app.env + deploy/helm/
platform components ... platform/<category>/<provider>/   install.sh/uninstall.sh/status.sh/values.yaml
scenarios ............. scenarios/<name>/        scenario.yaml + manifests/values/dashboards
runtimes .............. runtimes/<profile>/      up.sh/down.sh/runtime.env  (k3d|aks|eks)
build/deploy engine ... engine/                  strategy dispatch (docker|acr|ecr; helm)
services .............. services/<name>/         shared deps (redis)
IaC ................... foundation/terraform/    modules/{aks,eks} + environments/{dev,staging}
make includes ......... make/                    modular *.mk
config ................ .env(.example), versions.env, apps/*/app.env, runtimes/*/runtime.env
tasks / backlog ....... .ai/tasks/, .ai/state.json
```

## Where to look for what (don't load everything)

| You need to… | Open |
|---|---|
| Understand the plan / what to build next | `docs/ROADMAP.md` + `.ai/state.json` |
| Verify a feature by hand / test manually | `docs/runbooks/` |
| Know a `labctl` command | `docs/cli-reference.md` |
| Deep architecture / patterns | `docs/architecture.md` (large — skim, don't load whole) |
| Scenario format & authoring | `docs/scenarios.md` |
| Cloud (AKS/EKS) setup | `docs/cloud-runtimes.md` |
| CI/CD workflows | `docs/ci-cd.md` |
| How tasks/agents work | `AGENTS.md` |

## Conventions cheat-sheet

- **Shell:** `#!/usr/bin/env bash`, `set -euo pipefail`, idempotent, portable (rule 1).
- **Platform provider** = a dir with `install.sh`, `uninstall.sh`, `status.sh`, `values.yaml`.
  Selected by env var (`INGRESS_PROVIDER`, `METRICS_PROVIDER`, …); registry routes to it.
- **App contract** = `apps/<name>/app.env` (`BUILD_STRATEGY`, `DEPLOY_STRATEGY`, `HELM_VALUES`).
- **Helm values** = `values-{dev,prod-like,cloud,test}.yaml`.
- **Commit prefixes:** `feat:`, `fix:`, `ci:`, `docs:`, `chore:`. Conventional commits.
- **Tasks** live in `.ai/tasks/NNN-kebab-title.md`; status in `.ai/state.json`.

## Build & test (do this to verify your changes)

```bash
# Build the CLI for the current OS/arch (no committed binary)
cd cmd/labctl && go build -o ../../bin/labctl . && cd ../..

# Or via make
make cli-build

# Run CLI tests
cd cmd/labctl && go test ./... && cd ../..

# App tests
cd apps/go-api && go test ./... ; cd ../..
```

`make help` lists all targets. `bin/labctl --help` lists all commands.

## Working agreement for AI sessions

- Pick the next unblocked task from `docs/ROADMAP.md` / `.ai/state.json`. One task at a time.
- Keep changes scoped to the task. Don't redesign the API surface unless the task says so.
- After implementing: run the relevant tests, then update the task's status in
  `.ai/state.json` and the matching runbook in `docs/runbooks/`.
- If you discover a new bug, add a task file rather than silently expanding scope.
- This is a tool-agnostic repo: Claude, Codex, Cursor, etc. all read `CLAUDE.md` +
  `AGENTS.md`. Keep instructions tool-neutral.
