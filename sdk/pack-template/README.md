# Pack template

The starting point for a new scenario pack. The easy path is the scaffolder:

```bash
labctl pack init <name>                         # creates <name>/ from this template
cd <name> && labctl scenario verify <scenario>  # verify the bundled scenario
labctl pack publish . oci://<registry>/<repo>:0.1.0
```

Layout: `pack.yaml` at the root, one directory per scenario beside it. Copy this
directory manually if you prefer. Authoring guide: `docs/authoring/first-pack.md`.
