# Archive — v1 documents

These documents described Flightdeck v1. They are kept for history and are **not
the plan of record**. The current plan is `docs/ROADMAP.md`.

| File | What it was |
|---|---|
| `ROADMAP-v1.md` | Part I (homelab hardening, phases 0–7), Part II (simulator, M1–M6), Part III (OSS/commercial, M7–M9) |
| `SIMULATOR-v1.md` | The v1 vision and feature catalogue |
| `architecture-v1.md` | The v1 architecture reference |
| `runbooks-v1/` | v1 runbooks, several covering features v2 removes |
| `strategy-v1/` | OSS/commercial strategy, maintainer actions, SaaS RFC |

Related: `.ai/archive/` holds the v1 state file and its 79 task descriptions.

## Why v2 departs from these

Summarised in `docs/ROADMAP.md` and decided in `docs/adr/0001-cut-cloud-and-commercial-scope.md`.
In short: v1 optimised for feature breadth and shipped a marketplace, entitlement
tiers and three cloud runtimes before the core loop had cancellation, durable
state, or any shell or UI tests. v2 inverts that priority.
