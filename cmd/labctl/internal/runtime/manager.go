package runtime

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sagars-lab/labctl/internal/executor"
	"github.com/sagars-lab/labctl/internal/k8s"
)

// RuntimeInfo describes a discovered runtime and its status.
type RuntimeInfo struct {
	Name    string `json:"name"`
	Active  bool   `json:"active"`  // cluster exists and is reachable
	Current bool   `json:"current"` // is the current kubectl context
}

// Manager discovers and manages cluster runtimes (k3d, aks, eks, etc.).
type Manager struct {
	ProjectRoot string
	ClusterName string
	runtimes    []string // discovered runtime directory names
}

// NewManager scans the runtimes/ directory for available runtimes.
func NewManager(projectRoot, clusterName string) *Manager {
	m := &Manager{
		ProjectRoot: projectRoot,
		ClusterName: clusterName,
	}
	m.scan()
	return m
}

// List returns all discovered runtimes with their current status.
func (m *Manager) List() []RuntimeInfo {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	currentCtx, _ := k8s.GetCurrentContext(ctx)

	var infos []RuntimeInfo
	for _, name := range m.runtimes {
		info := RuntimeInfo{Name: name}

		// Determine expected context name for this runtime
		expectedCtx := m.expectedContext(name)

		// Check if this is the current context
		if expectedCtx != "" && currentCtx == expectedCtx {
			info.Current = true
			info.Active = true
		} else if expectedCtx != "" {
			// Check if context exists in kubeconfig (local check, no network)
			checkCtx, checkCancel := context.WithTimeout(context.Background(), 2*time.Second)
			out, err := k8s.RunKubectl(checkCtx, "config", "get-contexts", expectedCtx, "--no-headers")
			checkCancel()
			if err == nil && strings.TrimSpace(out) != "" {
				info.Active = true
			}
		}

		infos = append(infos, info)
	}
	return infos
}

// Names returns the names of all discovered runtimes.
func (m *Manager) Names() []string {
	return m.runtimes
}

// Activate provisions a runtime by running its up.sh script.
func (m *Manager) Activate(name string, exec *executor.Executor) error {
	if !m.exists(name) {
		return fmt.Errorf("runtime %q not found", name)
	}
	scriptPath := filepath.Join("runtimes", name, "up.sh")
	return exec.RunScriptStreamed(fmt.Sprintf("Activate %s", name), scriptPath, m.ClusterName)
}

// Deactivate tears down a runtime by running its down.sh script.
func (m *Manager) Deactivate(name string, exec *executor.Executor) error {
	if !m.exists(name) {
		return fmt.Errorf("runtime %q not found", name)
	}
	scriptPath := filepath.Join("runtimes", name, "down.sh")
	return exec.RunScriptStreamed(fmt.Sprintf("Deactivate %s", name), scriptPath, m.ClusterName)
}

func (m *Manager) scan() {
	runtimesDir := filepath.Join(m.ProjectRoot, "runtimes")
	entries, err := os.ReadDir(runtimesDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Only include directories that have an up.sh
		upScript := filepath.Join(runtimesDir, entry.Name(), "up.sh")
		if _, err := os.Stat(upScript); err == nil {
			m.runtimes = append(m.runtimes, entry.Name())
		}
	}
}

func (m *Manager) exists(name string) bool {
	for _, r := range m.runtimes {
		if r == name {
			return true
		}
	}
	return false
}

// expectedContext returns the kubectl context name expected for a runtime.
func (m *Manager) expectedContext(name string) string {
	switch name {
	case "k3d":
		return "k3d-" + m.ClusterName
	default:
		// For cloud runtimes (aks, eks), the context name varies.
		// Return the runtime name as a best-effort match.
		return name
	}
}
