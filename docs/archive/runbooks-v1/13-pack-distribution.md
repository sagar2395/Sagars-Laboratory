# Runbook 13 — Pack Distribution over OCI (M7)

Publish a scenario pack to an OCI registry as a signed, content-addressed
artifact and install it back, verifying the signature fails closed on tampering —
then discover packs through the registry index. This covers tasks **068**
(OCI distribution + signing) and **069** (registry index + discovery) of
milestone **M7 — OSS & Ecosystem Foundation**.

## Prereqs

- `bin/labctl` built locally (`make cli-build`).
- [`oras`](https://oras.land) on your `PATH` (push/pull).
- [`cosign`](https://docs.sigstore.dev) on your `PATH` (signing/verification).
- A registry you can push to. For a fully local loop:
  ```bash
  docker run -d --rm -p 5001:5000 --name pack-registry registry:2
  ```
- A valid pack to publish, e.g. the bundled example `packs/examples/hello-pack/`.

## Steps

### 1. Publish (unsigned) to a local registry

```bash
bin/labctl pack publish packs/examples/hello-pack oci://localhost:5001/hello-pack:0.1.0
```

**Expected:** the command validates `pack.yaml`, archives the pack, pushes it,
and prints a `layer digest: sha256:...`.

### 2. Install it back

```bash
bin/labctl pack add oci://localhost:5001/hello-pack:0.1.0
bin/labctl pack list
```

**Expected:** `pack list` shows `hello-pack` with its version and tier, and the
`hello-flightdeck` scenario it provides. Install by digest works too:

```bash
bin/labctl pack add --force oci://localhost:5001/hello-pack@sha256:<digest-from-step-1>
```

### 3. Sign and verify (keyless or key pair)

Generate a key pair and publish signed:

```bash
cosign generate-key-pair                       # creates cosign.key / cosign.pub
bin/labctl pack publish packs/examples/hello-pack \
  oci://localhost:5001/hello-pack:0.2.0 --sign --cosign-key cosign.key
```

Install with verification required:

```bash
bin/labctl pack add --force \
  oci://localhost:5001/hello-pack:0.2.0 \
  --require-signature --cosign-key cosign.pub
```

**Expected:** `cosign verify` runs **before** extraction; install succeeds only
because the signature is valid.

### 4. Confirm fail-closed behavior

- **Unsigned + required:** add an *unsigned* ref with `--require-signature` →
  install aborts with a verification error and nothing lands in the catalog.
- **Wrong key:** verify a signed pack with a different `--cosign-key` → aborts.
- **Tamper:** `oras pull` checks each layer digest against the manifest, so a
  mutated layer fails the transfer. Extraction also refuses any tar entry whose
  path escapes the destination (zip-slip) or any non-regular file.

```bash
# Negative check — expect a non-zero exit and "verification failed"
bin/labctl pack add --force oci://localhost:5001/hello-pack:0.1.0 --require-signature --cosign-key cosign.pub
```

### 5. Discover via the registry index (task 069)

Point the CLI at the seed index (a local fixture works) and search/resolve:

```bash
export PACK_REGISTRY_INDEX="file://$(pwd)/registry/index.json"

bin/labctl pack search                 # list every pack
bin/labctl pack search hello           # filter by term
bin/labctl pack info hello-pack        # registry entry (not installed yet)
bin/labctl pack validate-index registry/index.json   # the PR-gate check
```

**Expected:** `search` prints NAME/VERSION/TIER/VERIFIED/DESCRIPTION; `info`
shows the entry with its resolvable `ref`; `validate-index` reports the index is
valid. `bin/labctl pack add hello-pack` would resolve the name to its `ref` and
install it (needs the OCI ref reachable / `oras` installed). The index caches to
`.labctl/cache/registry-index.json` with a 1h TTL; a failed refresh falls back to
the cache with a warning.

## Expected (acceptance)

- A pack pushes to and pulls from a registry by tag **and** by digest.
- Checksum/digest integrity holds; tampering fails the install.
- `--require-signature` enforces a valid cosign signature before any content is
  written; unsigned/invalid → fail closed.
- Nothing is registry-vendor specific — `oci://<any-registry>/...` works.

## Cleanup

```bash
bin/labctl pack remove hello-pack
docker rm -f pack-registry 2>/dev/null || true
rm -f cosign.key cosign.pub
```

## Troubleshooting

- **`oras not found on PATH` / `cosign not found`** — install the binary; the CLI
  intentionally wraps these tools rather than embedding them.
- **`401 Unauthorized`** — `oras login <registry>` first (this is the same gate
  that will protect premium packs once the entitlement interface, task 070,
  lands).
- **Keyless verify** — supply both `--certificate-identity` and
  `--certificate-oidc-issuer`; without them keyless verification refuses to run.
- **First-party packs** publish automatically to GHCR on release via
  `.github/workflows/pack-publish.yml`.
