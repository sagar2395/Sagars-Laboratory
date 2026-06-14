# Publishing Scenario Packs (OCI)

A pack can be distributed two ways:

| Encoding | Identity | Integrity | Auth gating | Use it for |
|---|---|---|---|---|
| **git** | repo URL `[@ref]` | commit | repo visibility | the simple community path |
| **OCI** | `oci://registry/repo[:tag]` | content digest (+ optional cosign) | registry auth (401 without a token) | versioned, signed, gated packs |

OCI is the same mechanism that later gates premium/enterprise packs: the engine
runs every tier identically, so access control lives entirely at the
distribution layer (registry auth) and the entitlement interface (task 070) —
never in the engine.

> The CLI does not embed an OCI client or sigstore. It shells out to **`oras`**
> and **`cosign`** (golden rule #2: wrap tools, don't reimplement them). Install
> both to publish or to install signed packs:
> [oras.land](https://oras.land), [docs.sigstore.dev](https://docs.sigstore.dev).

## Artifact format

`labctl pack publish` deterministically tars the pack directory into a single
gzipped layer and pushes it as an OCI artifact:

- **artifactType:** `application/vnd.flightdeck.pack.v1+json`
- **layer media type:** `application/vnd.flightdeck.pack.layer.v1.tar+gzip`
- **layer title:** `pack.tar.gz`

The tar is reproducible (sorted entries, normalized metadata), so the same pack
contents always yield the same digest regardless of who builds it or when.
`*.sh` files keep their executable bit so checks/inject scripts run after pull.

## Publish

```bash
# Unsigned (community)
labctl pack publish ./packs/examples/hello-pack oci://ghcr.io/snowops/hello-pack:0.1.0

# Signed — keyless (OIDC, e.g. in CI) or with a key pair
labctl pack publish ./pack oci://ghcr.io/snowops/hello-pack:0.1.0 --sign
labctl pack publish ./pack oci://ghcr.io/snowops/hello-pack:0.1.0 --sign --cosign-key cosign.key
```

The directory must contain a valid `pack.yaml` (it is validated before anything
is pushed). The command prints the layer digest, which you can pin downstream.

First-party community packs under `packs/examples/` are published automatically
to GHCR and keyless-signed on every GitHub release by
[`.github/workflows/pack-publish.yml`](../../.github/workflows/pack-publish.yml).

## Install

```bash
# By tag or by digest (immutable)
labctl pack add oci://ghcr.io/snowops/hello-pack:0.1.0
labctl pack add oci://ghcr.io/snowops/hello-pack@sha256:<digest>

# Require a valid signature before any content is extracted (fail closed)
labctl pack add oci://ghcr.io/snowops/hello-pack:0.1.0 \
  --require-signature \
  --certificate-identity 'https://github.com/snowops/flightdeck/.github/workflows/pack-publish.yml@refs/tags/v0.1.0' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com

# Or verify against a fixed public key
labctl pack add oci://ghcr.io/snowops/hello-pack:0.1.0 --require-signature --cosign-key cosign.pub
```

### Verification, fail-closed

On install, in order:

1. **Signature** — when `--require-signature` is set, `cosign verify` runs
   **before** anything is written. Unsigned or invalid → install aborts. Keyless
   verification requires both `--certificate-identity` and
   `--certificate-oidc-issuer`. Community packs omit the flag.
2. **Digest** — `oras pull` verifies each layer against the manifest digest, so a
   tampered layer fails the transfer.
3. **Extraction** — the tarball is extracted with path-traversal ("zip-slip")
   protection; symlinks and escaping paths are refused. Only on full success is
   the pack atomically renamed into the catalog — a failed install leaves nothing
   behind.

> **Security:** packs run scripts and apply manifests on your cluster. Treat an
> unsigned pack like any untrusted code. Prefer signed packs and pin by digest.

## Roadmap

- **069** — a searchable registry index so `labctl pack search` finds packs
  without knowing the exact ref.
- **070** — the entitlement interface that turns `--require-signature` and
  registry 401s into a policy-driven gate for premium/enterprise tiers.
