// SPDX-License-Identifier: Apache-2.0

package pack

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeEvilTar writes a gzip tar containing a single regular file at the given
// (malicious) name, used to exercise UntarGz's path-traversal guard.
func writeEvilTar(t *testing.T, buf *bytes.Buffer, name string) {
	t.Helper()
	gz := gzip.NewWriter(buf)
	tw := tar.NewWriter(gz)
	body := []byte("pwned")
	if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(body)), Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(body); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
}

// writePack lays down a minimal valid pack (pack.yaml + one scenario + a script)
// under a fresh temp dir and returns it.
func writePack(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "pack.yaml"), validYAML)
	mustWrite(t, filepath.Join(dir, "kafka-broker-outage", "scenario.yaml"), "name: kafka-broker-outage\n")
	mustWrite(t, filepath.Join(dir, "kafka-broker-outage", "scripts", "ready.sh"), "#!/usr/bin/env bash\necho ok\n")
	return dir
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestTarGz_RoundTripAndExecBit(t *testing.T) {
	src := writePack(t)
	var buf bytes.Buffer
	if err := TarGz(src, &buf); err != nil {
		t.Fatal(err)
	}
	dest := t.TempDir()
	if err := UntarGz(bytes.NewReader(buf.Bytes()), dest); err != nil {
		t.Fatal(err)
	}
	// pack.yaml round-trips.
	got, err := os.ReadFile(filepath.Join(dest, "pack.yaml"))
	if err != nil || !strings.Contains(string(got), "kafka-drills") {
		t.Fatalf("pack.yaml not extracted: %v", err)
	}
	// *.sh keeps its executable bit.
	info, err := os.Stat(filepath.Join(dest, "kafka-broker-outage", "scripts", "ready.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o100 == 0 {
		t.Errorf("ready.sh should be executable, got mode %v", info.Mode())
	}
}

func TestTarGz_Deterministic(t *testing.T) {
	src := writePack(t)
	var a, b bytes.Buffer
	if err := TarGz(src, &a); err != nil {
		t.Fatal(err)
	}
	if err := TarGz(src, &b); err != nil {
		t.Fatal(err)
	}
	if Sha256Hex(a.Bytes()) != Sha256Hex(b.Bytes()) {
		t.Error("TarGz is not deterministic: digests differ across builds")
	}
}

func TestUntarGz_RejectsPathTraversal(t *testing.T) {
	// Hand-craft a tar with an escaping path.
	var buf bytes.Buffer
	writeEvilTar(t, &buf, "../escape.txt")
	err := UntarGz(bytes.NewReader(buf.Bytes()), t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "escapes the destination") {
		t.Fatalf("expected path-traversal rejection, got: %v", err)
	}
}

func TestVerifyChecksum_DetectsTamper(t *testing.T) {
	f := filepath.Join(t.TempDir(), "pack.tar.gz")
	mustWrite(t, f, "original content")
	good, err := Sha256File(f)
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyChecksum(f, good); err != nil {
		t.Fatalf("matching checksum should pass: %v", err)
	}
	// Tamper with the file.
	mustWrite(t, f, "tampered content")
	if err := VerifyChecksum(f, good); err == nil || !strings.Contains(err.Error(), "tampered") {
		t.Fatalf("expected tamper detection, got: %v", err)
	}
}

func TestPublish_BuildsArtifactAndSigns(t *testing.T) {
	src := writePack(t)
	var orasArgs, cosignArgs []string
	var stageDir string
	oras := func(_ context.Context, dir string, args ...string) error {
		stageDir = dir
		orasArgs = args
		// Confirm the tarball was staged for the push.
		if _, err := os.Stat(filepath.Join(dir, LayerFileName)); err != nil {
			t.Errorf("oras push staged without %s present: %v", LayerFileName, err)
		}
		return nil
	}
	cosign := func(_ context.Context, _ string, args ...string) error {
		cosignArgs = args
		return nil
	}

	digest, err := Publish(context.Background(), src, "oci://ghcr.io/snowops/kafka-drills:1.4.2", PublishOptions{
		Oras: oras, Cosign: cosign, Sign: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(digest, "sha256:") {
		t.Errorf("digest = %q, want sha256: prefix", digest)
	}
	if stageDir == "" {
		t.Error("oras was not run in the stage dir")
	}
	wantContains := []string{"push", "ghcr.io/snowops/kafka-drills:1.4.2", "--artifact-type", ArtifactType}
	joined := strings.Join(orasArgs, " ")
	for _, w := range wantContains {
		if !strings.Contains(joined, w) {
			t.Errorf("oras args %v missing %q", orasArgs, w)
		}
	}
	if !strings.Contains(joined, LayerFileName+":"+LayerMediaType) {
		t.Errorf("oras args %v missing layer:mediatype spec", orasArgs)
	}
	if len(cosignArgs) == 0 || cosignArgs[0] != "sign" {
		t.Errorf("cosign sign not invoked: %v", cosignArgs)
	}
}

func TestPublish_RequiresManifest(t *testing.T) {
	dir := t.TempDir() // no pack.yaml
	_, err := Publish(context.Background(), dir, "oci://r/x:1", PublishOptions{
		Oras: func(context.Context, string, ...string) error { return nil },
	})
	if err == nil || !strings.Contains(err.Error(), "required to publish") {
		t.Fatalf("expected missing-manifest error, got: %v", err)
	}
}

func TestPull_VerifiesThenExtracts(t *testing.T) {
	// Build a real artifact tarball that the stub oras "pull" will drop into -o.
	srcPack := writePack(t)
	var tarball bytes.Buffer
	if err := TarGz(srcPack, &tarball); err != nil {
		t.Fatal(err)
	}

	var verified bool
	cosign := func(_ context.Context, _ string, args ...string) error {
		if args[0] != "verify" {
			t.Errorf("expected cosign verify, got %v", args)
		}
		verified = true
		return nil
	}
	oras := func(_ context.Context, _ string, args ...string) error {
		if !verified {
			t.Error("oras pull ran before signature verification (not fail-closed)")
		}
		// args: pull <ref> -o <download>
		outDir := args[len(args)-1]
		return os.WriteFile(filepath.Join(outDir, LayerFileName), tarball.Bytes(), 0o644)
	}

	dest := t.TempDir()
	m, err := Pull(context.Background(), "oci://ghcr.io/snowops/kafka-drills:1.4.2", dest, PullOptions{
		Oras: oras, Cosign: cosign, RequireSignature: true,
		CosignKey: "cosign.pub",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !verified {
		t.Error("signature was not verified")
	}
	if m == nil || m.Metadata.Name != "acme/kafka-drills" {
		t.Errorf("pulled manifest = %+v", m)
	}
	if _, err := os.Stat(filepath.Join(dest, "kafka-broker-outage", "scenario.yaml")); err != nil {
		t.Errorf("scenario not extracted: %v", err)
	}
}

func TestPull_KeylessRequiresIdentity(t *testing.T) {
	_, err := Pull(context.Background(), "oci://r/x:1", t.TempDir(), PullOptions{
		Oras:             func(context.Context, string, ...string) error { return nil },
		Cosign:           func(context.Context, string, ...string) error { return nil },
		RequireSignature: true, // keyless but no identity/issuer
	})
	if err == nil || !strings.Contains(err.Error(), "certificate identity") {
		t.Fatalf("expected keyless identity error, got: %v", err)
	}
}

func TestPackNameFromOCIRef(t *testing.T) {
	cases := map[string]string{
		"oci://ghcr.io/snowops/kafka-drills:1.4.2":      "kafka-drills",
		"ghcr.io/snowops/kafka-drills":                  "kafka-drills",
		"oci://ghcr.io/snowops/kafka-drills@sha256:abc": "kafka-drills",
		"oci://localhost:5000/hello:dev":                "hello",
	}
	for in, want := range cases {
		if got := PackNameFromOCIRef(in); got != want {
			t.Errorf("PackNameFromOCIRef(%q) = %q, want %q", in, got, want)
		}
	}
}
