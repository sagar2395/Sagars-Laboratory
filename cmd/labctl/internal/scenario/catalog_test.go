package scenario

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const packScenarioYAML = `name: pack-scenario
displayName: Pack Scenario
description: From a catalog pack
category: testing
components: []
checks:
  - name: ping
    type: http
    url: "http://pack.{{.DomainSuffix}}"
`

// writeScenarioDir writes a scenario.yaml into dir.
func writeScenarioDir(t *testing.T, dir, yamlStr string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "scenario.yaml"), []byte(yamlStr), 0644); err != nil {
		t.Fatal(err)
	}
}

// stubGit simulates `git clone <url> <dest>` by copying a local fixture dir.
// fixtures maps URL → source dir.
func stubGit(t *testing.T, fixtures map[string]string) GitFunc {
	t.Helper()
	return func(args ...string) error {
		if len(args) < 2 || args[0] != "clone" {
			return fmt.Errorf("stub git: unexpected args %v", args)
		}
		url, dest := args[len(args)-2], args[len(args)-1]
		src, ok := fixtures[url]
		if !ok {
			return fmt.Errorf("stub git: repository %q not found", url)
		}
		return os.CopyFS(dest, os.DirFS(src))
	}
}

func TestSplitRef(t *testing.T) {
	tests := []struct{ in, url, ref string }{
		{"https://github.com/org/pack.git", "https://github.com/org/pack.git", ""},
		{"https://github.com/org/pack.git@v1.2.0", "https://github.com/org/pack.git", "v1.2.0"},
		{"git@github.com:org/pack.git", "git@github.com:org/pack.git", ""},
		{"git@github.com:org/pack.git@main", "git@github.com:org/pack.git", "main"},
		{"@weird", "@weird", ""},
	}
	for _, tt := range tests {
		url, ref := SplitRef(tt.in)
		if url != tt.url || ref != tt.ref {
			t.Errorf("SplitRef(%q) = (%q, %q), want (%q, %q)", tt.in, url, ref, tt.url, tt.ref)
		}
	}
}

func TestPackNameFromURL(t *testing.T) {
	tests := []struct{ in, want string }{
		{"https://github.com/org/chaos-pack.git", "chaos-pack"},
		{"https://github.com/org/chaos-pack", "chaos-pack"},
		{"git@github.com:org/chaos-pack.git", "chaos-pack"},
		{"https://example.com/deep/path/pack/", "pack"},
	}
	for _, tt := range tests {
		if got := PackNameFromURL(tt.in); got != tt.want {
			t.Errorf("PackNameFromURL(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestValidatePackDir(t *testing.T) {
	t.Run("single scenario at root", func(t *testing.T) {
		dir := t.TempDir()
		writeScenarioDir(t, dir, packScenarioYAML)
		names, err := ValidatePackDir(dir)
		if err != nil || len(names) != 1 || names[0] != "pack-scenario" {
			t.Fatalf("got (%v, %v)", names, err)
		}
	})

	t.Run("collection of scenarios", func(t *testing.T) {
		dir := t.TempDir()
		writeScenarioDir(t, filepath.Join(dir, "one"), packScenarioYAML)
		writeScenarioDir(t, filepath.Join(dir, "two"),
			strings.Replace(packScenarioYAML, "pack-scenario", "second-scenario", 1))
		os.MkdirAll(filepath.Join(dir, "docs"), 0755) // non-scenario dir is fine
		names, err := ValidatePackDir(dir)
		if err != nil || len(names) != 2 {
			t.Fatalf("got (%v, %v)", names, err)
		}
	})

	t.Run("empty pack rejected", func(t *testing.T) {
		if _, err := ValidatePackDir(t.TempDir()); err == nil || !strings.Contains(err.Error(), "no scenario.yaml") {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("invalid scenario rejects whole pack", func(t *testing.T) {
		dir := t.TempDir()
		writeScenarioDir(t, filepath.Join(dir, "ok"), packScenarioYAML)
		writeScenarioDir(t, filepath.Join(dir, "bad"),
			strings.Replace(packScenarioYAML, "type: http", "type: gopher", 1))
		if _, err := ValidatePackDir(dir); err == nil || !strings.Contains(err.Error(), "gopher") {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("path traversal rejected", func(t *testing.T) {
		dir := t.TempDir()
		evil := strings.Replace(packScenarioYAML, "components: []",
			"components:\n  - name: evil\n    type: manifest\n    path: ../../outside.yaml", 1)
		writeScenarioDir(t, dir, evil)
		if _, err := ValidatePackDir(dir); err == nil || !strings.Contains(err.Error(), "relative path inside") {
			t.Fatalf("traversal must be rejected, got %v", err)
		}
	})
}

func TestInstallPack_EndToEnd(t *testing.T) {
	root := t.TempDir()
	fixture := t.TempDir()
	writeScenarioDir(t, filepath.Join(fixture, "pack-scenario"), packScenarioYAML)
	git := stubGit(t, map[string]string{"https://example.com/org/community-pack.git": fixture})

	engine := NewEngine(root, "k3d.local", "k3d")
	pack, err := engine.InstallPack("https://example.com/org/community-pack.git", "", false, git)
	if err != nil {
		t.Fatalf("InstallPack: %v", err)
	}
	if pack.Name != "community-pack" || len(pack.Scenarios) != 1 {
		t.Fatalf("pack = %+v", pack)
	}

	// The scenario is now visible, attributed, and usable.
	s, err := engine.Get("pack-scenario")
	if err != nil {
		t.Fatalf("installed scenario not discoverable: %v", err)
	}
	if s.Source != "community-pack" {
		t.Errorf("source = %q, want community-pack", s.Source)
	}
	if !strings.Contains(s.Dir, filepath.Join(".labctl", "catalog", "community-pack")) {
		t.Errorf("scenario dir outside the catalog: %q", s.Dir)
	}

	// A fresh engine (new process) sees it too.
	if _, err := NewEngine(root, "k3d.local", "k3d").Get("pack-scenario"); err != nil {
		t.Fatalf("pack must survive engine restarts: %v", err)
	}

	// Reinstall: refused without force, allowed with it.
	if _, err := engine.InstallPack("https://example.com/org/community-pack.git", "", false, git); err == nil ||
		!strings.Contains(err.Error(), "--force") {
		t.Fatalf("expected already-installed error, got %v", err)
	}
	if _, err := engine.InstallPack("https://example.com/org/community-pack.git", "", true, git); err != nil {
		t.Fatalf("force reinstall: %v", err)
	}

	// Uninstall removes the scenarios.
	if err := engine.UninstallPack("community-pack"); err != nil {
		t.Fatalf("UninstallPack: %v", err)
	}
	if _, err := engine.Get("pack-scenario"); err == nil {
		t.Fatal("scenario must disappear with its pack")
	}
	if len(engine.Packs()) != 0 {
		t.Fatalf("packs should be empty, got %v", engine.Packs())
	}
}

func TestInstallPack_InvalidPackNeverLands(t *testing.T) {
	root := t.TempDir()
	fixture := t.TempDir()
	writeScenarioDir(t, fixture, strings.Replace(packScenarioYAML, "type: http", "type: gopher", 1))
	git := stubGit(t, map[string]string{"https://example.com/bad-pack.git": fixture})

	engine := NewEngine(root, "k3d.local", "k3d")
	if _, err := engine.InstallPack("https://example.com/bad-pack.git", "", false, git); err == nil {
		t.Fatal("invalid pack must be rejected")
	}
	if _, err := os.Stat(filepath.Join(engine.CatalogDir(), "bad-pack")); !os.IsNotExist(err) {
		t.Fatal("rejected pack must not remain in the catalog")
	}
	if len(engine.LoadErrors()) != 0 {
		t.Fatalf("rejected pack must leave no load errors: %v", engine.LoadErrors())
	}
}

func TestInstallPack_NameCollisionWithRepoScenario(t *testing.T) {
	root := t.TempDir()
	createTestScenario(t, root, "test-scenario", testScenarioYAML) // in-repo

	fixture := t.TempDir()
	writeScenarioDir(t, filepath.Join(fixture, "clash"),
		strings.Replace(packScenarioYAML, "pack-scenario", "test-scenario", 1))
	git := stubGit(t, map[string]string{"https://example.com/clash.git": fixture})

	engine := NewEngine(root, "k3d.local", "k3d")
	_, err := engine.InstallPack("https://example.com/clash.git", "", false, git)
	if err == nil || !strings.Contains(err.Error(), "collides") {
		t.Fatalf("expected collision error, got %v", err)
	}
	// The in-repo scenario is untouched.
	if s, err := engine.Get("test-scenario"); err != nil || s.Source != "" {
		t.Fatalf("in-repo scenario must win: %v %+v", err, s)
	}
}

func TestUninstallPack_Errors(t *testing.T) {
	engine := NewEngine(t.TempDir(), "k3d.local", "k3d")
	if err := engine.UninstallPack("ghost"); err == nil || !strings.Contains(err.Error(), "not installed") {
		t.Fatalf("got %v", err)
	}
	if err := engine.UninstallPack("../escape"); err == nil || !strings.Contains(err.Error(), "invalid pack name") {
		t.Fatalf("got %v", err)
	}
}

func TestUninstallPack_RefusesWhileScenarioActive(t *testing.T) {
	root := t.TempDir()
	fixture := t.TempDir()
	writeScenarioDir(t, filepath.Join(fixture, "pack-scenario"), packScenarioYAML)
	git := stubGit(t, map[string]string{"https://example.com/p.git": fixture})

	engine := NewEngine(root, "k3d.local", "k3d")
	if _, err := engine.InstallPack("https://example.com/p.git", "p", false, git); err != nil {
		t.Fatal(err)
	}
	if err := engine.markActive("pack-scenario"); err != nil {
		t.Fatal(err)
	}
	if err := engine.UninstallPack("p"); err == nil || !strings.Contains(err.Error(), "active") {
		t.Fatalf("expected active-scenario refusal, got %v", err)
	}
	engine.markInactive("pack-scenario")
	if err := engine.UninstallPack("p"); err != nil {
		t.Fatalf("uninstall after deactivation: %v", err)
	}
}
