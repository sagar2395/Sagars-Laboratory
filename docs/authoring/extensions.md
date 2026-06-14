# Extension Seams (entitlement, resolvers, hooks)

Flightdeck is designed so premium content and hosted/SaaS builds plug in
**without forking the engine**. The open core ships only interfaces and
allow-everything/no-op defaults; richer implementations are *injected at
construction* and live in a private repo. Building these seams early is the key
anti-lock-in decision (task 070) — no premium or business logic ever enters the
OSS engine.

These packages are part of the public SDK and are CODEOWNERS-locked to the lead
maintainer:

- [`pkg/entitlement`](../../cmd/labctl/pkg/entitlement/) — who may use a pack.
- [`pkg/extension`](../../cmd/labctl/pkg/extension/) — where a pack comes from
  (resolvers) and optional lifecycle hooks.

## Entitlement

```go
type Entitlement interface {
    Authorize(ctx context.Context, req Request) error // nil = allowed
}
```

The OSS default is `entitlement.AllowAll` (`entitlement.Default()`): every pack
is permitted, every tier runs identically. The engine consults it on the
install path (`installFetched`), so the seam is *exercised* in OSS but changes
nothing.

A premium build injects its own implementation:

```go
engine := scenario.NewEngine(root, domain, profile)
engine.Entitlement = mylicense.New(token) // private impl; gates premium/enterprise
```

A denial should return an `*entitlement.DeniedError` (use `entitlement.Denied`)
so callers can distinguish policy from operational failures. Example test-double
(see `pkg/entitlement/entitlement_test.go`):

```go
type denyTier struct{ hasToken bool }
func (d denyTier) Authorize(_ context.Context, req entitlement.Request) error {
    if req.Tier == "" || req.Tier == "community" || d.hasToken {
        return nil
    }
    return entitlement.Denied(req.Name, "tier "+req.Tier+" requires a license token")
}
```

License tokens themselves are verified through the `TokenVerifier` *seam* — the
OSS core ships the interface only, never a concrete verifier.

## Resolvers — where a pack comes from

```go
type Resolver interface {
    CanResolve(ref string) bool
    Resolve(ctx context.Context, ref, destDir string) error
}
```

Built-in open resolvers: `OCIResolver` (oci://, via oras/cosign), `GitResolver`
(git URLs), `LocalResolver` (dirs / file://). `DefaultResolver(opts)` returns the
chain `OCI → git → local`. The engine installs through the seam with
`engine.InstallVia(ctx, resolver, ref, name, force)`, which resolves into a temp
dir then runs the shared validate → entitle → rename install path.

A private/hosted source is just another resolver added to a `Chain` — no engine
change:

```go
chain := append(extension.DefaultResolver(opts), myHostedCatalogResolver{})
engine.InstallVia(ctx, chain, "hub://acme/kafka-drills", "kafka-drills", false)
```

## Lifecycle hooks

```go
type Hooks interface {
    PreStage(ctx, Event) error;  PostStage(ctx, Event) error
    PreCheck(ctx, Event) error;  PostCheck(ctx, Event) error
}
```

The OSS default is `extension.NoopHooks` (`extension.DefaultHooks()`); the engine
invokes them around each scenario stage (`Up`) and around checks (`Verify`). A
`Pre*` hook returning an error aborts the phase, so premium policy can fail
closed; the no-op default never errors and never changes behavior.

```go
engine.Hooks = myTelemetryHooks{} // record stage timings, gate checks, etc.
```

Hooks are advisory and data/script-driven by design — Flightdeck never loads
compiled Go plugins.

## Rules

- OSS defaults must be **open** (allow-all / no-op) and behavior-preserving.
- Premium/business logic lives in a **private** repo and is injected — never
  compiled into the OSS binary.
- These interfaces are **stable SDK**; changes go through an RFC
  (`docs/rfcs/`) and the lead maintainer (CODEOWNERS).
