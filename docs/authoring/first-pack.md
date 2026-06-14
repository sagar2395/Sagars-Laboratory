# Your First Scenario Pack

This walks the full loop — **scaffold → edit → verify → publish** — in a few
minutes. It is the fast path for a first contribution.

> Prereqs: `bin/labctl` built (`make cli-build`). For publishing: `oras` (and
> `cosign` if signing). For inline editor validation: VS Code + the Red Hat YAML
> extension (the repo's `.vscode/settings.json` wires the schemas).

## 1. Scaffold

A **scenario** is one declarative playground; a **pack** bundles one or more
scenarios for distribution.

```bash
# A single scenario inside this repo:
labctl scenario new my-first-scenario     # -> scenarios/my-first-scenario/

# Or a standalone pack (its own directory / repo):
labctl pack init my-org/my-first-pack     # -> ./my-first-pack/
```

Both scaffolds are **valid and verify-ready out of the box**: a v2
`scenario.yaml`, a passing `checks/ready.sh`, and (for packs) a `pack.yaml` +
README. Each YAML carries a `# yaml-language-server: $schema=…` modeline so your
editor validates as you type.

## 2. Edit

Open `scenario.yaml` and make it real:

- `description`, `objectives` — what the learner does and learns.
- `stages[].components` — what to deploy (helm / manifest / grafana-dashboard /
  script). See [scenarios.md](../scenarios.md) for the component reference.
- `checks` — machine-verifiable assertions. Replace the scaffolded `script`
  check with `http`, `kubectl`, or `promql` checks (or a real script).

Reference: [packs.md](packs.md) for the `pack.yaml` manifest fields.

## 3. Verify

Run the checks against a live cluster (`make init` first if you need one):

```bash
labctl scenario up my-first-scenario       # activate it
labctl scenario verify my-first-scenario   # run checks; --watch to retry
```

`verify` is green immediately on the scaffold; keep it green as you add real
components and checks.

## 4. Publish (packs)

```bash
labctl pack validate-index registry/index.json    # if adding to the registry
labctl pack publish ./my-first-pack oci://ghcr.io/my-org/my-first-pack:0.1.0
# add --sign to cosign-sign it
```

Then open a PR adding your pack to the registry index — see
[registry.md](registry.md). Distribution and signing details:
[publishing.md](publishing.md).

## Contributing back

- Good first issues are labelled **`good first issue`** / **`help wanted`** (see
  [`.github/labels.yml`](../../.github/labels.yml)).
- Read [`CONTRIBUTING.md`](../../CONTRIBUTING.md); content PRs must be
  cross-platform, idempotent, and declarative (the golden rules).
