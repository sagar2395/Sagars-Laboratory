# AGENTS.md — Lightweight Workflow for AI Tools

> This file is read automatically by most AI coding tools (Claude Code, Codex,
> Cursor, Windsurf, Cline, …). It defines a **simple, tool-agnostic** workflow.
> For project context and rules, read `CLAUDE.md` first.

The old version of this file defined an elaborate 5-role agent protocol
(Architect / Feature / Bug / DevOps / Reviewer) with worktrees and inter-agent
file messaging. **That has been removed** in favour of the lighter loop below.
Any AI tool — or you — can run it.

---

## The loop

```
1. PICK    a task from .ai/state.json "todo" (respect phase order in docs/ROADMAP.md)
2. CLAIM   move its id from "todo" -> "in_progress" in .ai/state.json
3. READ    the task file .ai/tasks/NNN-*.md (scope, files, acceptance criteria)
4. BUILD   implement only what the task asks; follow CLAUDE.md golden rules
5. VERIFY  run the tests named in the task; do the manual runbook in docs/runbooks/
6. RECORD  update the runbook if behaviour changed; move id -> "done"
7. COMMIT  conventional message (feat:/fix:/ci:/docs:/chore:)
```

One task per change. Keep diffs small and reviewable.

## Task files

- Location: `.ai/tasks/NNN-kebab-title.md`
- Template: `.ai/task-template.md`
- Every task states: **scope**, **files to touch**, **acceptance criteria**,
  **how to test**, **dependencies**.
- Found a new problem? Don't expand the current task — write a new task file and
  add its id to `.ai/state.json` `todo`.

## State

`.ai/state.json` is the single source of truth for progress:

```json
{
  "todo":        ["007-fix-echo-server-readiness-probe"],
  "in_progress": [],
  "done":        ["001-fix-platform-category-resolution-and-status"],
  "blocked":     []
}
```

`docs/ROADMAP.md` groups these task ids into phases and defines exit criteria.

## Running tasks in parallel (optional)

If you want multiple AI sessions working at once, give each its own git worktree
so they don't collide:

```bash
git worktree add ../lab-phase1 -b work/phase-1
git worktree add ../lab-ui     -b work/ui
```

Then point one tool at each worktree. Merge via normal PRs. No special protocol —
just keep tasks independent (the phase grouping in ROADMAP is designed for that).

## Roles are now just hints, not gates

You don't need separate "agents." A task file may carry a `Type` hint
(feature / bug / infra / review) so you know what kind of change it is, but any
session may implement any task. The only hard rules are the golden rules in
`CLAUDE.md`.

## Definition of done

- [ ] Acceptance criteria in the task file are met.
- [ ] Tests named in the task pass (`go test ./...` where relevant).
- [ ] The change is portable (macOS + Linux) per CLAUDE.md rule 1.
- [ ] The matching `docs/runbooks/` file exists and is accurate.
- [ ] `.ai/state.json` updated and a conventional commit made.
