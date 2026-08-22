# Flightdeck v2 — Delivery Roadmap

> The plan of record. Product rationale: `docs/PRODUCT.md`. Target design:
> `docs/architecture/ARCHITECTURE.md`. Live status: `.ai/state.json`.
> The v1 plan is archived at `docs/archive/ROADMAP-v1.md`.
> Last updated: 2026-07-26

---

## How this plan works

Work ships in **waves**. A wave is a self-contained, independently mergeable
increment that leaves `main` in a working state. Each wave has:

- a **goal** — one sentence,
- **tasks** with IDs (`W3-T02`), tracked in `.ai/state.json`,
- **exit criteria** — objective, checkable statements,
- **runbooks** — the manual validation the maintainer performs before merging.

Waves are sequential; tasks within a wave may be parallel. The wave is not done
until every exit criterion holds *and* its runbooks pass on real hardware.

### Definition of Done (applies to every task, no exceptions)

A task is done when all of these are true:

1. **Code** implements the task and nothing beyond it.
2. **Go tests** — table-driven, hermetic (`t.TempDir()`, `toolchain.Fake`, no
   live cluster), covering the happy path, ≥2 error/edge cases per exported
   function, and context cancellation wherever a `context.Context` is accepted.
   Package statement coverage **≥ 80%**. Race detector clean.
3. **Shell tests** — any script with branching logic has `bats` coverage with
   `kubectl`/`helm` stubbed on `PATH`.
4. **API/contract tests** — any new endpoint or CLI command has an integration
   test against a fake toolchain, including its error envelope.
5. **UI tests** — any new component has Vitest + Testing Library coverage; any
   new user journey has a Playwright test against a mocked API.
6. **Docs** updated in the same commit — reference docs, `docs/PRODUCT.md` or
   `docs/architecture/` if behaviour changed, and an ADR if a decision was made.
7. **Runbook** written or updated so a human can verify the feature by hand.
8. **State** updated — task status and the `next` pointer in `.ai/state.json`.

Full detail on the four test layers and CI gates: `docs/TESTING.md`.

---

## What v2 changes from v1

| Decision | v1 | v2 |
|---|---|---|
| Focus | 79 tasks across 9 milestones incl. marketplace, editions, certificates | Rock-solid core simulator; commercial surfaces removed |
| Execution | `cmd.Run()`, no context, in-memory job map (cap 100), dropped log lines | Durable run engine: cancellation, timeouts, locking, persisted replayable logs |
| Persistence | JSON files under `.labctl/` | SQLite (pure Go), migrated, transactional |
| Runtimes | k3d, kind, incluster, **aks, eks, gke** (never verified) | k3d, kind, incluster only |
| UI | Hand-rolled SPA, no router, no tests | Full rebuild: router, query layer, design system, tested at two layers |
| Module root | `cmd/labctl/go.mod` | repo root |
| Shell tests | none (128 scripts) | bats required for all scripts with logic |
| Content | 13 scenarios, none e2e-verified | Curated verified set, proven in CI on kind; rest marked unverified |

**Cut in v2:** `pkg/pack`, `pkg/entitlement`, `pkg/edition`, `pkg/credential`,
marketplace API + UI, `registry/`, `packs/`, `sdk/pack-template`,
`runtimes/{aks,eks,gke}`, `foundation/terraform`, ACR/ECR build strategies,
`docs/cloud-runtimes.md`, `docs/strategy/`, and the associated CI workflows.

**No migration tooling is provided** — there are no production users, and the
migration guide would cost more than it saves.

---

## Wave 0 — Ground clearing & foundations

**Goal:** A clean, correctly-structured repository with CI that enforces the new
quality bar, before any feature work begins.

| ID | Task |
|---|---|
| W0-T01 | Delete cloud runtimes, Terraform, and cloud build strategies |
| W0-T02 | Delete marketplace/entitlement/edition/credential surfaces; keep `pkg/extension` seam |
| W0-T03 | Move Go module to repo root; restructure into `internal/{cli,httpapi,service,...}` |
| W0-T04 | Rebuild `Makefile` + `make/*.mk` for the v2 target set |
| W0-T05 | CI pipeline: golangci-lint, gosec, govulncheck, shellcheck, shfmt, race tests, coverage gate |
| W0-T06 | Test harness scaffolding: bats runner + kubectl/helm stubs, Vitest + MSW, Playwright |
| W0-T07 | Rewrite `.ai/state.json` to wave-based v2 schema; archive v1 tasks |
| W0-T08 | Rewrite `CLAUDE.md`, `README.md`, `CONTRIBUTING.md` for v2 |

**Exit criteria**
- `go build ./...` and `go test -race ./...` pass from the repo root.
- CI fails a PR that drops any package below 80% coverage.
- CI fails a PR with a lint, `shellcheck`, `gosec` or `govulncheck` finding.
- No reference to AKS/EKS/GKE/marketplace remains outside `docs/archive/`.
- `bats`, `vitest` and `playwright` each run (with placeholder tests) in CI.

**Runbooks:** `R00-environment-and-build`

---

## Wave 1 — Durable core: store, run engine, toolchain

**Goal:** Every operation Flightdeck performs is cancellable, time-bounded,
serialised against conflicts, and durably recorded.

| ID | Task |
|---|---|
| W1-T01 | `internal/store`: SQLite open/migrate/close, embedded forward-only migrations, schema-version guard |
| W1-T02 | Store repositories: runs, run_logs, run_steps, audit — with transactional writes |
| W1-T03 | `internal/toolchain`: bash/kubectl/helm/k3d/kind adapters, argv-only, containment-checked script resolution |
| W1-T04 | Toolchain preflight (min versions, actionable errors) + `labctl doctor` |
| W1-T05 | `toolchain.Fake` for hermetic tests across all layers |
| W1-T06 | `internal/run`: submit/queue/worker pool, exclusive lock keys, 409-on-conflict |
| W1-T07 | Cancellation: process groups, SIGTERM → grace → SIGKILL, graceful server shutdown |
| W1-T08 | Timeouts per run kind; `timed_out` terminal state |
| W1-T09 | Durable log capture with monotonic sequence + cursor replay API |
| W1-T10 | Step marker parsing → structured step timeline |
| W1-T11 | `labctl runs list/logs/cancel` |

**Exit criteria**
- A run cancelled mid-`helm install` leaves no orphaned child processes
  (verified by process-group assertion in test and by runbook on real hardware).
- Killing the server mid-run and restarting shows the run as `cancelled` with
  its partial log intact and readable.
- A second run with a conflicting lock key is rejected in <100ms naming the
  holder.
- Reconnecting a log stream with `?after=<seq>` returns zero gaps and zero
  duplicates — asserted by a test that drops the connection mid-stream.
- `labctl doctor` detects a missing binary, an outdated binary and an
  unreachable cluster, with a distinct actionable message for each.

**Runbooks:** `R01-run-engine-and-cancellation`, `R02-doctor-and-preflight`

---

## Wave 2 — Content model & check engine v2

**Goal:** All declarative content is schema-validated, cross-referenced and
verifiable, and checks report *why* they failed.

| ID | Task |
|---|---|
| W2-T01 | v2 YAML schemas for scenario, incident, path, challenge (`sdk/schemas/`) |
| W2-T02 | `internal/catalog`: load, validate, index, atomic hot-reload |
| W2-T03 | Cross-reference integrity checking (path→scenario, challenge→incident, check→namespace) |
| W2-T04 | Typed template resolution; unknown key is an error |
| W2-T05 | `pkg/checks` v2: `eventually` with backoff + deadline, per-check timeout, independent execution |
| W2-T06 | Check results carry observed vs expected and a human-readable explanation |
| W2-T07 | `labctl validate` + CI gate; fuzz tests on all YAML parsers |
| W2-T08 | Migrate existing content to v2 schemas |
| W2-T09 | `FLIGHTDECK_CONTENT_PATH` external content roots |

**Exit criteria**
- Every file under `scenarios/`, `incidents/`, `learn/`, `challenges/` passes
  `labctl validate`; CI fails otherwise.
- A path referencing a missing scenario fails at load with the file, line and
  the missing reference named.
- A failing check prints observed and expected values, not just `FAIL`.
- Fuzz corpus runs clean — no panic on malformed, truncated or adversarial YAML.
- An external content root is discovered and usable without forking the repo.

**Runbooks:** `R03-content-authoring-and-validation`

---

## Wave 3 — Lab & platform lifecycle on the new engine

**Goal:** Cluster and platform-component lifecycle runs entirely through the
durable engine, with exact teardown and instant status.

| ID | Task |
|---|---|
| W3-T01 | `internal/service/lab`: up/down/status/snapshot/reset on the run engine |
| W3-T02 | Runtime profiles reduced to k3d, kind, incluster; profile contract documented |
| W3-T03 | `internal/service/platform`: install/uninstall/status per category and provider |
| W3-T04 | Component state recorded in the store; `down` uninstalls exactly what was installed |
| W3-T05 | Idempotency + partial-failure recovery (resume a half-installed stack) |
| W3-T06 | bats coverage for every `platform/*/*/{install,uninstall,status}.sh` with stubbed helm/kubectl |
| W3-T07 | bats coverage for `runtimes/*` and `bootstrap/` scripts |
| W3-T08 | `labctl lab`/`labctl platform` CLI with golden-file output tests |

**Exit criteria**
- `platform up` interrupted at any point, then re-run, converges — proven by a
  fault-injection test that kills the run at each step boundary.
- `lab down` removes every recorded component and reports anything it could not
  remove, rather than exiting 0 silently.
- `lab status` answers in <200ms from the store, with a `--live` flag for a
  real cluster query.
- Every shell script with branching logic has a bats test; `make test-shell`
  passes.

**Runbooks:** `R04-lab-lifecycle`, `R05-platform-components`

---

## Wave 4 — Simulation: scenarios, incidents, traffic

**Goal:** The core product loop — stage a situation, break it, verify the fix —
works end to end and is graded objectively.

| ID | Task |
|---|---|
| W4-T01 | `internal/service/scenario`: activate by stage, deactivate, status, verify |
| W4-T02 | Objectives and per-stage checks surfaced in results |
| W4-T03 | `internal/service/incident`: inject, resolve, status, auto-detect resolution via checks |
| W4-T04 | Progressive hints with scoring cost; MTTD/MTTR measurement |
| W4-T05 | Traffic generator (k6 profiles: steady, spike, soak) as a managed run |
| W4-T06 | Snapshot/reset fast-path so a scenario retry takes seconds |
| W4-T07 | Curate and harden the verified content set (~6 scenarios, ~8 incidents) |
| W4-T08 | Mark remaining content `unverified`; UI and CLI display the distinction |
| W4-T09 | On-call drill: alerts route through Alertmanager to the pager service |

**Exit criteria**
- Each verified scenario activates, passes its checks, and tears down cleanly on
  a fresh k3d cluster — automated in the nightly kind e2e job.
- Each verified incident injects, is detected by its own check, and resolves —
  same automation.
- MTTD/MTTR are recorded and reproducible across two runs of the same incident.
- Hint consumption reduces the recorded score by the declared amount.
- Traffic generator runs are cancellable and leave no stray pods.

**Runbooks:** `R06-scenario-loop`, `R07-game-day-and-incidents`

---

## Wave 5 — API v2 & security hardening

**Goal:** A versioned, consistent, hardened HTTP surface that the UI and third
parties can rely on.

| ID | Task |
|---|---|
| W5-T01 | `/api/v2` routing, request IDs, structured access logging |
| W5-T02 | RFC 7807 `problem+json` error envelope with stable type slugs, applied uniformly |
| W5-T03 | Cursor pagination + ETag/`If-None-Match` on catalog reads |
| W5-T04 | WS + SSE streaming from cursor; reconnect semantics documented and tested |
| W5-T05 | Auth v2: Argon2id, SQLite-backed sessions, secure cookie flags |
| W5-T06 | CSRF protection, auth rate limiting, constant-time comparison |
| W5-T07 | Refuse non-loopback bind without auth; TLS support |
| W5-T08 | Audit log of all mutating operations, queryable |
| W5-T09 | Optional Prometheus `/metrics` |
| W5-T10 | Full contract test suite + OpenAPI document generated from it |

**Exit criteria**
- Every endpoint has a contract test asserting success shape *and* error
  envelope; the OpenAPI doc is generated, not hand-written.
- `--bind 0.0.0.0` without `FLIGHTDECK_AUTH=true` exits non-zero with an
  explanation.
- Sessions survive a server restart; logout invalidates server-side.
- Rate limiting demonstrably blocks a credential-stuffing loop in test.
- `gosec` and `govulncheck` clean; a security review runbook is completed.

**Runbooks:** `R08-api-and-security`

---

## Wave 6 — UI rebuild: shell, design system, operational views

**Goal:** A genuinely pleasant operational interface — deep-linkable, honest
about state, with a run console people want to watch.

| ID | Task |
|---|---|
| W6-T01 | New Vite/React/TS-strict app skeleton; React Router with real URLs |
| W6-T02 | TanStack Query data layer: caching, retry, invalidation on run completion |
| W6-T03 | Design system: Tailwind tokens, light/dark, typography scale, Radix primitives |
| W6-T04 | App shell: navigation, command palette, connection status, notifications |
| W6-T05 | Run console: step timeline, live cursor-based log stream, cancel, failure summary |
| W6-T06 | Overview view: cluster health, active scenario, recent runs, next suggested action |
| W6-T07 | Labs view: lifecycle controls, component inventory, snapshot/reset |
| W6-T08 | Catalog view: browse/search scenarios and incidents with rich descriptions |
| W6-T09 | Designed loading/empty/error/offline states for every view |
| W6-T10 | Accessibility pass: keyboard nav, ARIA, WCAG AA contrast both themes |
| W6-T11 | Vitest + Testing Library component tests; MSW API mocking |
| W6-T12 | Playwright journeys: run a scenario, watch logs, cancel a run |

**Exit criteria**
- Every view is reachable by URL and survives a page refresh with state intact.
- Killing the backend mid-session produces a designed offline state and
  automatic recovery, not a blank page or an infinite spinner.
- Log streaming has no gaps across a forced disconnect (Playwright asserts it).
- Axe accessibility scan reports zero serious/critical violations.
- Vitest ≥80% statement coverage on components; Playwright journeys green in CI.

**Runbooks:** `R09-ui-operational-walkthrough`

---

## Wave 7 — UI: learning, challenges, results

**Goal:** The assessment loop is visible, motivating and clear.

| ID | Task |
|---|---|
| W7-T01 | Learn view: path browser, module progression, inline prose and objectives |
| W7-T02 | Guided scenario walkthrough: objectives, live check status, contextual hints |
| W7-T03 | Challenges view: timer, hidden hints with cost, submission and grading |
| W7-T04 | Results view: history, score breakdown, per-check detail, MTTR trend charts |
| W7-T05 | Leaderboard (server mode) with team aggregation |
| W7-T06 | Progress visualisation: completion, streaks, weak-area suggestions |
| W7-T07 | Component + Playwright coverage for the full assessment journey |

**Exit criteria**
- A user can complete a learning path start to finish entirely in the UI.
- A challenge run records score, elapsed time and hints used, and the score is
  reproducible from the stored result.
- Charts render correctly with 0, 1 and 1000 results (empty/degenerate/scale).
- Playwright covers: start path → run scenario → verify → complete → see result.

**Runbooks:** `R10-learning-and-assessment`

---

## Wave 8 — Team server, packaging, release

**Goal:** Flightdeck ships as a trustworthy artifact and runs for a team.

| ID | Task |
|---|---|
| W8-T01 | In-cluster server mode on the v2 stack; PVC-backed SQLite |
| W8-T02 | Helm chart rebuild with SA/RBAC least-privilege, probes, resource limits |
| W8-T03 | Multi-arch container image; non-root, distroless-based |
| W8-T04 | Release automation: goreleaser, checksums, SBOM, cosign signing |
| W8-T05 | Nightly e2e on kind gating the verified content set |
| W8-T06 | Upgrade path: schema migration test across versions |
| W8-T07 | Install docs, quickstart, troubleshooting guide |

**Exit criteria**
- `helm install` on a fresh kind cluster yields a working UI in under 5 minutes.
- Two users on the shared server see consistent leaderboard state and cannot
  clobber each other's runs (lock keys enforced).
- Release artifacts verify: checksums match, cosign signature validates, SBOM
  present.
- Upgrading from the previous release migrates the database without data loss —
  asserted by an automated cross-version test.

**Runbooks:** `R11-team-server`, `R12-release-verification`

---

## Runbook index

Runbooks are the human validation gate. They live in `docs/runbooks/` and are
written to be followed by a person on real hardware, with expected output stated
for every step and explicit failure signatures.

| ID | Runbook | Wave |
|---|---|---|
| R00 | Environment & build | W0 |
| R01 | Run engine & cancellation | W1 |
| R02 | Doctor & preflight | W1 |
| R03 | Content authoring & validation | W2 |
| R04 | Lab lifecycle | W3 |
| R05 | Platform components | W3 |
| R06 | Scenario loop | W4 |
| R07 | Game day & incidents | W4 |
| R08 | API & security | W5 |
| R09 | UI operational walkthrough | W6 |
| R10 | Learning & assessment | W7 |
| R11 | Team server | W8 |
| R12 | Release verification | W8 |

## Open items for the maintainer

Not automatable — they need a human with accounts or authority:

- **`go.flightdeck.dev` domain.** The module path depends on a domain that is not
  yet owned. Either register and host the vanity import metadata, or accept a
  rename before the first tagged release.
- **Repository name.** The repo is still `Sagars-Laboratory` while the product
  is Flightdeck. Rename when convenient; v2 assumes the Flightdeck name.
- **Branch protection and CI required-checks** on `main`.
- **Trademark search** before promoting the Flightdeck name publicly.
