# Contributing to Flightdeck

Thanks for your interest in **Flightdeck** — the flight simulator for platform
engineers. This project thrives on community-contributed **scenarios, platform
modules, documentation, and code**. This guide explains how to contribute and
what to expect.

> New here? The highest-leverage place to start is **content** — a new scenario
> or incident. Start with
> [Your First Scenario](docs/authoring/first-scenario.md) and look for issues
> labelled **good first issue**.

## Code of Conduct

By participating you agree to our [Code of Conduct](CODE_OF_CONDUCT.md).

## Contributor License Agreement (CLA)

Flightdeck requires every contributor to sign our **CLA** before their first
contribution can be merged. This keeps the project relicensable and protects both
you and the project. The **CLA Assistant** bot will prompt you automatically on
your first pull request — it's a one-time, one-click step. See [CLA.md](CLA.md).

## What you can contribute

| Type | Where it lives | Review owner |
|---|---|---|
| **Scenarios** | `scenarios/` | scenario maintainers |
| **Platform modules** | `platform/<category>/<provider>/` | platform maintainers |
| **Incidents / learning / challenges** | `incidents/`, `learn/`, `challenges/` | scenario maintainers |
| **Documentation** | `docs/` | docs maintainers |
| **Engine / CLI / SDK** | `cmd/`, `internal/`, `pkg/`, `engine/` | **lead maintainer** (see GOVERNANCE.md) |

Changes to the engine, the public SDK (`pkg/`), or the scenario schema require an
**RFC** first — a short markdown PR under `docs/rfcs/` that a maintainer approves
before implementation. This keeps architectural direction coherent.

## Development setup

```bash
# Build the CLI, UI embedded (no committed binary)
make cli-build

# Run the gates. All four test layers are mandatory — see docs/TESTING.md.
make test          # Go unit + shell (bats), race detector, coverage gate
make test-ui       # vitest component tests
make test-e2e      # playwright journeys
make lint          # gofmt, golangci-lint, gosec, govulncheck, shellcheck,
                   # shfmt, the portability gate, and TypeScript strict

# Bring up a local cluster + the platform to test changes end to end
bin/labctl lab up && make platform-up
```

New to the repo? Work through
[R00 — Environment & Build](docs/runbooks/R00-environment-and-build.md) once; it
verifies your setup and shows you each gate biting.

See [CLAUDE.md](CLAUDE.md) and [AGENTS.md](AGENTS.md) for the project's working
agreement and the golden rules (cross-platform, CLI-wraps-scripts, declarative
content, every feature ships with a runbook, docs + state updated in the same PR).

## The golden rules (please read before a code/content PR)

1. **Cross-platform.** Must run on macOS (Apple Silicon + Intel) and modern
   Linux. No GNU-only shell flags (`grep -P`, `sed -i` without a backup suffix,
   `readlink -f`, `date -d`), and **no cgo** — it breaks cross-compilation.
   `make lint-shell` enforces this automatically.
2. **Go orchestrates, scripts do the work.** Don't move shell/helm/kubectl logic
   into Go.
3. **Declarative content.** Scenarios, faults, learning paths and checks are
   YAML + scripts — never hardcoded in Go.
4. **Idempotent everything.** `helm upgrade --install`, `kubectl apply`, safe
   re-activation. Interrupting an operation and re-running it must converge.
5. **Every operation is cancellable and durable.** Anything that shells out goes
   through `internal/run` with a context, a timeout and a lock key. Never call
   `exec.Command(...).Run()` directly.
6. **Tests at every applicable layer.** Go unit (table-driven, hermetic, ≥80%,
   cancellation covered), bats for scripts with logic, contract tests for
   endpoints and commands, Vitest + Playwright for UI. See
   [docs/TESTING.md](docs/TESTING.md).
7. **Docs, runbook and state in the same PR.** Update the relevant `docs/`, add
   or update the runbook, write an ADR if you made a decision, and update
   `.ai/state.json`.

## Pull-request workflow

```
fork → branch → make your change → run lint + tests →
open a PR → sign the CLA (bot) → CI green → CODEOWNERS review → maintainer merge
```

- Keep PRs focused and small where possible.
- Use **Conventional Commits** (`feat:`, `fix:`, `docs:`, `ci:`, `chore:`).
- Fill out the PR template (what/why/how-tested).
- CI must pass every gate: Go build/vet/test on macOS **and** Linux with the
  race detector, the per-package coverage gate, golangci-lint, gosec,
  govulncheck, a fuzz smoke run, bats, shellcheck, the portability gate,
  TypeScript strict, vitest, Playwright, the license scan, and content
  validation. None of them are advisory.
- A maintainer or domain code owner reviews; the **lead maintainer** holds final
  merge authority on engine/SDK/schema changes (see [GOVERNANCE.md](GOVERNANCE.md)).

## Reporting bugs / requesting features

Use the issue templates. For **security vulnerabilities, do not open a public
issue** — follow [SECURITY.md](SECURITY.md).

## Licensing of contributions

All contributions are licensed under [Apache-2.0](LICENSE) (the same license as
the project) and are subject to the CLA. New source files should carry an SPDX
header:

```
// SPDX-License-Identifier: Apache-2.0
```

Thank you for helping build the open platform-engineering simulator. ✈️
