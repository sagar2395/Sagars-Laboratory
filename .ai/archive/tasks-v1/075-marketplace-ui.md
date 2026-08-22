# Task 075: Marketplace UI in the web app

## Phase
M8 — Marketplace (deferred)

## Type
feature

## Priority
P2

## Description
Surface pack discovery/install/management in the SPA: browse the catalog, view
pack details (publisher, version, tier, verified badge), install/remove, and see
entitlement status. Mirrors `labctl pack` over the existing REST/WS API.

## Implementation Notes
- Read from the same index/catalog the CLI uses (static index or hosted API).
- Premium listings show entitlement state; never expose premium bits to
  unentitled users.

## Acceptance Criteria
- [ ] Browse/search/install/remove packs from the UI
- [ ] Pack detail view with publisher/version/tier/verified
- [ ] Entitlement-aware premium listings

## Dependencies
069, 073
