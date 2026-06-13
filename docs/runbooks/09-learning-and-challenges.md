# Runbook 09 — Learning Paths & Challenges

Covers milestone **M3** (tasks 050–053). All tasks complete.

## Prereqs

- A healthy lab (runbook 01) — not all modules require it, but the
  `kubernetes-foundations` path does.
- CLI built: `make cli-build`.

## Steps

### 1. List available paths

```bash
bin/labctl learn list
```

**Expected:** `kubernetes-foundations` appears with `0/4 modules` and a
description. If the path does not appear, run
`go test ./cmd/labctl/internal/learn/...` — a broken `path.yaml` is
the most likely cause.

### 2. Start a path

```bash
bin/labctl learn start kubernetes-foundations
```

**Expected:** "Started: Kubernetes Foundations (4 modules). Run `labctl
learn next kubernetes-foundations` to begin."

### 3. Show the first module

```bash
bin/labctl learn next kubernetes-foundations --show-only
```

**Expected:** Module 1's intro (from `intros/01-init-cluster.md`) is
printed, followed by the action (`labctl runtime up`). No check is run
with `--show-only`.

### 4. Verify module completion

Run the module's action, then verify:

```bash
bin/labctl runtime up                          # or skip if cluster is already up
bin/labctl learn next kubernetes-foundations   # runs the check
```

**Expected:** "✓ Check passed. Module complete! Next: module 2 — Deploy
go-api."

Repeat for modules 2–4.

### 5. Completion

After all four modules pass:

```bash
bin/labctl learn next kubernetes-foundations
```

**Expected:** "All 4 modules complete! Path 'kubernetes-foundations'
finished."

### 6. Progress display

```bash
bin/labctl learn progress kubernetes-foundations
```

**Expected:** Shows each module with a [✓] or [ ] marker and a count
like "Kubernetes Foundations  4/4 modules".

### 7. Progress survives restarts

```bash
# Kill the CLI mid-path and re-run
bin/labctl learn progress kubernetes-foundations
```

**Expected:** Progress is exactly where you left it — it's persisted in
`.labctl/learn/kubernetes-foundations.json`.

### 8. Validation gate (what CI runs)

```bash
cd cmd/labctl && go test ./internal/learn/... && cd ../..
```

**Expected:** All green. `TestRepo_AllPathsValid` loads every `path.yaml`
in `learn/` and fails CI if any is invalid.

Break a path (e.g. remove `displayName` from a module) and re-run: CI
fails naming the field. Revert afterwards.

## REST API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/learn/paths` | GET | List all paths with progress |
| `/api/learn/paths/{name}` | GET | Path details + module list |
| `/api/learn/paths/{name}/start` | POST | Start (or restart) a path |
| `/api/learn/paths/{name}/progress` | GET | Current progress |
| `/api/learn/paths/{name}/complete` | POST | Mark module complete (body: `{"moduleIdx":N}`) |

---

## Part B — Challenge Mode (task 051)

### 9. List challenges

```bash
bin/labctl challenge list
```

**Expected:** three challenges: `restore-broken-deploy`, `find-the-memory-leak`,
`make-the-slo-green`.

### 10. Full challenge loop

```bash
bin/labctl challenge start restore-broken-deploy
```

**Expected:** `bad-deploy-rollout` injected, timer started, problem description
printed. Pods go into ImagePullBackOff.

Diagnose and fix:
```bash
kubectl get pods -n go-api   # ImagePullBackOff
bin/labctl incident hint     # use a hint if stuck (deducts score)
# Fix: update the deployment's image tag to a valid one
kubectl -n go-api set image deployment/go-api go-api=go-api:latest
bin/labctl challenge submit  # runs detection check + prints score
```

**Expected:** "✓ rollout-healthy … Checks: 1/1 passed … Score: N/100"

### 11. Abort escape hatch

```bash
bin/labctl challenge start find-the-memory-leak
bin/labctl challenge status    # shows elapsed time
bin/labctl challenge abort     # undo: resolves oom-kill, records "aborted"
```

**Expected:** `incident resolve` runs, score is 0, outcome is `aborted`.

### 12. History

```bash
bin/labctl challenge history
```

**Expected:** rows for completed and aborted runs with time, score, hints, outcome.

### 13. Validation gate

```bash
cd cmd/labctl && go test ./internal/challenge/... && cd ../..
```

**Expected:** all green. Break a `challenge.yaml` (e.g. remove `grading`)
and re-run: CI fails naming the field.

## Cleanup

```bash
bin/labctl challenge abort 2>/dev/null || true
bin/labctl incident resolve 2>/dev/null || true
rm -f .labctl/learn/kubernetes-foundations.json
rm -f .labctl/challenges/active.json
```

---

## Part C — UI Views (task 053)

The web UI (`labctl ui`) now includes three new tabs alongside Dashboard,
Scenarios, Platform, and Apps.

### 14. Learn tab

Open `http://localhost:8080` → click **Learn**.

**Expected:**
- A card lists all paths (e.g. `kubernetes-foundations`) with a progress
  bar showing `0%` (or current completion if already started).
- Click **Start** on a not-yet-started path — the server creates the
  progress record and the button disappears, replaced by the progress bar.
- Click **Details** to open a modal showing all modules with `✓`, `▶`
  (next), `Pending`, or `Locked` badges.

### 15. Challenges tab

Open the **Challenges** tab.

**Expected:**
- Challenge cards list each challenge with category, par time, and best
  score if previously attempted.
- Click **Start** — a confirmation dialog warns about the scoring
  penalty for hints. Confirm to start the timer.
- A live timer banner appears at the top while a challenge is active.
- **Hint (–10pts)** button is available in the active challenge's row;
  confirming shows the hint text in a notification and deducts 10 points.
- **Submit** (in the active banner) grades the challenge and shows the
  score in a notification.
- **Abort** (in the active banner) records score 0 / outcome `aborted`.
- The **Run History** table below shows all past runs with time, score,
  and outcome.

### 16. Results tab

Open the **Results** tab.

**Expected:**
- Summary stat cards show total runs, incident count, challenge count,
  module count, and average challenge score.
- The run history table lists all records (newest first) with type, name,
  date, duration, score, and outcome badges.
- The **All types** dropdown filters the table by kind
  (`Incident`, `Challenge`, `Learn Module`).

## Troubleshooting

- **Path not in `learn list`** — check `path.yaml` validity with
  `go test ./cmd/labctl/internal/learn/...`. The YAML error will name the
  field.
- **Check not passing** — the check is run with the same environment as
  the CLI. For `http` checks, make sure `/etc/hosts` has the ingress entry
  (runbook 00). For `script` checks, run the script directly to see
  the raw error.
- **Progress lost** — `.labctl/` is gitignored but local-only. If you
  cloned fresh or deleted it, re-run `labctl learn start <path>`.
