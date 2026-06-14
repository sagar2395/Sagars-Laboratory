# Authoring Scenario Packs

A **scenario pack** is a versioned, installable bundle of scenarios. Packs are how
the community (and, later, the marketplace) distributes content without touching
the engine.

## Layout

The engine loads the **flat** layout — one directory per scenario at the pack
root, with `pack.yaml` alongside (see `packs/examples/hello-pack/`):

```
my-pack/
  pack.yaml                 # the manifest
  hello-flightdeck/         # one dir per scenario
    scenario.yaml
    scripts/  manifests/  values/  dashboards/
  README.md
```

## The manifest — `pack.yaml`

```yaml
apiVersion: packs.flightdeck.dev/v1      # required
kind: ScenarioPack                       # required
metadata:
  name: acme/kafka-drills                # required (optionally publisher/name)
  version: 1.4.2                         # required, SemVer
  license: Apache-2.0                    # required
  publisher: acme
  displayName: "Kafka Incident Drills"
  description: "..."
  tier: community                        # community | premium | enterprise
  keywords: [kafka, incident]
  categories: [data, incident-response]
spec:
  engine:
    scenarioApiVersions: ["scenario.flightdeck.dev/v2"]
    minLabctlVersion: "0.1.0"
  requires:
    platform: [data/kafka]               # platform categories that must be present
    packs:                               # other packs (semver range)
      - name: core-observability
        version: ">=1.0.0 <2.0.0"
  provides:
    scenarios: [kafka-broker-outage]
```

- **Required:** `apiVersion`, `kind`, `metadata.name`, `metadata.version`
  (SemVer), `metadata.license`.
- **`tier`** is metadata only — the open engine runs every tier identically;
  premium access is gated at distribution (registry auth) + the entitlement
  interface, never by the engine.
- **`spec.engine`** lets the CLI refuse a pack that needs a newer schema or a
  newer `labctl` than you're running, with an actionable error.
- Packs without a `pack.yaml` still install (treated as legacy community git
  packs) — but publish one so users get versioning and compatibility checks.

Validate against [`sdk/schemas/pack.schema.json`](../../sdk/schemas/pack.schema.json)
in your editor.

## Install / inspect / remove

```bash
labctl pack add <git-url>[@ref]          # install from git
labctl pack add oci://<reg>/<repo>[:tag] # install from an OCI registry
labctl pack publish <dir> oci://...      # publish a pack (see publishing.md)
labctl pack list                         # name, version, tier, scenarios
labctl pack info <pack-name>             # manifest metadata
labctl scenario up <scenario>            # activate a scenario the pack provides
labctl pack remove <pack-name>
```

`labctl scenario install|packs|uninstall` remain as aliases.

For OCI distribution, signing, and verification, see
[publishing.md](./publishing.md). A searchable registry index lands in task 069.

> **Security:** packs run scripts and apply manifests on your cluster. Only
> install sources you trust — prefer signed OCI packs pinned by digest.
