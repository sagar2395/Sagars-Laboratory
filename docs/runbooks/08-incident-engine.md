# Runbook 08 — Incident Engine: Faults, Game Days, Hints & MTTR

Covers milestone **M2** (tasks 045–048; the on-call drill — 049 — will
extend this runbook when it ships).

## Prereqs

- A healthy lab (runbook 01) with **go-api and echo-server deployed** and
  ingress + monitoring up — faults target the demo apps.
- `/etc/hosts` entries for `*.k3d.local` (runbook 00).
- CLI built: `make cli-build`.

## Steps

### 1. Browse the fault library

```bash
bin/labctl incident list
```

**Expected:** six faults with category, severity, and target
(`crashloop-bad-config`, `bad-deploy-rollout`, `oom-kill`,
`network-blackhole`, `service-selector-broken`, `noisy-neighbor`).

### 2. The full loop on a known fault

```bash
bin/labctl incident inject service-selector-broken
curl -s -o /dev/null -w '%{http_code}\n' http://go-api.k3d.local/health   # 503/404 — broken
bin/labctl incident status                                                # NOT RESOLVED
```

Diagnose it like production (don't peek at the solution):

```bash
kubectl get pods -n go-api          # healthy — suspicious
kubectl get endpoints go-api -n go-api   # <none> — there's the lead
```

Fix it, then prove it:

```bash
kubectl -n go-api patch svc go-api -p '{"spec":{"selector":{"app.kubernetes.io/name":"go-api"}}}'
bin/labctl incident status
```

**Expected:** `RESOLVED — detection check "service-reachable" passes`, and
the incident state clears (a second `incident status` says no incident is
active).

### 3. Game-day mode: random and silent

```bash
bin/labctl incident inject --random --silent --seed 42
bin/labctl incident status     # shows "(hidden — silent mode)" while unresolved
```

**Expected:** the fault name is not revealed; you diagnose from symptoms.
The same `--seed` picks the same fault on another machine — useful for
running the same surprise across a team. Resolution (or `incident
resolve`) reveals what it was.

### 4. The escape hatch

```bash
bin/labctl incident inject oom-kill
bin/labctl incident resolve
bin/labctl incident status     # no incident active
kubectl get pods -n echo-server   # recovering / healthy
```

**Expected:** `resolve` always restores the lab, even if you half-fixed
things manually first. Every fault's `resolve.sh` tolerates prior fixes.

### 5. Injection guards

```bash
bin/labctl incident inject crashloop-bad-config
bin/labctl incident inject oom-kill          # refused: incident already active
bin/labctl incident inject oom-kill --force  # allowed
bin/labctl incident resolve                  # cleans up oom-kill
bin/labctl incident resolve crashloop-bad-config   # explicit-name resolve for the first one
```

### 6. Hints when you're stuck

With an incident active:

```bash
bin/labctl incident hint     # gentle nudge
bin/labctl incident hint     # more specific
bin/labctl incident hint     # near-answer
bin/labctl incident hint     # "no more hints"
bin/labctl incident solution # full walkthrough (asks for confirmation)
```

**Expected:** hints reveal in order, survive CLI restarts, and the count is
recorded on the run. `solution` prints the fault's complete diagnosis +
fix.

### 7. MTTR history

After resolving a couple of incidents (one manually, one via `resolve`):

```bash
bin/labctl incident history
```

**Expected:** one row per run with TIME-TO-CHECK (injection → your first
`status` call), MTTR (injection → resolution), HINTS used, and RESOLVED BY
(`manual` when the detection check passed, `auto` when you used the escape
hatch). Records live in `.labctl/history/incidents.jsonl` and survive
restarts.

### 8. Make it visible (recommended)

Run traffic during a fault and watch Grafana:

```bash
bin/labctl traffic start --profile steady --rps 20
bin/labctl incident inject noisy-neighbor
# node CPU climbs; go-api p99 latency rises in Grafana
bin/labctl incident resolve && bin/labctl traffic stop
```

### 9. Validation gate (what CI runs)

```bash
cd cmd/labctl && go test ./internal/incident/ && cd ../..
```

**Expected:** all green. This validates every fault's contract files,
schema, and that inject/resolve/detection scripts execute cleanly against
a stubbed kubectl. Break a `fault.yaml` (e.g. `severity: apocalyptic`) and
re-run: the repo test fails naming the field. Revert afterwards.

## Cleanup

```bash
bin/labctl incident resolve 2>/dev/null || true
bin/labctl traffic stop
```

## Troubleshooting

- **`inject` fails with prerequisite errors** — deploy the demo apps first
  (`labctl app deploy go-api`, `labctl app deploy echo-server`).
- **Detection check errors with DNS failures** — http-based detections
  reach the app via the ingress hostname; check `/etc/hosts` (runbook 00).
- **You lost track of what's broken** — `labctl incident resolve <name>`
  works even without active state; worst case run `resolve` for each
  fault name, all resolve scripts are no-ops when already healthy.
