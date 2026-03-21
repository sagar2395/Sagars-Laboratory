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

- provision and switch between runtimes such as `k3d`, `AKS`, and `EKS`
- install opinionated platform capabilities such as ingress, metrics, logs, traces, GitOps, policy, TLS, and chaos tooling
- deploy reusable services and sample workloads
- activate guided scenarios that teach real platform and SRE workflows
- observe the system through dashboards, logs, traces, and topology views
- break things safely and recover quickly
- reset the environment to a known-good state

---

## Target Users

### 1. Platform Engineers

Users who want to design and validate internal platform capabilities such as runtime provisioning, service enablement, observability, policy, and GitOps.

### 2. Platform Teams

Teams who need a shared sandbox for architecture spikes, golden path experiments, onboarding, demos, and operational training.

### 3. DevOps and SRE Learners

Individuals who want a practical, hands-on environment to understand Kubernetes, cloud runtimes, observability, reliability, security controls, and failure handling.

---

## Architectural Goals

The architecture must optimize for the following:

1. **Learnability**
   A new user should be able to understand the platform model and get value quickly.

2. **Reproducibility**
   A lab environment should be rebuildable, resettable, and deterministic.

3. **Composability**
   Runtimes, platform components, services, apps, and scenarios should be mix-and-match.

4. **Operational Realism**
   The system should feel like a small but credible platform engineering environment, not a toy demo.

5. **Safe Failure**
   Users should be able to introduce faults, drift, and policy violations without damaging their local machine or cloud environment beyond the declared lab boundary.

6. **Contributor Clarity**
   The architecture must make ownership boundaries obvious so multiple agents can work in parallel with low coordination overhead.

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

Each layer should have one job and one reason to change. Runtime provisioning, platform installation, service deployment, app deployment, and scenario activation must remain separate concerns.

### 3. Idempotent Operations

Repeated `up`, `down`, `install`, `uninstall`, and `status` actions should be safe and predictable.

### 4. Pluggable Providers

Equivalent capabilities should be swappable. Examples:

- `traefik` or `nginx` for ingress
- `k3d`, `AKS`, or `EKS` for runtime
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
  +-- Scenario Engine
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

- **Intent** is expressed in configuration, scenario manifests, values files, and future blueprint definitions.
- **Orchestration** is handled by `labctl` and supporting internal packages.
- **Execution** is performed through scripts, Helm charts, Kubernetes manifests, Terraform, and cloud tooling.

---

## Core Domains

### 1. Experience Layer

The experience layer includes:

- CLI: `labctl`
- Embedded web UI
- HTTP API used by the UI

Its responsibilities are:

- accept user intent
- present current state
- surface guided workflows
- explain what will happen before changes are applied

It must not contain infrastructure logic directly. It should call internal orchestration services.

### 2. Control Plane

The control plane is the brain of the system and should remain the most stable architectural layer.

Its responsibilities are:

- discover available runtimes, platform components, services, apps, and scenarios
- validate prerequisites
- orchestrate install and teardown order
- maintain lightweight state for what is active
- expose a consistent contract to both CLI and UI

Today this responsibility lives primarily inside `cmd/labctl/internal/`.

### 3. Runtime Domain

The runtime domain provides the execution substrate where workloads run.

Examples:

- local Kubernetes via `k3d`
- managed Kubernetes via `AKS`
- managed Kubernetes via `EKS`

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

Repository mapping:

- `platform/`

### 6. Service Domain

The service domain represents reusable dependencies consumed by workloads and scenarios.

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

- `go-api`
- `echo-server`

Responsibilities:

- provide realistic workloads with health endpoints and platform integration hooks
- expose telemetry and failure modes useful for observability, GitOps, policy, and chaos scenarios
- remain small enough to teach from

Repository mapping:

- `apps/`
- `engine/`

### 8. Scenario Domain

The scenario domain is the core teaching mechanism of the platform.

A scenario is a curated learning or validation experience that composes platform capabilities, services, applications, configuration, and exploration guidance around a real theme.

Examples:

- observability and SRE
- GitOps and CI/CD
- security and compliance
- chaos engineering

Responsibilities:

- declare prerequisites
- install additional components in the right order
- expose guided exploration steps
- make complex concepts easy to activate repeatedly

Repository mapping:

- `scenarios/`
- `cmd/labctl/internal/scenario/`

### 9. Topology and Insight Domain

The topology and insight domain explains what is running, how components relate, and what is degraded.

Desired-state responsibilities:

- generate a topology graph from runtime, platform, service, app, and scenario data
- show health and dependency relationships
- support investigation workflows for SRE-style debugging

Current-state note:

This domain is conceptually important but only partially implemented today. Contributors should treat it as a target capability, not as a fully realized subsystem.

---

## Desired Resource Model

The desired state of the application should revolve around these first-class concepts:

| Resource | Meaning |
|----------|---------|
| Runtime | Execution environment such as `k3d`, `AKS`, or `EKS` |
| Foundation | Underlying cloud and infrastructure resources |
| Platform Component | A reusable capability such as ingress, metrics, tracing, policy, or chaos |
| Service | A reusable backing dependency such as Redis |
| Application | A sample workload deployed into the lab |
| Scenario | A guided experience that composes components and teaches a concept |
| Session | The currently active lab state for a user |
| Topology | A computed view of relationships, health, and dependencies |

Future capabilities such as `Lab Blueprints` or `Stack Blueprints` should be introduced as explicit resources only when they have a real execution model and repository representation.

---

## Control Plane Responsibilities

The control plane should eventually standardize the following workflow:

1. Discover available resources from the repository
2. Validate compatibility and prerequisites
3. Resolve desired state into an ordered execution plan
4. Execute through adapters
5. Capture resulting state
6. Present health, status, and exploration guidance back to the user

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

---

## Repository Mapping

The repository should continue to align to the architecture like this:

```text
apps/                  Sample workloads for learning and experiments
bootstrap/             Developer and environment setup
cmd/labctl/            CLI, API, internal control-plane orchestration
delivery/              CI/CD workflow definitions
docs/                  Architecture, operational guides, contributor references
engine/                Build and deploy strategy dispatchers
foundation/terraform/  Cloud and infrastructure provisioning
platform/              Platform capability modules
runtimes/              Runtime lifecycle adapters
scenarios/             Guided learning and failure simulation packages
services/              Reusable service modules
ui/                    Web UI assets
agents/                Agent role definitions and collaboration prompts
```

### Important Contributor Rule

Contributors should prefer extending the layer that owns a capability instead of bypassing it.

Examples:

- a new ingress provider belongs under `platform/ingress/`
- a new runtime belongs under `runtimes/`
- a new learning experience belongs under `scenarios/`
- a new sample workload belongs under `apps/`

---

## Current State vs Desired State

This document intentionally describes the **desired state** of the product, but contributors need to know what exists now.

### Implemented today

- runtime lifecycle for `k3d`, `AKS`, and `EKS`
- platform component installation through provider directories and scripts
- reusable services with lifecycle scripts
- sample applications with Helm deployment
- scenario discovery and activation
- CLI and embedded UI entry points
- Terraform modules for cloud runtimes

### Not yet fully realized

- topology as a first-class subsystem
- richer persisted session state
- formal blueprint resources for reusable lab compositions
- deeper contract standardization across all adapters
- stronger dependency graphing across apps, services, and scenarios

Contributors should not invent undocumented abstractions casually. If a new resource type or subsystem is needed, it should be added deliberately and documented here or in a decision record.

---

## Architecture Rules for Contributors and Agents

When modifying this repository, contributors and agents should follow these rules:

1. Keep runtime, foundation, platform, service, application, and scenario concerns separate.
2. Prefer declarative configuration over hard-coded behavior.
3. Do not couple the UI directly to shell scripts or infrastructure commands.
4. Add new capabilities through existing extension points before creating new top-level structures.
5. Document new resource types, lifecycle contracts, and cross-layer dependencies.
6. Treat scenarios as guided platform exercises, not just bundles of random YAML.
7. Preserve the educational value of the system. A feature is better when it teaches.

---

## Quality Attributes

The architecture should be evaluated against these quality attributes:

- **Clarity:** contributors can find the right place to make a change
- **Portability:** the same concepts work across local and cloud runtimes
- **Safety:** users can reset or destroy environments predictably
- **Extensibility:** new providers and scenarios can be added without refactoring the whole system
- **Observability:** users can inspect what the system is doing and why
- **Educational Value:** features reinforce platform engineering and SRE concepts

---

## Recommended Next Architectural Documents

This document is the north star, but it is not enough on its own for parallel multi-agent work. The repository should also maintain:

1. **System Context**
   A short document showing the external actors, cloud dependencies, local tooling, and major system boundaries.

2. **Domain Model**
   Definitions for Runtime, Platform Component, Service, Application, Scenario, Session, and Topology, including lifecycle and ownership.

3. **Contributor Map**
   A file that explains which directories own which concerns, what contracts exist, and where new work should go.

4. **Decision Records**
   Small ADR-style documents for major architectural choices such as scenario format, state storage, topology design, and provider contracts.

5. **Agent Operating Guide**
   A guide for AI agents describing workflow, task boundaries, file ownership, coordination rules, and handoff expectations.

6. **Execution Contracts**
   Contracts for `install.sh`, `uninstall.sh`, `status.sh`, scenario manifests, runtime adapters, and future blueprint definitions.

7. **Roadmap / Capability Matrix**
   A status document showing what is implemented, planned, experimental, or intentionally deferred.

---

## Summary

Sagars-Laboratory should evolve into a scenario-driven platform engineering simulator with a clear control plane, pluggable execution adapters, strong learning workflows, and explicit architectural boundaries.

If contributors align their changes to the domains and rules in this document, the repository will remain understandable, extensible, and much easier for multiple agents to evolve in parallel.
