# Task 073: Hosted catalog API + verified-publisher program

## Phase
M8 — Marketplace (deferred; build on real demand)

## Type
feature

## Priority
P2

## Description
A hosted superset of the static registry index (task 069): search, ratings,
download counts, verified badges, private/entitled listings, and a
verified-publisher onboarding flow. The CLI prefers it when configured and falls
back to the static index otherwise.

## Implementation Notes
- Do NOT build before pack supply/demand justifies it — the static index covers
  ~90% of ecosystem value at ~1% of the cost.
- Keep the static index as the source of truth / fallback; the API augments it.
- Verified publishers: signing key on file, reserved namespace, basic review.

## Acceptance Criteria
- [ ] CLI uses the API when configured, static index otherwise
- [ ] Verified-publisher onboarding + badge
- [ ] No regression for purely-static/offline use

## Dependencies
069
