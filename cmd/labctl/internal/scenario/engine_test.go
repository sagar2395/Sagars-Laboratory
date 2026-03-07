package scenario

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testScenarioYAML = `name: test-scenario
displayName: Test Scenario
description: A test scenario for unit testing
category: testing
prerequisites:
  platform:
    - ingress/traefik
  apps:
    - go-api
components:
  - name: test-helm
    type: helm
    chart: test/chart
    namespace: test-ns
    valuesFile: values/test.yaml
  - name: test-manifest
    type: manifest
    path: manifests/test.yaml
explore:
  urls:
    - label: Test UI
      url: "http://test.{{.DomainSuffix}}"
  commands:
    - label: Check pods
      command: kubectl get pods -n test-ns
  tips:
    - This is a test tip
`

func createTestScenario(t *testing.T, root, name, yaml string) {
	t.Helper()
	dir := filepath.Join(root, "scenarios", name)
	os.MkdirAll(filepath.Join(dir, "values"), 0755)
	os.MkdirAll(filepath.Join(dir, "manifests"), 0755)
	os.WriteFile(filepath.Join(dir, "scenario.yaml"), []byte(yaml), 0644)
	os.WriteFile(filepath.Join(dir, "values", "test.yaml"), []byte("# test"), 0644)
	os.WriteFile(filepath.Join(dir, "manifests", "test.yaml"), []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: test"), 0644)
}

func TestNewEngine_Discovery(t *testing.T) {
	root := t.TempDir()

	createTestScenario(t, root, "test-scenario", testScenarioYAML)

	second := strings.Replace(testScenarioYAML, "test-scenario", "second-scenario", 1)
	second = strings.Replace(second, "Test Scenario", "Second Scenario", 1)
	createTestScenario(t, root, "second-scenario", second)

	engine := NewEngine(root, "k3d.local")
	scenarios := engine.List()

	if len(scenarios) != 2 {
		t.Fatalf("expected 2 scenarios, got %d", len(scenarios))
	}
}

func TestGet(t *testing.T) {
	root := t.TempDir()
	createTestScenario(t, root, "test-scenario", testScenarioYAML)

	engine := NewEngine(root, "k3d.local")

	s, err := engine.Get("test-scenario")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if s.Name != "test-scenario" {
		t.Errorf("Name: got %q, want %q", s.Name, "test-scenario")
	}
	if s.DisplayName != "Test Scenario" {
		t.Errorf("DisplayName: got %q, want %q", s.DisplayName, "Test Scenario")
	}
	if s.Category != "testing" {
		t.Errorf("Category: got %q, want %q", s.Category, "testing")
	}
	if len(s.Components) != 2 {
		t.Errorf("expected 2 components, got %d", len(s.Components))
	}
	if len(s.Prerequisites.Platform) != 1 {
		t.Errorf("expected 1 platform prerequisite, got %d", len(s.Prerequisites.Platform))
	}
	if len(s.Explore.Tips) != 1 {
		t.Errorf("expected 1 explore tip, got %d", len(s.Explore.Tips))
	}
}

func TestGet_NotFound(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "scenarios"), 0755)

	engine := NewEngine(root, "k3d.local")
	_, err := engine.Get("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent scenario")
	}
}

func TestResolveTemplate(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "scenarios"), 0755)

	engine := NewEngine(root, "k3d.local")

	tests := []struct {
		input    string
		expected string
	}{
		{"http://grafana.{{.DomainSuffix}}", "http://grafana.k3d.local"},
		{"{{.ProjectRoot}}/apps", root + "/apps"},
		{"no templates here", "no templates here"},
		{"{{.DomainSuffix}} and {{.DomainSuffix}}", "k3d.local and k3d.local"},
	}

	for _, tt := range tests {
		got := engine.ResolveTemplate(tt.input)
		if got != tt.expected {
			t.Errorf("ResolveTemplate(%q): got %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestStatus(t *testing.T) {
	root := t.TempDir()
	createTestScenario(t, root, "test-scenario", testScenarioYAML)

	engine := NewEngine(root, "k3d.local")
	statuses := engine.Status()

	if len(statuses) != 1 {
		t.Fatalf("expected 1 status, got %d", len(statuses))
	}

	s := statuses[0]
	if s.Name != "test-scenario" {
		t.Errorf("Name: got %q, want %q", s.Name, "test-scenario")
	}
	if s.Active {
		t.Error("expected scenario to be inactive")
	}
}

func TestIsActive_MarkActive(t *testing.T) {
	root := t.TempDir()
	createTestScenario(t, root, "test-scenario", testScenarioYAML)

	engine := NewEngine(root, "k3d.local")

	// Initially inactive
	if engine.isActive("test-scenario") {
		t.Error("expected scenario to be inactive initially")
	}

	// Mark active
	err := engine.markActive("test-scenario")
	if err != nil {
		t.Fatalf("markActive: %v", err)
	}

	// Should now be active
	if !engine.isActive("test-scenario") {
		t.Error("expected scenario to be active after markActive")
	}

	// Mark inactive
	engine.markInactive("test-scenario")
	if engine.isActive("test-scenario") {
		t.Error("expected scenario to be inactive after markInactive")
	}
}

func TestNewEngine_InvalidYAML(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "scenarios", "bad-scenario")
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "scenario.yaml"), []byte("invalid: [yaml: {"), 0644)

	engine := NewEngine(root, "k3d.local")
	scenarios := engine.List()
	if len(scenarios) != 0 {
		t.Errorf("expected 0 scenarios for invalid YAML, got %d", len(scenarios))
	}
}

func TestNewEngine_MissingName(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "scenarios", "no-name")
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "scenario.yaml"), []byte("description: no name field"), 0644)

	engine := NewEngine(root, "k3d.local")
	scenarios := engine.List()
	if len(scenarios) != 0 {
		t.Errorf("expected 0 scenarios for missing name, got %d", len(scenarios))
	}
}

func TestManifestHasExplicitNamespace(t *testing.T) {
	tests := []struct {
		name     string
		manifest string
		want     bool
	}{
		{
			name: "single document with namespace",
			manifest: `apiVersion: v1
kind: ConfigMap
metadata:
  name: test
  namespace: go-api
`,
			want: true,
		},
		{
			name: "single document without namespace",
			manifest: `apiVersion: v1
kind: ConfigMap
metadata:
  name: test
`,
			want: false,
		},
		{
			name: "multi document with one namespace",
			manifest: `apiVersion: v1
kind: ConfigMap
metadata:
  name: one
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: two
  namespace: echo-server
`,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := manifestHasExplicitNamespace(tt.manifest)
			if got != tt.want {
				t.Fatalf("manifestHasExplicitNamespace() = %v, want %v", got, tt.want)
			}
		})
	}
}
