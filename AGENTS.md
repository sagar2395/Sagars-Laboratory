# AGENTS.md — Lightweight Workflow for AI Tools

> This file is read automatically by most AI coding tools (Claude Code, Codex,
> Cursor, Windsurf, Cline, …). It defines a **simple, tool-agnostic** workflow.
> For project context and the golden rules, read `CLAUDE.md` first.

---

## The loop

Work ships in **waves** (W0–W8). Each wave is an independently mergeable
increment that leaves `main` working.

```
1. PICK    the task named by "next" in .ai/state.json
2. CLAIM   set its status to "in_progress"
3. READ    its wave's goal and exit criteria in docs/ROADMAP.md
4. BUILD   implement only what the task asks; follow the CLAUDE.md golden rules
5. TEST    every applicable layer (see below) — not just the one that's easy
6. VERIFY  update or write the wave's runbook so a human can check it by hand
7. RECORD  set status "done", advance "next", keep the wave's status in sync
8. COMMIT  feat(W1/T06): <what changed>
```

One task per change. Keep diffs small and reviewable.

**Do not start a new wave** until the previous wave's exit criteria all hold and
its runbooks have been signed off by the maintainer.

## State

`.ai/state.json` is the single source of truth. Shape:

```json
{
  "currentWave": "W1",
  "next": "W1-T03",
  "waves": {
    "W1": {
      "title": "...", "goal": "...", "status": "in_progress",
      "exitCriteria": ["..."],
      "tasks": { "W1-T03": { "title": "...", "status": "todo" } }
    }
  }
}
```

`docs/ROADMAP.md` holds each wave's goal, task list and exit criteria. The v1
state file and its 79 task descriptions are archived under `.ai/archive/`.

Found a new problem mid-task? **Don't expand the current task.** Add a task to
the wave in `.ai/state.json` and `docs/ROADMAP.md`, and carry on.

## Definition of done

A task is done only when every applicable box is ticked. There are no
exceptions, and "the wave is running late" is not one.

- [ ] The code implements the task and nothing beyond it.
- [ ] **Go tests** — table-driven where cases share a shape; hermetic
      (`t.TempDir()`, `toolchain.Fake`, never a live cluster or network); happy
      path, ≥2 error/edge cases per exported function, and cancellation +
      deadline cases wherever a `context.Context` is accepted.
- [ ] **`make test-go` passes**, including the race detector and the
      per-package coverage gate (≥80%; `.coverage-exceptions` is a ratchet, not
      an escape hatch — read its header before touching it).
- [ ] **Shell tests** — any script with branching logic has bats coverage using
      `test/shell/helpers/stub.bash`, asserting idempotency and failure
      propagation. `make test-shell` passes.
- [ ] **Contract tests** — any new endpoint or CLI command is tested including
      its error envelope.
- [ ] **UI tests** — any new component has Vitest coverage; any new journey has
      a Playwright test. `make test-ui` and `make test-e2e` pass.
- [ ] **`make lint` passes** — gofmt, golangci-lint, gosec, govulncheck,
      shellcheck, shfmt, the portability gate, TypeScript strict.
- [ ] **Portable** (macOS + Linux, no cgo) per golden rule 1.
- [ ] **Docs updated in the same commit** — reference docs, plus
      `docs/PRODUCT.md` or `docs/architecture/` if behaviour changed, plus an
      **ADR** in `docs/adr/` if you made a decision a future maintainer would
      otherwise have to reverse-engineer.
- [ ] **Runbook** written or updated in `docs/runbooks/`.
- [ ] **`.ai/state.json` updated** and a conventional commit made.

## Running work in parallel (optional)

Give each session its own git worktree so they don't collide:

```bash
git worktree add ../flightdeck-w2 -b work/w2
```

Tasks within a wave are usually independent; tasks across waves are not — the
wave order encodes real dependencies.

## Roles are hints, not gates

There is no multi-agent protocol. Any session may implement any task. The only
hard rules are the golden rules in `CLAUDE.md` and the Definition of Done above.
