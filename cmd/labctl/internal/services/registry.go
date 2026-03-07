package services

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sagars-lab/labctl/internal/executor"
)

// Service represents a shared service (e.g., redis, postgres).
type Service struct {
	Name string // directory name under services/
	Path string // absolute filesystem path
}

// HasScript checks if the service has a specific script.
func (s *Service) HasScript(name string) bool {
	_, err := os.Stat(filepath.Join(s.Path, name))
	return err == nil
}

// Registry discovers and manages shared services.
type Registry struct {
	ProjectRoot string
	services    []Service
}

// NewRegistry scans the services/ directory for available services.
func NewRegistry(projectRoot string) *Registry {
	r := &Registry{
		ProjectRoot: projectRoot,
	}
	r.scan()
	return r
}

// List returns all discovered services.
func (r *Registry) List() []Service {
	return r.services
}

// Get returns a specific service by name.
func (r *Registry) Get(name string) (*Service, error) {
	for _, s := range r.services {
		if s.Name == name {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("service %q not found", name)
}

// Install runs install.sh for a service.
func (r *Registry) Install(name string, exec *executor.Executor) error {
	s, err := r.Get(name)
	if err != nil {
		return err
	}
	script := filepath.Join(s.Path, "install.sh")
	if _, err := os.Stat(script); os.IsNotExist(err) {
		return fmt.Errorf("install.sh not found for service %q", name)
	}
	rel, err := filepath.Rel(r.ProjectRoot, script)
	if err != nil {
		return err
	}
	return exec.RunScript(rel)
}

// Uninstall runs uninstall.sh for a service.
func (r *Registry) Uninstall(name string, exec *executor.Executor) error {
	s, err := r.Get(name)
	if err != nil {
		return err
	}
	script := filepath.Join(s.Path, "uninstall.sh")
	if _, err := os.Stat(script); os.IsNotExist(err) {
		return fmt.Errorf("uninstall.sh not found for service %q", name)
	}
	rel, err := filepath.Rel(r.ProjectRoot, script)
	if err != nil {
		return err
	}
	return exec.RunScript(rel)
}

// Status runs status.sh for a service.
func (r *Registry) Status(name string, exec *executor.Executor) error {
	s, err := r.Get(name)
	if err != nil {
		return err
	}
	script := filepath.Join(s.Path, "status.sh")
	if _, err := os.Stat(script); os.IsNotExist(err) {
		return fmt.Errorf("status.sh not found for service %q", name)
	}
	rel, err := filepath.Rel(r.ProjectRoot, script)
	if err != nil {
		return err
	}
	return exec.RunScript(rel)
}

func (r *Registry) scan() {
	servicesDir := filepath.Join(r.ProjectRoot, "services")
	entries, err := os.ReadDir(servicesDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		fullPath := filepath.Join(servicesDir, entry.Name())
		// A service must have install.sh
		if _, err := os.Stat(filepath.Join(fullPath, "install.sh")); err == nil {
			r.services = append(r.services, Service{
				Name: entry.Name(),
				Path: fullPath,
			})
		}
	}
}
