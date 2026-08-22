# RFC-0002: SaaS / Hosted Control-Plane Architecture

> **Status:** Accepted — spike artefact, no production code yet.
> **Milestone:** M9 — Commercial & Hosted (task 077)
> **Author:** Maintainer team
> **Date:** 2026-06-16

---

## 1 Problem statement

The OSS engine (`labctl`) runs on a user's laptop or a self-hosted VM. To deliver
browser-based labs at scale — with multi-tenancy, pay-per-use billing, and no
local install — we need a hosted control plane that wraps the same OSS engine
without forking it.

**Core invariant from the strategy doc (§6):** The OSS engine binary is identical
for all editions. The control plane is additive, never a source fork.

---

## 2 Architecture overview

```
Browser (learner)
     │ HTTPS/WS
     ▼
┌─────────────────────────────────────────────────┐
│  CONTROL PLANE  (private repo; Kubernetes)      │
│                                                 │
│  ┌──────────┐  ┌───────────┐  ┌────────────┐   │
│  │ Auth /   │  │  Billing  │  │  Catalog / │   │
│  │ SSO (01) │  │  API (02) │  │  Entitle. │   │
│  └──────────┘  └───────────┘  └────────────┘   │
│          │            │              │          │
│          └────────────┼──────────────┘          │
│                       │                         │
│  ┌────────────────────▼────────────────────┐    │
│  │        Lab Lifecycle Manager (03)       │    │
│  │  provision → ready → expire → teardown  │    │
│  └──────────────────┬──────────────────────┘    │
└─────────────────────┼───────────────────────────┘
                      │ runs
              ┌───────▼────────┐
              │  OSS ENGINE    │  (same binary — go.flightdeck.dev/labctl)
              │  labctl serve  │
              │                │
              │  + pkg/entitlement injected: TokenEntitlement
              │  + pkg/extension injected: premium Resolver
              └───────┬────────┘
                      │
              ┌───────▼────────┐
              │  Tenant        │
              │  Kubernetes    │  (one per isolation unit; see §4)
              │  Cluster /     │
              │  Namespace     │
              └────────────────┘
```

### Key design decisions

| Decision | Choice | Rationale |
|---|---|---|
| Engine | unmodified OSS binary | validates §6 invariant |
| Auth | OIDC (GitHub / Google) + JWT | integrates with existing `pkg/entitlement.TokenVerifier` seam |
| Tenancy | namespace (default) / vcluster (isolated) | see §4 |
| Lab lifecycle | ephemeral: 4-hour TTL, extendable | cost control |
| State | PVC per tenant for `.labctl/` history | lab-session persistence |
| Registry | private OCI (GAR/ECR) | premium pack distribution |

---

## 3 Control-plane components

### 3.1 Auth / SSO service
- OIDC proxy — accepts GitHub/Google tokens, issues a signed `LABCTL_TOKEN`
  JWT that the OSS `pkg/entitlement.TokenEntitlement` can verify.
- Claims include `sub` (user), `tiers` (entitled tiers), `exp`.
- The OSS engine never imports this service; it only sees the JWT via env.

### 3.2 Billing API
- Wraps Stripe or Paddle.
- On subscription, issues a long-lived license token; on cancellation, token
  expires and the engine's entitlement check starts denying premium packs.
- No billing logic enters the OSS engine.

### 3.3 Catalog / Entitlement service
- Hosts the premium-pack OCI registry (private GHCR or GAR).
- The license JWT acts as the OCI pull credential (OAuth2 device flow).
- Premium packs are pulled through `pkg/extension.OCIResolver`; the OSS engine
  never holds the registry password itself.

### 3.4 Lab Lifecycle Manager
- Watches a `Lab` CRD; provisioner reconciles.
- `Lab` states: `Pending → Provisioning → Ready → Expired → Terminating`.
- Expiry triggers snapshot (optional) → teardown → PVC deletion.
- User budget cap: maximum concurrent labs per subscription tier.

---

## 4 Tenancy model

Three isolation levels, chosen per subscription tier:

| Tier | Isolation | How |
|---|---|---|
| CE / trial | Shared namespace | `labctl` runs as a Deployment in `labs-shared`; namespaces prefixed `lab-<user>` |
| Pro | vcluster | One vcluster per user session; stronger isolation, slower provisioning (~30 s) |
| Enterprise | Dedicated cluster | Separate node pool or project; provisioned via Terraform; highest isolation |

**Default (CE / trial):** namespace isolation is sufficient for a lab that owns
a k3d-style workload. Tenant namespaces are created per lab, garbage-collected
at expiry.

### Pod security

All shared-namespace lab pods run with `runAsNonRoot`, `readOnlyRootFilesystem`,
and restricted PodSecurityPolicy. No cluster-admin is granted to lab pods.

---

## 5 Lab lifecycle

```
User clicks "Start Lab"
   │
   ▼
Control plane issues Lab CRD with:
  - userID, subscriptionTier
  - scenarioRef (pack name + scenario name)
  - ttl (default: 4 h)
   │
   ▼
Lifecycle Manager provisions:
  - Namespace (or vcluster) for the session
  - labctl Deployment with:
      - LABCTL_TOKEN from Auth service
      - LABCTL_EDITION=professional (or enterprise)
      - PVC mounted at /home/labctl/.labctl
   │
   ▼
Lab → Ready (webhook notifies browser; WebSocket URL returned)
   │
   ▼
User works in browser IDE / terminal
   │
   ▼
TTL expires OR user clicks "End Lab"
   │
   ▼
Results saved to PVC → control plane exports to user's history
Namespace / vcluster deleted → cost stops immediately
```

---

## 6 Validating the §6 invariant

> "The OSS engine binary is identical for all editions; the control plane
> is additive, never a source fork."

Checklist:
- [x] `pkg/entitlement.TokenEntitlement` already exists in OSS; injected at
      construction time via env `LABCTL_TOKEN`.
- [x] `pkg/extension.OCIResolver` already exists in OSS; premium-pack OCI
      registry injected via env or flags — no source change.
- [x] `pkg/edition.Current()` reads `LABCTL_EDITION` env — no source change.
- [x] Auth middleware in `internal/auth` already supports JWT sessions via the
      `TokenVerifier` seam — no source change needed.
- [x] The control plane wraps the OSS binary as a Kubernetes Deployment; it
      never recompiles or forks it.

**Conclusion:** The architecture requires zero OSS engine changes. All commercial
and hosted features are assembled by the control plane at deployment time.

---

## 7 Cost model (indicative)

| Isolation | Spin-up | $/session-hour | Notes |
|---|---|---|---|
| Shared namespace | < 5 s | ~$0.02 | GKE Autopilot spot |
| vcluster | ~30 s | ~$0.08 | shared node pool |
| Dedicated cluster | ~5 min | ~$0.40 | Autopilot cluster |

At 1 000 concurrent trial users (namespace isolation) the fleet is ~20 shared
cluster nodes — well within a single GKE Autopilot cluster.

---

## 8 What must happen next (manual actions)

These are maintainer manual actions, not engine code:

| Action | When |
|---|---|
| Stand up the private `flightdeck-saas` repo | Before any production code |
| Design the `Lab` CRD spec | Before lifecycle manager |
| Choose Stripe vs Paddle | Before billing API |
| Register OIDC application (GitHub + Google) | Before auth service |
| Negotiate GKE pricing and run a cost model | Before provisioning |
| Legal: ToS, Privacy Policy, DPA | Before public launch |

See `docs/strategy/MAINTAINER-MANUAL-ACTIONS.md` for the canonical list.

---

## 9 Out of scope for this RFC

- The actual production control plane code (private repo).
- A specific Kubernetes distribution for the hosted fleet.
- Multi-region / HA design (follow-on RFC).
- Certification exam proctoring (RFC-0003, not yet written).
