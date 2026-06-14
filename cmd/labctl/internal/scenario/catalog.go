// SPDX-License-Identifier: Apache-2.0
package scenario

// Scenario catalog (task 044): install scenario packs from git into
// .labctl/catalog/<pack>/, validate them against the v2 schema before they
// become visible, and remove them again. Packs are content snapshots —
// their .git directory is dropped and there is no auto-update; reinstall
// with --force to upgrade. Packs run scripts and manifests on your cluster:
// only install sources you trust.

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/sagars-lab/labctl/pkg/entitlement"
	"github.com/sagars-lab/labctl/pkg/extension"
	"github.com/sagars-lab/labctl/pkg/pack"
	schema "github.com/sagars-lab/labctl/pkg/scenario"
)

// entitlement returns the engine's entitlement, defaulting to the open OSS
// allow-all when unset (e.g. an Engine built without NewEngine in tests).
func (e *Engine) entitlement() entitlement.Entitlement {
	if e.Entitlement == nil {
		return entitlement.Default()
	}
	return e.Entitlement
}

// Pack describes one installed catalog pack.
type Pack struct {
	Name      string         `json:"name"`
	Dir       string         `json:"-"`
	Scenarios []string       `json:"scenarios"`
	Manifest  *pack.Manifest `json:"manifest,omitempty"` // nil for legacy packs without pack.yaml
}

// GitFunc runs a git command; tests inject a stub.
type GitFunc func(args ...string) error

// DefaultGit shells out to git with output passed through.
func DefaultGit(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return nil
}

var validPackName = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)

// CatalogDir is where installed packs live (runtime state, gitignored).
func (e *Engine) CatalogDir() string {
	return filepath.Join(e.ProjectRoot, ".labctl", "catalog")
}

// SplitRef separates an optional "@ref" suffix from a git URL. The "@" only
// counts as a ref separator when it appears after the last "/" and ":" —
// so ssh URLs like git@github.com:org/repo.git survive intact.
func SplitRef(src string) (url, ref string) {
	i := strings.LastIndex(src, "@")
	if i <= 0 {
		return src, ""
	}
	if i < strings.LastIndex(src, "/") || i < strings.LastIndex(src, ":") {
		return src, ""
	}
	return src[:i], src[i+1:]
}

// PackNameFromURL derives a default pack name from a git URL
// (the repository basename, minus .git).
func PackNameFromURL(url string) string {
	base := strings.TrimSuffix(url, "/")
	if i := strings.LastIndexAny(base, "/:"); i >= 0 {
		base = base[i+1:]
	}
	return strings.TrimSuffix(base, ".git")
}

// ValidatePackDir checks every scenario in a pack directory against the
// schema and returns their names. A pack with no scenarios, an invalid
// scenario, or duplicate names is rejected wholesale.
func ValidatePackDir(dir string) ([]string, error) {
	var candidates []string
	if _, err := os.Stat(filepath.Join(dir, "scenario.yaml")); err == nil {
		candidates = append(candidates, dir)
	} else {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("reading pack: %w", err)
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			if _, err := os.Stat(filepath.Join(dir, entry.Name(), "scenario.yaml")); err == nil {
				candidates = append(candidates, filepath.Join(dir, entry.Name()))
			}
		}
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no scenario.yaml found in the pack (expected at its root or in scenario directories)")
	}

	seen := map[string]bool{}
	var names []string
	for _, c := range candidates {
		s, err := loadScenarioFile(filepath.Join(c, "scenario.yaml"))
		if err != nil {
			return nil, fmt.Errorf("pack validation failed: %w", err)
		}
		if seen[s.Name] {
			return nil, fmt.Errorf("pack declares scenario %q twice", s.Name)
		}
		seen[s.Name] = true
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return names, nil
}

// InstallPack clones a git source, validates it, and makes its scenarios
// available. Nothing invalid ever lands in the catalog: the clone happens
// in a temp dir and is only renamed into place after validation.
func (e *Engine) InstallPack(src, name string, force bool, git GitFunc) (*Pack, error) {
	if git == nil {
		git = DefaultGit
	}
	url, ref := SplitRef(src)
	if name == "" {
		name = PackNameFromURL(url)
	}
	if !validPackName.MatchString(name) {
		return nil, fmt.Errorf("invalid pack name %q: must match ^[a-zA-Z0-9_-]{1,64}$ (use --name)", name)
	}

	dest, err := e.prepareDest(name, force)
	if err != nil {
		return nil, err
	}

	tmp := dest + ".installing"
	os.RemoveAll(tmp)
	defer os.RemoveAll(tmp)

	args := []string{"clone", "--depth", "1"}
	if ref != "" {
		args = append(args, "--branch", ref)
	}
	args = append(args, url, tmp)
	if err := git(args...); err != nil {
		return nil, fmt.Errorf("cloning %s: %w", url, err)
	}
	os.RemoveAll(filepath.Join(tmp, ".git")) // packs are content snapshots

	return e.installFetched(name, dest, tmp)
}

// InstallVia fetches a pack with the given extension.Resolver into a temp dir,
// then installs it through the shared validate → entitle → rename path. This is
// the seam that lets premium/private/hosted sources slot in (task 070): the
// engine never needs to know where a pack came from, only how to install it.
func (e *Engine) InstallVia(ctx context.Context, resolver extension.Resolver, ref, name string, force bool) (*Pack, error) {
	tmp, err := os.MkdirTemp("", "labctl-pack-resolve-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)
	if err := resolver.Resolve(ctx, ref, tmp); err != nil {
		return nil, err
	}
	return e.InstallPackFromDir(tmp, name, force)
}

// InstallPackFromDir installs a pack whose contents have already been fetched
// into srcDir (e.g. by an OCI pull). It shares the same validation, manifest,
// and collision rules as the git path — nothing invalid ever lands in the
// catalog because everything is checked in a staging copy before the rename.
func (e *Engine) InstallPackFromDir(srcDir, name string, force bool) (*Pack, error) {
	if name == "" {
		return nil, fmt.Errorf("a pack name is required (use --name)")
	}
	if !validPackName.MatchString(name) {
		return nil, fmt.Errorf("invalid pack name %q: must match ^[a-zA-Z0-9_-]{1,64}$ (use --name)", name)
	}

	dest, err := e.prepareDest(name, force)
	if err != nil {
		return nil, err
	}

	tmp := dest + ".installing"
	os.RemoveAll(tmp)
	defer os.RemoveAll(tmp)
	if err := copyTree(srcDir, tmp); err != nil {
		return nil, fmt.Errorf("staging pack: %w", err)
	}
	os.RemoveAll(filepath.Join(tmp, ".git"))

	return e.installFetched(name, dest, tmp)
}

// prepareDest resolves the catalog destination for a pack, enforcing the
// already-installed/--force rule and ensuring the catalog dir exists.
func (e *Engine) prepareDest(name string, force bool) (string, error) {
	dest := filepath.Join(e.CatalogDir(), name)
	if _, err := os.Stat(dest); err == nil {
		if !force {
			return "", fmt.Errorf("pack %q is already installed (use --force to replace it)", name)
		}
		if err := os.RemoveAll(dest); err != nil {
			return "", err
		}
		e.rescan() // drop the old pack's scenarios before collision checks
	}
	if err := os.MkdirAll(e.CatalogDir(), 0755); err != nil {
		return "", err
	}
	return dest, nil
}

// copyTree recursively copies src into dst, preserving file permissions. Used to
// stage an already-fetched pack (OCI pull) into the catalog's .installing dir so
// it goes through the same validate-then-rename path as a git clone.
func copyTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode().Perm())
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, in); err != nil {
			out.Close()
			return err
		}
		return out.Close()
	})
}

// installFetched validates a staged pack in tmp and, if it is sound, atomically
// renames it into dest. It is the shared tail of every install path (git, OCI,
// local dir).
func (e *Engine) installFetched(name, dest, tmp string) (*Pack, error) {
	names, err := ValidatePackDir(tmp)
	if err != nil {
		return nil, err
	}

	// Read the optional pack.yaml manifest. Legacy packs without one still
	// install (a warning is surfaced by the caller); a present manifest must be
	// valid and compatible with this engine before anything lands.
	manifest, err := pack.Load(tmp)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", pack.ManifestFile, err)
	}
	if manifest != nil {
		if err := manifest.Validate(); err != nil {
			return nil, err
		}
		if err := manifest.CheckEngineCompat(e.LabctlVersion, schema.SupportedScenarioAPIVersions); err != nil {
			return nil, err
		}
	}

	// Entitlement seam (task 070): the OSS default allows everything, so this is a
	// no-op for the open engine; an injected premium entitlement can gate the
	// install here. Premium/private sources are gated upstream too (registry auth).
	req := entitlement.Request{Name: name}
	if manifest != nil {
		req.Name = manifest.Metadata.Name
		req.Tier = manifest.Tier()
	}
	if err := e.entitlement().Authorize(context.Background(), req); err != nil {
		return nil, err
	}

	for _, n := range names {
		if existing, ok := e.scenarios[n]; ok {
			from := "the repository"
			if existing.Source != "" {
				from = "pack " + existing.Source
			}
			return nil, fmt.Errorf("pack scenario %q collides with %s — rename the pack scenario or remove the conflict", n, from)
		}
	}

	if err := os.Rename(tmp, dest); err != nil {
		return nil, err
	}
	e.rescan()
	return &Pack{Name: name, Dir: dest, Scenarios: names, Manifest: manifest}, nil
}

// UninstallPack removes an installed pack and its scenarios.
func (e *Engine) UninstallPack(name string) error {
	if !validPackName.MatchString(name) {
		return fmt.Errorf("invalid pack name %q", name)
	}
	dir := filepath.Join(e.CatalogDir(), name)
	if _, err := os.Stat(dir); err != nil {
		return fmt.Errorf("pack %q is not installed", name)
	}
	for _, s := range e.scenarios {
		if s.Source == name && e.isActive(s.Name) {
			return fmt.Errorf("scenario %q from pack %q is active — run 'labctl scenario down %s' first", s.Name, name, s.Name)
		}
	}
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	e.rescan()
	return nil
}

// Packs lists installed packs and the scenarios they currently provide.
func (e *Engine) Packs() []Pack {
	entries, err := os.ReadDir(e.CatalogDir())
	if err != nil {
		return nil
	}
	var packs []Pack
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasSuffix(entry.Name(), ".installing") {
			continue
		}
		p := Pack{Name: entry.Name(), Dir: filepath.Join(e.CatalogDir(), entry.Name())}
		if m, err := pack.Load(p.Dir); err == nil {
			p.Manifest = m // nil for legacy packs without pack.yaml
		}
		for _, s := range e.scenarios {
			if s.Source == p.Name {
				p.Scenarios = append(p.Scenarios, s.Name)
			}
		}
		sort.Strings(p.Scenarios)
		packs = append(packs, p)
	}
	sort.Slice(packs, func(i, j int) bool { return packs[i].Name < packs[j].Name })
	return packs
}
