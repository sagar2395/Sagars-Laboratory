# Task 076: Edition packaging (CE / Pro / Enterprise)

## Phase
M9 — Commercial & Hosted (deferred)

## Type
feature

## Priority
P2

## Description
Package the same engine into editions via content + entitlement, never via
separate forks: Community (OSS + community packs), Professional (premium pack
bundles + support), Enterprise (SSO/RBAC from task 062, team mode from 063,
air-gapped registry mirror, compliance/audit packs, private hosting).

## Implementation Notes
- Editions differ only by entitled content + optional closed services injected
  through the public interfaces — the OSS engine binary is identical.
- Never move an existing free feature behind a paywall (community trust).

## Acceptance Criteria
- [ ] Edition = engine + entitled content/services; no engine fork
- [ ] CE remains fully functional standalone
- [ ] Documented, public CE-vs-paid feature line

## Dependencies
070, 074, 062, 063
