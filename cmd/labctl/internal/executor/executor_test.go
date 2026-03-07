package executor

import (
	"bytes"
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	exec := New("/tmp/test")
	if exec.ProjectRoot != "/tmp/test" {
		t.Errorf("ProjectRoot: got %q, want %q", exec.ProjectRoot, "/tmp/test")
	}
	if exec.Env == nil {
		t.Error("Env map should be initialized")
	}
	if exec.Stdout != os.Stdout {
		t.Error("Stdout should default to os.Stdout")
	}
	if exec.Stderr != os.Stderr {
		t.Error("Stderr should default to os.Stderr")
	}
}

func TestSetEnv(t *testing.T) {
	exec := New("/tmp/test")
	exec.SetEnv("FOO", "bar")
	exec.SetEnv("BAZ", "qux")

	if exec.Env["FOO"] != "bar" {
		t.Errorf("Env[FOO]: got %q, want %q", exec.Env["FOO"], "bar")
	}
	if exec.Env["BAZ"] != "qux" {
		t.Errorf("Env[BAZ]: got %q, want %q", exec.Env["BAZ"], "qux")
	}
}

func TestSetEnv_Overwrite(t *testing.T) {
	exec := New("/tmp/test")
	exec.SetEnv("KEY", "first")
	exec.SetEnv("KEY", "second")

	if exec.Env["KEY"] != "second" {
		t.Errorf("Env[KEY]: got %q, want %q", exec.Env["KEY"], "second")
	}
}

func TestRunCommand_Echo(t *testing.T) {
	var buf bytes.Buffer
	exec := New(t.TempDir())
	exec.Stdout = &buf
	exec.Stderr = &buf

	err := exec.RunCommand("echo", "hello")
	if err != nil {
		t.Fatalf("RunCommand echo: %v", err)
	}

	output := buf.String()
	if output != "hello\n" {
		t.Errorf("output: got %q, want %q", output, "hello\n")
	}
}

func TestCaptureOutput(t *testing.T) {
	exec := New(t.TempDir())

	out, err := exec.CaptureOutput("echo", "captured")
	if err != nil {
		t.Fatalf("CaptureOutput echo: %v", err)
	}

	if out != "captured" {
		t.Errorf("output: got %q, want %q", out, "captured")
	}
}

func TestRunScript(t *testing.T) {
	root := t.TempDir()
	script := "test-script.sh"
	os.WriteFile(root+"/"+script, []byte("#!/bin/bash\necho script-ran"), 0755)

	var buf bytes.Buffer
	exec := New(root)
	exec.Stdout = &buf

	err := exec.RunScript(script)
	if err != nil {
		t.Fatalf("RunScript: %v", err)
	}

	if buf.String() != "script-ran\n" {
		t.Errorf("output: got %q, want %q", buf.String(), "script-ran\n")
	}
}

func TestRunScript_NotFound(t *testing.T) {
	exec := New(t.TempDir())
	err := exec.RunScript("nonexistent.sh")
	if err == nil {
		t.Error("expected error for missing script")
	}
}

func TestRunCommand_Failure(t *testing.T) {
	exec := New(t.TempDir())
	exec.Stdout = &bytes.Buffer{}
	exec.Stderr = &bytes.Buffer{}

	err := exec.RunCommand("false")
	if err == nil {
		t.Error("expected error for failed command")
	}
}

func TestBuildEnv(t *testing.T) {
	exec := New(t.TempDir())
	exec.SetEnv("CUSTOM_VAR", "custom_value")

	env := exec.buildEnv()

	found := false
	for _, e := range env {
		if e == "CUSTOM_VAR=custom_value" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected CUSTOM_VAR=custom_value in build environment")
	}
}
