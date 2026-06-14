# Releasing Flightdeck

Releases are cut by the **lead maintainer** (release authority is reserved — see
[GOVERNANCE.md](GOVERNANCE.md)). This document is the runbook.

## Versioning

- **SemVer 2.0** for the engine/CLI: `MAJOR.MINOR.PATCH`.
- The **scenario schema** carries its own `apiVersion` (e.g. `scenario.lab.dev/v2`)
  and evolves independently; the CLI supports the current and previous schema
  versions (N and N-1).
- **Scenario packs** are versioned independently (SemVer), and declare the engine
  and schema versions they need (see the pack format, task 067).
- A **compatibility matrix** (which CLI supports which schema/pack apiVersions) is
  published in the docs and updated each release.

## Pre-1.0

While pre-1.0, minor versions may include breaking changes, but each is called
out in the changelog. The public SDK (`pkg/`) stability policy applies from the
first 1.0 release.

## Release steps (maintainer)

1. Ensure `main` is green (CI: lint, tests, license-scan, pack-validate).
2. Update the changelog (`CHANGELOG.md`) — automated from Conventional Commits.
3. Bump versions where applicable; update the compatibility matrix.
4. Tag: `git tag -s vX.Y.Z` (**signed** — the maintainer holds the signing key,
   manual action C6) and push the tag.
5. CI builds cross-platform `labctl` binaries, publishes the GitHub Release, and
   publishes first-party community packs as signed OCI artifacts (task 068).
6. Announce; update docs if needed.

## Signing

- Release artifacts and tags are **signed**; the private signing key is held by
  the lead maintainer and stored as a CI secret (manual action C6).
- Community packs published by CI are signed with cosign so installers can verify
  provenance.

## Hotfixes

Patch releases branch from the release tag, cherry-pick the fix, and follow the
same signed-release flow.
