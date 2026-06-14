# RFCs (Requests for Comments)

Architectural changes to Flightdeck — the **engine**, the **public SDK (`pkg/`)**,
the **scenario/pack schema**, CI policy, or licensing — require an RFC before
implementation. This keeps the project's direction coherent and the reasoning
transparent (see [GOVERNANCE.md](../../GOVERNANCE.md)).

Routine content (scenarios, packs, docs, bug fixes) does **not** need an RFC.

## Process

1. Copy `0000-template.md` to `NNNN-short-title.md` (next free number).
2. Open it as a pull request.
3. Discussion happens on the PR. The **lead maintainer** approves, requests
   changes, or declines.
4. On approval, implementation proceeds (often in follow-up PRs that reference
   the RFC).

## Status values

`draft` → `accepted` → `implemented` (or `declined` / `superseded`).

## Index

| RFC | Title | Status |
|-----|-------|--------|
| [0001](0001-public-sdk-boundary.md) | Public SDK boundary (`pkg/`) | accepted |
