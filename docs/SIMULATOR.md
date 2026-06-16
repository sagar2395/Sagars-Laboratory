# Platform Engineering Simulator — Vision & Feature Catalog

> The "why and what" of the next era of Sagars-Laboratory. The "when and in what
> order" lives in `docs/ROADMAP.md` (Part II); per-task detail in `.ai/tasks/`.
> Last updated: 2026-06-11

---

## Vision

Evolve the homelab toolkit into a **Platform Engineering Simulator**: an
environment where DevOps/SRE/platform teams can **replicate real-world
production scenarios across different technical stacks** — and break them,
fix them, and measure themselves doing it.

Four use cases drive every feature decision:

| Use case | Who | What they need |
|---|---|---|
| **Learning** | Individual engineers | Guided paths, scenarios with objectives, hints, verifiable outcomes |
| **PoC / evaluation** | Platform teams | Swap stacks (mesh A vs B, queue A vs B) on identical workloads, compare |
| **Experimentation** | SRE / staff engineers | Inject realistic faults, run game days, snapshot & reset state fast |
| **Skills assessment** | Teams / leads | Timed challenges, auto-graded checks, scores, team leaderboards |

The architecture stays the same — **`labctl` orchestrates, shell scripts and
declarative YAML do the work** — but three new engines sit on top of it.

## The three new engines

### 1. Simulation engine (scenario format v2)

Today's scenarios install components and print tips. v2 makes them
**interactive simulations** by adding three blocks to `scenario.yaml`:

```yaml
stages:        # ordered steps, each installable/activatable independently
  - name: baseline
    components: [...]
  - name: inject-failure
    components: [...]
objectives:    # human-readable goals per stage
  - "Restore p99 latency below 300ms"
checks:        # machine-verifiable assertions (the grading primitive)
  - type: http        # http | kubectl | promql | script
    url: "http://go-api.{{.DomainSuffix}}/health"
    expectStatus: 200
  - type: promql
    query: 'histogram_quantile(0.99, ...)'
    operator: "<"
    value: 0.3
```

`labctl scenario verify <name>` runs the checks and reports pass/fail. The
same checks power CI validation, challenge grading, and incident resolution
detection. **Checks are the single most important new primitive** — every
other feature builds on them.

Plus: a built-in **traffic generator** (k6-based profiles: steady / spike /
soak) so every simulation runs under realistic load, **lab snapshot/reset**
for fast iteration, and a **scenario catalog** (`labctl scenario install
<git-url>`) so scenarios become shareable packs, not just in-repo dirs.

### 2. Incident engine (game days)

A library of realistic, reversible production faults — each one a small
dir with `inject.sh`, `resolve.sh`, `hints.md`, and a `fault.yaml` declaring
its detection check:

```
incidents/<name>/      e.g. crashloop-bad-config, oom-kill, dns-blackhole,
                       cert-expiry, disk-pressure, noisy-neighbor,
                       bad-deploy-rollout, pvc-full, network-partition
```

- `labctl incident inject <name>` / `--random` / `--category network`
- `labctl incident hint` — progressive hints (each hint costs score)
- `labctl incident status` — uses the fault's check to detect resolution,
  records **time-to-detect and time-to-resolve (MTTR)**
- **On-call drills**: faults fire real alerts through Alertmanager to a
  webhook receiver, so teams practice the full page → triage → fix loop.

### 3. Assessment engine (learning & scoring)

- **Learning paths** (`learn/<path>/path.yaml`): ordered modules that chain
  scenarios + incidents with prose, objectives, and completion checks.
  `labctl learn start kubernetes-foundations`, `labctl learn progress`.
- **Challenge mode**: a scenario or incident run with a timer and hidden
  hints; grade = checks passed, time, hints used. `labctl challenge start <name>`.
- **Score persistence**: results stored in `.labctl/` locally and exposed via
  the REST API; the web UI gets Learn / Challenges / Leaderboard views.

## Stack expansion (swap & compare real-world stacks)

New platform categories, same provider contract (`install.sh` /
`uninstall.sh` / `status.sh` / `values.yaml`, selected by env var):

| Category | Providers | Example simulation it unlocks |
|---|---|---|
| `platform/mesh` | istio, linkerd | Traffic shifting, canary, mTLS, fault injection at mesh level |
| `platform/data` | kafka (strimzi), postgres (cnpg), rabbitmq | Event-driven arch, consumer lag incidents, DB failover |
| `platform/secrets` | vault, external-secrets | Secret rotation, leaked-secret incident response |
| `platform/autoscaling` | keda, vpa | Scale-on-queue-depth, load-spike survival challenges |
| `platform/cost` | opencost | Right-sizing exercises, cost-per-namespace visibility |

Each new category ships with **at least one scenario that uses it** — a
category without a simulation is dead weight.

## Multi-environment & day-2 operations

- **Env promotion simulation** ✅ `env-promotion` scenario + `labctl env`
  commands: dev → staging → prod as three namespaces in one cluster (default,
  laptop-friendly) or as separate k3d clusters (flag). `labctl env promote dev
  staging` updates the declared image-tag ConfigMap — a lightweight stand-in
  for a GitOps manifest commit. `labctl env list` shows the release train.
  Full runbook: `docs/runbooks/11-multi-env-day2.md`.
- **Day-2 drills**: cluster version upgrade, node drain/cordon under load,
  etcd/state backup & restore, certificate rotation — each as a scenario
  with checks proving zero (or measured) downtime.

## Team mode (later)

- Optional auth + per-user RBAC on the API/UI; team sessions on a shared
  remote deployment (labctl server as a Helm chart); shared leaderboards.
- More runtimes: kind (CI-friendly), GKE (third cloud).

## What stays true (non-negotiable)

1. Everything above is **declarative-first**: scenarios, faults, paths and
   checks are YAML + scripts, never Go logic. Go only orchestrates.
2. **Cross-platform** (macOS + Linux) and **idempotent** — golden rules in
   `CLAUDE.md` apply to every new engine.
3. A feature without **docs + runbook + checks** does not ship.

## Out of scope (deliberately)

- A hosted SaaS / multi-tenant cloud service.
- Reimplementing existing tools (k6, Chaos Mesh, Litmus) — we wrap them.
- Offensive-security attack labs beyond defensive misconfiguration drills.

## Roadmap

Milestones M1–M6 with priorities and task ids: see **`docs/ROADMAP.md`,
Part II**. Live status: `.ai/state.json`.
