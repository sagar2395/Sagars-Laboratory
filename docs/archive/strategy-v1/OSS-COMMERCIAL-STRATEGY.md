# Open-Source & Commercial Strategy

> Status: **DECISIONS LOCKED 2026-06-14** — core license **Apache-2.0**,
> **CLA** for contributions, GitHub org **`snowops`** with a **monorepo core +
> 2 satellites**. One sub-decision open: the **trademark-able product name**
> (the repo/module identity, task 071). Author: AI session. Owner: @sagar2395.
>
> This document proposes how to evolve Sagars-Laboratory from a personal homelab
> project into a community-driven open-source **platform-engineering simulator**
> with optional future commercial offerings — without a future redesign.
>
> It is deliberately long. It optimises for long-term sustainability,
> contributor adoption, extensibility, and commercial flexibility — not
> short-term implementation convenience.

---

## 0. TL;DR and the 4 decisions only you can make

**The vision: "Flight Simulator for Platform Engineers."** A stable open-source
**engine** (the simulator), an open **content ecosystem** of scenario packs (the
"aircraft and missions"), a **marketplace** (the add-on store), and later
**certification** (the "pilot ratings"). The engine is the moat-free commons;
the money is in convenience (hosted), curation (premium/enterprise packs), and
assurance (certification, support) — never in crippling the core.

**Headline recommendations**

| Area | Recommendation |
|---|---|
| Repo strategy | **Monorepo for the entire OSS core**, a **separate private repo** for premium content, a **separate public index repo** for the marketplace catalog. Do *not* split engine/CLI into many repos yet. |
| Extension model | Keep it **100% data-driven** (declarative YAML + sandboxed scripts), never compiled Go plugins. Add a formal **pack format** + **stable schema apiVersion** + **entitlement interface**. |
| Distribution | **Git packs now** (already built), **OCI artifacts next** (versioned, signable, auth-gatable). A **static signed index** for discovery. |
| Licensing | **Apache-2.0** for engine + CLI + SDK (max adoption, patent + trademark safety). **Open-core**: premium content under a separate proprietary EULA, never mixed into the OSS tree. Protect the SaaS moat with **trademark + hosted convenience + premium content**, not license restrictions. |
| Governance | **Maintainer-led (you hold architectural + release authority)** via `CODEOWNERS` + `GOVERNANCE.md`; open contribution for scenarios/docs/modules; **CLA** to preserve relicensing/commercial flexibility. |

**The 4 pivotal decisions — RESOLVED 2026-06-14:**

1. **Core license** — ✅ **Apache-2.0**.
2. **Contribution agreement** — ✅ **CLA** (via CLA Assistant).
3. **Brand & GitHub home** — ✅ dedicated repo **`snowops/flightdeck`**; product
   name **Flightdeck** (pending a formal trademark knock-out search); vanity
   import path `go.flightdeck.dev/...` (drives task 071).
4. **Repo strategy** — ✅ **monorepo core + 2 satellites** (`registry`,
   private premium).

In the trees below, `flightdeck` is the chosen product name (pending a trademark
knock-out search); the core monorepo is `snowops/flightdeck`.

Everything below is structured so that only a handful of tasks (§7) depend on
these answers; the rest can start immediately.

---

## 1. Recommended repository strategy

### 1.1 Monorepo vs multi-repo — recommendation: **monorepo core + 2 satellites**

A young ecosystem dies from fragmentation: split repos multiply CI, versioning,
issue triage, and "where do I file this?" confusion, and they make atomic
cross-cutting changes (engine + CLI + scenario schema) painful. Conversely,
*everything* in one repo blocks clean licensing separation and makes the premium
boundary leak.

The sustainable middle path is **three repositories**:

| Repo | Visibility | Contains | Why separate |
|---|---|---|---|
| **`snowops/flightdeck`** (core monorepo) | Public, Apache-2.0 | engine, `labctl` CLI, SDK (`pkg/`), platform modules, runtimes, **community scenario packs**, docs, web UI | One place to contribute, one CI, atomic cross-cutting changes |
| **`snowops/registry`** (marketplace index) | Public, Apache-2.0/CC0 | the **catalog index** (signed JSON/YAML) listing community + verified third-party packs, publisher metadata | The catalog evolves on its own cadence; third parties PR their pack entries here without touching core |
| **`snowops/flightdeck-premium`** (premium content + entitlement) | **Private**, proprietary EULA | enterprise/premium scenario packs, certification content, entitlement/license-key service, SaaS control-plane | Must never be in the OSS tree; different license; restricted access |

Rationale: this mirrors what worked for **Grafana** (core OSS + enterprise repo),
**HashiCorp pre-BSL** (core + enterprise), **Argo/Backstage** (core + a plugin
index), and **kubectl Krew** (index repo pattern). It gives you a clean OSS story
(goal 1), independent installability (goal 5), and a marketplace seam (goal 6)
without prematurely shattering the contributor experience (goal 7).

### 1.2 What lives where (core monorepo internal layout)

See §8 for the full tree. The key principle: **a clear public/SDK boundary**.
Everything an external pack author or downstream tool depends on lives under a
versioned `pkg/` with a documented compatibility policy; everything else stays in
`internal/` and can change freely.

### 1.3 Migration plan from the current structure

The current repo is already ~90% of the core monorepo. Migration is mostly
**addition and renaming**, not restructuring — do it in safe, reversible steps:

1. **Identity & licensing first (M7, immediate).** Add `LICENSE`, `NOTICE`,
   governance files, `CODEOWNERS`, CLA/DCO. Decide the canonical module path /
   brand (decision 3). *Until this lands, every external contribution is a legal
   liability.*
2. **Carve the SDK boundary.** Move the stable schema types (scenario, check,
   pack, results) from `cmd/labctl/internal/...` into a public `pkg/` module with
   an explicit `apiVersion`. Keep `internal/` for everything else.
3. **Formalise the pack format** (`pack.yaml`) on top of the existing
   `catalog.go` machinery. Existing git-pack installs keep working.
4. **Extract community scenarios into `packs/community/`** within the monorepo
   (still shipped in-tree, but addressable as packs). The four built-in scenarios
   become the canonical first packs.
5. **Stand up `registry`** as a separate public repo with a seed index pointing
   at the in-repo community packs.
6. **Stand up `lab-premium`** as a private repo with the entitlement interface
   implemented (the OSS engine ships a no-op default).
7. **(Optional, decision 3) Rename/move** to a neutral org. Use a vanity import
   path (e.g. `go.lab.dev/...`) so the module path never has to change again even
   if the GitHub home moves.

No step requires a "big bang"; each is independently shippable and reversible.

---

## 2. Scenario plugin architecture

The engine is already correctly **content-driven** (declarative `scenario.yaml`
+ `manifest`/`helm`/`grafana-dashboard`/`script` components + `checks`). We
formalise the *packaging, identity, dependency, versioning, discovery, and
install* layers around that — without changing the golden rule that the CLI
orchestrates while YAML + scripts do the work.

### 2.1 Packaging format

A **scenario pack** is a self-contained, versioned unit of content. Two
distribution encodings, one logical format:

- **Git pack** (today): a repo/dir with scenarios. Simple, great for community.
- **OCI artifact** (next): the pack tarball pushed to *any* OCI registry (GHCR,
  ECR, ACR, Docker Hub, Harbor). This gives free versioning, content addressing,
  signing (cosign), and — crucially — **auth-gated access** that doubles as the
  premium entitlement mechanism (a registry that returns 401 without a license
  token). Same pattern as Helm OCI charts and Argo CD config-management plugins.

A pack is a directory with this shape (the flat layout the engine loads today —
one directory per scenario at the pack root, with `pack.yaml` alongside; see
`packs/examples/hello-pack/`):

```
my-pack/
  pack.yaml                 # the manifest (see §2.2) — at the pack root
  <scenario-name>/          # one directory per scenario (scenario.yaml + assets)
    scenario.yaml
    manifests/  values/  dashboards/  scripts/
  LICENSE                   # the pack's own license (community or premium)
  README.md
```

> Planned (later): an optional `scenarios/` subdir plus bundled `platform/`,
> `incidents/`, `learn/`, `challenges/` directories so a pack can ship modules and
> content together. The loader and `ValidatePackDir` gain the deeper scan then.

### 2.2 Scenario / pack metadata model (`pack.yaml`)

```yaml
apiVersion: packs.lab.dev/v1          # the PACK manifest schema version
kind: ScenarioPack
metadata:
  name: kafka-incident-drills          # unique within a publisher namespace
  version: 1.4.2                        # SEMVER for the pack
  publisher: acme-sre                   # verified-publisher namespace
  displayName: "Kafka Incident Drills"
  description: "..."
  homepage: https://...
  license: Apache-2.0                   # or "LicenseRef-ACME-Commercial"
  tier: community                       # community | premium | enterprise
  keywords: [kafka, incident, sre]
  categories: [data, incident-response]
spec:
  # What the engine must support to run this pack:
  engine:
    scenarioApiVersions: ["scenario.lab.dev/v2"]   # schema compatibility
    minLabctlVersion: "1.2.0"
  # Hard dependencies the pack needs present on the cluster/registry:
  requires:
    platform: [data/kafka, monitoring/metrics]     # platform categories
    packs:                                          # other packs (semver ranges)
      - name: core-observability
        version: ">=1.0.0 <2.0.0"
  provides:
    scenarios: [kafka-broker-outage, consumer-lag-storm]
    incidents: [kafka-isr-shrink]
  # Integrity / provenance (filled by the publish tooling):
  checksum: "sha256:..."
  signature: "cosign:..."               # optional, required for verified packs
```

Design choices and why:

- **Pack version is independent of engine version.** The engine advertises which
  *scenario schema apiVersions* it supports; packs declare which they need. This
  is the single most important decoupling for long-term sustainability — it lets
  content and engine evolve on separate clocks (goal 4).
- **`tier`** is metadata, not enforcement. The OSS engine treats all tiers
  identically; *entitlement* (can you fetch/run a premium pack) is enforced at
  the **distribution** layer (registry auth) and an optional **entitlement
  interface** (§2.6 of the roadmap), never by crippling the engine.
- **`publisher`** namespaces every pack, enabling a verified-publisher program
  and third-party packs without name collisions (goal 6).

### 2.3 Dependency model

- **Platform dependencies** reuse the existing prerequisite machinery (`platform:
  [mesh, data/kafka]`) — already validated by preflight (task 002).
- **Pack-to-pack dependencies** use **semver ranges**; the resolver installs a
  compatible set or fails with a clear conflict (no silent surprises).
- **Engine compatibility** is a constraint, not a dependency: a pack states the
  schema versions and min CLI version it needs; the CLI refuses incompatible
  packs with an actionable message.

### 2.4 Versioning strategy

- **Packs:** strict **SemVer 2.0**. Breaking scenario changes ⇒ major bump.
- **Scenario schema:** explicit `apiVersion: scenario.lab.dev/v2`. New schema
  features that break old engines ⇒ new apiVersion (`v3`), with the engine
  supporting **N and N-1** for a deprecation window.
- **Engine/CLI:** SemVer; a published **compatibility matrix** (which CLI
  supports which scenario apiVersions and pack apiVersions).
- **SDK (`pkg/`):** SemVer with a written stability policy — this is what third
  parties compile against.

### 2.5 Discovery mechanism

- **Index-based:** a static, **signed catalog index** (one JSON/YAML per pack
  version) hosted from the `registry` repo via GitHub Pages / a CDN. `labctl pack
  search <term>`, `labctl pack info <name>`. Same shape as Krew's index and
  Artifact Hub — cheap, cacheable, no server to run on day one.
- **Direct refs:** `labctl pack add oci://ghcr.io/acme/kafka-drills:1.4.2` or
  `labctl pack add git+https://...@v1.4.2` for packs not in the index.
- **Later:** a hosted catalog **API** (search, ratings, downloads, verified
  badges) — a superset of the static index, not a replacement.

### 2.6 Installation mechanism

Evolve the existing `InstallPack` flow:

1. Resolve the ref (index name → OCI/git URL, or a direct ref).
2. **Verify** integrity (checksum) and, for verified/premium packs, **signature**
   (cosign) and **entitlement** (license token → registry auth).
3. Validate every scenario against the supported schema apiVersion (reuse
   `ValidatePackDir`).
4. Materialise into `~/.labctl/packs/<publisher>/<name>@<version>/` (versioned,
   multiple versions can coexist).
5. Record provenance in a lockfile so installs are reproducible and auditable.

Community packs need none of the auth/entitlement steps — they degrade
gracefully to today's git flow.

---

## 3. Governance model

### 3.1 Roles

- **Lead maintainer / BDFL (you).** Final authority on architecture, roadmap,
  releases, the schema, and what merges to `main`. Holds the trademark and the
  release-signing keys. This is explicit and legitimate for an open-core project
  (cf. early Kubernetes/Linux/Vue models).
- **Maintainers (invited).** Trusted reviewers with merge rights to specific
  areas (e.g. a scenario domain). Added by you.
- **Reviewers / domain owners.** Can approve PRs in their area (via `CODEOWNERS`)
  but not override architectural decisions.
- **Contributors.** Anyone; contribute via fork + PR under the CLA/DCO.

### 3.2 CODEOWNERS strategy

The mechanism that *guarantees* you retain architectural and release authority
(goal 3) while opening the content surface (goal 2):

```
# Engine, CLI, SDK, schema, CI, licensing — YOU must review (locked)
/cmd/                        @sagar2395
/pkg/                        @sagar2395
/engine/                     @sagar2395
/.github/                    @sagar2395
/LICENSE  /GOVERNANCE.md     @sagar2395
/docs/strategy/              @sagar2395

# Content surfaces — open to domain maintainers (community can own these)
/packs/community/            @sagar2395 @scenario-maintainers
/scenarios/                  @sagar2395 @scenario-maintainers
/docs/                       @sagar2395 @docs-maintainers
/platform/                   @sagar2395 @platform-maintainers
```

Combined with **branch protection** (required review from code owners, required
CI, no direct pushes to `main`), this makes "you control the engine; the
community grows the content" a *mechanical* guarantee, not a social one.

### 3.3 CLA vs DCO (decision 2)

- **CLA (recommended).** A Contributor License Agreement (e.g. via the
  **CLA Assistant** GitHub app) grants you a license to relicense/redistribute
  contributions. This **preserves your ability to dual-license, ship premium
  derivatives, and offer SaaS** without chasing every past contributor for
  permission. Cost: a small contributor-friction (one click on first PR).
- **DCO.** A lighter `Signed-off-by` certification of origin. More
  contributor-friendly, but it **does not grant relicensing rights** — if you
  later need to change the license or build a closed derivative of community code,
  you may be blocked. Choose DCO only if you commit to never relicensing the
  core.

For a project that explicitly wants future commercial flexibility (goal 4), the
CLA is the safer long-term choice. It is a one-time setup.

### 3.4 Contribution workflow

```
fork → branch → PR → CLA check (bot) → CI (lint/test/build) →
CODEOWNERS review → maintainer approval → squash-merge to main
```

- **RFC process** for anything that touches the engine, schema, or public SDK:
  a short markdown PR under `docs/rfcs/` that you approve before implementation.
  This keeps architectural authority with you while making the *reasoning*
  transparent (great for contributor trust).
- **"Good first issue" / "help wanted"** labels and a `pack init` scaffold to
  funnel newcomers toward the content surface, where contribution is safe and
  high-value.
- **Release process:** SemVer tags, signed releases, an automated changelog, and
  a `RELEASING.md` only you can execute (release authority, goal 3).

---

## 4. Licensing strategy

### 4.1 Core (engine + CLI + SDK) — recommendation: **Apache-2.0**

- **Maximises adoption** (goal 2/7): enterprises and other OSS projects can
  depend on it without legal review friction.
- **Explicit patent grant** protects you and users — important for an
  infra/security tool.
- **Trademark-safe:** Apache-2.0 grants *no* trademark rights, so your brand
  remains your lever (§4.3).
- **Compatible with open-core monetisation** (goal 4): you can build proprietary
  premium content and hosted services *around* an Apache core without conflict.

**Alternative considered — AGPL-3.0.** Strong protection against a hosted
competitor free-riding your engine (they'd have to open their modifications).
But AGPL deters enterprise adoption and *other* OSS projects, complicates your
*own* SaaS (you'd dual-license via the CLA), and adds compliance fear. Given you
already plan to monetise via **premium content + hosted convenience + trademark**
rather than by restricting the engine, Apache-2.0's adoption upside outweighs
AGPL's defensive upside. **If protecting against cloud-vendor strip-mining is a
top priority**, the dual-license route (AGPL core + commercial license via CLA)
is viable — but it taxes adoption, so I do not recommend it as the default.

**Not recommended:** MIT (no patent grant), BSL/SSPL/Elastic (not OSI-approved —
violates goal 1's "fully open source").

### 4.2 Premium content — proprietary EULA, physically separated

- Premium/enterprise packs and certification content live **only** in the private
  `lab-premium` repo under a proprietary license (`LicenseRef-Lab-Commercial`).
- They are **never** copied into the OSS tree, so there is no license
  contamination and the OSS story stays clean.
- They are distributed via an **authenticated OCI registry**; access = a license
  token. The OSS engine runs them through the *same* public interfaces — no fork,
  no special engine build (goal: "premium must not require forking the engine").

### 4.3 Trademark ownership

- **Register the product name + logo** as trademarks, owned by you or a holding
  entity (an LLC is cleanest once revenue appears). This is your durable moat:
  Apache-2.0 lets anyone use the *code*, but only you can use the *brand*.
- Publish a **Trademark Usage Policy** (model on Kubernetes/Linux Foundation):
  forks may use the code, must not imply official endorsement, must rename if
  they diverge. This is what stops a competitor shipping "Sagars-Laboratory
  Cloud" off your code.

### 4.4 Protection against future licensing conflicts

- **CLA** (§3.3) keeps relicensing optional and lawful.
- **`pkg/` SDK boundary** with a stability policy prevents accidental API lock-in.
- **No GPL/AGPL dependencies** in the Apache core (a CI license-scan gate) — so
  you never get "infected" into copyleft.
- **Per-pack `license` field** + an index policy: the catalog records and
  displays each pack's license, and rejects license-incompatible community packs.
- **`NOTICE` + SPDX headers** on every source file for clean provenance.

---

## 5. Marketplace architecture

### 5.1 Community packs

- Authored in any public repo (or contributed to `packs/community/` in core).
- Listed in the public `registry` index via a PR adding a signed entry.
- Free; installed with `labctl pack add <name>`; no auth.
- A lightweight review bar (schema-valid, lints clean, CI green, code-of-conduct).

### 5.2 Enterprise packs

- Authored in `lab-premium` (or a customer's private repo).
- Distributed via an **authenticated OCI registry**; the customer's license token
  is the access credential.
- The engine resolves them through the same install path; the only difference is
  an auth header and a signature/entitlement check.

### 5.3 Third-party packs

- Independent publishers host their own OCI/git packs and submit an index entry.
- A **verified-publisher** program (signing key on file, namespace reserved,
  basic review) yields a "verified" badge — trust without you hosting their bits.
- Optional future **revenue share** for paid third-party packs sold through a
  hosted marketplace.

### 5.4 Registry / catalog design

- **Phase 1 (now-ish): static signed index.** One file per pack/version in the
  `registry` repo, served via Pages/CDN. `labctl pack search/info/add` read it.
  Zero infra, fully auditable, PR-based moderation.
- **Phase 2 (deferred): hosted catalog API.** Search, ratings, download counts,
  verified badges, private/entitled listings, telemetry-opt-in popularity. A
  superset of the static index; the CLI prefers it if configured, else falls back
  to the static index. This is also where a paid marketplace and billing attach.

---

## 6. Future commercial architecture (editions)

One engine, gated by **content + entitlement + hosting** — never by separate
forks. Each edition is the same OSS core plus progressively more packs/services.

| Edition | What it is | How it's delivered |
|---|---|---|
| **Community** | Full OSS engine + CLI + community packs | Apache-2.0, free, self-hosted |
| **Professional** | Community + premium pack bundles (advanced incidents, certification prep), priority support | License token unlocks premium OCI registry; same engine |
| **Enterprise** | Pro + SSO/RBAC (task 062), air-gapped registry mirror, compliance/audit packs, team mode (task 063), private pack hosting, SLAs | License token + the optional closed entitlement/identity service plugged into the OSS interfaces |
| **SaaS / Hosted** | Managed control-plane: browser labs, multi-tenant clusters, billing, no local setup | A hosting layer **around** the same engine + a control-plane in `lab-premium`; the engine itself is unchanged |

The critical invariant: **moving up an edition adds content/services; it never
swaps the engine.** That is what makes the project "monetisable without a
redesign" (goal 4) and keeps community and premium cleanly separable (goal 5).

---

## 7. Technical implementation roadmap

### 7.1 Immediate (do now — cheap, high-leverage, mostly decision-independent)

These are the M7 tasks (§ task files 065–072). They are the foundation that
everything else assumes; they get *more* expensive the longer the project grows
without them.

1. **OSS governance baseline** (065): LICENSE, NOTICE, CONTRIBUTING, CODE_OF_CONDUCT,
   GOVERNANCE, SECURITY, MAINTAINERS, CODEOWNERS, PR/issue templates, CLA/DCO bot,
   SPDX headers, license-scan CI gate.
2. **Public SDK boundary** (066): move schema types (scenario/check/pack/results)
   to a versioned `pkg/`; document a stability policy; explicit
   `apiVersion: scenario.lab.dev/v2`.
3. **Formal pack format** (067): `pack.yaml` manifest + validation + a `labctl
   pack` command group (evolve `scenario install/packs`); back-compatible with
   today's git packs.
4. **Entitlement & extension interface** (070): a default **no-op** entitlement
   interface + content-source resolver hooks in the engine, so premium/hosted can
   plug in later **without a fork**. Shipping this early is what prevents future
   lock-in.
5. **Module path & brand alignment** (071): pick a **vanity import path** and
   product name now so they never have to change again (decision 3).

### 7.2 Soon (next — enables the ecosystem)

6. **OCI pack distribution + signing** (068).
7. **Registry index + discovery** (069): static signed index + `pack search/info`.
8. **Contributor experience** (072): `pack init` / `scenario new` scaffolds,
   authoring guide, docs site, good-first-issues.

### 7.3 Deferred (real demand required — do not build speculatively)

- Hosted catalog API + verified-publisher program (M8: 073).
- Premium pack repo + entitlement service (M8: 074, private).
- Marketplace UX in the web UI (M8: 075).
- Edition packaging / feature-flagging (M9: 076).
- SaaS/hosted control-plane spike (M9: 077).
- Certification/training framework (M9: 078).

### 7.4 Avoid now (future lock-in traps)

- ❌ **Compiled Go plugins** for scenarios. They break portability and force
  engine recompiles. Stay declarative-data + sandboxed-script; add WASM later if
  richer logic is ever needed.
- ❌ **Hardcoding premium logic** into the engine. All gating goes through the
  entitlement interface + distribution auth.
- ❌ **A non-OSI license for the core** (BSL/SSPL/Elastic) — violates goal 1 and
  poisons adoption.
- ❌ **Premature multi-repo split** of engine/CLI — fragments CI and contributors.
- ❌ **Tying pack identity to GitHub** specifically — use registry-neutral
  `publisher/name@version` so OCI/self-hosting works.
- ❌ **A schema that can't carry signing/entitlement/provenance metadata** — bake
  those fields into `pack.yaml` v1 even if unused at first.
- ❌ **Vendor-locking the registry** — keep the index a portable static artifact.

---

## 8. Proposed repository structure (core monorepo)

```
snowops/flightdeck   (public, Apache-2.0)
├── LICENSE  NOTICE  CODE_OF_CONDUCT.md  CONTRIBUTING.md  GOVERNANCE.md
├── SECURITY.md  MAINTAINERS.md  RELEASING.md  TRADEMARKS.md
├── .github/
│   ├── CODEOWNERS                    # engine/CLI/SDK locked to lead maintainer
│   ├── workflows/                    # ci, release, license-scan, pack-validate
│   ├── ISSUE_TEMPLATE/  PULL_REQUEST_TEMPLATE.md
│   └── cla.yml                       # CLA Assistant config (or DCO)
│
├── cmd/
│   └── labctl/                       # CLI entrypoint (internal wiring stays here)
│       └── internal/                 # NON-public: command glue, executors, api
│
├── pkg/                              # ★ PUBLIC SDK — semver-stable, third-party facing
│   ├── scenario/                     # scenario schema + loader (apiVersion v2)
│   ├── checks/                       # check types + runner contract
│   ├── pack/                         # pack.yaml schema + validate + resolve
│   ├── entitlement/                  # no-op-by-default entitlement interface
│   ├── extension/                    # content-source resolver + hook interfaces
│   └── results/                      # results/score schema
│
├── engine/                          # build/deploy strategy dispatch (existing)
├── runtimes/                        # k3d | aks | eks (existing)
├── platform/                        # swappable platform modules (existing)
│   └── <category>/<provider>/       # install/uninstall/status/values + _interface.yaml
│
├── packs/                           # ★ NEW: first-party content as installable packs
│   ├── community/                   # the built-in, Apache-2.0 community packs
│   │   ├── observability-sre/  ...  # (today's scenarios, repackaged)
│   │   └── stack-expansion/         # mesh/data/secrets/autoscaling scenarios
│   └── examples/                    # reference packs for authors
│
├── scenarios/                       # (transitional) in-tree scenarios → migrate into packs/
├── incidents/  learn/  challenges/  # fault library, learning paths, challenges
├── services/  apps/  foundation/    # shared deps, demo apps, IaC (existing)
│
├── sdk/                             # ★ authoring toolkit
│   ├── pack-template/               # `labctl pack init` scaffold
│   ├── scenario-template/           # `labctl scenario new` scaffold
│   └── schemas/                     # JSON Schemas for scenario.yaml & pack.yaml (editor validation)
│
├── ui/                              # web UI (existing)
├── docs/
│   ├── strategy/                    # THIS doc + manual-actions
│   ├── rfcs/                        # ★ architecture RFCs (maintainer-approved)
│   ├── authoring/                   # ★ "write a scenario pack" guide (SDK docs)
│   ├── runbooks/  architecture.md  scenarios.md  cli-reference.md  ...
│   └── SIMULATOR.md  ROADMAP.md
├── make/  versions.env  .ai/        # build includes, pinned versions, task state
└── go.work / go.mod                 # workspace tying cmd + pkg modules

snowops/registry   (public)            # marketplace catalog
├── index/<publisher>/<pack>/<version>.yaml   # signed pack entries
├── publishers/<publisher>.yaml               # verified-publisher metadata
├── schema/                                    # index entry JSON Schema
└── .github/workflows/validate-index.yml       # PR gate for new entries

snowops/flightdeck-premium   (PRIVATE, proprietary)
├── packs/<enterprise-packs>/        # premium/enterprise content
├── certification/                   # cert tracks & exams
├── entitlement-service/             # license-key issuance/verification (impl of pkg/entitlement)
└── control-plane/                   # SaaS/hosted (impl around the OSS engine)
```

---

## 9. Risk assessment

### 9.1 Architectural risks

| Risk | Mitigation |
|---|---|
| Schema churn breaks community packs | Explicit `apiVersion`, N/N-1 support window, deprecation policy, JSON Schemas in CI |
| Accidental public-API lock-in | Hard `pkg/` vs `internal/` boundary + written stability policy |
| Premium logic leaks into the engine | Entitlement interface (no-op default) + license-scan + CODEOWNERS on `pkg/entitlement` |
| Scripts in packs are an RCE surface | Sandboxing posture, signed/verified packs, explicit "packs run code on your cluster" consent (already noted in `catalog.go`) |
| Over-engineering the marketplace before demand | Static index first; hosted API only on real traction |

### 9.2 Open-source governance risks

| Risk | Mitigation |
|---|---|
| You lose architectural control as contributors grow | CODEOWNERS + branch protection + RFC process = mechanical authority |
| "Personal project" perception limits adoption | Neutral org/brand (decision 3), GOVERNANCE.md, public roadmap |
| Maintainer bus-factor / burnout | Invite area maintainers early; document RELEASING; automate CI/release |
| Contributor disputes / conduct issues | CODE_OF_CONDUCT + enforcement path |

### 9.3 Monetisation risks

| Risk | Mitigation |
|---|---|
| Cloud vendor strip-mines the engine | Trademark moat + premium content + hosted convenience; AGPL is the nuclear option if ever needed (CLA keeps it available) |
| Can't relicense later | **CLA now** — this is the cheapest insurance you can buy |
| Community feels "bait-and-switched" by premium tier | Draw the CE/premium line **publicly and early** (this doc); never move existing free features behind a paywall |
| Premium and OSS get entangled | Physical repo separation + interface-only coupling |

### 9.4 Contributor-experience risks

| Risk | Mitigation |
|---|---|
| High barrier to a first contribution | `pack init`/`scenario new` scaffolds, authoring guide, good-first-issues, JSON-Schema editor validation |
| Confusing "where does this go?" | Monorepo + a clear `docs/authoring` + CODEOWNERS routing |
| Slow/opaque reviews discourage PRs | SLA expectations in CONTRIBUTING, CI that gives fast actionable feedback |
| CLA friction scares casual contributors | CLA Assistant one-click; DCO fallback documented if friction proves real |

---

## 10. Final recommendation — the "Flight Simulator for Platform Engineers"

Build it the way flight simulators won their ecosystems:

- **A rock-solid, open core engine** (the simulator) that *never* gets crippled
  to sell upgrades. Its stability and openness are the reason people build on it.
- **An open content format** (scenarios = aircraft/missions) with a real SDK, so
  the community produces far more content than you ever could — the actual engine
  of growth (goal 2, 7).
- **A neutral brand + marketplace** (the add-on store) where community packs are
  free and discoverable, and verified/enterprise packs are a click away (goal 6).
- **Monetise the edges, not the center:** convenience (hosted), curation
  (premium/enterprise packs), and assurance (certification + support) — all of
  which *increase* with a thriving commons rather than fighting it (goal 4, 5).
- **Keep authority where it belongs:** you own the engine, the schema, the brand,
  and the releases (goal 3); the community owns an ever-growing library of
  content.

Do the five immediate foundation tasks now (licensing, SDK boundary, pack format,
entitlement interface, brand/module path) — they are cheap today and ruinously
expensive to retrofit later. Defer the marketplace API, premium service, and SaaS
until real demand appears. Avoid the lock-in traps in §7.4 religiously.

### Where I challenge the brief

1. **"May eventually have commercial offerings" → decide the CE/premium line
   *now*, publicly.** The expensive mistake isn't charging money; it's drawing the
   line late and appearing to claw back free features. This doc draws it.
2. **Don't over-index on the marketplace yet.** A static signed index gets you
   90% of the ecosystem value at ~1% of the cost. Build the hosted marketplace
   only when pack supply/demand justifies it.
3. **The biggest risk isn't architecture — it's the missing LICENSE/CLA today.**
   Every contribution merged without them is a compounding legal liability. This
   is the one thing that cannot wait.
4. **Consider a neutral brand sooner than feels comfortable.** "Sagars-Laboratory"
   is a great personal/homelab name but a community/commercial project benefits
   from a neutral, brandable, trademark-able identity. Keep the personal repo as
   the origin story; launch the ecosystem under a neutral org.
5. **Resist compiled-plugin temptation.** It will be requested ("can my pack run
   custom Go?"). Say no; offer declarative checks + sandboxed scripts + (later)
   WASM. Native plugins are a portability and security tar pit.

---

## Appendix A — proposed milestones & tasks

See `.ai/state.json` (Part III) and `.ai/tasks/065-078`. Summary:

- **M7 — OSS & Ecosystem Foundation (P0):** 065 governance baseline · 066 public
  SDK boundary · 067 pack format · 068 OCI distribution + signing · 069 registry
  index + discovery · 070 entitlement/extension interface · 071 module/brand
  alignment · 072 contributor experience.
- **M8 — Marketplace (P2, deferred):** 073 hosted catalog API + verified
  publishers · 074 premium pack repo + entitlement service (private) · 075
  marketplace UI.
- **M9 — Commercial & Hosted (P2, deferred):** 076 edition packaging · 077 SaaS
  control-plane spike · 078 certification/training framework.

Manual, maintainer-only actions are in
[`MAINTAINER-MANUAL-ACTIONS.md`](MAINTAINER-MANUAL-ACTIONS.md).
