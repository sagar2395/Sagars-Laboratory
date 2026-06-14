# Contributing to Flightdeck

Thanks for your interest in **Flightdeck** — the flight simulator for platform
engineers. This project thrives on community-contributed **scenarios, platform
modules, documentation, and code**. This guide explains how to contribute and
what to expect.

> New here? The highest-leverage place to start is **content** — a new scenario
> or scenario pack. See [docs/authoring](docs/authoring) (the authoring guide)
> and look for issues labelled **good first issue**.

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
| **Scenarios / packs** | `packs/`, `scenarios/` | scenario maintainers |
| **Platform modules** | `platform/<category>/<provider>/` | platform maintainers |
| **Incidents / learning / challenges** | `incidents/`, `learn/`, `challenges/` | scenario maintainers |
| **Documentation** | `docs/` | docs maintainers |
| **Engine / CLI / SDK** | `cmd/`, `pkg/`, `engine/` | **lead maintainer** (see GOVERNANCE.md) |

Changes to the engine, the public SDK (`pkg/`), or the scenario schema require an
**RFC** first — a short markdown PR under `docs/rfcs/` that a maintainer approves
before implementation. This keeps architectural direction coherent.

## Development setup

```bash
# Build the CLI (no committed binary)
make cli-build           # or: cd cmd/labctl && go build -o ../../bin/labctl .

# Run tests
cd cmd/labctl && go test ./...

# Bring up a local cluster + the platform to test changes end to end
bin/labctl init && make platform-up
```

See [CLAUDE.md](CLAUDE.md) and [AGENTS.md](AGENTS.md) for the project's working
agreement and the golden rules (cross-platform, CLI-wraps-scripts, declarative
content, every feature ships with a runbook, docs + state updated in the same PR).

## The golden rules (please read before a code/content PR)

1. **Cross-platform.** Must run on macOS (Apple Silicon + Intel) and modern
   Linux. No GNU-only shell flags (`grep -oP`, `sed -i` without backup,
   `readlink -f`, `date -d`). Detect OS/arch.
2. **CLI wraps scripts.** Don't move shell/helm/kubectl logic into Go.
3. **Declarative content.** Scenarios, faults, learning paths, and checks are
   YAML + scripts — never hardcoded in Go.
4. **Idempotent everything.** `helm upgrade --install`, `kubectl apply`, safe
   re-activation.
5. **Docs + state in the same PR.** Update the relevant `docs/`, the runbook, and
   `.ai/state.json` alongside the change.

## Pull-request workflow

```
fork → branch → make your change → run lint + tests →
open a PR → sign the CLA (bot) → CI green → CODEOWNERS review → maintainer merge
```

- Keep PRs focused and small where possible.
- Use **Conventional Commits** (`feat:`, `fix:`, `docs:`, `ci:`, `chore:`).
- Fill out the PR template (what/why/how-tested).
- CI must pass: shellcheck, portability check, yamllint, Go build/test, license
  scan, and (for content) scenario/pack validation.
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
