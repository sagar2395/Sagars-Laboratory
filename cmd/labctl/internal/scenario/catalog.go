// SPDX-License-Identifier: Apache-2.0
package scenario

// Scenario catalog (task 044): install scenario packs from git into
// .labctl/catalog/<pack>/, validate them against the v2 schema before they
// become visible, and remove them again. Packs are content snapshots —
// their .git directory is dropped and there is no auto-update; reinstall
// with --force to upgrade. Packs run scripts and manifests on your cluster:
// only install sources you trust.

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/sagars-lab/labctl/pkg/pack"
	schema "github.com/sagars-lab/labctl/pkg/scenario"
)

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

	dest := filepath.Join(e.CatalogDir(), name)
	if _, err := os.Stat(dest); err == nil {
		if !force {
			return nil, fmt.Errorf("pack %q is already installed (use --force to replace it)", name)
		}
		if err := os.RemoveAll(dest); err != nil {
			return nil, err
		}
		e.rescan() // drop the old pack's scenarios before collision checks
	}
	if err := os.MkdirAll(e.CatalogDir(), 0755); err != nil {
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
