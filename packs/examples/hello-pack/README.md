# Hello Pack — example Flightdeck scenario pack

The reference layout for a scenario pack. Copy it to start your own.

```
hello-pack/
  pack.yaml                 # the manifest (packs.flightdeck.dev/v1)
  hello-flightdeck/         # one directory per scenario
    scenario.yaml           # the scenario (scenario.flightdeck.dev/v2)
    scripts/ready.sh
  README.md
```

## Try it

```bash
# Packs install from a git source. From a local clone you can serve this dir
# as a throwaway git repo, then install it:
git -C packs/examples/hello-pack init -q && git -C packs/examples/hello-pack add -A \
  && git -C packs/examples/hello-pack commit -qm pack
bin/labctl pack add "file://$(pwd)/packs/examples/hello-pack"
bin/labctl pack list
bin/labctl pack info hello-pack
bin/labctl scenario up hello-flightdeck     # then: scenario verify hello-flightdeck
bin/labctl pack remove hello-pack
```

## The manifest

See [`pack.yaml`](pack.yaml). Required: `apiVersion`, `kind`, `metadata.name`,
`metadata.version` (SemVer), `metadata.license`. The `spec.engine` block declares
the scenario schema versions and minimum `labctl` version the pack needs; the CLI
refuses incompatible packs with a clear error.

Validate against the JSON Schemas in [`sdk/schemas/`](../../../sdk/schemas/) for
editor support.
