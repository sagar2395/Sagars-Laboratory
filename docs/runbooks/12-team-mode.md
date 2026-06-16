# Runbook 12 — Team Mode: Auth & RBAC

**Feature:** optional authentication + per-user roles for the labctl API/UI
**Commands:** `labctl users add|list|remove`, `LABCTL_AUTH=true labctl ui`
**Task:** 062

---

## What this provides

The labctl server (`labctl ui`) is single-user and unauthenticated by default —
perfect for a laptop. Team mode makes it safe to share a running lab with others:

- **Optional auth**, off unless `LABCTL_AUTH=true`. When off, behaviour is
  byte-identical to today (no login, no checks).
- **Two roles:**
  - `operator` — full control (platform, runtime, lab, apps, services, …).
  - `participant` — run challenges/incidents/learn and read status, but **cannot**
    uninstall platform, switch runtime, reset the lab, or build/deploy apps.
- **Identity in results:** scored runs (challenges, incidents, learn modules) are
  attributed to the authenticated user in `.labctl/history/results.jsonl`.

> Passwords are stored as PBKDF2-HMAC-SHA256 hashes in `.labctl/users.yaml`
> (mode 0600), never in plain text. OIDC/SSO is intentionally out of scope for
> v1 — see "Future work" below.

---

## 1. Create users

`labctl users` edits `.labctl/users.yaml`. It works whether or not auth is
currently enabled — it only edits the file; the server enforces it.

```bash
# An operator (full control)
labctl users add alice --role operator --password 's3cret-op'

# A participant (run labs, read status)
labctl users add bob --role participant --password 's3cret-part'

labctl users list
# NAME    ROLE
# alice   operator
# bob     participant
```

Password input order of precedence: `--password` flag → `LABCTL_PASSWORD` env
var → interactive stdin prompt. For scripted setup prefer the env var:

```bash
LABCTL_PASSWORD='s3cret' labctl users add carol --role participant
```

Remove a user:

```bash
labctl users remove bob
```

---

## 2. Start the server with auth enabled

```bash
LABCTL_AUTH=true labctl ui --port 3939
```

On start the server logs how many users loaded. If auth is on but no users
exist, it warns (nobody could log in) — add a user first.

---

## 3. Two-browser manual test (operator + participant)

**Browser A — operator (alice):**

1. Open `http://localhost:3939`. You should see a login screen.
2. Log in as `alice`. The header shows `alice (operator)` and a Log out button.
3. Go to Platform → install a component. It works.
4. Go to Challenges → start and complete a challenge. It works.

**Browser B — participant (bob):**

1. Open the same URL in a private/incognito window. Log in as `bob`.
2. Go to Challenges → start and complete a challenge. **It works.**
3. Try a platform install or **Lab → Reset**. The server returns **403**
   ("this action requires the operator role") and the UI surfaces the error.

**Verify attribution:**

```bash
labctl results            # CLI view, or GET /api/results
# The challenge bob completed should show "user": "bob";
# alice's run should show "user": "alice".
```

---

## 4. API access from the CLI / curl

The server accepts the session via cookie (browser) or
`Authorization: Bearer <token>` (scripts):

```bash
# Log in, capture the token from the Set-Cookie or response
TOKEN=$(curl -s -X POST http://localhost:3939/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"s3cret-op"}' -c - | awk '/labctl_session/ {print $7}')

curl -s http://localhost:3939/api/status -H "Authorization: Bearer $TOKEN"
```

---

## Role policy reference

| Endpoint group (mutating) | operator | participant |
|---|---|---|
| `/api/platform/*` (up/down/component) | ✅ | ❌ 403 |
| `/api/runtimes/*` (activate/deactivate) | ✅ | ❌ 403 |
| `/api/lab/*` (snapshots, reset, restore) | ✅ | ❌ 403 |
| `/api/apps/*` (build/deploy/destroy) | ✅ | ❌ 403 |
| `/api/services/*` (up/down) | ✅ | ❌ 403 |
| `/api/scenarios/*`, `/api/challenges/*`, `/api/incidents/*`, `/api/learn/*` | ✅ | ✅ |
| All `GET` reads | ✅ | ✅ |

---

## Troubleshooting

| Symptom | Fix |
|---|---|
| Login screen never appears | Auth is off. Start with `LABCTL_AUTH=true labctl ui`. |
| "auth enabled but no users defined" warning | `labctl users add <name> --role operator`. |
| Login always fails | Wrong password, or the users file was edited by hand with a bad hash. Re-add the user. |
| Participant gets 403 they shouldn't | Check the role: `labctl users list`. Operator routes are listed above. |
| Everyone logged out after restart | Sessions are in-memory by design; log in again. |
| Results show OS username, not the login | That run was performed via the CLI (or with auth off). API runs under auth carry the authenticated user. |

---

## Security notes & future work

- Sessions are in-memory (cleared on restart) and tokens are random 256-bit
  values. The cookie is `HttpOnly` + `SameSite=Strict`.
- Serve behind TLS (a reverse proxy) for any non-localhost deployment — the
  session cookie is not marked `Secure` because the lab commonly runs on plain
  `http://localhost`.
- **Future work:** OIDC/SSO, `Secure` cookie toggle, per-team leaderboards
  (builds on task 063), and password rotation policy. Not in v1.
