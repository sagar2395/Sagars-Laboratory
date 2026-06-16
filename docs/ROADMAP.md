# ROADMAP — Path to the Desired State

> The single plan of record. **Part I** (phases 0–7) is the original homelab
> hardening plan — nearly complete. **Part II** (milestones M1–M6) is the
> Platform Engineering Simulator era — the long-term build-out. Vision and
> feature rationale: `docs/SIMULATOR.md`. Status of each task lives in
> `.ai/state.json`; details in `.ai/tasks/`.
> Last updated: 2026-06-11

---

# Part I — Homelab hardening (phases 0–7)

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

## Task ↔ phase index (Part I)

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

---

# Part II — Platform Engineering Simulator (milestones M1–M6)

> Vision and feature rationale: `docs/SIMULATOR.md`. Read it once before
> working any Part II task. Milestones are ordered by priority; **M1 is the
> foundation everything else builds on** (the `checks` primitive). M2–M3 are
> the core simulator value. M4–M6 expand breadth and can be parallelized once
> M1 is done.

```
M1  Scenario Engine v2 (checks, traffic, snapshot, catalog)   P0  ── gate
                │
   ┌────────────┼──────────────┐
M2  Incident    M3  Learning &  M4  Stack
    Engine          Assessment      Expansion        P0/P1
   └────────────┴──────────────┘
                │
M5  Multi-Env & Day-2 Ops                            P2
M6  Team Mode & New Runtimes                         P2
```

---

## M1 — Scenario Engine v2 `[P0 — the gate for everything]`

**Goal:** scenarios become verifiable simulations. Introduce `stages`,
`objectives`, and machine-checkable `checks` to the scenario format; add the
traffic generator, lab snapshot/reset, and external scenario packs.

**Tasks**

| Id | Title | Priority |
|----|-------|----------|
| 040 | Scenario format v2: stages, objectives, checks schema + parser | P0 |
| 041 | Verification engine: `labctl scenario verify` with http/kubectl/promql/script check runners | P0 |
| 042 | Traffic generator: k6-based load profiles as a platform service | P1 |
| 043 | Lab snapshot & reset: `labctl lab snapshot/restore/reset` | P1 |
| 044 | Scenario catalog: `labctl scenario install <git-url>` for external packs | P2 |

**Exit criterion:** every existing scenario has at least 3 checks and
`labctl scenario verify <name>` passes on a freshly activated scenario and
fails when a component is deleted. Traffic profiles run against go-api.
`lab reset` returns a dirty cluster to the post-init state in < 5 minutes.

**Runbook:** `docs/runbooks/07-scenario-engine-v2.md`

---

## M2 — Incident Engine (game days) `[P0/P1 — depends on M1]`

**Goal:** realistic, reversible production faults with guided resolution,
MTTR measurement, and on-call drills.

**Tasks**

| Id | Title | Priority |
|----|-------|----------|
| 045 | Fault library: `incidents/<name>/` contract (inject.sh, resolve.sh, fault.yaml, hints.md) + first 6 faults | P0 |
| 046 | `labctl incident` command group: inject / list / status / resolve / --random | P0 |
| 047 | Progressive hints + solution walkthroughs (`labctl incident hint`) | P1 |
| 048 | MTTR tracking: time-to-detect / time-to-resolve recorded per run | P1 |
| 049 | On-call drill: Alertmanager → webhook receiver, full page-triage-fix loop | P2 |

**Exit criterion:** `labctl incident inject --random` breaks the lab in a
way that fires an alert; `incident status` detects resolution via the
fault's check; MTTR is recorded and queryable. All faults are reversible
via `resolve.sh` (the escape hatch).

**Runbook:** `docs/runbooks/08-incident-engine.md`

---

## M3 — Learning & Assessment `[P1 — depends on M1; M2 enriches it]`

**Goal:** guided learning paths and timed, auto-graded challenges with
persistent scores — usable by individuals and by teams testing skills.

**Tasks**

| Id | Title | Priority |
|----|-------|----------|
| 050 | Learning path format (`learn/<path>/path.yaml`) + `labctl learn` command | P1 |
| 051 | Challenge mode: timed runs, hidden hints, grade from checks (`labctl challenge`) | P1 |
| 052 | Score & progress persistence in `.labctl/` + REST API endpoints | P1 |
| 053 | Web UI: Learn, Challenges, and Leaderboard views | P2 |

**Exit criterion:** a new user can `labctl learn start kubernetes-foundations`
and complete a module with verified checks; a challenge produces a score
(checks passed, time, hints used) that survives restarts and shows in the UI.

**Runbook:** `docs/runbooks/09-learning-and-challenges.md`

---

## M4 — Stack Expansion `[P1/P2 — depends on M1]`

**Goal:** swap and compare real-world stacks on identical workloads. Every
new platform category ships with a scenario that uses it.

**Tasks**

| Id | Title | Priority |
|----|-------|----------|
| 054 | `platform/mesh` category: istio + linkerd providers | P1 |
| 055 | `platform/data` category: kafka (strimzi) + postgres (cnpg) providers | P1 |
| 056 | `platform/secrets` category: vault + external-secrets providers | P2 |
| 057 | `platform/autoscaling` category: keda provider + scale-on-load scenario | P2 |
| 058 | New scenarios: mesh-traffic-management, event-driven-arch, secrets-management | P1 |

**Exit criterion:** `MESH_PROVIDER=istio labctl platform up mesh` then swap
to linkerd on the same workload; the three new scenarios activate, verify,
and deactivate cleanly on k3d.

**Runbook:** `docs/runbooks/10-stack-expansion.md`

---

## M5 — Multi-Env & Day-2 Ops `[P2 — depends on M1, M4 (gitops)]`

**Goal:** simulate release engineering across environments and the scary
day-2 operations, locally and for free.

**Tasks**

| Id | Title | Priority |
|----|-------|----------|
| 059 | Multi-cluster env promotion: dev → staging → prod via GitOps, as a scenario | P2 | ✅ done |
| 060 | Day-2 drills: cluster upgrade, node drain under load, backup/restore — as checked scenarios | P2 | ✅ done |
| 061 | Cost & capacity: opencost provider + right-sizing exercise scenario | P2 | ✅ done |

**Exit criterion:** an image promoted dev → prod purely via Git; the upgrade
drill records measured downtime via checks; opencost shows per-namespace cost.

**Runbook:** `docs/runbooks/11-multi-env-day2.md`

---

## M6 — Team Mode & New Runtimes `[P2 — last]`

**Goal:** teams share one simulator deployment; broaden runtime support.

**Tasks**

| Id | Title | Priority |
|----|-------|----------|
| 062 | Optional auth + per-user RBAC on REST API / UI | P2 | ✅ done |
| 063 | Team sessions: labctl server Helm chart for shared remote deploy + shared leaderboard | P2 | ✅ done |
| 064 | New runtimes: kind (CI-friendly) and GKE | P2 |

**Exit criterion:** two users with separate identities use one deployed
simulator; their challenge scores appear on a shared leaderboard; `labctl
runtime up --profile kind` works headless in CI.

**Runbook:** `docs/runbooks/12-team-mode.md`

---

## Task ↔ milestone index (Part II)

| Milestone | Priority | Task ids |
|-----------|----------|----------|
| M1 | P0 | 040, 041, 042, 043, 044 |
| M2 | P0/P1 | 045, 046, 047, 048, 049 |
| M3 | P1 | 050, 051, 052, 053 |
| M4 | P1/P2 | 054, 055, 056, 057, 058 |
| M5 | P2 | 059, 060, 061 |
| M6 | P2 | 062, 063, 064 |

**Picking the next task:** take the lowest-numbered unblocked task of the
highest-priority milestone in `.ai/state.json` (`milestones` block tracks
per-milestone status; `next` points at the recommended pick).

---

# Part III — OSS & Commercial Evolution (milestones M7–M9)

> Full plan of record: [`docs/strategy/OSS-COMMERCIAL-STRATEGY.md`](strategy/OSS-COMMERCIAL-STRATEGY.md).
> Part III turns the simulator into a community OSS project (Apache-2.0, CLA,
> "Flightdeck", org `snowops`) with a clean public SDK and a pack ecosystem, then
> layers entitled premium content/services on the **same engine** — never a fork.

## M7 — OSS & Ecosystem Foundation `[P0 — ✅ complete]`

| Task | Title | Status |
|------|-------|--------|
| 065 | OSS governance & licensing baseline (Apache-2.0, CLA) | ✅ done |
| 066 | Public SDK boundary — `pkg/checks`, `pkg/scenario`, schemas, RFC 0001 | ✅ done |
| 067 | Scenario-pack format (`pack.yaml`) + `labctl pack` | ✅ done |
| 068 | OCI pack distribution + cosign signing | ✅ done |
| 069 | Registry index & discovery (`labctl pack search`) | ✅ done |
| 070 | Entitlement / extension interface | ✅ done |
| 071 | Module path & brand alignment | ✅ done |
| 072 | Contributor experience (scaffolds, walkthrough) | ✅ done |

## M8 — Marketplace `[P2 — deferred, depends on M7]`

Tasks 073–075: hosted catalog API, premium pack repo + entitlement, marketplace
UI. Premium content lives outside the OSS tree.

## M9 — Commercial / Hosted `[P2 — deferred, depends on M7]`

Tasks 076–078: edition packaging, SaaS control-plane spike, certification &
training framework. Same engine + entitled content/services.

## Task ↔ milestone index (Part III)

| Milestone | Priority | Task ids |
|-----------|----------|----------|
| M7 | P0 | 065, 066, 067, 068, 069, 070, 071, 072 |
| M8 | P2 | 073, 074, 075 |
| M9 | P2 | 076, 077, 078 |
