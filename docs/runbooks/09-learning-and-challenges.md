# Runbook 09 — Learning Paths & Challenges

Covers milestone **M3** (tasks 050–053).

> Tasks 051–053 (challenge mode, score persistence, UI views) are
> still in progress. This runbook covers task 050: the learning path
> engine and `labctl learn` commands.

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

## Cleanup

```bash
rm -f .labctl/learn/kubernetes-foundations.json
```

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
