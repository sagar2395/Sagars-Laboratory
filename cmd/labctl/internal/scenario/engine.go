// SPDX-License-Identifier: Apache-2.0
package scenario

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"

	"github.com/sagars-lab/labctl/internal/checks"
	"github.com/sagars-lab/labctl/internal/executor"
)

// ErrAlreadyActive is returned by Up when the scenario is already active.
// Callers should treat this as a no-op, not a failure.
var ErrAlreadyActive = errors.New("scenario already active")

// ErrNoChecks is returned by Verify when the scenario declares no checks.
var ErrNoChecks = errors.New("scenario defines no checks")

// Scenario represents a lab scenario loaded from scenario.yaml.
//
// Format v1 declares a flat `components` list. Format v2 may instead group
// components into ordered `stages`, and add human-readable `objectives` and
// machine-verifiable `checks` (run by `labctl scenario verify`). A scenario
// must use either `components` or `stages`, never both.
type Scenario struct {
	// APIVersion declares the scenario schema version. Optional for backward
	// compatibility: an empty value is treated as the current default
	// (DefaultScenarioAPIVersion). The engine accepts the current and previous
	// schema versions; see SupportedScenarioAPIVersions.
	APIVersion string `yaml:"apiVersion,omitempty" json:"apiVersion,omitempty"`

	Name          string        `yaml:"name" json:"name"`
	DisplayName   string        `yaml:"displayName" json:"displayName"`
	Description   string        `yaml:"description" json:"description"`
	Category      string        `yaml:"category" json:"category"`
	Prerequisites Prerequisites `yaml:"prerequisites" json:"prerequisites"`
	Runtimes      []string      `yaml:"runtimes" json:"runtimes"`
	Components    []Component   `yaml:"components" json:"components"`
	Explore       Explore       `yaml:"explore" json:"explore"`

	// Format v2 (optional)
	Objectives []string       `yaml:"objectives,omitempty" json:"objectives,omitempty"`
	Stages     []Stage        `yaml:"stages,omitempty" json:"stages,omitempty"`
	Checks     []checks.Check `yaml:"checks,omitempty" json:"checks,omitempty"`

	// Runtime fields (not from YAML)
	Dir    string `yaml:"-" json:"-"`
	Active bool   `yaml:"-" json:"active"`
	Source string `yaml:"-" json:"source,omitempty"` // "" = in-repo; else the catalog pack name
}

// Stage is an ordered group of components that can be reasoned about
// independently (baseline, inject-failure, …). Stages install in declaration
// order; uninstall runs in reverse.
type Stage struct {
	Name        string      `yaml:"name" json:"name"`
	Description string      `yaml:"description,omitempty" json:"description,omitempty"`
	Components  []Component `yaml:"components" json:"components"`
}

// AllComponents returns the scenario's components in install order,
// regardless of whether they are declared flat (v1) or in stages (v2).
func (s *Scenario) AllComponents() []Component {
	if len(s.Stages) == 0 {
		return s.Components
	}
	var all []Component
	for _, st := range s.Stages {
		all = append(all, st.Components...)
	}
	return all
}

// stagesOrDefault returns the stage list, wrapping a v1 flat component list
// in a single anonymous stage so install logic has one code path.
func (s *Scenario) stagesOrDefault() []Stage {
	if len(s.Stages) > 0 {
		return s.Stages
	}
	return []Stage{{Components: s.Components}}
}

var validComponentTypes = map[string]bool{
	"helm": true, "manifest": true, "grafana-dashboard": true, "script": true,
}

// Scenario schema versioning. The engine supports the current and previous
// schema versions; unknown versions are rejected with an actionable error so a
// pack built for a newer engine fails clearly instead of misbehaving.
const DefaultScenarioAPIVersion = "scenario.flightdeck.dev/v2"

// SupportedScenarioAPIVersions lists the schema versions this engine understands
// (newest first). An empty apiVersion in a scenario is treated as the default.
var SupportedScenarioAPIVersions = []string{
	"scenario.flightdeck.dev/v2",
}

func scenarioAPIVersionSupported(v string) bool {
	if v == "" {
		return true // back-compat: defaults to DefaultScenarioAPIVersion
	}
	for _, s := range SupportedScenarioAPIVersions {
		if v == s {
			return true
		}
	}
	return false
}

// unsafePath reports whether a scenario asset path could escape the
// scenario directory (absolute, or containing a ".." segment).
func unsafePath(p string) bool {
	if p == "" {
		return false
	}
	if filepath.IsAbs(p) {
		return true
	}
	for _, seg := range strings.Split(filepath.ToSlash(p), "/") {
		if seg == ".." {
			return true
		}
	}
	return false
}

// Validate reports every schema problem in the scenario at once, naming the
// offending fields. It accepts all valid v1 scenarios unchanged.
func (s *Scenario) Validate() error {
	var errs []string
	add := func(format string, args ...interface{}) {
		errs = append(errs, fmt.Sprintf(format, args...))
	}

	if strings.TrimSpace(s.Name) == "" {
		add("name is required")
	}
	if !scenarioAPIVersionSupported(s.APIVersion) {
		add("unsupported apiVersion %q (supported: %s)",
			s.APIVersion, strings.Join(SupportedScenarioAPIVersions, ", "))
	}
	if len(s.Components) > 0 && len(s.Stages) > 0 {
		add("declare either components (v1) or stages (v2), not both")
	}

	stageNames := map[string]bool{}
	for i, st := range s.Stages {
		if strings.TrimSpace(st.Name) == "" {
			add("stage %d: name is required", i+1)
		} else if stageNames[st.Name] {
			add("stage %q: duplicate stage name", st.Name)
		} else {
			stageNames[st.Name] = true
		}
	}

	compNames := map[string]bool{}
	for _, c := range s.AllComponents() {
		where := fmt.Sprintf("component %q", c.Name)
		if strings.TrimSpace(c.Name) == "" {
			add("component with empty name")
			continue
		}
		if compNames[c.Name] {
			add("%s: duplicate component name", where)
		}
		compNames[c.Name] = true
		switch {
		case !validComponentTypes[c.Type]:
			add("%s: unknown type %q (expected helm | manifest | grafana-dashboard | script)", where, c.Type)
		case c.Type == "helm" && c.Chart == "":
			add("%s: helm component requires chart", where)
		case (c.Type == "manifest" || c.Type == "grafana-dashboard") && c.Path == "":
			add("%s: %s component requires path", where, c.Type)
		case c.Type == "script" && c.Script == "":
			add("%s: script component requires script", where)
		}
		// Asset paths must stay inside the scenario directory — external
		// packs (task 044) are untrusted, and in-repo scenarios have no
		// business escaping their dir either.
		for field, p := range map[string]string{"valuesFile": c.ValuesFile, "path": c.Path, "script": c.Script} {
			if unsafePath(p) {
				add("%s: %s %q must be a relative path inside the scenario directory", where, field, p)
			}
		}
	}

	checkNames := map[string]bool{}
	for _, c := range s.Checks {
		if err := c.Validate(); err != nil {
			add("%s", err.Error())
		}
		if c.Name != "" {
			if checkNames[c.Name] {
				add("check %q: duplicate check name", c.Name)
			}
			checkNames[c.Name] = true
		}
		if unsafePath(c.Script) {
			add("check %q: script %q must be a relative path inside the scenario directory", c.Name, c.Script)
		}
	}

	if len(errs) > 0 {
		name := s.Name
		if name == "" {
			name = "(unnamed)"
		}
		return fmt.Errorf("invalid scenario %q:\n  - %s", name, strings.Join(errs, "\n  - "))
	}
	return nil
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
	ProjectRoot         string
	DomainSuffix        string
	Profile             string // active runtime profile (k3d|aks|eks), used for preflight
	MonitoringNamespace string // namespace for monitoring/logging/tracing (default: "monitoring")
	scenarios           map[string]*Scenario
	loadErrors          map[string]error // scenario dir name → why it failed to load
	stateDir            string
}

// NewEngine creates a scenario engine by scanning the scenarios/ directory.
func NewEngine(projectRoot, domainSuffix, profile string, monitoringNamespace ...string) *Engine {
	ns := "monitoring"
	if len(monitoringNamespace) > 0 && monitoringNamespace[0] != "" {
		ns = monitoringNamespace[0]
	}
	e := &Engine{
		ProjectRoot:         projectRoot,
		DomainSuffix:        domainSuffix,
		Profile:             profile,
		MonitoringNamespace: ns,
		scenarios:           make(map[string]*Scenario),
		loadErrors:          make(map[string]error),
		stateDir:            filepath.Join(projectRoot, ".labctl", "scenarios"),
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

// Preflight validates a scenario before activation: checks runtime compatibility,
// prerequisite directory existence, and component asset file existence.
// It returns a combined error listing all failures so the user can fix them all at once.
func (e *Engine) Preflight(s *Scenario) error {
	var errs []string

	// 1. Runtime compatibility — only checked when the scenario restricts runtimes.
	if len(s.Runtimes) > 0 && e.Profile != "" {
		ok := false
		for _, r := range s.Runtimes {
			if r == e.Profile {
				ok = true
				break
			}
		}
		if !ok {
			errs = append(errs, fmt.Sprintf(
				"active profile %q is not in supported runtimes %v", e.Profile, s.Runtimes))
		}
	}

	// 2. Prerequisite apps — check apps/<name>/app.env exists.
	for _, app := range s.Prerequisites.Apps {
		appEnv := filepath.Join(e.ProjectRoot, "apps", app, "app.env")
		if _, err := os.Stat(appEnv); err != nil {
			errs = append(errs, fmt.Sprintf("prerequisite app %q not found (expected %s)", app, appEnv))
		}
	}

	// 3. Prerequisite platform components — check platform/<category>/ directory exists.
	for _, p := range s.Prerequisites.Platform {
		platformDir := filepath.Join(e.ProjectRoot, "platform", p)
		if _, err := os.Stat(platformDir); err != nil {
			errs = append(errs, fmt.Sprintf("prerequisite platform %q not found (expected %s)", p, platformDir))
		}
	}

	// 4. Component asset files.
	for _, comp := range s.AllComponents() {
		if comp.ValuesFile != "" {
			p := filepath.Join(s.Dir, comp.ValuesFile)
			if _, err := os.Stat(p); err != nil {
				errs = append(errs, fmt.Sprintf("component %q: valuesFile %q not found", comp.Name, p))
			}
		}
		if comp.Path != "" && (comp.Type == "manifest" || comp.Type == "grafana-dashboard") {
			p := filepath.Join(s.Dir, comp.Path)
			if _, err := os.Stat(p); err != nil {
				errs = append(errs, fmt.Sprintf("component %q: path %q not found", comp.Name, p))
			}
		}
		if comp.Script != "" {
			p := filepath.Join(s.Dir, comp.Script)
			if _, err := os.Stat(p); err != nil {
				errs = append(errs, fmt.Sprintf("component %q: script %q not found", comp.Name, p))
			}
		}
	}

	// 5. Check script files.
	for _, c := range s.Checks {
		if c.Type == checks.TypeScript && c.Script != "" && !filepath.IsAbs(c.Script) {
			p := filepath.Join(s.Dir, c.Script)
			if _, err := os.Stat(p); err != nil {
				errs = append(errs, fmt.Sprintf("check %q: script %q not found", c.Name, p))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("preflight failed for scenario %q:\n  - %s", s.Name, strings.Join(errs, "\n  - "))
	}
	return nil
}

// Up activates a scenario by installing all its components.
func (e *Engine) Up(name string, exec *executor.Executor) error {
	s, err := e.Get(name)
	if err != nil {
		return err
	}

	if e.isActive(name) {
		return fmt.Errorf("%w: %s", ErrAlreadyActive, name)
	}

	if err := e.Preflight(s); err != nil {
		return err
	}

	fmt.Printf("Activating scenario: %s\n", s.DisplayName)
	fmt.Printf("  %s\n\n", s.Description)

	if len(s.Objectives) > 0 {
		fmt.Println("Objectives:")
		for _, o := range s.Objectives {
			fmt.Printf("  - %s\n", o)
		}
		fmt.Println()
	}

	total := len(s.AllComponents())
	i := 0
	for _, st := range s.stagesOrDefault() {
		if st.Name != "" {
			fmt.Printf("=== Stage: %s ===\n", st.Name)
			if st.Description != "" {
				fmt.Printf("    %s\n", st.Description)
			}
		}
		for _, comp := range st.Components {
			i++
			fmt.Printf("[%d/%d] Installing %s (%s)...\n", i, total, comp.Name, comp.Type)
			if err := e.installComponent(s, &comp, exec); err != nil {
				return fmt.Errorf("installing component %s: %w", comp.Name, err)
			}
		}
	}

	// Mark as active
	if err := e.markActive(name); err != nil {
		return fmt.Errorf("marking scenario active: %w", err)
	}

	if len(s.Checks) > 0 {
		fmt.Printf("\nThis scenario has %d verifiable checks. Run: labctl scenario verify %s\n", len(s.Checks), s.Name)
	}

	// Print explore hints
	e.printExploreHints(s)

	return nil
}

// Verify runs the scenario's checks (template-resolved) through the supplied
// runner and returns one result per check. It does not require the scenario
// to be active — verifying an inactive scenario is how you prove it is down.
func (e *Engine) Verify(ctx context.Context, name string, runner *checks.Runner) ([]checks.Result, error) {
	s, err := e.Get(name)
	if err != nil {
		return nil, err
	}
	if len(s.Checks) == 0 {
		return nil, fmt.Errorf("%w: %s (add a checks block — see docs/scenarios.md)", ErrNoChecks, name)
	}

	runner.ScriptDir = s.Dir
	resolved := make([]checks.Check, len(s.Checks))
	for i, c := range s.Checks {
		resolved[i] = e.resolveCheck(c)
	}
	return runner.RunAll(ctx, resolved), nil
}

// resolveCheck resolves template variables in a check's templatable fields.
func (e *Engine) resolveCheck(c checks.Check) checks.Check {
	c.URL = e.resolveTemplate(c.URL)
	c.BodyContains = e.resolveTemplate(c.BodyContains)
	c.Resource = e.resolveTemplate(c.Resource)
	c.Namespace = e.resolveTemplate(c.Namespace)
	c.JSONPath = e.resolveTemplate(c.JSONPath)
	c.Query = e.resolveTemplate(c.Query)
	c.Value = e.resolveTemplate(c.Value)
	return c
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

	// Uninstall in reverse order (across all stages)
	all := s.AllComponents()
	for i := len(all) - 1; i >= 0; i-- {
		comp := all[i]
		fmt.Printf("[%d/%d] Uninstalling %s...\n", len(all)-i, len(all), comp.Name)
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
			Description: s.Description,
			Category:    s.Category,
			Runtimes:    s.Runtimes,
			Active:      e.isActive(s.Name),
		})
	}
	return result
}

// ScenarioStatus is a lightweight status for listing scenarios.
type ScenarioStatus struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"displayName"`
	Description string   `json:"description,omitempty"`
	Category    string   `json:"category"`
	Runtimes    []string `json:"runtimes,omitempty"`
	Active      bool     `json:"active"`
}

func (e *Engine) scan() {
	// In-repo scenarios scan first, so on name collisions they always win
	// over catalog packs.
	scenariosDir := filepath.Join(e.ProjectRoot, "scenarios")
	if entries, err := os.ReadDir(scenariosDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			e.loadInto(filepath.Join(scenariosDir, entry.Name()), "", entry.Name())
		}
	}

	// Installed catalog packs (task 044) live under .labctl/catalog/<pack>/.
	packs, err := os.ReadDir(e.CatalogDir())
	if err != nil {
		return
	}
	for _, pack := range packs {
		if pack.IsDir() {
			e.scanPack(filepath.Join(e.CatalogDir(), pack.Name()), pack.Name())
		}
	}
}

// scanPack loads scenarios from an installed pack: either a single scenario
// at the pack root, or a collection of scenario directories.
func (e *Engine) scanPack(dir, packName string) {
	if _, err := os.Stat(filepath.Join(dir, "scenario.yaml")); err == nil {
		e.loadInto(dir, packName, packName)
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Non-scenario dirs inside a pack (docs/, .github/, …) are fine.
		if _, err := os.Stat(filepath.Join(dir, entry.Name(), "scenario.yaml")); err != nil {
			continue
		}
		e.loadInto(filepath.Join(dir, entry.Name()), packName, packName+"/"+entry.Name())
	}
}

// loadInto loads one scenario dir, recording failures and collisions under
// the given key. One broken scenario must not hide the rest; LoadErrors()
// surfaces these (and the repo test fails CI on in-repo ones).
func (e *Engine) loadInto(dir, source, key string) {
	s, err := e.loadScenario(filepath.Join(dir, "scenario.yaml"))
	if err != nil {
		e.loadErrors[key] = err
		return
	}
	if existing, ok := e.scenarios[s.Name]; ok {
		from := "the repository"
		if existing.Source != "" {
			from = "pack " + existing.Source
		}
		e.loadErrors[key] = fmt.Errorf("scenario name %q already provided by %s — skipped", s.Name, from)
		return
	}
	s.Dir = dir
	s.Source = source
	e.scenarios[s.Name] = s
}

// rescan rebuilds the scenario index from disk (used after pack changes).
func (e *Engine) rescan() {
	e.scenarios = make(map[string]*Scenario)
	e.loadErrors = make(map[string]error)
	e.scan()
}

// LoadErrors returns scenarios that were discovered but failed to load or
// validate, keyed by directory name.
func (e *Engine) LoadErrors() map[string]error {
	out := make(map[string]error, len(e.loadErrors))
	for k, v := range e.loadErrors {
		out[k] = v
	}
	return out
}

func (e *Engine) loadScenario(path string) (*Scenario, error) {
	return loadScenarioFile(path)
}

// loadScenarioFile parses and validates a scenario.yaml. It is package-level
// so pack validation (catalog.go) can use it without an Engine.
func loadScenarioFile(path string) (*Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var s Scenario
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
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
	ns := e.componentNamespace(comp, "default")

	// Add helm repo if specified
	if comp.Repo != "" {
		repoName := strings.Split(comp.Chart, "/")[0]
		exec.RunCommandStreamed("Helm repo add "+repoName, "helm", "repo", "add", repoName, comp.Repo, "--force-update")
		exec.RunCommandStreamed("Helm repo update", "helm", "repo", "update")
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
			resolvedValuesPath, cleanup, err := e.resolveFileTemplate(valuesPath, "labctl-values-*.yaml")
			if err != nil {
				return err
			}
			defer cleanup()
			args = append(args, "-f", resolvedValuesPath)
		}
	}

	for k, v := range comp.Set {
		resolved := e.resolveTemplate(v)
		args = append(args, "--set", k+"="+resolved)
	}

	_, err := exec.RunCommandStreamed("Helm install "+comp.Name, "helm", args...)
	return err
}

func (e *Engine) uninstallHelm(comp *Component, exec *executor.Executor) error {
	ns := comp.Namespace
	if ns == "" {
		ns = "default"
	}
	_, err := exec.RunCommandStreamed("Helm uninstall "+comp.Name, "helm", "uninstall", comp.Name, "--namespace", ns)
	return err
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
	ns := e.componentNamespace(comp, "")
	if ns != "" && !manifestHasExplicitNamespace(resolved) {
		args = append(args, "--namespace", ns)
	}

	_, err = exec.RunCommandStreamed("Apply manifest "+comp.Name, "kubectl", args...)
	return err
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
	ns := e.componentNamespace(comp, "")
	if ns != "" && !manifestHasExplicitNamespace(resolved) {
		args = append(args, "--namespace", ns)
	}

	_, err = exec.RunCommandStreamed("Delete manifest "+comp.Name, "kubectl", args...)
	return err
}

func (e *Engine) installGrafanaDashboard(s *Scenario, comp *Component, exec *executor.Executor) error {
	dashDir := filepath.Join(s.Dir, comp.Path)
	entries, err := os.ReadDir(dashDir)
	if err != nil {
		return fmt.Errorf("reading dashboard dir: %w", err)
	}

	ns := comp.Namespace
	if ns == "" {
		ns = e.MonitoringNamespace
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
		exec.RunCommandStreamed("Apply dashboard "+entry.Name(), "kubectl", "apply", "-f", tmpFile.Name())
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
		ns = e.MonitoringNamespace
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		cmName := fmt.Sprintf("scenario-%s-%s", s.Name, strings.TrimSuffix(entry.Name(), ".json"))
		exec.RunCommandStreamed("Delete dashboard "+entry.Name(), "kubectl", "delete", "configmap", cmName, "--namespace", ns, "--ignore-not-found")
	}

	return nil
}

func (e *Engine) runScript(s *Scenario, comp *Component, exec *executor.Executor) error {
	// Compute the path relative to the project root for RunScriptStreamed
	relPath, err := filepath.Rel(e.ProjectRoot, filepath.Join(s.Dir, comp.Script))
	if err != nil {
		// Fallback to absolute path
		relPath = filepath.Join(s.Dir, comp.Script)
	}
	_, err = exec.RunScriptStreamed("Run script "+comp.Name, relPath)
	return err
}

func (e *Engine) componentNamespace(comp *Component, defaultNamespace string) string {
	ns := strings.TrimSpace(comp.Namespace)
	if ns == "" {
		return defaultNamespace
	}
	return e.resolveTemplate(ns)
}

func (e *Engine) resolveFileTemplate(path, pattern string) (string, func(), error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", func() {}, fmt.Errorf("reading values file %s: %w", path, err)
	}
	resolved := e.resolveTemplate(string(data))
	if resolved == string(data) {
		return path, func() {}, nil
	}

	tmpFile, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", func() {}, err
	}
	if _, err := tmpFile.WriteString(resolved); err != nil {
		name := tmpFile.Name()
		tmpFile.Close()
		os.Remove(name)
		return "", func() {}, err
	}
	if err := tmpFile.Close(); err != nil {
		name := tmpFile.Name()
		os.Remove(name)
		return "", func() {}, err
	}
	return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }, nil
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
		"DomainSuffix":        e.DomainSuffix,
		"ProjectRoot":         e.ProjectRoot,
		"MonitoringNamespace": e.MonitoringNamespace,
		"LokiRetentionPeriod": lokiRetentionPeriod(),
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return input
	}
	return buf.String()
}

func lokiRetentionPeriod() string {
	hours := strings.TrimSpace(os.Getenv("LOKI_RETENTION_HOURS"))
	if hours == "" {
		hours = "168"
	}
	return hours + "h"
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
			fmt.Printf("  - %s\n", e.resolveTemplate(t))
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
