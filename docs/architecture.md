# Sagars-Laboratory Architecture

## Purpose

This document defines the desired-state architecture for Sagars-Laboratory.

It is written for two audiences:

- Platform engineers, SREs, DevOps learners, and platform teams using the product
- Human and AI contributors who need a clear architectural source of truth before making changes

This document should be treated as the architectural north star for the repository. If the codebase and this document disagree, contributors should assume the codebase is the current implementation and this document is the intended direction unless a newer decision document says otherwise.

---

## Product Definition

Sagars-Laboratory is a **Platform Engineering Simulator and Distributed Systems Playground**.

It is a safe environment for learning and demonstrating how modern platforms are built, operated, observed, secured, and intentionally stressed.

The platform should enable users to:

- provision and switch between runtimes such as `k3d`, `AKS`, `EKS`, and `GKE`
- install opinionated platform capabilities such as ingress, metrics, logs, traces, GitOps, policy, TLS, and chaos tooling
- deploy reusable services and sample workloads
- compose and deploy multi-service stacks that form realistic distributed systems
- activate guided scenarios that teach real platform and SRE workflows
- follow structured learning paths with difficulty levels and progress tracking
- observe the system through dashboards, logs, traces, topology views, and interactive service graphs
- inject application-level failure modes without requiring external chaos tooling
- validate prerequisites before scenario activation with pre-flight health checks
- boot entire lab environments from blueprints with a single command
- break things safely and recover quickly
- reset the environment to a known-good state

---

## Target Users

### 1. Platform Engineers

Users who want to design and validate internal platform capabilities such as runtime provisioning, service enablement, observability, policy, and GitOps. They need topology views, dry-run execution plans, and pluggable provider interfaces.

### 2. Platform Teams

Teams who need a shared sandbox for architecture spikes, golden path experiments, onboarding, demos, and operational training. They need lab blueprints for one-command environment provisioning and reset capabilities.

### 3. DevOps and SRE Learners

Individuals who want a practical, hands-on environment to understand Kubernetes, cloud runtimes, observability, reliability, security controls, and failure handling. They need guided learning paths, difficulty-graded scenarios, progress tracking, and interactive walkthroughs.

---

## Architectural Goals

The architecture must optimize for the following:

1. **Learnability**
   A new user should be able to understand the platform model and get value quickly.

2. **Reproducibility**
   A lab environment should be rebuildable, resettable, and deterministic.

3. **Composability**
   Runtimes, platform components, services, apps, stacks, and scenarios should be mix-and-match.

4. **Operational Realism**
   The system should feel like a small but credible platform engineering environment, not a toy demo.

5. **Safe Failure**
   Users should be able to introduce faults, drift, and policy violations without damaging their local machine or cloud environment beyond the declared lab boundary.

6. **Contributor Clarity**
   The architecture must make ownership boundaries obvious so multiple agents can work in parallel with low coordination overhead.

7. **Progressive Learning**
   The platform should guide users from beginner to advanced concepts in a structured sequence. Scenarios should have difficulty levels, learning objectives, and checkpoints so users can build knowledge incrementally.

8. **Realistic Workloads**
   Sample applications should form realistic distributed systems with inter-service communication, not just isolated single-service deployments. Multi-service stacks should enable scenarios that stress service-to-service interactions, distributed tracing, network policies, and cascading failure modes.

---

## Non-Goals

Sagars-Laboratory is not intended to be:

- a production PaaS
- a multi-tenant enterprise control plane
- a replacement for Terraform Cloud, Argo CD, Backstage, or a commercial internal developer platform
- a generic cluster management product

Its purpose is simulation, experimentation, education, and reference implementation.

---

## Core Architectural Principles

### 1. Declarative First

Desired state should be described through files and metadata where possible. Imperative scripts are allowed, but they should be thin execution adapters around declarative inputs.

### 2. Clear Layer Boundaries

Each layer should have one job and one reason to change. Runtime provisioning, platform installation, service deployment, app deployment, stack composition, and scenario activation must remain separate concerns.

### 3. Idempotent Operations

Repeated `up`, `down`, `install`, `uninstall`, and `status` actions should be safe and predictable.

### 4. Pluggable Providers

Equivalent capabilities should be swappable. Examples:

- `traefik` or `nginx` for ingress
- `k3d`, `AKS`, `EKS`, or `GKE` for runtime
- Kubernetes-native or cloud-managed implementations for services over time

### 5. Scenario-Driven Learning

The core user experience is not just "install tools." It is "activate a realistic scenario, investigate behavior, and learn from the system."

### 6. Observable by Default

The platform should teach users how systems behave. Metrics, logs, traces, events, health, and topology should be first-class concepts.

### 7. Repository as Control Surface

The repository layout should mirror the product architecture closely enough that contributors can infer where new capabilities belong.

---

## Desired-State System Model

At the highest level, the system is a control plane for building and exploring ephemeral platform environments.

```text
User
  |
  v
UI / CLI / API
  |
  v
Control Plane
  |
  +-- Catalog & State
  +-- Runtime Orchestrator
  +-- Foundation Orchestrator
  +-- Platform Orchestrator
  +-- Service Orchestrator
  +-- Application Orchestrator
  +-- Stack Orchestrator
  +-- Scenario Engine
  +-- Learning Engine
  +-- Blueprint Orchestrator
  +-- Topology & Insight Engine
  |
  v
Execution Adapters
  |
  +-- Shell scripts
  +-- Helm
  +-- kubectl
  +-- Terraform
  +-- Cloud CLIs
```

This architecture deliberately separates **intent**, **orchestration**, and **execution**:

- **Intent** is expressed in configuration, scenario manifests, stack definitions, blueprint definitions, and values files.
- **Orchestration** is handled by `labctl` and supporting internal packages.
- **Execution** is performed through scripts, Helm charts, Kubernetes manifests, Terraform, and cloud tooling.

---

## Core Domains

### 1. Experience Layer

The experience layer includes:

- CLI: `labctl`
- Web UI: React-based single-page application
- HTTP API that serves both CLI and UI

Its responsibilities are:

- accept user intent
- present current state and topology
- surface guided learning workflows and scenario walkthroughs
- explain what will happen before changes are applied (dry-run / execution plan)
- display real-time operation progress via WebSocket

It must not contain infrastructure logic directly. It should call internal orchestration services.

The web UI is described in detail in the **Web UI Architecture** section below.

### 2. Control Plane

The control plane is the brain of the system and should remain the most stable architectural layer.

Its responsibilities are:

- discover available runtimes, platform components, services, apps, stacks, blueprints, and scenarios
- validate prerequisites and perform pre-flight health checks
- orchestrate install and teardown order with dependency resolution
- maintain lightweight state for what is active
- expose a consistent contract to both CLI and UI
- generate execution plans before applying changes (dry-run mode)
- track user progress through learning paths

Today this responsibility lives primarily inside `cmd/labctl/internal/`.

### 3. Runtime Domain

The runtime domain provides the execution substrate where workloads run.

Examples:

- local Kubernetes via `k3d`
- managed Kubernetes via `AKS`
- managed Kubernetes via `EKS`
- managed Kubernetes via `GKE`

Responsibilities:

- create and destroy clusters
- switch active runtime context
- publish runtime-specific settings such as ingress class, domain suffix, storage class, and registry type

Repository mapping:

- `runtimes/`

### 4. Foundation Domain

The foundation domain manages cloud or infrastructure resources that sit underneath or beside the runtime.

Examples:

- VPCs, node groups, registries, IAM roles
- Azure resource groups, ACR, Log Analytics
- GCP projects, GKE node pools, Artifact Registry
- future managed services such as databases, queues, or object stores

Responsibilities:

- provision prerequisite infrastructure
- encode cloud-environment differences
- keep infrastructure lifecycle separate from workload lifecycle

Repository mapping:

- `foundation/terraform/`

### 5. Platform Domain

The platform domain installs reusable platform capabilities on top of a runtime.

Examples:

- ingress
- metrics
- dashboards
- logging
- tracing
- GitOps
- policy
- TLS
- secrets
- chaos tooling

Responsibilities:

- install, uninstall, and report status for platform components
- keep provider implementations behind a common category model
- expose capabilities that scenarios and apps can depend on
- enforce a standardized `_interface.yaml` contract for every category

Every platform category must define an `_interface.yaml` that declares:

```yaml
category: ingress
providers:
  - name: traefik
    default: true
  - name: nginx
namespace: traefik          # default namespace for the default provider
capabilities:
  - http-routing
  - tls-termination
```

Repository mapping:

- `platform/`

### 6. Service Domain

The service domain represents reusable dependencies consumed by workloads, stacks, and scenarios.

Examples:

- Redis today
- PostgreSQL, Kafka, MinIO, RabbitMQ in the future

Responsibilities:

- deploy and remove reusable backing services
- standardize local-vs-cloud service provisioning over time
- make service dependencies explicit instead of hidden in app charts

Repository mapping:

- `services/`

### 7. Application Domain

The application domain contains sample workloads used for learning, validation, and scenario exercises.

Examples in the current repository:

- `go-api` — HTTP API with Prometheus metrics, OpenTelemetry tracing, and injectable failure modes
- `echo-server` — HTTP echo service with Redis caching and Prometheus metrics

Planned applications:

- `traffic-generator` — configurable load generator with named traffic profiles
- Stack-specific applications (see Stack Domain below)

Responsibilities:

- provide realistic workloads with health endpoints and platform integration hooks
- expose telemetry and failure modes useful for observability, GitOps, policy, and chaos scenarios
- implement the Failure Mode Library API contract (see below)
- remain small enough to teach from

#### Failure Mode Library

All sample applications should implement a standardized failure injection API that enables chaos experimentation without requiring external tooling like Chaos Mesh:

```text
POST /debug/failure/enable   { "mode": "high-latency" | "error-rate" | "oom" | "cpu-spike" | "connection-drop" }
POST /debug/failure/disable
GET  /debug/failure/status
```

This API is guarded behind a debug flag and disabled by default. It allows scenarios to inject application-level failures that are visible in metrics, logs, and traces without requiring cluster-level chaos controllers. The `go-api` already has a `simulateFailure` flag; this pattern should be formalized and expanded across all apps.

#### Traffic Generator

The `traffic-generator` app is a dedicated workload that generates sustained, configurable HTTP traffic against other apps and stacks:

- supports named traffic profiles: `steady`, `spike`, `ramp`, `random`, `checkout-heavy`
- targets are configurable via environment variables or API
- emits its own Prometheus metrics so the generator itself is observable
- enables meaningful Grafana dashboards during observability and chaos scenarios

Repository mapping:

- `apps/`
- `engine/`

### 8. Scenario Domain

The scenario domain is the core teaching mechanism of the platform.

A scenario is a curated learning or validation experience that composes platform capabilities, services, applications, stacks, configuration, and exploration guidance around a real theme.

Examples:

- observability and SRE
- GitOps and CI/CD
- security and compliance
- chaos engineering

Responsibilities:

- declare prerequisites including platform components, apps, services, and stacks
- install additional components in the right order
- expose guided exploration steps with learning objectives and checkpoints
- support pre-flight validation of all prerequisites before activation
- make complex concepts easy to activate repeatedly
- assign difficulty levels to enable progressive learning

#### Enhanced Scenario Manifest Format

The `scenario.yaml` format gains these fields:

```yaml
name: observability-sre
displayName: "Observability & SRE"
description: "Full observability stack with log aggregation, distributed tracing, alerting rules, and SLO dashboards."
category: observability
difficulty: intermediate           # beginner | intermediate | advanced

prerequisites:
  platform:
    - ingress
    - monitoring/metrics
    - monitoring/grafana
  apps:
    - go-api
  stacks:                          # NEW — scenarios can require a full stack
    - microservices-demo

learning:                          # NEW — structured educational content
  objectives:
    - "Understand the three pillars of observability: metrics, logs, and traces"
    - "Configure Prometheus alerting rules and verify they fire"
    - "Build an SLO dashboard in Grafana using histogram data"
    - "Correlate a distributed trace across multiple services in Tempo"
  concepts:
    - "SLO"
    - "Error Budget"
    - "Golden Signals"
    - "Distributed Tracing"
    - "Log Aggregation"
  checkpoints:
    - step: 1
      description: "Verify Loki is receiving logs from all stack services"
      verify: "kubectl logs -n monitoring -l app=promtail --tail=5 | grep go-api"
    - step: 2
      description: "Confirm Tempo shows traces with multiple spans across services"
      verify: "curl -s http://grafana.{{.DomainSuffix}}/api/datasources | grep Tempo"
    - step: 3
      description: "Check that alerting rules are loaded in Prometheus"
      verify: "curl -s http://prometheus.{{.DomainSuffix}}/api/v1/rules | grep alerting"
  reflection:
    - "What changed in Grafana when you injected high-latency via the failure mode API?"
    - "How would you set an error budget for the go-api /health endpoint?"
    - "Which service in the stack had the highest p99 latency and why?"

components:
  - name: loki
    type: helm
    # ... (existing component definitions unchanged)

explore:
  urls:
    - label: "Grafana Dashboards"
      url: "http://grafana.{{.DomainSuffix}}"
  commands:
    - label: "Generate API traffic"
      command: "labctl traffic start --profile steady --target go-api"
  tips:
    - "Start by generating background traffic, then explore logs in Grafana"
```

#### Scenario Validation

The `labctl scenario validate <name>` command performs a pre-flight check before activation:

1. Verify each prerequisite platform component has a healthy Helm release (not just namespace existence)
2. Verify each prerequisite app is deployed and passing health checks
3. Verify each prerequisite stack is fully deployed with all constituent services healthy
4. Report exactly which prerequisites are missing and suggest the commands to fix them

This saves learners significant debugging time and ensures scenarios start from a known-good state.

Repository mapping:

- `scenarios/`
- `cmd/labctl/internal/scenario/`

### 9. Topology and Insight Domain

The topology and insight domain explains what is running, how components relate, and what is degraded.

Desired-state responsibilities:

- generate a topology graph from runtime, platform, service, app, stack, and scenario data
- show health and dependency relationships with color-coded status
- support investigation workflows for SRE-style debugging
- provide the data model for the interactive topology view in the web UI
- combine declared topology edges (from `stack.yaml`) with live cluster state

Topology data model:

```text
TopologyGraph
  +-- nodes[]
  |     +-- id: string
  |     +-- type: "app" | "service" | "platform" | "runtime"
  |     +-- name: string
  |     +-- namespace: string
  |     +-- health: "healthy" | "degraded" | "unhealthy" | "unknown"
  |     +-- metadata: { replicas, ready, version, ... }
  +-- edges[]
        +-- source: node-id
        +-- target: node-id
        +-- type: "depends-on" | "routes-to" | "monitors"
        +-- protocol: "http" | "grpc" | "tcp" | "redis" | "kafka"
```

Initially, topology edges come from declared `topology[]` entries in `stack.yaml`. Future iterations may augment this with live network data from service mesh sidecars or eBPF-based traffic observation.

Repository mapping:

- `cmd/labctl/internal/topology/` (future)

### 10. Stack Domain

The stack domain introduces multi-service composable workloads as a first-class resource. A stack is a named composition of multiple applications and services that together form a realistic distributed system.

Stacks exist between the Application domain (individual apps) and the Scenario domain (educational experiences). A stack provides the deployable substrate; a scenario provides the learning structure on top of it.

#### Why Stacks

Current scenarios can only reference single apps like `go-api`. This limits what can be taught:

- no inter-service communication patterns to observe
- no distributed tracing across service boundaries
- no meaningful network policy scenarios (requires multiple services to isolate)
- chaos experiments on a single pod are not realistic

Stacks directly unlock richer scenarios by providing multi-tier systems where service interactions, cascading failures, and cross-service observability become possible.

#### Stack Manifest Format

Each stack is defined by a `stack.yaml` in its directory under `stacks/`:

```yaml
name: e-commerce
displayName: "E-Commerce Platform"
description: "A multi-tier e-commerce system with API gateway, product catalog, order processing, and caching layers."
apps:
  - name: api-gateway
    image: sagars-lab/api-gateway
    port: 8080
    replicas: 2
  - name: product-service
    image: sagars-lab/product-service
    port: 8081
    replicas: 2
  - name: order-service
    image: sagars-lab/order-service
    port: 8082
    replicas: 2
  - name: frontend
    image: sagars-lab/frontend
    port: 3000
    replicas: 1
services:
  - redis
  - postgres
topology:
  - source: frontend
    target: api-gateway
    protocol: http
  - source: api-gateway
    target: product-service
    protocol: http
  - source: api-gateway
    target: order-service
    protocol: http
  - source: order-service
    target: postgres
    protocol: tcp
  - source: product-service
    target: redis
    protocol: redis
healthCheck:
  timeout: 120s
  endpoints:
    - name: api-gateway
      url: "http://api-gateway.{{.DomainSuffix}}/health"
    - name: product-service
      url: "http://product-service.{{.DomainSuffix}}/health"
```

#### Stack Lifecycle

```text
labctl stack list                    # list available stacks
labctl stack up <name>               # deploy all apps and services in order
labctl stack down <name>             # tear down all apps and services
labctl stack status <name>           # show health of all stack components
labctl stack topology <name>         # print the topology graph
```

The Stack Orchestrator in the control plane resolves the dependency order from the `topology` edges, deploys services first, then apps in dependency order, and waits for health checks to pass before reporting success.

#### Planned Stacks

| Stack | Apps | Services | Primary Scenario Use |
|-------|------|----------|---------------------|
| `e-commerce` | frontend, api-gateway, product-service, order-service | Redis, PostgreSQL | All four scenarios — observability, chaos, security, GitOps |
| `data-pipeline` | producer, consumer, stream-processor | Kafka, MinIO | Observability, chaos — event-driven failure modes |
| `microservices-demo` | service-a, service-b, service-c, service-d, service-e | Redis | Observability, chaos — cascading failure chains |
| `event-driven` | publisher, subscriber, dlq-handler | RabbitMQ, PostgreSQL | Chaos, observability — dead letter queue patterns |
| `api-gateway-stack` | gateway, auth-service, rate-limiter, backend-a, backend-b | Redis | Security, observability — rate limiting, auth flows |
| `ml-inference` | model-server, feature-store, request-router | MinIO, Redis | Observability — latency-sensitive workload patterns |

Each planned stack's constituent apps should be built as small Go services following the same patterns as `go-api` and `echo-server`: Prometheus metrics, structured logging, OpenTelemetry tracing, health endpoints, and the Failure Mode Library API.

Repository mapping:

- `stacks/`
- `cmd/labctl/internal/stack/` (future)

### 11. Learning Domain

The learning domain formalizes education as a first-class architectural concern. While scenarios provide the content, the learning domain provides the structure, progression, and tracking.

#### Learning Paths

A learning path is an ordered sequence of scenarios with progressive difficulty. Each path guides a user from foundational concepts to advanced techniques:

```yaml
name: platform-engineering-fundamentals
displayName: "Platform Engineering Fundamentals"
description: "A structured path from basic Kubernetes concepts to advanced platform engineering techniques."
paths:
  - order: 1
    scenario: observability-sre
    difficulty: beginner
    unlocks: "Understanding metrics, logs, and traces"
  - order: 2
    scenario: security-compliance
    difficulty: intermediate
    unlocks: "Policy enforcement and TLS management"
  - order: 3
    scenario: gitops-cicd
    difficulty: intermediate
    unlocks: "Declarative deployment and drift detection"
  - order: 4
    scenario: chaos-engineering
    difficulty: advanced
    unlocks: "Failure injection and resilience validation"
```

#### Progress Tracking

User progress is persisted locally in `~/.labctl/progress.json`:

```json
{
  "scenarios": {
    "observability-sre": {
      "activated": true,
      "completedCheckpoints": [1, 2],
      "lastActive": "2026-03-20T14:30:00Z"
    }
  },
  "learningPaths": {
    "platform-engineering-fundamentals": {
      "currentStep": 2,
      "completedScenarios": ["observability-sre"]
    }
  }
}
```

Progress tracking is local-first. For Platform Teams sharing environments, a future iteration may add server-side progress storage, but local-only is sufficient for v1.

#### Interactive Guide

The `labctl scenario guide <name>` command provides an interactive walkthrough:

1. Displays learning objectives for the scenario
2. Walks through checkpoints one at a time
3. Shows the relevant `explore` commands inline with copy-paste support
4. Runs verification commands to confirm checkpoint completion
5. Presents reflection prompts after all checkpoints pass
6. Updates progress tracking on completion

#### Difficulty Levels

Scenarios must declare one of three difficulty levels:

| Level | Audience | Prerequisites |
|-------|----------|---------------|
| `beginner` | New to Kubernetes and platform concepts | Basic terminal skills |
| `intermediate` | Familiar with Kubernetes, learning platform engineering | Completed at least one beginner scenario |
| `advanced` | Experienced with Kubernetes, learning SRE and chaos practices | Completed intermediate scenarios |

Repository mapping:

- `learning-paths/`
- `cmd/labctl/internal/learning/` (future)

### 12. Blueprint Domain

The blueprint domain enables one-command lab environment provisioning. A blueprint is a reusable composition of runtime, platform components, services, stacks, and scenarios that together define a complete lab experience.

Blueprints are the primary mechanism for Platform Teams who need to spin up identical environments for onboarding, demos, and training without manually configuring each layer.

#### Blueprint Manifest Format

Each blueprint is defined by a `lab.yaml` in its directory under `blueprints/`:

```yaml
name: sre-training
displayName: "SRE Training Lab"
description: "Complete SRE training environment with observability stack, chaos tooling, and multi-service workloads."
runtime: k3d
platform:
  - ingress
  - monitoring/metrics
  - monitoring/grafana
  - logging/loki
  - tracing/tempo
  - chaos/chaos-mesh
services:
  - redis
stacks:
  - microservices-demo
scenarios:
  - observability-sre
  - chaos-engineering
learningPath: platform-engineering-fundamentals
```

#### Blueprint Lifecycle

```text
labctl lab list                      # list available blueprints
labctl lab up <name>                 # provision runtime + install platform + deploy stacks + activate scenarios
labctl lab down <name>               # tear down everything in reverse order
labctl lab reset <name>              # soft reset — tear down stacks and scenarios, keep runtime and platform
labctl lab status <name>             # show health of all blueprint components
labctl lab plan <name>               # dry-run — show execution plan without applying changes
```

The `lab plan` command shows the full execution plan before any changes are made. This serves Platform Engineers who want to understand what will happen before committing, and serves as a teaching tool that explains the dependency ordering.

#### Planned Blueprints

| Blueprint | Runtime | Platform | Stacks | Scenarios |
|-----------|---------|----------|--------|-----------|
| `sre-training` | k3d | ingress, metrics, grafana, loki, tempo, chaos-mesh | microservices-demo | observability-sre, chaos-engineering |
| `security-workshop` | k3d | ingress, metrics, grafana, kyverno, cert-manager | api-gateway-stack | security-compliance |
| `gitops-demo` | k3d | ingress, argocd | e-commerce | gitops-cicd |
| `full-platform` | k3d | all components | e-commerce, microservices-demo | all scenarios |

Repository mapping:

- `blueprints/`
- `cmd/labctl/internal/blueprint/` (future)

---

## Desired Resource Model

The desired state of the application should revolve around these first-class concepts:

| Resource | Meaning |
|----------|---------|
| Runtime | Execution environment such as `k3d`, `AKS`, `EKS`, or `GKE` |
| Foundation | Underlying cloud and infrastructure resources |
| Platform Component | A reusable capability such as ingress, metrics, tracing, policy, or chaos |
| Service | A reusable backing dependency such as Redis |
| Application | A sample workload deployed into the lab |
| Stack | A named composition of multiple apps and services forming a multi-tier distributed system |
| Scenario | A guided experience that composes components and teaches a concept |
| Blueprint | A reusable lab environment definition composing runtime, platform, stacks, and scenarios |
| Learning Path | An ordered sequence of scenarios with progressive difficulty |
| Progress | User's completion state for scenarios and learning paths |
| Session | The currently active lab state for a user |
| Topology | A computed view of relationships, health, and dependencies |

---

## Control Plane Responsibilities

The control plane should standardize the following workflow:

1. Discover available resources from the repository (runtimes, platform, services, apps, stacks, blueprints, scenarios, learning paths)
2. Validate compatibility and prerequisites with pre-flight health checks
3. Resolve desired state into an ordered execution plan (with dry-run support via `labctl plan`)
4. Execute through adapters in dependency order
5. Capture resulting state
6. Present health, status, topology, and exploration guidance back to the user
7. Track learning progress and checkpoint completion

This makes the control plane the stable contract, while providers and scripts remain replaceable implementation details.

---

## Execution Model

Execution should remain adapter-based.

### Adapter Types

- shell scripts for simple lifecycle actions
- Helm for packaged Kubernetes capabilities
- raw manifests for focused Kubernetes resources
- Terraform for cloud infrastructure
- cloud CLIs where provider interaction is unavoidable

### Rules

- adapters must be idempotent
- adapters must have clear input and output expectations
- adapters should not own unrelated state
- control-plane code should prefer orchestration over embedding command details everywhere

### Execution Contracts

Every adapter must follow a documented contract. The following contracts should be standardized:

| Adapter | Required Scripts | Expected Env Vars | Exit Codes |
|---------|-----------------|-------------------|------------|
| Runtime | `up.sh`, `down.sh` | `CLUSTER_NAME`, `PROFILE` | 0 = success, 1 = failure |
| Platform component | `install.sh`, `uninstall.sh`, `status.sh` | `NAMESPACE`, `VALUES_FILE` | 0 = success, 1 = failure |
| Service | `install.sh`, `uninstall.sh`, `status.sh` | `NAMESPACE` | 0 = success, 1 = failure |
| Scenario component | defined in `scenario.yaml` | `PROJECT_ROOT`, `DOMAIN_SUFFIX` | handled by scenario engine |

---

## Web UI Architecture

The web UI is the primary visual interface for Sagars-Laboratory. It consumes the HTTP API served by `labctl` and provides an interactive experience for managing the lab environment, exploring topology, learning through scenarios, and monitoring system health.

### Technology

- **Framework**: React 18+ with TypeScript
- **Build tool**: Vite
- **Styling**: Tailwind CSS
- **Theme**: Dark theme by default (matching the current design language)
- **Data fetching**: REST API calls + WebSocket for real-time updates
- **Graph visualization**: React Flow (for topology view)
- **Bundling**: Production build outputs to `ui/dist/` for Go embed

The UI is embedded into the `labctl` binary via Go's `embed` package. During development, Vite's dev server proxies API requests to the running `labctl` backend. For production, `vite build` outputs static assets to `ui/dist/` which the Go server serves.

### Migration Strategy

The current UI is a single `index.html` with vanilla JavaScript (~900 lines). The migration to React should be phased:

1. **Phase 1**: Build the React app alongside the existing UI. The Go server serves the React app at `/` and the legacy UI at `/legacy`. Both share the same API.
2. **Phase 2**: Achieve feature parity in the React app for all existing functionality (status, platform, apps, scenarios, WebSocket, command output panel).
3. **Phase 3**: Remove the legacy UI once the React app is validated.

### Views

The UI has six primary views accessible from a persistent sidebar navigation.

#### View 1: Dashboard (Home)

The landing page that gives an at-a-glance summary of the entire lab environment.

**Layout**:

```text
+------------------------------------------------------------------+
| Header: labctl Dashboard    [Runtime Selector] [WS Status Dot]   |
+--------+---------------------------------------------------------+
|        |                                                         |
| Side-  |  Health Summary Bar                                     |
| bar    |  [Cluster: Connected] [Platform: 4/4] [Apps: 2/3]       |
|        |  [Stacks: 1/2]       [Scenarios: 1 active]              |
| Icons: |                                                         |
|  Home  |  Quick Action Cards (3-column grid)                     |
|  Topo  |  +----------------+ +----------------+ +---------------+|
|  Scene |  | Start Scenario | | Deploy Stack   | | Boot Lab      ||
|  Stack |  | Browse 4       | | Browse 6       | | Browse 4      ||
|  Plat  |  | scenarios      | | stacks         | | blueprints    ||
|  Blue  |  +----------------+ +----------------+ +---------------+|
|        |                                                         |
|        |  Active Operations Feed (live via WebSocket)            |
|        |  [14:30:01] Installing loki... ████████░░ 80%           |
|        |  [14:29:45] Deployed go-api ✓                           |
|        |                                                         |
|        |  Quick Links (dynamic — from /api/dashboards)           |
|        |  [Grafana] [Prometheus] [ArgoCD] [Traefik]              |
|        |                                                         |
+--------+---------------------------------------------------------+
| Command Output Panel (resizable, collapsible)                    |
+------------------------------------------------------------------+
```

**Data sources**: `GET /api/status`, `GET /api/dashboards`, `WebSocket /api/ws`

**Key behaviors**:
- Health summary bar updates every 5 seconds via WebSocket
- Quick action cards show counts and link to relevant views
- Active operations feed shows real-time progress of async actions
- Quick links only show dashboards that are currently available

#### View 2: Topology

The interactive service graph is the visual differentiator of the platform. It allows users to see every component, how they connect, and what is healthy or degraded.

**Layout**:

```text
+--------+---------------------------------------------------------+
|        |  Topology                                    [Filters]  |
| Side-  |                                                         |
| bar    |  +---------------------------------------------------+  |
|        |  |                                                   |  |
|        |  |    [k3d cluster]                                  |  |
|        |  |         |                                         |  |
|        |  |    [traefik] ──── [grafana]                        |  |
|        |  |         |              |                          |  |
|        |  |    [api-gateway] ── [prometheus]                   |  |
|        |  |      /       \                                    |  |
|        |  | [product-svc] [order-svc]                          |  |
|        |  |      |              |                             |  |
|        |  |   [redis]      [postgres]                          |  |
|        |  |                                                   |  |
|        |  +---------------------------------------------------+  |
|        |                                                         |
|        |  Detail Panel (appears on node click)                   |
|        |  +---------------------------------------------------+  |
|        |  | product-service          [healthy] [2/2 ready]     |  |
|        |  | Namespace: e-commerce    Image: sagars-lab/prod:v1 |  |
|        |  | [View Logs] [Inject Failure] [Open in Grafana]     |  |
|        |  +---------------------------------------------------+  |
+--------+---------------------------------------------------------+
```

**Data sources**: `GET /api/topology`, `GET /api/stacks/{name}`, `GET /api/status`

**Key behaviors**:
- Interactive pan/zoom graph using React Flow or D3-force layout
- Nodes represent apps, services, and platform components
- Edges represent dependency relationships from stack topology declarations
- Node colors: green = healthy, yellow = degraded, red = unhealthy, gray = unknown
- Clicking a node opens a detail panel with: replica count, health status, relevant dashboard links, failure mode controls, and log viewer link
- Filter controls: show/hide by type (app, service, platform), by stack, by scenario
- Active scenario overlay: highlights which nodes are participating in the current scenario
- Health data refreshes via polling every 10 seconds

#### View 3: Scenarios

The learning hub where users discover, activate, and work through guided scenarios.

**Layout**:

```text
+--------+---------------------------------------------------------+
|        |  Scenarios                                              |
| Side-  |                                                         |
| bar    |  [All] [Beginner] [Intermediate] [Advanced]  [Search]   |
|        |  [observability] [security] [delivery] [reliability]    |
|        |                                                         |
|        |  Card Grid (filterable by difficulty + category)         |
|        |  +-------------------+  +-------------------+           |
|        |  | Observability&SRE |  | Security&Comply   |           |
|        |  | ★★☆ Intermediate  |  | ★★☆ Intermediate  |           |
|        |  | Learn metrics,    |  | Policy, TLS,      |           |
|        |  | logs, traces      |  | network isolation |           |
|        |  | ✓ 2/3 checkpoints |  | Not started       |           |
|        |  | [Continue] [Info] |  | [Start] [Info]    |           |
|        |  +-------------------+  +-------------------+           |
|        |  +-------------------+  +-------------------+           |
|        |  | GitOps & CI/CD    |  | Chaos Engineering |           |
|        |  | ★★☆ Intermediate  |  | ★★★ Advanced      |           |
|        |  | ArgoCD, drift     |  | Pod kill, network |           |
|        |  | detection         |  | delay, CPU stress |           |
|        |  | Not started       |  | 🔒 Complete       |           |
|        |  |                   |  | intermediate first|           |
|        |  | [Start] [Info]    |  | [Locked] [Info]   |           |
|        |  +-------------------+  +-------------------+           |
|        |                                                         |
|        |  Learning Paths Tab                                     |
|        |  Platform Engineering Fundamentals                      |
|        |  [✓ Observability] → [● Security] → [○ GitOps] → [🔒]  |
|        |                                                         |
+--------+---------------------------------------------------------+
```

**Scenario Detail Modal** (opens on [Info] click):

```text
+----------------------------------------------------------+
| Observability & SRE                      [×]              |
| ★★☆ Intermediate | Category: observability               |
|                                                          |
| DESCRIPTION                                              |
| Full observability stack with log aggregation,           |
| distributed tracing, alerting rules, SLO dashboards.     |
|                                                          |
| LEARNING OBJECTIVES                                      |
| • Understand the three pillars of observability          |
| • Configure Prometheus alerting rules                    |
| • Build an SLO dashboard in Grafana                      |
| • Correlate a distributed trace across services          |
|                                                          |
| KEY CONCEPTS                                             |
| [SLO] [Error Budget] [Golden Signals] [Dist. Tracing]   |
|                                                          |
| PREREQUISITES                                           |
| ✓ ingress (installed)                                    |
| ✓ monitoring/metrics (installed)                         |
| ✗ microservices-demo stack (not deployed)                |
|                                                          |
| CHECKPOINTS                                              |
| ✓ 1. Verify Loki receiving logs from all services        |
| ✓ 2. Confirm Tempo shows multi-span traces               |
| ○ 3. Check alerting rules loaded in Prometheus           |
|                                                          |
| EXPLORE                                                  |
| [Grafana Dashboards ↗] [Prometheus ↗]                    |
|                                                          |
| COMMANDS                                                 |
| ┌─────────────────────────────────────────────┐          |
| │ Generate API traffic                    [📋] │          |
| │ labctl traffic start --profile steady       │          |
| │            --target go-api                  │          |
| └─────────────────────────────────────────────┘          |
|                                                          |
| REFLECTION                                               |
| • What changed in Grafana when you injected failure?     |
| • How would you set an error budget for /health?         |
|                                                          |
| [Validate Prerequisites] [Activate Scenario]             |
+----------------------------------------------------------+
```

**Data sources**: `GET /api/scenarios`, `GET /api/scenarios/{name}`, `POST /api/scenarios/{name}/validate`, `GET /api/progress`, `GET /api/learning-paths`

**Key behaviors**:
- Filter by difficulty level and category
- Search by keyword
- Progress indicators on each card (checkpoints completed / total)
- Learning paths tab shows ordered sequences with completion state
- Advanced scenarios show a lock icon until prerequisites (intermediate scenarios) are completed
- Detail modal shows full learning content, prerequisites with live status, and explore commands with copy-to-clipboard
- "Validate Prerequisites" button runs pre-flight checks and reports results inline

#### View 4: Stacks

Multi-service workload management view.

**Layout**:

```text
+--------+---------------------------------------------------------+
|        |  Stacks                                                 |
| Side-  |                                                         |
| bar    |  Card Grid                                              |
|        |  +-------------------+  +-------------------+           |
|        |  | E-Commerce        |  | Data Pipeline     |           |
|        |  | 4 apps, 2 services|  | 3 apps, 2 services|           |
|        |  | [topology preview]|  | [topology preview]|           |
|        |  |                   |  |                   |           |
|        |  | Status: Deployed  |  | Status: Available |           |
|        |  | [Details] [Down]  |  | [Details] [Up]    |           |
|        |  +-------------------+  +-------------------+           |
|        |                                                         |
+--------+---------------------------------------------------------+
```

**Stack Detail View** (opens on [Details] click):

```text
+----------------------------------------------------------+
| E-Commerce Platform                          [×]          |
| 4 apps | 2 services | Status: Deployed                   |
|                                                          |
| TOPOLOGY                                                 |
| (Interactive graph — same component as Topology view)    |
| frontend → api-gateway → product-service → redis         |
|                       → order-service → postgres          |
|                                                          |
| APPS                                                     |
| api-gateway    2/2 ready  [healthy]  [Logs] [Failure]    |
| product-svc    2/2 ready  [healthy]  [Logs] [Failure]    |
| order-service  2/2 ready  [healthy]  [Logs] [Failure]    |
| frontend       1/1 ready  [healthy]  [Logs]              |
|                                                          |
| SERVICES                                                 |
| redis          1/1 ready  [healthy]                      |
| postgres       1/1 ready  [healthy]                      |
|                                                          |
| HEALTH CHECK                                             |
| All endpoints passing (last checked 10s ago)             |
|                                                          |
| [Deploy Stack] [Tear Down] [View in Topology]            |
+----------------------------------------------------------+
```

**Data sources**: `GET /api/stacks`, `GET /api/stacks/{name}`, `POST /api/stacks/{name}/up`, `POST /api/stacks/{name}/down`

**Key behaviors**:
- Card shows a small topology preview (simplified graph thumbnail)
- Deploy/teardown actions with real-time progress in command output panel
- Detail view shows full interactive topology, constituent app/service health, and failure mode controls
- "View in Topology" navigates to the Topology view filtered to this stack

#### View 5: Platform

Component management organized by category.

**Layout**:

```text
+--------+---------------------------------------------------------+
|        |  Platform Components                                    |
| Side-  |                                                         |
| bar    |  [Install All] [Remove All]                             |
|        |                                                         |
|        |  INGRESS                                                |
|        |  traefik        [Running ●]  [Config ▾] [Remove]        |
|        |   └─ capabilities: http-routing, tls-termination        |
|        |   └─ dashboard: http://traefik.k3d.local ↗              |
|        |                                                         |
|        |  MONITORING                                             |
|        |  metrics (prom)  [Running ●]  [Config ▾] [Remove]       |
|        |  grafana         [Running ●]  [Config ▾] [Remove]       |
|        |   └─ dashboard: http://grafana.k3d.local ↗              |
|        |                                                         |
|        |  LOGGING                                                |
|        |  loki            [Stopped ○]  [Install]                 |
|        |                                                         |
|        |  TRACING                                                |
|        |  tempo           [Stopped ○]  [Install]                 |
|        |                                                         |
|        |  GITOPS                                                 |
|        |  argocd          [Stopped ○]  [Install]                 |
|        |                                                         |
|        |  SECURITY                                               |
|        |  kyverno         [Stopped ○]  [Install]                 |
|        |  cert-manager    [Stopped ○]  [Install]                 |
|        |                                                         |
|        |  CHAOS                                                  |
|        |  chaos-mesh      [Stopped ○]  [Install]                 |
|        |                                                         |
+--------+---------------------------------------------------------+
```

**Data sources**: `GET /api/platform`, `POST /api/platform/component/{category}/{name}/up`, `POST /api/platform/component/{category}/{name}/down`

**Key behaviors**:
- Grouped by category with collapsible sections
- Each component shows: provider name, install status, capabilities, related dashboard links
- Expandable config detail showing active values
- Install/remove actions with real-time progress
- "Install All" deploys the full platform stack in dependency order

#### View 6: Blueprints

One-click lab environment management.

**Layout**:

```text
+--------+---------------------------------------------------------+
|        |  Lab Blueprints                                         |
| Side-  |                                                         |
| bar    |  Card Grid                                              |
|        |  +-------------------+  +-------------------+           |
|        |  | SRE Training      |  | Security Workshop |           |
|        |  | Runtime: k3d      |  | Runtime: k3d      |           |
|        |  | Platform: 6 comp  |  | Platform: 4 comp  |           |
|        |  | Stacks: micro-demo|  | Stacks: api-gw    |           |
|        |  | Scenarios: 2      |  | Scenarios: 1      |           |
|        |  | Path: PE Fundmntl |  |                   |           |
|        |  |                   |  |                   |           |
|        |  | [Preview Plan]    |  | [Preview Plan]    |           |
|        |  | [Boot Lab]        |  | [Boot Lab]        |           |
|        |  +-------------------+  +-------------------+           |
|        |                                                         |
|        |  Active Lab: SRE Training                               |
|        |  +---------------------------------------------------+  |
|        |  | Status: Running | Uptime: 2h 15m                  |  |
|        |  | Platform: 6/6 healthy | Stack: deployed            |  |
|        |  | Scenarios: observability-sre (active)               |  |
|        |  | [Reset Lab] [Tear Down Lab]                         |  |
|        |  +---------------------------------------------------+  |
|        |                                                         |
+--------+---------------------------------------------------------+
```

**Execution Plan Modal** (opens on [Preview Plan]):

```text
+----------------------------------------------------------+
| Execution Plan: SRE Training                 [×]          |
|                                                          |
| The following actions will be performed in order:         |
|                                                          |
| 1. ○ Create k3d cluster "sagars-lab"                     |
| 2. ○ Install traefik (ingress)                           |
| 3. ○ Install prometheus (metrics)                        |
| 4. ○ Install grafana (dashboards)                        |
| 5. ○ Install loki (logging)                              |
| 6. ○ Install tempo (tracing)                             |
| 7. ○ Install chaos-mesh (chaos)                          |
| 8. ○ Deploy redis (service)                              |
| 9. ○ Deploy microservices-demo stack                     |
|    └ service-a, service-b, service-c, service-d, svc-e   |
| 10. ○ Activate observability-sre scenario                |
| 11. ○ Activate chaos-engineering scenario                |
|                                                          |
| Estimated components: 6 platform, 1 service, 5 apps,     |
| 2 scenarios                                              |
|                                                          |
| [Cancel] [Execute Plan]                                  |
+----------------------------------------------------------+
```

**Data sources**: `GET /api/blueprints`, `POST /api/blueprints/{name}/up`, `POST /api/blueprints/{name}/down`, `POST /api/blueprints/{name}/reset`, `POST /api/plan`

**Key behaviors**:
- Each blueprint card shows what it includes: runtime, platform components, stacks, scenarios, and learning path
- "Preview Plan" shows the full execution plan (dry-run) before committing
- "Boot Lab" executes the plan with real-time progress in the command output panel
- Active lab section shows health status and uptime
- "Reset Lab" performs a soft reset: tears down stacks and scenarios, preserves runtime and platform
- "Tear Down Lab" destroys everything including the runtime

### Shared UI Components

These components appear across all views:

| Component | Description | Behavior |
|-----------|-------------|----------|
| **Sidebar** | Persistent left navigation with icons for each view | Active view highlighted; collapses to icon-only on mobile |
| **Header** | App title, runtime selector dropdown, WebSocket connection indicator | Runtime selector triggers runtime switch; connection dot shows green/yellow/red |
| **Command Output Panel** | Bottom panel showing real-time output from async operations | Resizable via drag handle, collapsible, auto-scroll toggle, clear button |
| **Toast Notifications** | Floating notifications for operation results | Auto-dismiss for success/info (8s); persist for errors until manually dismissed |
| **Health Badge** | Colored status indicator used across all views | Green = healthy/running, yellow = degraded/pending, red = unhealthy/stopped, gray = unknown |
| **Topology Graph** | Reusable interactive graph component | Used in Topology view (full), Stack detail (filtered), and Blueprint preview (static) |

### API Endpoints

The web UI requires the following API endpoints. Existing endpoints are listed for completeness; new endpoints are marked.

#### Existing Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/api/status` | Overall lab status (cluster, platform, apps) |
| `GET` | `/api/apps` | List applications with deploy status |
| `POST` | `/api/apps/{name}/build` | Trigger app build |
| `POST` | `/api/apps/{name}/deploy` | Trigger app deploy |
| `POST` | `/api/apps/{name}/destroy` | Trigger app teardown |
| `GET` | `/api/platform` | Platform component status by category |
| `POST` | `/api/platform/up` | Install all platform components |
| `POST` | `/api/platform/down` | Remove all platform components |
| `POST` | `/api/platform/component/{category}/{name}/up` | Install specific component |
| `POST` | `/api/platform/component/{category}/{name}/down` | Remove specific component |
| `GET` | `/api/dashboards` | Available dashboard URLs |
| `GET` | `/api/scenarios` | List all scenarios with status |
| `GET` | `/api/scenarios/{name}` | Scenario detail with explore content |
| `POST` | `/api/scenarios/{name}/up` | Activate scenario |
| `POST` | `/api/scenarios/{name}/down` | Deactivate scenario |
| `GET` | `/api/services` | List available services |
| `POST` | `/api/services/{name}/up` | Install service |
| `POST` | `/api/services/{name}/down` | Remove service |
| `GET` | `/api/runtimes` | List available runtimes |
| `POST` | `/api/runtimes/{name}/activate` | Activate runtime |
| `POST` | `/api/runtimes/{name}/deactivate` | Deactivate runtime |
| `WS` | `/api/ws` | WebSocket for real-time status and action events |

#### New Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/api/topology` | Returns topology graph (nodes + edges) for all active resources |
| `GET` | `/api/stacks` | List available stacks with deploy status |
| `GET` | `/api/stacks/{name}` | Stack detail with topology and component health |
| `POST` | `/api/stacks/{name}/up` | Deploy stack |
| `POST` | `/api/stacks/{name}/down` | Tear down stack |
| `GET` | `/api/blueprints` | List available blueprints |
| `POST` | `/api/blueprints/{name}/up` | Boot lab from blueprint |
| `POST` | `/api/blueprints/{name}/down` | Tear down lab |
| `POST` | `/api/blueprints/{name}/reset` | Reset lab to known-good state |
| `GET` | `/api/progress` | User's scenario and learning path progress |
| `POST` | `/api/progress/{scenario}` | Update progress for a scenario |
| `POST` | `/api/scenarios/{name}/validate` | Pre-flight validation of scenario prerequisites |
| `GET` | `/api/learning-paths` | List learning paths with completion state |
| `POST` | `/api/plan` | Dry-run execution plan for a blueprint or action |

### UI Project Structure

```text
ui/
  src/
    api/              API client functions with TypeScript types
      client.ts       Base fetch wrapper with error handling
      status.ts       Status, cluster, platform API calls
      scenarios.ts    Scenario CRUD and validation
      stacks.ts       Stack lifecycle API calls
      blueprints.ts   Blueprint lifecycle API calls
      topology.ts     Topology graph data fetching
      progress.ts     Learning progress API calls
    components/       Shared UI components
      Sidebar.tsx     Navigation sidebar with route icons
      Header.tsx      Runtime selector, connection status
      LogPanel.tsx    Command output panel (resizable, collapsible)
      Notification.tsx Toast notification system
      HealthBadge.tsx  Status indicator component
      TopologyGraph.tsx Reusable React Flow graph component
      CardGrid.tsx    Reusable card layout component
      Modal.tsx       Reusable modal/dialog component
    hooks/            Custom React hooks
      useWebSocket.ts WebSocket connection with auto-reconnect
      useApi.ts       Data fetching hook with loading/error states
      useProgress.ts  Learning progress hook
    pages/            Page-level components
      Dashboard.tsx   Home view — health summary, quick actions
      Topology.tsx    Interactive topology graph view
      Scenarios.tsx   Scenario discovery and learning hub
      Stacks.tsx      Stack management view
      Platform.tsx    Platform component management
      Blueprints.tsx  Blueprint management and lab provisioning
    types/            TypeScript type definitions
      api.ts          Types matching Go API response structs
      topology.ts     Node, Edge, Graph types
      scenario.ts     Scenario, Learning, Progress types
    App.tsx           Root component with router and layout
    main.tsx          Entry point
  public/
    favicon.svg
  index.html
  vite.config.ts      Vite config with API proxy for dev
  tailwind.config.js  Tailwind with dark theme tokens
  tsconfig.json
  package.json
```

### UI Design Principles

1. **Information hierarchy**: Dashboard shows summary; detail views show depth. Users should never feel overwhelmed on the landing page.
2. **Progressive disclosure**: Beginners see simple cards and guided actions. Advanced users access topology views, failure injection controls, and execution plans.
3. **Real-time feedback**: Every async operation (deploy, install, activate) shows immediate progress via WebSocket events and the command output panel.
4. **Context preservation**: Navigating between views should not lose state. Active operations continue and are visible from any view via the command output panel.
5. **Teachable interface**: The UI itself should teach — scenario cards explain concepts, blueprint previews show execution order, topology views show how systems connect.

---

## Repository Mapping

The repository should align to the architecture like this:

```text
apps/                  Sample workloads for learning and experiments
blueprints/            Lab environment blueprints (one-command provisioning)
bootstrap/             Developer and environment setup
cmd/labctl/            CLI, API, internal control-plane orchestration
delivery/              CI/CD workflow definitions
docs/                  Architecture, operational guides, contributor references
engine/                Build and deploy strategy dispatchers
foundation/terraform/  Cloud and infrastructure provisioning
learning-paths/        Ordered scenario sequences for structured learning
platform/              Platform capability modules
runtimes/              Runtime lifecycle adapters
scenarios/             Guided learning and failure simulation packages
services/              Reusable service modules
stacks/                Multi-service composable workload definitions
ui/                    React + Vite web UI application
agents/                Agent role definitions and collaboration prompts
```

### Important Contributor Rule

Contributors should prefer extending the layer that owns a capability instead of bypassing it.

Examples:

- a new ingress provider belongs under `platform/ingress/`
- a new runtime belongs under `runtimes/`
- a new learning experience belongs under `scenarios/`
- a new sample workload belongs under `apps/`
- a new multi-service composition belongs under `stacks/`
- a new lab environment definition belongs under `blueprints/`
- a new learning path belongs under `learning-paths/`

---

## Current State vs Desired State

This document intentionally describes the **desired state** of the product, but contributors need to know what exists now.

### Implemented today

- runtime lifecycle for `k3d`, `AKS`, and `EKS`
- platform component installation through provider directories and scripts
- reusable services with lifecycle scripts (Redis)
- sample applications with Helm deployment (`go-api`, `echo-server`)
- scenario discovery and activation with four scenarios
- CLI and embedded UI entry points
- HTTP API with status, platform, apps, scenarios, services, runtimes, and dashboards endpoints
- WebSocket for real-time action events
- Terraform modules for cloud runtimes
- CI/CD workflows for lint, test, build, deploy

### Not yet fully realized

- topology as a first-class subsystem with graph data model and UI visualization
- stacks as multi-service composable workloads
- lab blueprints for one-command environment provisioning
- learning domain with difficulty levels, progress tracking, and guided walkthroughs
- learning paths as ordered scenario sequences
- scenario pre-flight validation with health checks
- failure mode library standardized across all apps
- traffic generator workload
- `labctl plan` dry-run execution mode
- React-based web UI consuming the full API surface
- `_interface.yaml` standardized across all platform categories
- GKE runtime support
- stack-specific applications (e-commerce, data-pipeline, etc.)
- several new API endpoints (topology, stacks, blueprints, progress, learning paths, plan)
- deeper contract standardization across all adapters
- stronger dependency graphing across apps, services, and scenarios

Contributors should not invent undocumented abstractions casually. If a new resource type or subsystem is needed, it should be added deliberately and documented here or in a decision record.

---

## Architecture Rules for Contributors and Agents

When modifying this repository, contributors and agents should follow these rules:

1. Keep runtime, foundation, platform, service, application, stack, scenario, and blueprint concerns separate.
2. Prefer declarative configuration over hard-coded behavior.
3. Do not couple the UI directly to shell scripts or infrastructure commands.
4. Add new capabilities through existing extension points before creating new top-level structures.
5. Document new resource types, lifecycle contracts, and cross-layer dependencies.
6. Treat scenarios as guided platform exercises, not just bundles of random YAML.
7. Preserve the educational value of the system. A feature is better when it teaches.
8. Multi-service scenarios should reference stacks in `prerequisites.stacks[]`, not ad-hoc app lists.
9. New sample applications must implement the Failure Mode Library API contract (`/debug/failure/*`).
10. UI development follows React + Vite + TypeScript conventions. The UI must call the HTTP API only and never execute shell scripts directly.

---

## Quality Attributes

The architecture should be evaluated against these quality attributes:

- **Clarity:** contributors can find the right place to make a change
- **Portability:** the same concepts work across local and cloud runtimes
- **Safety:** users can reset or destroy environments predictably
- **Extensibility:** new providers, scenarios, stacks, and blueprints can be added without refactoring the whole system
- **Observability:** users can inspect what the system is doing and why
- **Educational Value:** features reinforce platform engineering and SRE concepts
- **Progressive Disclosure:** beginners see simple views and guided actions; advanced users access topology, chaos controls, and execution plans
- **Visual Debugging:** users can visually trace issues through interactive topology graphs and correlated dashboards

---

## Recommended Next Architectural Documents

This document is the north star, but it is not enough on its own for parallel multi-agent work. The repository should also maintain:

1. **System Context**
   A short document showing the external actors, cloud dependencies, local tooling, and major system boundaries.

2. **Domain Model**
   Definitions for Runtime, Platform Component, Service, Application, Stack, Scenario, Blueprint, Learning Path, Session, and Topology, including lifecycle and ownership.

3. **Contributor Map**
   A file that explains which directories own which concerns, what contracts exist, and where new work should go.

4. **Decision Records**
   Small ADR-style documents for major architectural choices such as scenario format, state storage, topology design, stack format, blueprint format, and provider contracts.

5. **Agent Operating Guide**
   A guide for AI agents describing workflow, task boundaries, file ownership, coordination rules, and handoff expectations.

6. **Execution Contracts**
   Contracts for `install.sh`, `uninstall.sh`, `status.sh`, `_interface.yaml`, scenario manifests, stack manifests, blueprint manifests, runtime adapters, and app failure mode APIs.

7. **Roadmap / Capability Matrix**
   A status document showing what is implemented, planned, experimental, or intentionally deferred.

8. **UI Component Specification**
   Detailed component specs, wireframes, and interaction patterns for each UI view. This document provides the high-level design; a detailed spec should cover edge cases, error states, responsive breakpoints, and accessibility requirements.

---

## Summary

Sagars-Laboratory should evolve into a scenario-driven platform engineering simulator with a clear control plane, pluggable execution adapters, strong learning workflows, multi-service stacks, one-command lab blueprints, an interactive topology-aware web UI, and explicit architectural boundaries.

The platform serves three audiences: Platform Engineers who need topology views and pluggable provider interfaces, Platform Teams who need one-command lab environments and reset capabilities, and DevOps/SRE Learners who need guided learning paths with progressive difficulty and progress tracking.

If contributors align their changes to the domains and rules in this document, the repository will remain understandable, extensible, and much easier for multiple agents to evolve in parallel.
