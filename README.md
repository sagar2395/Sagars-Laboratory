# Flightdeck

**Flightdeck is a Kubernetes platform-engineering simulator.** It stands up a
realistic, production-shaped cluster on your laptop in minutes, breaks it in
realistic ways, and grades you on how you fix it.

Practise the things production never lets you practise: diagnosing a
CrashLoopBackOff with a pager going off, draining a node under load without
breaking the SLO, deciding between two service meshes with evidence instead of
opinion.

> **Status:** mid-rebuild. Flightdeck is being redesigned to v2 —
> see [`docs/ROADMAP.md`](docs/ROADMAP.md) for the wave plan,
> [`docs/PRODUCT.md`](docs/PRODUCT.md) for what it is and who it is for, and
> [`docs/architecture/ARCHITECTURE.md`](docs/architecture/ARCHITECTURE.md) for
> how it works. AI sessions start from [`CLAUDE.md`](CLAUDE.md).
>
> Runs on **macOS (Apple Silicon and Intel) and Linux**. Local clusters only —
> cloud runtimes were cut in v2 ([ADR-0001](docs/adr/0001-cut-cloud-and-commercial-scope.md)).

## The four loops

| Loop | What you do | Command |
|---|---|---|
| **Build** | Stand up a cluster and install a real platform stack from swappable providers | `labctl lab up` |
| **Simulate** | Activate a declarative scenario with objectives and verifiable checks | `labctl scenario up <name>` |
| **Break** | Inject a realistic, reversible fault while traffic flows and alerts fire | `labctl incident inject <name>` |
| **Measure** | Grade the outcome — checks passed, time taken, hints used | `labctl scenario verify <name>` |

Everything you install is the real thing: actual Prometheus, actual Istio,
actual Kafka. Failures are injected into real systems, so the signal you debug
is the signal you would see in production.

## What's inside

| Layer | What it does |
|---|---|
| **Runtimes** | `k3d` (the golden path), `kind` (headless, powers CI), `incluster` (shared team server) |
| **Platform** | Ingress (Traefik/Nginx), monitoring (Prometheus + Grafana), logging (Loki), tracing (Tempo), GitOps (ArgoCD), mesh (Istio/Linkerd), data (Kafka/Postgres), secrets (Vault/ESO), autoscaling (KEDA), cost (OpenCost), security (Kyverno, cert-manager), chaos (Chaos Mesh) |
| **Scenarios** | Declarative playgrounds with objectives and machine-verifiable checks |
| **Incidents** | Reversible production faults with progressive hints and MTTR measurement |
| **Learn** | Ordered paths chaining scenarios and incidents into a curriculum |
| **Challenges** | Timed, graded runs with hidden hints |
| **Apps** | `go-api` (HTTP + metrics + tracing), `echo-server` (HTTP + Redis) |
| **CLI + UI** | `labctl` — one binary with the web dashboard embedded |

## Prerequisites

- Docker running
- `kubectl`, `helm` 3, and `k3d`
- Go 1.25+ and Node 22+ to build from source

```bash
make setup-tools              # installs tools for PROFILE (default: k3d)
make setup-tools PROFILE=kind # headless alternative
```

Versions are pinned in [`versions.env`](versions.env).

## Quickstart

```bash
git clone <repo-url> && cd flightdeck
cp .env.example .env

make cli-build                       # build bin/labctl (UI embedded)
make init                            # create the cluster + install the platform

./bin/labctl app deploy go-api
./bin/labctl scenario list
./bin/labctl scenario up observability-sre
./bin/labctl scenario verify observability-sre    # did you achieve the objective?

./bin/labctl ui                      # dashboard at http://localhost:3939
make teardown                        # remove everything
```

## Project structure

```
go.mod                    module go.flightdeck.dev/flightdeck (repo root)
cmd/labctl/               entrypoint
internal/cli/             cobra commands — thin adapters
internal/httpapi/         REST/WS — thin adapters
internal/run/             durable run engine (cancel, timeouts, locks, logs)
internal/store/           SQLite persistence
internal/toolchain/       kubectl/helm/k3d adapters + fakes for tests
pkg/                      public SDK: checks, scenario types, extension seam
ui/                       React SPA, embedded into the binary
platform/<cat>/<prov>/    install.sh / uninstall.sh / status.sh / values.yaml
scenarios/ incidents/     declarative content
learn/ challenges/
runtimes/<profile>/       k3d | kind | incluster
apps/<name>/              sample workloads
test/shell/               bats suites with kubectl/helm stubbed
docs/                     PRODUCT, ROADMAP, TESTING, architecture/, adr/, runbooks/
```

> `internal/run`, `internal/store` and `internal/toolchain` land in Wave 1.

## Documentation

| Document | What it covers |
|---|---|
| [Product](docs/PRODUCT.md) | What Flightdeck is, who it's for, what's out of scope |
| [Roadmap](docs/ROADMAP.md) | The wave plan, exit criteria, Definition of Done |
| [Architecture](docs/architecture/ARCHITECTURE.md) | How the system fits together |
| [Decisions](docs/adr/) | Why it was built this way, with alternatives considered |
| [Testing](docs/TESTING.md) | The four mandatory test layers and CI gates |
| [Runbooks](docs/runbooks/) | Hands-on validation a human performs before a wave merges |
| [CLI Reference](docs/cli-reference.md) | Every `labctl` command and flag |
| [Scenarios](docs/scenarios.md) | The scenario format and how to author one |
| [Authoring](docs/authoring/) | Your first scenario, extension seams, stability policy |
| [Contributing](CONTRIBUTING.md) | Golden rules and the PR bar |

## Testing

All four layers are mandatory for every change — see [docs/TESTING.md](docs/TESTING.md).

```bash
make test              # Go unit + shell, race detector, coverage gate
make test-ui           # vitest component tests
make test-e2e          # playwright journeys
make lint              # every static-analysis gate
```

## Configuration

Global settings live in `.env` (from [`.env.example`](.env.example)):

```bash
PROFILE=k3d                    # k3d | kind | incluster
CLUSTER_NAME=flightdeck
INGRESS_PROVIDER=traefik       # traefik | nginx
METRICS_PROVIDER=prometheus
```

Per-app settings live in `apps/<name>/app.env`:

```bash
BUILD_STRATEGY=docker
DEPLOY_STRATEGY=helm
HELM_VALUES=values-dev.yaml    # values-dev | values-prod-like | values-test
```

Own scenarios stay in your own repository — point `FLIGHTDECK_CONTENT_PATH` at
them and they appear alongside the built-in catalog. No pack format, no
registry ([ADR-0008](docs/adr/0008-content-extensibility-seam.md)).

## Key URLs (k3d)

| Service | URL | Credentials |
|---|---|---|
| go-api | http://go-api.k3d.local | — |
| echo-server | http://echo-server.k3d.local | — |
| Grafana | http://grafana.k3d.local | admin / admin |
| Prometheus | http://prometheus.k3d.local | — |
| ArgoCD | http://argocd.k3d.local | admin / (see install output) |
| Flightdeck UI | http://localhost:3939 | — |

Domains follow `${DOMAIN_SUFFIX:-k3d.local}` — never hardcoded.

## License

Apache-2.0 — see [LICENSE](LICENSE), [NOTICE](NOTICE) and
[TRADEMARKS.md](TRADEMARKS.md).
