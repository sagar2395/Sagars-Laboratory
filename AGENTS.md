# Multi-Agent Engineering System

## Overview

This project operates with a multi-agent workflow where each agent has a distinct role. Agents coordinate through files in the `.ai/` directory and never cross role boundaries.

This system is **tool-agnostic** — it works with any AI coding tool (Claude Code, GPT, Codex, Cursor, Windsurf, Cline, GitHub Copilot, or any other). See `agents/README.md` for tool-specific setup instructions.

## Agents

| Agent | Role | Instruction File | Permissions |
|-------|------|-------------------|-------------|
| **Architect** | Design, planning, task creation | `agents/architect.md` | Read-only — never writes code |
| **Feature** | Feature implementation | `agents/feature.md` | Read-write — implements code |
| **Bug** | Bug fixing, stability | `agents/bug.md` | Read-write — minimal fixes only |
| **DevOps** | Infra, CI/CD, Helm, Terraform | `agents/devops.md` | Read-write — infra code only |
| **Reviewer** | Code review, quality, security | `agents/reviewer.md` | Read-only — writes reviews only |

Each agent file in `agents/` is a **self-contained system prompt** with full project context, role definition, workflow, and rules. Load the appropriate file into your AI tool to activate that agent.

## Coordination Protocol

### Task Lifecycle

```
Architect creates task → .ai/tasks/{id}-{title}.md
     ↓
Feature/DevOps/Bug picks up task → marks in-progress in .ai/state.json
     ↓
Agent implements → commits with conventional message
     ↓
Agent logs work → .ai/logs/activity.log
     ↓
Reviewer reviews → .ai/reviews/{id}-review.md
     ↓
Human approves → task marked completed in .ai/state.json
```

### File Naming Conventions

- **Tasks**: `.ai/tasks/{NNN}-{kebab-case-title}.md`
- **Specs**: `.ai/specs/{YYYY-MM-DD}-{kebab-case-title}.md`
- **Reviews**: `.ai/reviews/{NNN}-review.md`
- **Logs**: `.ai/logs/activity.log` (append-only)

### State Tracking

`.ai/state.json` tracks all task states:
```json
{
  "current_tasks": ["001-task-name"],
  "completed_tasks": ["000-initial-task"],
  "blocked_tasks": []
}
```

### Commit Message Format

| Agent | Prefix | Example |
|-------|--------|---------|
| Architect | `docs:` / `chore:` | `docs: add API gateway spec` |
| Feature | `feat:` | `feat: implement health dashboard` |
| Bug | `fix:` | `fix: resolve nil pointer in status check` |
| DevOps | `ci:` / `chore:` | `ci: add helm validation workflow` |
| Reviewer | `docs:` | `docs: add review for task 003` |

### Log Entry Format

```
[AGENT] [YYYY-MM-DD]
Action: {what was done}
Task: {task ID if applicable}
Files: {files created/modified}
Notes: {any relevant context}
```

## Setup

### Option 1: Worktrees (Recommended for Parallel Agents)

Each agent operates in its own git worktree with its own branch:

```bash
# Create worktrees from the main repo root
git worktree add worktrees/architect -b agent/architect
git worktree add worktrees/feature -b agent/feature
git worktree add worktrees/bug -b agent/bug
git worktree add worktrees/devops -b agent/devops
git worktree add worktrees/reviewer -b agent/reviewer
```

Then load the agent instructions into your AI tool. See `agents/README.md` for tool-specific commands (Claude Code, Codex, Cursor, etc.).

### Option 2: Single Directory

Use one project directory and switch agent files as needed. Suitable when running one agent at a time.

### Option 3: Mixed Tools

You can use different AI tools for different agents — e.g., Claude Code for Feature, Codex for DevOps, Cursor for Architect. Each tool just needs to load the corresponding agent file from `agents/`.

## Rules

1. **Role isolation**: Agents only perform their assigned responsibilities
2. **Communication via files**: All inter-agent communication goes through `.ai/`
3. **No cross-boundary edits**: Architect never writes code; Reviewer never modifies source
4. **Atomic tasks**: Each task is small, independent, and completable in one session
5. **Log everything**: Every action is recorded in `.ai/logs/activity.log`
6. **State consistency**: `.ai/state.json` is the single source of truth for task status
