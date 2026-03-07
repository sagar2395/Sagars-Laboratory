package scenario

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"

	"github.com/sagars-lab/labctl/internal/executor"
)

// Scenario represents a lab scenario loaded from scenario.yaml.
type Scenario struct {
	Name          string        `yaml:"name" json:"name"`
	DisplayName   string        `yaml:"displayName" json:"displayName"`
	Description   string        `yaml:"description" json:"description"`
	Category      string        `yaml:"category" json:"category"`
	Prerequisites Prerequisites `yaml:"prerequisites" json:"prerequisites"`
	Runtimes      []string      `yaml:"runtimes" json:"runtimes"`
	Components    []Component   `yaml:"components" json:"components"`
	Explore       Explore       `yaml:"explore" json:"explore"`

	// Runtime fields (not from YAML)
	Dir    string `yaml:"-" json:"-"`
	Active bool   `yaml:"-" json:"active"`
}

// Prerequisites defines what must be running before a scenario can activate.
type Prerequisites struct {
	Platform []string `yaml:"platform" json:"platform"`
	Apps     []string `yaml:"apps" json:"apps"`
}

// Component defines a single deployable unit within a scenario.
type Component struct {
	Name       string            `yaml:"name" json:"name"`
	Type       string            `yaml:"type" json:"type"` // helm, manifest, grafana-dashboard, script
	Chart      string            `yaml:"chart,omitempty" json:"chart,omitempty"`
	Repo       string            `yaml:"repo,omitempty" json:"repo,omitempty"`
	Version    string            `yaml:"version,omitempty" json:"version,omitempty"`
	Namespace  string            `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	ValuesFile string            `yaml:"valuesFile,omitempty" json:"valuesFile,omitempty"`
	Path       string            `yaml:"path,omitempty" json:"path,omitempty"`
	Set        map[string]string `yaml:"set,omitempty" json:"set,omitempty"`
	Script     string            `yaml:"script,omitempty" json:"script,omitempty"`
}

// Explore contains hints for the user on how to explore the scenario.
type Explore struct {
	URLs     []ExploreURL     `yaml:"urls" json:"urls"`
	Commands []ExploreCommand `yaml:"commands" json:"commands"`
	Tips     []string         `yaml:"tips" json:"tips"`
}

// ExploreURL is a URL hint.
type ExploreURL struct {
	Label string `yaml:"label" json:"label"`
	URL   string `yaml:"url" json:"url"`
}

// ExploreCommand is a command hint.
type ExploreCommand struct {
	Label   string `yaml:"label" json:"label"`
	Command string `yaml:"command" json:"command"`
}

// Engine discovers, loads, and manages scenarios.
type Engine struct {
	ProjectRoot  string
	DomainSuffix string
	scenarios    map[string]*Scenario
	stateDir     string
}

// NewEngine creates a scenario engine by scanning the scenarios/ directory.
func NewEngine(projectRoot, domainSuffix string) *Engine {
	e := &Engine{
		ProjectRoot:  projectRoot,
		DomainSuffix: domainSuffix,
		scenarios:    make(map[string]*Scenario),
		stateDir:     filepath.Join(projectRoot, ".labctl", "scenarios"),
	}
	e.scan()
	return e
}

// List returns all discovered scenarios.
func (e *Engine) List() []*Scenario {
	var result []*Scenario
	for _, s := range e.scenarios {
		s.Active = e.isActive(s.Name)
		result = append(result, s)
	}
	return result
}

// Get returns a scenario by name.
func (e *Engine) Get(name string) (*Scenario, error) {
	s, ok := e.scenarios[name]
	if !ok {
		return nil, fmt.Errorf("scenario %q not found", name)
	}
	s.Active = e.isActive(name)
	return s, nil
}

// Up activates a scenario by installing all its components.
func (e *Engine) Up(name string, exec *executor.Executor) error {
	s, err := e.Get(name)
	if err != nil {
		return err
	}

	if e.isActive(name) {
		return fmt.Errorf("scenario %q is already active", name)
	}

	fmt.Printf("Activating scenario: %s\n", s.DisplayName)
	fmt.Printf("  %s\n\n", s.Description)

	for i, comp := range s.Components {
		fmt.Printf("[%d/%d] Installing %s (%s)...\n", i+1, len(s.Components), comp.Name, comp.Type)
		if err := e.installComponent(s, &comp, exec); err != nil {
			return fmt.Errorf("installing component %s: %w", comp.Name, err)
		}
	}

	// Mark as active
	if err := e.markActive(name); err != nil {
		return fmt.Errorf("marking scenario active: %w", err)
	}

	// Print explore hints
	e.printExploreHints(s)

	return nil
}

// Down deactivates a scenario by uninstalling all its components in reverse order.
func (e *Engine) Down(name string, exec *executor.Executor) error {
	s, err := e.Get(name)
	if err != nil {
		return err
	}

	if !e.isActive(name) {
		return fmt.Errorf("scenario %q is not active", name)
	}

	fmt.Printf("Deactivating scenario: %s\n\n", s.DisplayName)

	// Uninstall in reverse order
	for i := len(s.Components) - 1; i >= 0; i-- {
		comp := s.Components[i]
		fmt.Printf("[%d/%d] Uninstalling %s...\n", len(s.Components)-i, len(s.Components), comp.Name)
		if err := e.uninstallComponent(s, &comp, exec); err != nil {
			fmt.Printf("  Warning: %v\n", err)
		}
	}

	// Mark as inactive
	e.markInactive(name)
	fmt.Println("\nScenario deactivated.")
	return nil
}

// Status returns a summary of active scenarios.
func (e *Engine) Status() []ScenarioStatus {
	var result []ScenarioStatus
	for _, s := range e.scenarios {
		result = append(result, ScenarioStatus{
			Name:        s.Name,
			DisplayName: s.DisplayName,
			Category:    s.Category,
			Active:      e.isActive(s.Name),
		})
	}
	return result
}

// ScenarioStatus is a lightweight status for listing scenarios.
type ScenarioStatus struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Category    string `json:"category"`
	Active      bool   `json:"active"`
}

func (e *Engine) scan() {
	scenariosDir := filepath.Join(e.ProjectRoot, "scenarios")
	entries, err := os.ReadDir(scenariosDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		yamlPath := filepath.Join(scenariosDir, entry.Name(), "scenario.yaml")
		s, err := e.loadScenario(yamlPath)
		if err != nil {
			continue
		}
		s.Dir = filepath.Join(scenariosDir, entry.Name())
		e.scenarios[s.Name] = s
	}
}

func (e *Engine) loadScenario(path string) (*Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var s Scenario
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	if s.Name == "" {
		return nil, fmt.Errorf("scenario in %s has no name", path)
	}

	return &s, nil
}

func (e *Engine) installComponent(s *Scenario, comp *Component, exec *executor.Executor) error {
	switch comp.Type {
	case "helm":
		return e.installHelm(s, comp, exec)
	case "manifest":
		return e.installManifest(s, comp, exec)
	case "grafana-dashboard":
		return e.installGrafanaDashboard(s, comp, exec)
	case "script":
		return e.runScript(s, comp, exec)
	default:
		return fmt.Errorf("unknown component type: %s", comp.Type)
	}
}

func (e *Engine) uninstallComponent(s *Scenario, comp *Component, exec *executor.Executor) error {
	switch comp.Type {
	case "helm":
		return e.uninstallHelm(comp, exec)
	case "manifest":
		return e.uninstallManifest(s, comp, exec)
	case "grafana-dashboard":
		// Grafana dashboards are removed when grafana restarts or scenario configmap is deleted
		return e.uninstallGrafanaDashboard(s, comp, exec)
	case "script":
		// Scripts don't have a clean uninstall; skip
		return nil
	default:
		return nil
	}
}

func (e *Engine) installHelm(s *Scenario, comp *Component, exec *executor.Executor) error {
	ns := comp.Namespace
	if ns == "" {
		ns = "default"
	}

	// Create namespace
	if err := exec.RunCommand("kubectl", "create", "namespace", ns, "--dry-run=client", "-o", "yaml"); err == nil {
		exec.RunCommand("kubectl", "apply", "-f", "-")
	}
	exec.RunCommand("kubectl", "create", "namespace", ns, "--dry-run=client", "-o", "yaml")

	// Add helm repo if specified
	if comp.Repo != "" {
		repoName := strings.Split(comp.Chart, "/")[0]
		exec.RunHelm("repo", "add", repoName, comp.Repo)
		exec.RunHelm("repo", "update")
	}

	args := []string{
		"upgrade", "--install", comp.Name, comp.Chart,
		"--namespace", ns, "--create-namespace",
		"--wait", "--timeout", "5m",
	}

	if comp.Version != "" {
		args = append(args, "--version", comp.Version)
	}

	if comp.ValuesFile != "" {
		valuesPath := filepath.Join(s.Dir, comp.ValuesFile)
		if _, err := os.Stat(valuesPath); err == nil {
			args = append(args, "-f", valuesPath)
		}
	}

	for k, v := range comp.Set {
		resolved := e.resolveTemplate(v)
		args = append(args, "--set", k+"="+resolved)
	}

	return exec.RunHelm(args...)
}

func (e *Engine) uninstallHelm(comp *Component, exec *executor.Executor) error {
	ns := comp.Namespace
	if ns == "" {
		ns = "default"
	}
	return exec.RunHelm("uninstall", comp.Name, "--namespace", ns)
}

func (e *Engine) installManifest(s *Scenario, comp *Component, exec *executor.Executor) error {
	manifestPath := filepath.Join(s.Dir, comp.Path)

	// Template the manifest
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("reading manifest %s: %w", manifestPath, err)
	}

	resolved := e.resolveTemplate(string(data))

	// Write to temp file and apply
	tmpFile, err := os.CreateTemp("", "labctl-manifest-*.yaml")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(resolved); err != nil {
		tmpFile.Close()
		return err
	}
	tmpFile.Close()

	args := []string{"apply", "-f", tmpFile.Name()}
	if comp.Namespace != "" && !manifestHasExplicitNamespace(resolved) {
		args = append(args, "--namespace", comp.Namespace)
	}

	return exec.RunKubectl(args...)
}

func (e *Engine) uninstallManifest(s *Scenario, comp *Component, exec *executor.Executor) error {
	manifestPath := filepath.Join(s.Dir, comp.Path)

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil // Already gone
	}

	resolved := e.resolveTemplate(string(data))

	tmpFile, err := os.CreateTemp("", "labctl-manifest-*.yaml")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(resolved)
	tmpFile.Close()

	args := []string{"delete", "-f", tmpFile.Name(), "--ignore-not-found"}
	if comp.Namespace != "" && !manifestHasExplicitNamespace(resolved) {
		args = append(args, "--namespace", comp.Namespace)
	}

	return exec.RunKubectl(args...)
}

func (e *Engine) installGrafanaDashboard(s *Scenario, comp *Component, exec *executor.Executor) error {
	dashDir := filepath.Join(s.Dir, comp.Path)
	entries, err := os.ReadDir(dashDir)
	if err != nil {
		return fmt.Errorf("reading dashboard dir: %w", err)
	}

	ns := comp.Namespace
	if ns == "" {
		ns = "monitoring"
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dashDir, entry.Name()))
		if err != nil {
			continue
		}

		cmName := fmt.Sprintf("scenario-%s-%s", s.Name, strings.TrimSuffix(entry.Name(), ".json"))

		// Create ConfigMap with Grafana sidecar label
		cm := fmt.Sprintf(`apiVersion: v1
kind: ConfigMap
metadata:
  name: %s
  namespace: %s
  labels:
    grafana_dashboard: "1"
data:
  %s: |
%s`,
			cmName, ns, entry.Name(), indentJSON(string(data), "    "))

		tmpFile, err := os.CreateTemp("", "labctl-dashboard-*.yaml")
		if err != nil {
			continue
		}

		tmpFile.WriteString(cm)
		tmpFile.Close()
		exec.RunKubectl("apply", "-f", tmpFile.Name())
		os.Remove(tmpFile.Name())
	}

	return nil
}

func (e *Engine) uninstallGrafanaDashboard(s *Scenario, comp *Component, exec *executor.Executor) error {
	dashDir := filepath.Join(s.Dir, comp.Path)
	entries, err := os.ReadDir(dashDir)
	if err != nil {
		return nil
	}

	ns := comp.Namespace
	if ns == "" {
		ns = "monitoring"
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		cmName := fmt.Sprintf("scenario-%s-%s", s.Name, strings.TrimSuffix(entry.Name(), ".json"))
		exec.RunKubectl("delete", "configmap", cmName, "--namespace", ns, "--ignore-not-found")
	}

	return nil
}

func (e *Engine) runScript(s *Scenario, comp *Component, exec *executor.Executor) error {
	scriptPath := filepath.Join(s.Dir, comp.Script)
	return exec.RunScript(scriptPath)
}

// ResolveTemplate resolves Go template variables in a string (e.g., {{.DomainSuffix}}).
func (e *Engine) ResolveTemplate(input string) string {
	return e.resolveTemplate(input)
}

func (e *Engine) resolveTemplate(input string) string {
	tmpl, err := template.New("scenario").Parse(input)
	if err != nil {
		return input
	}

	data := map[string]string{
		"DomainSuffix": e.DomainSuffix,
		"ProjectRoot":  e.ProjectRoot,
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return input
	}
	return buf.String()
}

func (e *Engine) isActive(name string) bool {
	statePath := filepath.Join(e.stateDir, name+".active")
	_, err := os.Stat(statePath)
	return err == nil
}

func (e *Engine) markActive(name string) error {
	if err := os.MkdirAll(e.stateDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(e.stateDir, name+".active"), []byte("active"), 0644)
}

func (e *Engine) markInactive(name string) {
	os.Remove(filepath.Join(e.stateDir, name+".active"))
}

func (e *Engine) printExploreHints(s *Scenario) {
	if len(s.Explore.URLs) == 0 && len(s.Explore.Commands) == 0 && len(s.Explore.Tips) == 0 {
		return
	}

	fmt.Println("\n=== Explore This Scenario ===")

	if len(s.Explore.URLs) > 0 {
		fmt.Println("\nURLs:")
		for _, u := range s.Explore.URLs {
			resolved := e.resolveTemplate(u.URL)
			fmt.Printf("  %-30s %s\n", u.Label+":", resolved)
		}
	}

	if len(s.Explore.Commands) > 0 {
		fmt.Println("\nCommands to try:")
		for _, c := range s.Explore.Commands {
			resolved := e.resolveTemplate(c.Command)
			fmt.Printf("  %s:\n    %s\n", c.Label, resolved)
		}
	}

	if len(s.Explore.Tips) > 0 {
		fmt.Println("\nTips:")
		for _, t := range s.Explore.Tips {
			fmt.Printf("  - %s\n", t)
		}
	}

	fmt.Println()
}

func indentJSON(s, prefix string) string {
	var result strings.Builder
	for _, line := range strings.Split(s, "\n") {
		result.WriteString(prefix)
		result.WriteString(line)
		result.WriteString("\n")
	}
	return result.String()
}

func manifestHasExplicitNamespace(manifest string) bool {
	decoder := yaml.NewDecoder(strings.NewReader(manifest))

	for {
		var doc map[string]interface{}
		if err := decoder.Decode(&doc); err != nil {
			break
		}

		if len(doc) == 0 {
			continue
		}

		metadataRaw, ok := doc["metadata"]
		if !ok {
			continue
		}

		metadata, ok := metadataRaw.(map[string]interface{})
		if !ok {
			continue
		}

		nsRaw, ok := metadata["namespace"]
		if !ok {
			continue
		}

		ns, ok := nsRaw.(string)
		if ok && strings.TrimSpace(ns) != "" {
			return true
		}
	}

	return false
}
