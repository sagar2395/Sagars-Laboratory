# Runbook 13 — Marketplace & Credentials

> Covers tasks 073–078 (M8 Marketplace + M9 Commercial/Hosted engine-side features).
> Last updated: 2026-06-16

---

## Pack Marketplace (tasks 073, 075)

### CLI: search and install packs

```bash
# Browse the community registry
labctl pack search

# Filter by keyword
labctl pack search kafka

# Install by registry name
labctl pack add kafka-drills

# Install by OCI ref
labctl pack add oci://ghcr.io/snowops/kafka-drills:1.4.2

# List installed packs
labctl pack list

# Remove a pack
labctl pack remove kafka-drills
```

### CLI: hosted catalog (when LABCTL_CATALOG_URL is set)

When a hosted catalog is available, `labctl pack search` prefers it (download
counts, ratings, verified-publisher badges). The static registry index remains
the fallback for offline use.

```bash
export LABCTL_CATALOG_URL=https://catalog.flightdeck.dev
labctl pack search
```

### Web UI: Marketplace tab

The **Marketplace** tab in the labctl UI surfaces the same pack catalog:

1. Open the UI: `labctl ui`
2. Click the **Marketplace** tab.
3. Search by keyword or browse all packs.
4. Click a pack card to see details (publisher, version, tier, rating).
5. Click **Install** to install the pack (async — watch the log panel).
6. Switch to **Installed** sub-tab to see installed packs and remove them.

---

## Edition Info (task 076)

### CLI: check current edition

```bash
labctl edition
```

Output example (CE):
```
Edition: Community Edition (CE)
         Community Edition (CE): full open-source engine + community packs (Apache-2.0)

Feature              Available
──────               ─────────
auth-rbac            yes
challenge-mode       yes
...
premium-packs        no  (premium/enterprise content)
```

### Switching editions

Editions are assembled at deployment time, not at runtime. To use a premium
edition:

1. Obtain a license key from the Flightdeck team.
2. Set `LABCTL_EDITION=professional` (or `enterprise`) in the environment.
3. Set `LABCTL_TOKEN=<license-key>`.
4. Restart the labctl server.

The engine binary is identical for all editions — only entitled content and
services differ.

### Web UI: edition info

The `/api/edition` endpoint returns the active edition and feature line. The UI
uses this to show edition-aware messaging in future views.

---

## Credentials (task 078)

### CLI: issuing credentials

```bash
# Issue a credential for completing a certification track
labctl credential issue \
  --achievement kubernetes-foundations \
  --subject alice \
  --score 88

# Issue with signing key (verifiable)
export LABCTL_SIGNING_KEY=$(openssl rand -hex 32)
labctl credential issue \
  --achievement sre-on-call \
  --subject bob \
  --score 92 \
  --output json > credential.json

# List issued credentials
labctl credential list

# Verify a credential file
labctl credential verify .labctl/credentials/<id>.json
```

### Signing key

Set `LABCTL_SIGNING_KEY` (hex-encoded 32-byte key) to sign issued credentials.
Without it, credentials are unsigned — still valid within the same lab instance
but not independently verifiable.

```bash
# Generate a signing key
openssl rand -hex 32
# Save it securely; the same key is required to verify
```

### Web UI: credentials

```
GET /api/credentials           # list all credentials
POST /api/credentials/issue    # issue a new credential
GET /api/credentials/{id}      # get a specific credential
POST /api/credentials/{id}/verify  # verify a credential's signature
```

### Credential schema

Credentials are JSON files stored in `.labctl/credentials/<id>.json`:

```json
{
  "schema": "credential.flightdeck.dev/v1",
  "id": "a1b2c3d4...",
  "subject": "alice",
  "achievement": "kubernetes-foundations",
  "grade": "merit",
  "score": 88,
  "issuedAt": "2026-06-16T12:00:00Z",
  "issuer": "Flightdeck",
  "evidence": [],
  "signature": "base64-hmac-sha256-of-payload"
}
```

Grades: `pass` (≥ 60), `merit` (≥ 75), `distinction` (≥ 90).

---

## SaaS / Hosted (task 077)

The SaaS control-plane spike is a design artefact, not production code. Read the
RFC at `docs/strategy/RFC-0002-saas-control-plane.md`. No manual steps required
until the private `flightdeck-saas` repo is created (a future maintainer action).

---

## Troubleshooting

### Pack search returns nothing

Check `PACK_REGISTRY_INDEX` is set (defaults to the snowops GitHub Pages URL):

```bash
labctl pack search   # should hit the static index
# If empty: check network, or set PACK_REGISTRY_INDEX=file://path/to/index.json
```

### Credential verification fails with "certificate is unsigned"

Either:
- The credential was issued without `LABCTL_SIGNING_KEY`, or
- `LABCTL_SIGNING_KEY` is set but you're providing a key.

Verify consistently: both issue and verify must use the same key (or both must
omit it).

### Edition shows "community" but I set LABCTL_EDITION

Make sure the env var is exported and the server was restarted after the change:

```bash
export LABCTL_EDITION=professional
labctl edition   # should show Professional
```
