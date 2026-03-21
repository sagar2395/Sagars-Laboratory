# AI Engineering Coordination Hub

## Purpose

This directory is the inter-agent communication layer. All coordination between agents happens through files here.

## Directory Structure

```
.ai/
├── README.md           ← This file
├── state.json          ← Single source of truth for task status
├── task-template.md    ← Template for creating new tasks
├── tasks/              ← Task definitions (Architect creates, others consume)
├── specs/              ← Architecture specs and decisions (Architect only)
├── reviews/            ← Review reports and clarification requests
└── logs/
    └── activity.log    ← Append-only log of all agent activity
```

## Agents

| Agent | Writes To | Reads From |
|-------|-----------|------------|
| **Architect** | `tasks/`, `specs/`, `state.json`, `logs/` | Everything (read-only) |
| **Feature** | Source code, `state.json`, `logs/` | `tasks/` |
| **Bug** | Source code, `state.json`, `logs/`, `reviews/` | Errors, `tasks/`, `logs/` |
| **DevOps** | Infra code, `state.json`, `logs/` | `tasks/` |
| **Reviewer** | `reviews/`, `logs/` | Everything (read-only) |

## Workflow

1. **Architect** creates tasks in `tasks/` and updates `state.json`
2. **Feature/DevOps/Bug** picks a task, implements it, updates `state.json`
3. **Reviewer** reviews the work, writes report in `reviews/`
4. **Human** approves and merges

## File Naming

- Tasks: `{NNN}-{kebab-case-title}.md`
- Specs: `{YYYY-MM-DD}-{kebab-case-title}.md`
- Reviews: `{NNN}-review.md` or `{NNN}-clarification.md`

## Rules

- All agents log to `logs/activity.log` (append-only)
- `state.json` is the single source of truth
- Agents never cross role boundaries
- Tasks must be atomic and independently implementable