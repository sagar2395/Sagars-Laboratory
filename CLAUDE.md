# CLAUDE.md — Project Context for AI Sessions

> Read this first in every session. It is the lean entry point.
> Anything deeper lives in `docs/` — open those only when the task needs them.
> Last updated: 2026-07-26

---

## What this project is

**Flightdeck** is a Kubernetes **platform-engineering simulator**: it stands up a
realistic production-shaped cluster on your laptop in minutes, breaks it in
realistic ways, and grades you on how you fix it. Built for engineers learning
platform skills, teams evaluating stacks, SREs running game days, and leads
assessing incident-response ability.

A single Go binary (`labctl`) plus an embedded web UI wraps shell scripts and
declarative YAML to spin up a cluster, install swappable platform components,
activate scenarios with verifiable `checks`, inject incidents, run learning
paths and graded challenges, and tear it all down.

**The core boundary: Go orchestrates, records and grades. Shell scripts and YAML
do the domain work.** `labctl` never reimplements `helm`/`kubectl` logic.

## Current status — read before planning work

The project is in a **v2 redesign**. v1 shipped a broad feature set (79 tasks,
9 milestones) but with a weak core: no cancellation, no durable state, zero
shell tests, zero UI tests, and cloud runtimes that were never verified.

v2 cuts scope hard and rebuilds the core to production standard.

- **Plan of record: `docs/ROADMAP.md`** — nine waves, W0 through W8.
- **Live state: `.ai/state.json`** — its `next` field names the task to pick up.
- **Product rationale: `docs/PRODUCT.md`. Target design:
  `docs/architecture/ARCHITECTURE.md`. Decisions: `docs/adr/`.**
- **Test policy: `docs/TESTING.md`** — four mandatory layers.
- v1 plan, runbooks, strategy docs and task files: `docs/archive/`, `.ai/archive/`.

**Cut in v2:** cloud runtimes (AKS/EKS/GKE) and Terraform; the pack marketplace,
entitlement tiers, editions and certificate framework. See ADR-0001.

**Currently:** the v2 plan is written; implementation starts at wave W0.

## The golden rules (do not violate)

1. **Cross-platform.** macOS (Apple Silicon + Intel) and modern Linux. No
   GNU-only flags (`grep -oP`, `sed -i` without a backup suffix, `readlink -f`,
   `date -d`). No hardcoded `linux/amd64`. **No cgo** — it breaks
   cross-compilation.
2. **Go orchestrates, scripts do the work.** Don't move script logic into Go.
   Add a script, register it.
3. **Config flows through the environment.** Scripts read `${VAR:-default}` from
   the executor's env. They must NOT source `.env` themselves.
4. **No hardcoded domains.** Always `${DOMAIN_SUFFIX:-k3d.local}`.
5. **Idempotent everything.** `helm upgrade --install`, `kubectl apply`,
   skip-if-exists cluster creation, safe re-activation. Interrupting an
   operation and re-running it must converge.
6. **Never commit build artifacts.** No binaries, no runtime state.
7. **Every operation is cancellable and durable.** Anything that shells out goes
   through `internal/run` with a context, a timeout, a lock key and persisted
   logs. Never call `exec.Command(...).Run()` directly.
8. **Simulator content is declarative.** Scenarios, faults, paths, challenges and
   checks are YAML + scripts. Never bake their logic into Go.
9. **Docs, runbook and state are part of the change — mandatory.** Every change
   updates the relevant doc, adds or updates a runbook, writes an ADR if a
   decision was made, and updates `.ai/state.json`. An undocumented change is an
   unfinished change.
10. **Every feature ships with tests at every applicable layer — no exceptions.**
    Go unit (table-driven, hermetic, ≥80%, cancellation covered), bats for
    scripts with logic, contract tests for endpoints and CLI commands, Vitest +
    Playwright for UI. Full policy: `docs/TESTING.md`.

## Repository map (v2 target)

```
go.mod ................ repo root — module go.flightdeck.dev/flightdeck
cmd/labctl/ ........... entrypoint only
internal/cli/ ......... cobra commands — thin adapters
internal/httpapi/ ..... REST/WS/SSE — thin adapters
internal/service/ ..... use-cases and invariants (the only place logic lives)
internal/run/ ......... durable run engine: queue, locks, cancel, logs
internal/store/ ....... SQLite persistence + migrations
internal/catalog/ ..... declarative content loading + validation
internal/toolchain/ ... bash/kubectl/helm/k3d adapters + fakes
pkg/checks/ ........... public: the check engine
pkg/scenario/ ......... public: content types + schema
pkg/extension/ ........ public: third-party seam
ui/ ................... React SPA, embedded into the binary
platform/<cat>/<prov>/  install.sh/uninstall.sh/status.sh/values.yaml
scenarios/ incidents/ learn/ challenges/    declarative content
runtimes/<profile>/ ... k3d | kind | incluster
apps/<name>/ .......... sample workloads (own Go module)
docs/ ................. PRODUCT, ROADMAP, TESTING, architecture/, adr/, runbooks/
.ai/state.json ........ wave + task state
```

## Where to look for what (don't load everything)

| You need to… | Open |
|---|---|
| Know what to build next | `.ai/state.json` → `docs/ROADMAP.md` |
| Understand why a feature exists | `docs/PRODUCT.md` |
| Understand how the system fits together | `docs/architecture/ARCHITECTURE.md` |
| Know why something was built this way | `docs/adr/` |
| Know the test bar | `docs/TESTING.md` |
| Verify a feature by hand | `docs/runbooks/` |
| Know a `labctl` command | `docs/cli-reference.md` |
| Scenario format & authoring | `docs/scenarios.md`, `docs/authoring/` |

## Conventions cheat-sheet

- **Shell:** `#!/usr/bin/env bash`, `set -euo pipefail`, idempotent, portable,
  handles `SIGTERM`, emits `##flightdeck:step:<name>` markers for progress.
- **Platform provider** = a dir with `install.sh`, `uninstall.sh`, `status.sh`,
  `values.yaml`, selected by env var (`INGRESS_PROVIDER`, …).
- **Go:** every exported function that can block takes `context.Context` first.
  Errors wrap with `%w`. No `panic` outside `main`.
- **Commit prefixes:** `feat:`, `fix:`, `test:`, `ci:`, `docs:`, `chore:`,
  scoped by wave/task — `feat(W1/T06): run queue with exclusive lock keys`.

## Build & test

```bash
go build ./...          # from the repo root
make test               # layers 1-3, race detector, coverage gate
make test-shell         # bats
make test-ui            # vitest
make test-e2e           # playwright
make lint               # every static analysis gate
make test-coverage      # HTML report
```

## Working agreement for AI sessions

- Work is delivered in **waves**. Pick the task named by `next` in
  `.ai/state.json`. One task at a time; finish it completely before starting
  another.
- **A task is not done until the Definition of Done in `docs/ROADMAP.md` is
  satisfied** — code, tests at every applicable layer, docs, runbook, state.
- Keep changes scoped to the task. If you discover a new problem, add it to the
  wave's task list rather than silently expanding scope.
- Do not start a new wave until the previous wave's exit criteria hold and its
  runbooks have been signed off by the maintainer.
- This repo is tool-agnostic: Claude, Codex, Cursor all read `CLAUDE.md` +
  `AGENTS.md`. Keep instructions tool-neutral.
