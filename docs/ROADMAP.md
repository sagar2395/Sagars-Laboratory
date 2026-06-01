# ROADMAP — Path to the Desired State

> The single plan of record. Work phases top-to-bottom; within a phase, work by
> task id. Status of each task lives in `.ai/state.json`; details in `.ai/tasks/`.
> Last updated: 2026-06-01

---

## Desired end state

A platform-engineering homelab that a single person can run on **a MacBook (Apple
Silicon or Intel) or any modern Linux box**, with the *option* to push the same
workloads to **AKS/EKS** in the cloud:

- `make init` (or `labctl init`) brings up a local k3d cluster + platform stack on
  a clean machine, regardless of OS.
- `labctl` and the **web UI** drive everything: deploy apps, install/swap platform
  components, activate scenarios, watch live status.
- Four scenarios (observability, GitOps, security, chaos) activate and explore
  cleanly, each with a documented manual runbook.
- AKS/EKS provisioning via Terraform is verified against real accounts at least once.
- CI/CD validates lint, tests, image build, Helm, and shell portability on every PR.

## How the plan is organized

Each phase has a **goal**, a **task list** (ids map to `.ai/tasks/NNN-*.md`), an
**exit criterion** (objective, testable), and a **runbook** to verify it by hand.
Phases 1–2 are sequential (stability first). Phases 3–6 only depend on Phase 2 and
can be parallelized across worktrees/AI tools.

```
Phase 0  Cross-Platform Foundation   ─┐ (must be first)
Phase 1  Core Stability (P0)          │ sequential
Phase 2  Reliability & Correctness   ─┘
                │
   ┌────────────┼────────────┬───────────────┐
Phase 3      Phase 4       Phase 5         Phase 6
Web UI       Apps &        CI/CD &         Cloud
(SPA)        Observability Supply Chain    (AKS/EKS)
   └────────────┴────────────┴───────────────┘
                │
            Phase 7  Polish & Release
```

---

## Phase 0 — Cross-Platform Foundation `[macOS + Linux]`

**Goal:** the toolkit installs and runs identically on macOS (arm64/amd64) and
Linux. This is the gate for everything else — the project was authored on WSL and
carries Linux-only assumptions.

**Tasks**

| Id | Title |
|----|-------|
| 030 | Make `bootstrap/setup-tools.sh` OS/arch-aware (detect `uname -s/-m`, pick correct download URLs, replace `grep -oP` with portable parsing) |
| 031 | Remove committed `bin/labctl`; make `make cli-build` build for host OS/arch and document cross-compilation |
| 032 | Add `/etc/hosts` helper for `*.k3d.local` (a `labctl hosts add/remove` command or documented `make hosts`) |
| 033 | Support macOS container runtimes (Docker Desktop / Colima / OrbStack detection) in k3d up + preflight `check` |
| 034 | Portability sweep: audit all `*.sh` for GNU-only constructs; add a CI lint that fails on them (folds into Phase 5 CI) |

**Exit criterion:** on a clean macOS machine and a clean Linux machine,
`make setup-tools PROFILE=k3d` then `make init` produces a healthy cluster, and
`bin/labctl status` reports it. No manual edits required.

**Runbook:** `docs/runbooks/00-cross-platform-setup.md`

---

## Phase 1 — Core Stability (P0) `[local k3d]`

**Goal:** the everyday local loop — init, deploy an app, hit it, see status — is
correct and safe. These are the P0 bugs.

**Tasks**

| Id | Title |
|----|-------|
| 002 | Add scenario preflight validation |
| 007 | Fix echo-server readiness probe |
| 008 | Add request timeout + input validation to API handlers (security) |
| 015 | Fix Kubernetes version parsing in k8s client |
| 016 | Add echo-server request body size limit |

(001 — platform category resolution — already done.)

**Exit criterion:** `labctl init && labctl app deploy go-api && curl …/health`
succeeds; `labctl status` shows correct k8s version; API rejects malformed input
with 400 and times out long subprocesses with 504. All Phase-1 tests green.

**Runbook:** `docs/runbooks/01-local-cluster-and-apps.md`

---

## Phase 2 — Reliability & Correctness (P1) `[local k3d]`

**Goal:** remove the rough edges that make the tool feel flaky — status parsing,
idempotency, provider-awareness, structured errors.

**Tasks**

| Id | Title |
|----|-------|
| 003 | Stream action events for all API operations |
| 004 | Improve runtime detection and status reporting |
| 005 | Make platform make-targets provider-aware |
| 009 | Add Helm pre-deploy lint + post-deploy readiness wait |
| 010 | Add `uninstall.sh` for kubernetes-dashboard |
| 012 | Fix engine strategy script existence validation |
| 017 | Return job id from async API actions |
| 019 | Add scenario `up` idempotency check |
| 020 | Fix kubectl pod output parsing (use JSON) |
| 022 | Add structured error responses to API handlers |
| 025 | Fix ArgoCD values hardcoded `k3d.local` domain |
| 029 | Validate PROFILE and APP_NAME before use |

**Exit criterion:** every `labctl … status` command returns accurate state from
JSON (not text scraping); re-running `scenario up` / `platform up` is a no-op;
API actions return a job id and structured errors. Phase-2 tests green.

**Runbook:** `docs/runbooks/02-platform-and-status.md`

---

## Phase 3 — Web UI Rebuild (SPA) `[depends on Phase 2]`

**Goal:** replace the single hand-written `ui/dist/index.html` with a real
single-page app, embedded into the binary via `go:embed`, talking to the existing
REST + WebSocket API.

**Tasks**

| Id | Title |
|----|-------|
| 035 | Scaffold SPA (Vite + React or Svelte + TypeScript) in `ui/`, wired to the API; dev proxy to `labctl ui` |
| 036 | Build the four views: Dashboard (cluster/platform/apps), Scenarios, Platform (swap providers), Apps (deploy/logs); live updates over WebSocket |
| 037 | Production build → `ui/dist`, copied & embedded by `make cli-build`; server serves embedded assets with filesystem fallback for dev |

**Exit criterion:** `make cli-build && bin/labctl ui` serves the SPA at
`localhost:3939`; all four views render real data; activating a scenario from the
UI streams progress live; the single binary contains the UI (no external files).

**Runbook:** `docs/runbooks/03-web-ui.md`

---

## Phase 4 — Apps & Observability `[depends on Phase 2]`

**Goal:** make the apps and the observability scenario genuinely demonstrable.

**Tasks**

| Id | Title |
|----|-------|
| 013 | Add `/version` endpoint to go-api and echo-server |
| 014 | Wire `--verbose` flag to the structured logger |
| 018 | Make observability namespace configurable |
| 021 | Add Redis cache TTL configuration to echo-server |
| 028 | Add Loki log retention policy |
| 024 | Add Kyverno ClusterPolicy manifests (shared by security scenario) |

**Exit criterion:** `labctl scenario up observability-sre` produces queryable logs
(Loki) and traces (Tempo) in Grafana for both apps; `/version` reflects the built
image; `--verbose` changes log level. Phase-4 tests green.

**Runbook:** `docs/runbooks/04-observability-scenario.md`

---

## Phase 5 — CI/CD & Supply Chain `[depends on Phase 2]`

**Goal:** PRs are validated automatically and images are scanned.

**Tasks**

| Id | Title |
|----|-------|
| 006 | Expand CI validation for infra + shell assets (incl. portability lint from Phase 0) |
| 011 | Add container image security scanning to CI |
| 023 | Enable ArgoCD sync step in CD pipeline |

**Exit criterion:** CI runs lint + unit tests + Helm lint + shell portability +
image scan on PRs; CD builds/pushes images and (in GitOps scenario) ArgoCD syncs.

**Runbook:** `docs/runbooks/05-ci-cd.md`

---

## Phase 6 — Cloud Runtimes (AKS/EKS) `[depends on Phase 2]`

**Goal:** make cloud deployment real, not theoretical. Requires cloud accounts and
incurs spend — do this deliberately.

**Tasks**

| Id | Title |
|----|-------|
| 027 | Configure Terraform remote state backend |
| 026 | Fix Chaos-Mesh hardcoded containerd socket path (make runtime-aware for cloud) |
| 038 | Provision AKS dev env via Terraform; run go-api + one scenario; tear down |
| 039 | Provision EKS dev env via Terraform; run go-api + one scenario; tear down |

**Exit criterion:** `labctl runtime up --profile aks` (and `eks`) provisions a
cluster, an app deploys with the cloud Helm profile, at least one scenario runs,
and `runtime down` removes all billable resources. Verified once, cost recorded.

**Runbook:** `docs/runbooks/06-cloud-runtimes.md`

---

## Phase 7 — Polish & Release

**Goal:** the repo is something a stranger can clone and succeed with.

- Refresh `README.md` to match reality (quickstart that actually works on macOS).
- Ensure every runbook in `docs/runbooks/` is complete and accurate.
- Trim `docs/architecture.md` to the essentials; move long history out.
- Tag a `v0.1.0` release with cross-platform build instructions.
- Final portability pass; final `go test ./...` + `make help` sanity.

**Exit criterion:** clone → follow README → working lab on a fresh Mac, with no
tribal knowledge required.

---

## Task ↔ phase index

| Phase | Task ids |
|-------|----------|
| 0 | 030, 031, 032, 033, 034 |
| 1 | 002, 007, 008, 015, 016 |
| 2 | 003, 004, 005, 009, 010, 012, 017, 019, 020, 022, 025, 029 |
| 3 | 035, 036, 037 |
| 4 | 013, 014, 018, 021, 024, 028 |
| 5 | 006, 011, 023 |
| 6 | 026, 027, 038, 039 |
| 7 | (docs/release — no numbered tasks) |
| done | 001 |
