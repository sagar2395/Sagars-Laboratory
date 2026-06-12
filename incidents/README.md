# Incidents — Fault Library

Realistic, **reversible** production faults for game days, on-call practice,
and (later) graded challenges. Inject one, watch the lab break, practice
diagnosing it, fix it — and let the detection check confirm you actually
fixed it. See `docs/runbooks/08-incident-engine.md` for the full workflow.

## The contract

Every fault is a directory `incidents/<name>/` with:

| File | Purpose |
|------|---------|
| `fault.yaml` | Metadata + the **detection check** (passes ⇔ the fault is RESOLVED) |
| `inject.sh` | Breaks the lab. Idempotent: re-running while injected is a no-op. |
| `resolve.sh` | The escape hatch. Always restores the lab, even after a partial manual fix. |
| `hints.md` | Progressive hints, one `## Hint N` section each (revealed by `labctl incident hint`) |
| `solution.md` | Full diagnosis + fix walkthrough (spoiler — `labctl incident solution`) |
| `manifests/`, `checks/` | Optional supporting files |

### fault.yaml

```yaml
name: crashloop-bad-config        # must match the directory name
displayName: "CrashLoop: broken container command"
description: "What the victim experiences, not how it's injected"
category: workload                # workload | network | resources | storage | config
severity: medium                  # low | medium | high
target:
  namespace: go-api
  workload: go-api
prerequisites:
  apps: [go-api]                  # gated before injection
detection:                        # same schema as scenario checks
  name: rollout-healthy           # PASSES when the fault is RESOLVED
  type: script                    # http | kubectl | promql | script
  script: checks/resolved.sh
  timeoutSeconds: 30
```

## Rules for fault authors

1. **Target only the demo apps** (`go-api`, `echo-server`) or a dedicated
   fault namespace — never platform components or `kube-system`.
2. **Record what you change.** Annotate the touched resource
   (`labfault-<name>=...`, original values in `labfault-<name>-original-*`)
   so `resolve.sh` can always undo it without guessing.
3. **`resolve.sh` must never fail the user.** It runs after any amount of
   manual fixing; every step tolerates "already fixed" (`--ignore-not-found`,
   guards, `|| true` where safe).
4. **Portable shell** (CLAUDE.md golden rule 1) — these scripts are
   shellcheck'd and portability-linted in CI.
5. **Write the detection check as "what must be true when healthy"** — it
   doubles as the resolution detector and the challenge grader.
6. Hints go from gentle nudge to near-answer; the last hint may name the
   resource, the solution names the command.

## Current faults

| Fault | Category | Severity | What breaks |
|-------|----------|----------|-------------|
| `crashloop-bad-config` | workload | medium | go-api's container command is replaced with one that exits immediately — new pods crash-loop |
| `bad-deploy-rollout` | workload | medium | go-api is "deployed" with a nonexistent image tag — rollout sticks in ImagePullBackOff |
| `oom-kill` | resources | high | echo-server's memory limit is slashed — pods are OOMKilled at startup |
| `network-blackhole` | network | high | a deny-all-ingress NetworkPolicy lands in go-api's namespace — the service goes dark through the ingress |
| `service-selector-broken` | config | medium | go-api's Service selector stops matching its pods — endpoints empty, pods perfectly healthy (sneaky) |
| `noisy-neighbor` | resources | low | a CPU-burning deployment lands on the cluster with big requests and no limits |

Note: the original plan (task 045) listed `dns-blackhole` and `pvc-full`;
they were swapped for `network-blackhole` and `service-selector-broken`,
whose detection works reliably on a default k3d cluster (DNS exec probes
and PVC behavior vary too much with the local storage/CNI setup).

## Using faults

```bash
labctl incident list
labctl incident inject service-selector-broken     # or --random [--seed N] [--silent]
labctl incident status                             # runs the detection check
labctl incident hint                               # next hint (recorded)
labctl incident resolve                            # escape hatch
```
