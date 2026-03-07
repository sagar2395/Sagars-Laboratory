package executor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Executor runs shell scripts from the project, streaming output to the caller.
type Executor struct {
	ProjectRoot string
	Env         map[string]string
	Stdout      io.Writer
	Stderr      io.Writer
}

// New creates an Executor rooted at projectRoot.
func New(projectRoot string) *Executor {
	return &Executor{
		ProjectRoot: projectRoot,
		Env:         make(map[string]string),
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
	}
}

// RunScript executes a shell script relative to the project root.
func (e *Executor) RunScript(scriptPath string, args ...string) error {
	absPath := filepath.Join(e.ProjectRoot, scriptPath)
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("script not found: %s", absPath)
	}

	cmd := exec.Command("bash", append([]string{absPath}, args...)...)
	cmd.Dir = e.ProjectRoot
	cmd.Stdout = e.Stdout
	cmd.Stderr = e.Stderr
	cmd.Env = e.buildEnv()

	return cmd.Run()
}

// RunCommand executes an arbitrary command in the project root.
func (e *Executor) RunCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = e.ProjectRoot
	cmd.Stdout = e.Stdout
	cmd.Stderr = e.Stderr
	cmd.Env = e.buildEnv()

	return cmd.Run()
}

// RunMake executes a Make target.
func (e *Executor) RunMake(target string, vars ...string) error {
	args := []string{target}
	for _, v := range vars {
		args = append(args, v)
	}
	return e.RunCommand("make", args...)
}

// RunHelm executes a helm command.
func (e *Executor) RunHelm(args ...string) error {
	return e.RunCommand("helm", args...)
}

// RunKubectl executes a kubectl command.
func (e *Executor) RunKubectl(args ...string) error {
	return e.RunCommand("kubectl", args...)
}

// SetEnv adds an environment variable to pass to executed commands.
func (e *Executor) SetEnv(key, value string) {
	e.Env[key] = value
}

// CaptureOutput runs a command and returns its stdout as a string.
func (e *Executor) CaptureOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = e.ProjectRoot
	cmd.Env = e.buildEnv()

	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("%s: %s", err, string(exitErr.Stderr))
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (e *Executor) buildEnv() []string {
	env := os.Environ()
	for k, v := range e.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return env
}
