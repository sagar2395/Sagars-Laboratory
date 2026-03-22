package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sagars-lab/labctl/internal/executor"
)

// Provider represents a platform component provider (e.g., traefik, nginx).
type Provider struct {
	Category string // e.g., "ingress", "monitoring/metrics"
	Name     string // e.g., "traefik", "prometheus"
	Path     string // Filesystem path to the provider directory
}

// HasScript checks if the provider has a specific script.
func (p *Provider) HasScript(name string) bool {
	_, err := os.Stat(filepath.Join(p.Path, name))
	return err == nil
}

// Namespace returns the conventional Kubernetes namespace for this provider.
// Monitoring, logging, and tracing providers share the "monitoring" namespace.
// Other providers use their own name as the namespace.
func (p *Provider) Namespace() string {
	top := p.Category
	if i := strings.Index(top, "/"); i >= 0 {
		top = top[:i]
	}
	switch top {
	case "monitoring", "logging", "tracing":
		return "monitoring"
	default:
		return p.Name
	}
}

// Registry discovers and manages platform component providers.
type Registry struct {
	ProjectRoot string
	providers   map[string][]Provider // category -> providers
}

// NewRegistry scans the platform/ directory for available providers.
func NewRegistry(projectRoot string) *Registry {
	r := &Registry{
		ProjectRoot: projectRoot,
		providers:   make(map[string][]Provider),
	}
	r.scan()
	return r
}

// GetProviders returns all providers for a given category.
func (r *Registry) GetProviders(category string) []Provider {
	return r.providers[category]
}

// GetProvider returns a specific provider by category and name.
func (r *Registry) GetProvider(category, name string) (*Provider, error) {
	for _, p := range r.providers[category] {
		if p.Name == name {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("provider %s/%s not found", category, name)
}

// Categories returns all discovered categories.
func (r *Registry) Categories() []string {
	var cats []string
	for k := range r.providers {
		cats = append(cats, k)
	}
	return cats
}

// Install runs the install.sh for a provider.
func (r *Registry) Install(category, name string, exec *executor.Executor) error {
	p, err := r.GetProvider(category, name)
	if err != nil {
		return err
	}
	scriptPath, err := filepath.Rel(r.ProjectRoot, filepath.Join(p.Path, "install.sh"))
	if err != nil {
		return err
	}
	return exec.RunScript(scriptPath)
}

// InstallStreamed runs install.sh for a provider with output streaming.
func (r *Registry) InstallStreamed(category, name string, exec *executor.Executor) error {
	p, err := r.GetProvider(category, name)
	if err != nil {
		return err
	}
	scriptPath, err := filepath.Rel(r.ProjectRoot, filepath.Join(p.Path, "install.sh"))
	if err != nil {
		return err
	}
	return exec.RunScriptStreamed(fmt.Sprintf("Install %s/%s", category, name), scriptPath)
}

// Uninstall runs the uninstall.sh for a provider.
func (r *Registry) Uninstall(category, name string, exec *executor.Executor) error {
	p, err := r.GetProvider(category, name)
	if err != nil {
		return err
	}
	script := filepath.Join(p.Path, "uninstall.sh")
	if _, err := os.Stat(script); os.IsNotExist(err) {
		return fmt.Errorf("uninstall.sh not found for %s/%s", category, name)
	}
	scriptPath, err := filepath.Rel(r.ProjectRoot, script)
	if err != nil {
		return err
	}
	return exec.RunScript(scriptPath)
}

// UninstallStreamed runs uninstall.sh for a provider with output streaming.
func (r *Registry) UninstallStreamed(category, name string, exec *executor.Executor) error {
	p, err := r.GetProvider(category, name)
	if err != nil {
		return err
	}
	script := filepath.Join(p.Path, "uninstall.sh")
	if _, err := os.Stat(script); os.IsNotExist(err) {
		return fmt.Errorf("uninstall.sh not found for %s/%s", category, name)
	}
	scriptPath, err := filepath.Rel(r.ProjectRoot, script)
	if err != nil {
		return err
	}
	return exec.RunScriptStreamed(fmt.Sprintf("Uninstall %s/%s", category, name), scriptPath)
}

// Status runs the status.sh for a provider.
func (r *Registry) Status(category, name string, exec *executor.Executor) error {
	p, err := r.GetProvider(category, name)
	if err != nil {
		return err
	}
	script := filepath.Join(p.Path, "status.sh")
	if _, err := os.Stat(script); os.IsNotExist(err) {
		return fmt.Errorf("status.sh not found for %s/%s", category, name)
	}
	scriptPath, err := filepath.Rel(r.ProjectRoot, script)
	if err != nil {
		return err
	}
	return exec.RunScript(scriptPath)
}

func (r *Registry) scan() {
	platformDir := filepath.Join(r.ProjectRoot, "platform")
	r.scanDir(platformDir, "")
}

func (r *Registry) scanDir(dir, prefix string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "_schema" {
			continue
		}

		fullPath := filepath.Join(dir, entry.Name())
		category := entry.Name()
		if prefix != "" {
			category = prefix + "/" + entry.Name()
		}

		// Check if this directory is a provider (has install.sh)
		if _, err := os.Stat(filepath.Join(fullPath, "install.sh")); err == nil {
			r.providers[prefix] = append(r.providers[prefix], Provider{
				Category: prefix,
				Name:     entry.Name(),
				Path:     fullPath,
			})
		} else {
			// Recurse one level deeper
			r.scanDir(fullPath, category)
		}
	}
}
