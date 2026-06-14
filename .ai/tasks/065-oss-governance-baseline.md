# Task 065: OSS governance & licensing baseline

## Phase
M7 — OSS & Ecosystem Foundation (see docs/strategy/OSS-COMMERCIAL-STRATEGY.md)

## Type
foundation

## Priority
P0

## Description
Establish the legal + governance baseline that must exist before the project
accepts external contributions. This is the single highest-leverage, most
time-sensitive task: every contribution merged without it is a compounding
liability.

Depends on maintainer decisions A1 (license) and A2 (CLA vs DCO) in
docs/strategy/MAINTAINER-MANUAL-ACTIONS.md.

## Files to Modify
- `LICENSE` (chosen core license — recommended Apache-2.0), `NOTICE`
- `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `GOVERNANCE.md`, `SECURITY.md`,
  `MAINTAINERS.md`, `RELEASING.md`, `TRADEMARKS.md` (draft)
- `.github/CODEOWNERS` (engine/CLI/SDK/CI locked to lead maintainer)
- `.github/ISSUE_TEMPLATE/`, `.github/PULL_REQUEST_TEMPLATE.md`
- `.github/cla.yml` (if CLA) or DCO check in CI
- `.github/workflows/` — add a license-scan/SPDX gate
- SPDX headers across source files

## Implementation Notes
- CODEOWNERS must make architectural authority mechanical: `/cmd`, `/pkg`,
  `/engine`, `/.github`, `LICENSE`, `/docs/strategy/` require the lead maintainer.
- License-scan CI: fail the build on GPL/AGPL transitive deps entering the
  Apache core.
- Keep CONTRIBUTING pointed at the content surface (packs/scenarios) for newcomers.

## Acceptance Criteria
- [ ] LICENSE + NOTICE present; SPDX headers added; license-scan gate green
- [ ] All governance docs present and internally consistent
- [ ] CODEOWNERS + (documented) branch-protection settings enforce maintainer review
- [ ] CLA bot config (or DCO check) in place per decision A2

## Testing Instructions
CI runs the license-scan + (CLA/DCO) checks on a test PR. Manual: confirm
CODEOWNERS routes a sample engine PR to the maintainer.

## Dependencies
Maintainer decisions A1, A2 (manual-actions doc)
