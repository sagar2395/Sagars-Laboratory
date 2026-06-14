// SPDX-License-Identifier: Apache-2.0

package pack

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// OCI distribution (task 068). A pack can be published to and installed from any
// OCI registry alongside the git source. OCI gives versioning, content
// addressing, and signature/auth gating — the same mechanism that later gates
// premium/enterprise packs (a registry returns 401 without a license token).
//
// In keeping with golden rule #2 ("the CLI wraps tools, it does not reimplement
// them") this package does not vendor an OCI client or sigstore; it shells out
// to the `oras` and `cosign` binaries through injectable Runner funcs, exactly
// as the catalog shells out to `git`. Tests inject stubs; production uses the
// DefaultOras / DefaultCosign runners. oras itself verifies layer digests
// against the manifest on pull, so transport tampering already fails closed;
// cosign adds publisher provenance on top.

// OCIScheme is the prefix that marks a pack reference as an OCI registry ref.
const OCIScheme = "oci://"

// Runner executes an external CLI (oras or cosign). dir, when non-empty, is the
// working directory for the command; args are passed verbatim. Returning an
// error means the tool failed or is missing.
type Runner func(ctx context.Context, dir string, args ...string) error

// IsOCIRef reports whether src is an OCI reference (oci://registry/repo[:tag]).
func IsOCIRef(src string) bool {
	return strings.HasPrefix(src, OCIScheme)
}

// TrimOCIScheme removes the oci:// prefix; oras takes a bare registry reference.
func TrimOCIScheme(src string) string {
	return strings.TrimPrefix(src, OCIScheme)
}

// PackNameFromOCIRef derives a default catalog name from an OCI ref — the
// repository's last path segment, minus any tag or digest.
func PackNameFromOCIRef(ref string) string {
	r := TrimOCIScheme(ref)
	// Strip digest, then tag.
	if i := strings.IndexByte(r, '@'); i >= 0 {
		r = r[:i]
	}
	if i := strings.LastIndexByte(r, ':'); i > strings.LastIndexByte(r, '/') {
		r = r[:i]
	}
	if i := strings.LastIndexByte(r, '/'); i >= 0 {
		r = r[i+1:]
	}
	return r
}

// DefaultOras runs the real `oras` binary, streaming its output.
func DefaultOras(ctx context.Context, dir string, args ...string) error {
	return runTool(ctx, dir, "oras", args...)
}

// DefaultCosign runs the real `cosign` binary, streaming its output.
func DefaultCosign(ctx context.Context, dir string, args ...string) error {
	return runTool(ctx, dir, "cosign", args...)
}

func runTool(ctx context.Context, dir, name string, args ...string) error {
	if _, err := exec.LookPath(name); err != nil {
		return fmt.Errorf("%s not found on PATH — install it to use OCI packs (https://oras.land, https://docs.sigstore.dev): %w", name, err)
	}
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return nil
}

// PublishOptions configures Publish.
type PublishOptions struct {
	// Oras runs the oras CLI; defaults to DefaultOras when nil.
	Oras Runner
	// Cosign runs the cosign CLI; defaults to DefaultCosign when nil.
	Cosign Runner
	// Sign, when true, signs the pushed artifact with cosign after the push.
	Sign bool
	// CosignKey is the path to a cosign private key. Empty means keyless
	// (OIDC) signing — what CI uses on GitHub via the ambient token.
	CosignKey string
}

// Publish packages the pack in dir into a single tarball layer and pushes it to
// ref as an OCI artifact. It validates the pack.yaml manifest first so nothing
// malformed reaches a registry, and optionally signs the result with cosign.
// It returns the sha256 digest of the tarball layer.
func Publish(ctx context.Context, dir, ref string, opts PublishOptions) (string, error) {
	oras := opts.Oras
	if oras == nil {
		oras = DefaultOras
	}

	m, err := Load(dir)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", ManifestFile, err)
	}
	if m == nil {
		return "", fmt.Errorf("%s not found in %s: a manifest is required to publish", ManifestFile, dir)
	}
	if err := m.Validate(); err != nil {
		return "", err
	}

	stage, err := os.MkdirTemp("", "labctl-pack-publish-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(stage)

	tarball := filepath.Join(stage, LayerFileName)
	f, err := os.Create(tarball)
	if err != nil {
		return "", err
	}
	if err := TarGz(dir, f); err != nil {
		f.Close()
		return "", fmt.Errorf("archiving pack: %w", err)
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	digest, err := Sha256File(tarball)
	if err != nil {
		return "", err
	}

	target := TrimOCIScheme(ref)
	// Run from the stage dir so the layer's title annotation is just the
	// filename (LayerFileName), not an absolute path.
	args := []string{
		"push", target,
		"--artifact-type", ArtifactType,
		fmt.Sprintf("%s:%s", LayerFileName, LayerMediaType),
	}
	if err := oras(ctx, stage, args...); err != nil {
		return "", fmt.Errorf("pushing pack to %s: %w", target, err)
	}

	if opts.Sign {
		cosign := opts.Cosign
		if cosign == nil {
			cosign = DefaultCosign
		}
		signArgs := []string{"sign", "--yes"}
		if opts.CosignKey != "" {
			signArgs = append(signArgs, "--key", opts.CosignKey)
		}
		signArgs = append(signArgs, target)
		if err := cosign(ctx, "", signArgs...); err != nil {
			return "", fmt.Errorf("signing %s: %w", target, err)
		}
	}

	return digest, nil
}

// PullOptions configures Pull.
type PullOptions struct {
	// Oras runs the oras CLI; defaults to DefaultOras when nil.
	Oras Runner
	// Cosign runs the cosign CLI; defaults to DefaultCosign when nil.
	Cosign Runner
	// RequireSignature, when true, runs `cosign verify` before extracting and
	// fails closed if the artifact is unsigned or the signature is invalid.
	// Community packs leave this false; "verified"/premium packs set it true
	// (the entitlement interface, task 070, will drive this from policy).
	RequireSignature bool
	// CosignKey verifies against a fixed public key. Empty uses keyless
	// verification, in which case CertIdentity and CertOIDCIssuer are required.
	CosignKey string
	// CertIdentity / CertOIDCIssuer pin the expected keyless signer identity.
	CertIdentity   string
	CertOIDCIssuer string
}

// Pull fetches the OCI pack artifact ref into destDir. Order is fail-closed:
// when a signature is required it is verified before any content is written,
// then oras pulls the tarball (verifying layer digests against the manifest),
// and finally the tarball is extracted with path-traversal protection. After
// extraction the pack.yaml manifest is validated. destDir is left populated only
// on full success.
func Pull(ctx context.Context, ref, destDir string, opts PullOptions) (*Manifest, error) {
	oras := opts.Oras
	if oras == nil {
		oras = DefaultOras
	}
	target := TrimOCIScheme(ref)

	if opts.RequireSignature {
		cosign := opts.Cosign
		if cosign == nil {
			cosign = DefaultCosign
		}
		verifyArgs := []string{"verify"}
		switch {
		case opts.CosignKey != "":
			verifyArgs = append(verifyArgs, "--key", opts.CosignKey)
		default:
			if opts.CertIdentity == "" || opts.CertOIDCIssuer == "" {
				return nil, fmt.Errorf("keyless verification of %s requires a certificate identity and OIDC issuer", target)
			}
			verifyArgs = append(verifyArgs,
				"--certificate-identity", opts.CertIdentity,
				"--certificate-oidc-issuer", opts.CertOIDCIssuer)
		}
		verifyArgs = append(verifyArgs, target)
		if err := cosign(ctx, "", verifyArgs...); err != nil {
			return nil, fmt.Errorf("signature verification failed for %s: %w", target, err)
		}
	}

	download, err := os.MkdirTemp("", "labctl-pack-pull-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(download)

	if err := oras(ctx, "", "pull", target, "-o", download); err != nil {
		return nil, fmt.Errorf("pulling pack from %s: %w", target, err)
	}

	tarball := filepath.Join(download, LayerFileName)
	if _, err := os.Stat(tarball); err != nil {
		return nil, fmt.Errorf("pulled artifact %s does not contain %s — not a Flightdeck pack", target, LayerFileName)
	}

	tf, err := os.Open(tarball)
	if err != nil {
		return nil, err
	}
	defer tf.Close()
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return nil, err
	}
	if err := UntarGz(tf, destDir); err != nil {
		return nil, fmt.Errorf("extracting pack: %w", err)
	}

	m, err := Load(destDir)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", ManifestFile, err)
	}
	if m != nil {
		if err := m.Validate(); err != nil {
			return nil, err
		}
	}
	return m, nil
}
