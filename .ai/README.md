# .ai/ — Task Backlog & State

This directory is the lightweight coordination layer for AI-assisted work. The
old elaborate multi-agent protocol has been removed (see `AGENTS.md`). What
remains is simple and tool-agnostic.

```
.ai/
├── README.md         ← this file
├── state.json        ← single source of truth for task status (todo/in_progress/done/blocked)
├── task-template.md  ← template for new tasks
├── tasks/            ← one file per task: NNN-kebab-title.md
└── logs/             ← optional scratch logs (gitignored)
```

## How to use it

1. Open `docs/ROADMAP.md` to see phases and which task ids belong to each.
2. Pick the next unblocked task id from `state.json` `todo` (respect phase order).
3. Move it to `in_progress`, read `tasks/NNN-*.md`, implement, test.
4. Move it to `done`; update the matching `docs/runbooks/` file if behaviour changed.

See `AGENTS.md` for the full loop and the definition of done.

## Adding a task

Copy `task-template.md` to `tasks/NNN-kebab-title.md` (next free number), fill it
in, and add the id to `state.json` `todo`. Each task must be small, independent,
and completable in one session.

## Authoritative mapping

`docs/ROADMAP.md` is the authority for which phase a task belongs to and what the
exit criteria are. `state.json` is the authority for current status. Keep them in
sync.
