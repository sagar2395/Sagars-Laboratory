# Maintainer Manual Actions

> Things **only you (@sagar2395)** can do — they need a human, an account, money,
> or a legal decision. The AI/automation can prepare files and configs, but
> cannot click these buttons, sign these things, or spend on your behalf.
>
> Ordered by urgency. ⚠️ = do before accepting external contributions.

---

## A. Decisions (RESOLVED 2026-06-14)

The 4 pivotal choices from `OSS-COMMERCIAL-STRATEGY.md §0`:

- [x] **A1. Core license → Apache-2.0.**
- [x] **A2. Contribution agreement → CLA** (CLA Assistant).
- [x] **A3. Brand & GitHub home → `snowops/flightdeck`.** Product name
  **Flightdeck**; vanity import path `go.flightdeck.dev/...`.
- [x] **A4. Repo strategy → monorepo core + `snowops/registry` + private
  `snowops/flightdeck-premium`.**

### A0. Product name chosen → **Flightdeck** (verify before registering)

- [ ] **Trademark knock-out search for "Flightdeck"** before filing/registering.
  Note: "Flightdeck" has prior uses in software — confirm clearance in your
  classes/jurisdictions (consider counsel). If it doesn't clear, the only rework
  is re-running task 071 with a different name; nothing else depends on it.
- [ ] **Register `flightdeck.dev` + `go.flightdeck.dev`** (and `.io` if desired).
- [ ] **Host the vanity import page.** Task 071 applied the module path
  `go.flightdeck.dev/labctl`; deploy [`docs/vanity/labctl/index.html`](../vanity/labctl/index.html)
  at `https://go.flightdeck.dev/labctl` (GitHub Pages custom domain or any static
  host) and update its repo URL to the canonical git home. Full `go get`
  resolution also needs the org transfer (C1); internal builds are unaffected
  (built from source).

---

## B. Legal & brand (⚠️ before external contributions)

- [ ] **B1. Add the LICENSE** the AI prepares for your chosen license (A1). Only
  you should make the licensing commit / decision of record.
- [ ] **B2. Stand up the CLA** (if A2 = CLA): install the **CLA Assistant** GitHub
  App (or SAP/EasyCLA), host the CLA text, set the bot to gate PRs. (AI prepares
  the CLA text + `.github/cla.yml`; you install the app and authorise it.)
- [ ] **B3. Trademark.** Decide the name (A3), then **search** for conflicts and
  **register** the wordmark (+ logo) in your jurisdiction(s). Consider a lawyer
  for the filing. Owner = you now, a holding **LLC/company** once revenue starts.
- [ ] **B4. Publish a Trademark Usage Policy** (AI drafts `TRADEMARKS.md`; you
  approve and, ideally, have counsel review before launch).
- [ ] **B5. Domain(s).** Register the product domain(s) for docs site, vanity
  import path, and future SaaS (e.g. `<brand>.dev`, `go.<brand>.dev`).

---

## C. GitHub / org administration (you have the admin rights, AI does not)

- [ ] **C1. Create the GitHub org** (if A3 = neutral org) and transfer/mirror this
  repo, or rename it.
- [ ] **C2. Branch protection on `main`:** require PRs, required CODEOWNERS review,
  required status checks (CI, license-scan, pack-validate, CLA), no force-push,
  linear history. (AI provides the exact settings; you toggle them — they need
  repo-admin.)
- [ ] **C3. Create teams**: `@scenario-maintainers`, `@platform-maintainers`,
  `@docs-maintainers` referenced by CODEOWNERS; invite initial members.
- [ ] **C4. Create the `registry` public repo** and the **`lab-premium` private
  repo** (§1.1). Set their visibility and access.
- [ ] **C5. Enable GitHub Pages** (or a CDN) to serve the registry index.
- [ ] **C6. Release signing:** create a signing key (GPG/cosign) and store it as
  an Actions secret; you hold the private key. Configure `GITHUB_TOKEN`/OIDC for
  publishing OCI packs.
- [ ] **C7. Secrets** for any registries you publish to (GHCR is free; ECR/ACR if
  used) and, later, the catalog API.

---

## D. Foundational accounts (only when you reach that phase)

- [ ] **D1. OCI registry** for packs — GHCR is free and recommended to start.
- [ ] **D2. (Deferred) Hosted catalog API + DB** — only when pack supply/demand
  justifies it (M8).
- [ ] **D3. (Deferred) Entitlement / license-key service** + a billing provider
  (Stripe/Paddle) for premium tiers (M8/M9).
- [ ] **D4. (Deferred) SaaS infra** (managed clusters, multi-tenancy, auth/SSO
  provider) for the hosted edition (M9).
- [ ] **D5. (Deferred) Certification platform** (exam delivery, credentialing) for
  training content (M9).

---

## E. Community launch (when the foundation is in place)

- [ ] **E1. CODE_OF_CONDUCT enforcement contact** — an email/alias you monitor.
- [ ] **E2. Security disclosure contact** for `SECURITY.md` (a monitored alias;
  consider GitHub Private Vulnerability Reporting — you enable it).
- [ ] **E3. Seed "good first issues"** and announce the contribution path.
- [ ] **E4. Decide & publish the CE vs premium feature line** (the doc proposes
  one; you ratify it publicly so there's no "bait-and-switch" perception).
- [ ] **E5. Pick the company/entity** that will own trademark + commercial
  contracts before the first paid offering (LLC or equivalent).

---

## What the AI/automation will do for you (no action needed from you)

- Draft all governance files (CONTRIBUTING, GOVERNANCE, CODE_OF_CONDUCT,
  SECURITY, MAINTAINERS, CODEOWNERS, PR/issue templates, TRADEMARKS draft).
- Add the LICENSE/NOTICE + SPDX headers + a license-scan CI gate.
- Build the `pkg/` SDK boundary, `pack.yaml` format + validation, the entitlement
  interface (no-op default), OCI distribution, the static registry index + schema,
  and the `pack init`/`scenario new` scaffolds.
- Prepare exact branch-protection and CLA-bot settings for you to apply.
- Keep `.ai/state.json`, the roadmap, and docs in sync.

> Rule of thumb: **if it needs a credit card, a signature, a legal judgement, or
> GitHub-admin/org rights, it's on this list. Everything else, the automation
> can do.**
