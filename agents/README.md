# Agent Setup Guide

This directory contains **tool-agnostic** agent instruction files for the multi-agent engineering system. Each file is a complete, self-contained system prompt that works with **any AI coding tool**.

## Agents

| File | Role | Permissions |
|------|------|-------------|
| `architect.md` | Architecture design, task creation, specs | **Read-only** — never writes code |
| `feature.md` | Feature implementation from tasks | **Read-write** — implements code |
| `bug.md` | Bug fixing, stability, minimal patches | **Read-write** — fixes only |
| `devops.md` | CI/CD, Helm, Terraform, infra | **Read-write** — infra code only |
| `reviewer.md` | Code review, security, quality reports | **Read-only** — writes reviews only |

## Quick Start

1. Pick the agent role you need
2. Follow the setup instructions for your AI tool below
3. Open the project (or a worktree) and start working

---

## Setup by Tool

### Claude Code

Copy the agent file as `CLAUDE.md` at the project root (or worktree root):

```bash
cp agents/architect.md CLAUDE.md
```

Claude Code automatically reads `CLAUDE.md` as its system prompt. To switch agents, replace the file:

```bash
cp agents/feature.md CLAUDE.md
```

**With worktrees** (recommended for parallel agents):
```bash
git worktree add worktrees/architect -b agent/architect
cp agents/architect.md worktrees/architect/CLAUDE.md
cd worktrees/architect && claude
```

---

### OpenAI Codex (CLI)

Pass the agent file as the system prompt:

```bash
codex --instructions agents/architect.md
```

Or set per-worktree:
```bash
cp agents/feature.md worktrees/feature/AGENTS.md
cd worktrees/feature && codex
```

---

### Cursor

**Option A** — Copy into `.cursorrules` at the project/worktree root:
```bash
cp agents/architect.md .cursorrules
```

**Option B** — In Cursor Settings → Rules, paste the contents of the agent file.

**Option C** — Reference in `.cursor/rules`:
```
Read and follow the instructions in agents/architect.md as your system prompt.
```

---

### Windsurf (Cascade)

Copy the agent file into `.windsurfrules` at the project/worktree root:
```bash
cp agents/architect.md .windsurfrules
```

Or paste into Windsurf → Settings → AI Rules.

---

### Cline / Roo Code

In Cline settings, paste the contents of the agent file as "Custom Instructions" or "System Prompt".

For role-specific configs, use Cline's custom modes and paste each agent file into the corresponding mode.

---

### ChatGPT / GPT (Web or API)

Paste the agent file contents into:
- **Custom Instructions** (web) → "How would you like ChatGPT to respond?"
- **System message** (API) → Use file contents as the system prompt

---

### GitHub Copilot (VS Code)

Option A — Create chatmodes from agent files:
```bash
mkdir -p .github/chatmodes
# Add YAML frontmatter and copy content
echo '---
description: "Architect agent"
---' > .github/chatmodes/architect.chatmode.md
cat agents/architect.md >> .github/chatmodes/architect.chatmode.md
```

Option B — Reference in `.github/copilot-instructions.md`.

---

### Any Other Tool

Every agent file is a standalone markdown document. Use it as:
- System prompt
- Custom instructions
- Project-level configuration file
- Pasted into a chat session

Whatever your tool supports for providing persistent instructions.

---

## Worktree Strategy (Parallel Agents)

For running multiple agents simultaneously, use git worktrees so each agent has its own working directory and branch:

```bash
# Create worktrees from the main repo root
git worktree add worktrees/architect -b agent/architect
git worktree add worktrees/feature -b agent/feature
git worktree add worktrees/bug -b agent/bug
git worktree add worktrees/devops -b agent/devops
git worktree add worktrees/reviewer -b agent/reviewer
```

Then set up each worktree with the agent prompt for your chosen tool:

```bash
# Example: Claude Code
for agent in architect feature bug devops reviewer; do
  cp agents/${agent}.md worktrees/${agent}/CLAUDE.md
done

# Example: Cursor
for agent in architect feature bug devops reviewer; do
  cp agents/${agent}.md worktrees/${agent}/.cursorrules
done
```

Open each worktree in a separate terminal/editor window and start the AI tool.

**You can mix tools** — e.g., Claude Code for Feature, Codex for DevOps, Cursor for Architect.

---

## Coordination

All agents coordinate through the `.ai/` directory:
- `.ai/tasks/` — Task definitions (Architect creates, others consume)
- `.ai/specs/` — Architecture decisions (Architect only)
- `.ai/reviews/` — Review reports and clarification requests
- `.ai/logs/activity.log` — Activity audit trail
- `.ai/state.json` — Task state tracking (single source of truth)

See `AGENTS.md` at the project root for the full coordination protocol.

---

## Starter Prompts

Use these prompts to kick off each agent after setup:

### Architect
```
Read agents/architect.md for your role and AGENTS.md for the coordination protocol.
Check .ai/state.json for current state. Analyze the project and create a prioritized
backlog of tasks in .ai/tasks/. Each task must follow .ai/task-template.md, be assigned
to the right agent, and be independently implementable. Update state.json and log activity.
```

### Feature
```
Read agents/feature.md for your role. Check .ai/tasks/ for tasks assigned to "Feature".
Pick the highest priority unblocked task, implement it following project conventions,
verify it compiles/runs, then update .ai/state.json and log in .ai/logs/activity.log.
Commit with "feat: <description>".
```

### Bug
```
Read agents/bug.md for your role. Scan for bugs — check build errors, lint issues,
test failures, runtime problems. Also check .ai/tasks/ for Bug-assigned tasks.
Apply minimal targeted fixes. Log all fixes in .ai/logs/activity.log.
Commit with "fix: <description>".
```

### DevOps
```
Read agents/devops.md for your role. Check .ai/tasks/ for DevOps-assigned tasks.
Review CI/CD pipelines, Helm charts, Terraform modules, and runtime scripts.
Validate with helm lint, terraform validate, shellcheck.
Log work in .ai/logs/activity.log.
```

### Reviewer
```
Read agents/reviewer.md for your role. Review all recent changes across the codebase.
Check for security issues, convention violations, missing error handling.
Write review reports in .ai/reviews/ following the format in your agent file.
Log reviews in .ai/logs/activity.log.
```
