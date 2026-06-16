// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"bytes"
	"strings"
	"testing"
)

// executeHelp runs the cobra command tree with the given args and captures
// stdout/stderr output. Because --help causes cobra to exit 0 without running
// PersistentPreRunE, these tests are fully hermetic.
func executeHelp(t *testing.T, args ...string) string {
	t.Helper()
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs(args)
	// cobra's Execute returns nil for --help
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute(%v): %v", args, err)
	}
	return buf.String()
}

func TestScenarioCmd_Help(t *testing.T) {
	out := executeHelp(t, "scenario", "--help")
	if out == "" {
		t.Error("expected non-empty --help output for 'scenario'")
	}
	if !strings.Contains(out, "scenario") {
		t.Errorf("help output should mention 'scenario': %q", out)
	}
}

func TestIncidentCmd_Help(t *testing.T) {
	out := executeHelp(t, "incident", "--help")
	if out == "" {
		t.Error("expected non-empty --help output for 'incident'")
	}
	if !strings.Contains(strings.ToLower(out), "incident") {
		t.Errorf("help output should mention 'incident': %q", out)
	}
}

func TestLearnCmd_Help(t *testing.T) {
	out := executeHelp(t, "learn", "--help")
	if out == "" {
		t.Error("expected non-empty --help output for 'learn'")
	}
	if !strings.Contains(strings.ToLower(out), "learn") {
		t.Errorf("help output should mention 'learn': %q", out)
	}
}

func TestPlatformCmd_Help(t *testing.T) {
	out := executeHelp(t, "platform", "--help")
	if out == "" {
		t.Error("expected non-empty --help output for 'platform'")
	}
	if !strings.Contains(strings.ToLower(out), "platform") {
		t.Errorf("help output should mention 'platform': %q", out)
	}
}

func TestPackCmd_Help(t *testing.T) {
	out := executeHelp(t, "pack", "--help")
	if out == "" {
		t.Error("expected non-empty --help output for 'pack'")
	}
	if !strings.Contains(strings.ToLower(out), "pack") {
		t.Errorf("help output should mention 'pack': %q", out)
	}
}
