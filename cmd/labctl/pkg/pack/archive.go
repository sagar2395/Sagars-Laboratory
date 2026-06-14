// SPDX-License-Identifier: Apache-2.0

package pack

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// OCI artifact identity for a Flightdeck scenario pack. A pack is published as a
// single gzipped-tar layer so it is content-addressable, signable, and
// registry-neutral. Keeping one layer (instead of one file per pack entry) makes
// the digest stable and the publish/pull commands simple.
const (
	// ArtifactType is the OCI manifest artifactType for a pack.
	ArtifactType = "application/vnd.flightdeck.pack.v1+json"
	// LayerMediaType is the media type of the single tarball layer.
	LayerMediaType = "application/vnd.flightdeck.pack.layer.v1.tar+gzip"
	// LayerFileName is the title (filename) of the tarball layer inside the
	// artifact; oras restores it under this name on pull.
	LayerFileName = "pack.tar.gz"
)

// TarGz writes a deterministic gzip-compressed tar of srcDir to w. Entries are
// emitted in sorted order with normalized metadata (no mod times, fixed perms,
// no uid/gid) so the same pack contents always produce the same bytes — and thus
// the same sha256 — regardless of who builds it or when. The pack.yaml manifest
// at the root is included like any other file.
func TarGz(srcDir string, w io.Writer) error {
	gz := gzip.NewWriter(w)
	tw := tar.NewWriter(gz)

	var files []string
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		// Never archive a stray .git or runtime state — packs are content snapshots.
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if rel == "." || strings.HasPrefix(rel, ".git"+string(filepath.Separator)) || rel == ".git" {
			return nil
		}
		files = append(files, rel)
		return nil
	})
	if err != nil {
		return err
	}
	sort.Strings(files)

	for _, rel := range files {
		data, err := os.ReadFile(filepath.Join(srcDir, rel))
		if err != nil {
			return err
		}
		// Use forward slashes in the archive for cross-platform reproducibility.
		name := filepath.ToSlash(rel)
		hdr := &tar.Header{
			Name: name,
			Mode: 0o644,
			Size: int64(len(data)),
		}
		if isExecutableScript(name) {
			hdr.Mode = 0o755
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if _, err := tw.Write(data); err != nil {
			return err
		}
	}

	if err := tw.Close(); err != nil {
		return err
	}
	return gz.Close()
}

// isExecutableScript marks shell scripts executable so checks/inject scripts run
// after extraction. The pack contract keeps logic in scripts (golden rule #2),
// so preserving the +x bit on *.sh matters.
func isExecutableScript(name string) bool {
	return strings.HasSuffix(name, ".sh")
}

// UntarGz extracts a gzip-compressed tar from r into destDir. It rejects path
// traversal ("zip slip"): every entry must stay within destDir. Packs come from
// remote registries, so extraction fails closed on any escaping path.
func UntarGz(r io.Reader, destDir string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("reading gzip stream: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)

	cleanDest := filepath.Clean(destDir)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading tar stream: %w", err)
		}
		// Resolve the target and ensure it is inside destDir.
		target := filepath.Join(cleanDest, filepath.FromSlash(hdr.Name))
		if target != cleanDest && !strings.HasPrefix(target, cleanDest+string(os.PathSeparator)) {
			return fmt.Errorf("refusing to extract %q: path escapes the destination", hdr.Name)
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(hdr.Mode)&0o777)
			if err != nil {
				return err
			}
			// Bound the copy to the declared size to avoid decompression bombs
			// inflating beyond the header.
			if _, err := io.CopyN(f, tr, hdr.Size); err != nil && err != io.EOF {
				f.Close()
				return err
			}
			if err := f.Close(); err != nil {
				return err
			}
		default:
			// Skip symlinks, devices, etc. — packs are plain files only.
			return fmt.Errorf("refusing to extract %q: unsupported tar entry type %d", hdr.Name, hdr.Typeflag)
		}
	}
	return nil
}

// Sha256Hex returns the "sha256:<hex>" digest of data.
func Sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// Sha256File returns the "sha256:<hex>" digest of a file's contents.
func Sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}

// VerifyChecksum compares the digest of file against want ("sha256:<hex>"). It
// fails closed: any mismatch — i.e. tampering — returns an error.
func VerifyChecksum(path, want string) error {
	got, err := Sha256File(path)
	if err != nil {
		return err
	}
	if !strings.EqualFold(got, want) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s (artifact may be tampered)", want, got)
	}
	return nil
}
