<!-- Thanks for contributing to Flightdeck! Please fill this out. -->

## What

<!-- What does this PR do? One or two sentences. -->

## Why

<!-- The motivation / the issue it closes. Link issues: Closes #123 -->

## How it was tested

<!-- Commands you ran, cluster you tested on (k3d/AKS/EKS), scenario verify output, etc. -->

## Type of change

- [ ] Scenario / pack / content
- [ ] Platform module
- [ ] Docs
- [ ] Engine / CLI / SDK (requires an RFC under `docs/rfcs/` — link it)
- [ ] CI / tooling

## Checklist

- [ ] I signed the **CLA** (the bot will prompt on first PR)
- [ ] Conventional Commit title (`feat:`, `fix:`, `docs:`, `ci:`, `chore:`)
- [ ] Cross-platform & idempotent (no GNU-only shell flags; safe to re-run)
- [ ] Updated relevant **docs** and the **runbook** (golden rule 8)
- [ ] Updated `.ai/state.json` if task status changed
- [ ] New source files carry an SPDX header (`// SPDX-License-Identifier: Apache-2.0`)
- [ ] `go test ./...`, shellcheck, and yamllint pass locally
