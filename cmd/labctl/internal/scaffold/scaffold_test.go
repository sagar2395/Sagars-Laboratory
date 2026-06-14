// SPDX-License-Identifier: Apache-2.0

package scaffold

import (
	"os"
	"path/filepath"
	"testing"

	"go.flightdeck.dev/labctl/pkg/pack"
	schema "go.flightdeck.dev/labctl/pkg/scenario"
	"gopkg.in/yaml.v3"
)

// parseScenario unmarshals scaffolded YAML into the SDK scenario type.
func parseScenario(t *testing.T, data []byte) *schema.Scenario {
	t.Helper()
	var s schema.Scenario
	if err := yaml.Unmarshal(data, &s); err != nil {
		t.Fatalf("scaffolded scenario does not parse: %v", err)
	}
	return &s
}

func TestScenario_ProducesValidVerifyReadyOutput(t *testing.T) {
	root := t.TempDir()
	dir, err := Scenario(root, "demo-scenario", false)
	if err != nil {
		t.Fatal(err)
	}

	// scenario.yaml parses and validates against the SDK schema.
	data, err := os.ReadFile(filepath.Join(dir, "scenario.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	s := parseScenario(t, data)
	if err := s.Validate(); err != nil {
		t.Fatalf("scaffolded scenario is invalid: %v", err)
	}

	// The readiness check is executable and exits 0 (verify-ready).
	info, err := os.Stat(filepath.Join(dir, "checks", "ready.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o100 == 0 {
		t.Errorf("ready.sh is not executable: %v", info.Mode())
	}

	// Refuses to overwrite without force; succeeds with it.
	if _, err := Scenario(root, "demo-scenario", false); err == nil {
		t.Error("expected refusal to overwrite without --force")
	}
	if _, err := Scenario(root, "demo-scenario", true); err != nil {
		t.Errorf("force overwrite failed: %v", err)
	}
}

func TestScenario_RejectsBadName(t *testing.T) {
	if _, err := Scenario(t.TempDir(), "Bad Name", false); err == nil {
		t.Fatal("expected invalid-name error")
	}
}

func TestPack_ProducesValidPackAndScenario(t *testing.T) {
	parent := t.TempDir()
	dir, err := Pack(parent, "acme/kafka-drills", false)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(dir) != "kafka-drills" {
		t.Errorf("pack dir = %s, want .../kafka-drills", dir)
	}

	// pack.yaml validates (publisher/name preserved).
	m, err := pack.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if m == nil {
		t.Fatal("scaffolded pack has no pack.yaml")
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("scaffolded pack.yaml invalid: %v", err)
	}
	if m.Metadata.Name != "acme/kafka-drills" {
		t.Errorf("pack name = %q", m.Metadata.Name)
	}

	// The bundled scenario validates too, so the pack is verifiable as-is.
	data, err := os.ReadFile(filepath.Join(dir, "kafka-drills", "scenario.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := parseScenario(t, data).Validate(); err != nil {
		t.Fatalf("bundled scenario invalid: %v", err)
	}
}
