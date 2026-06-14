// SPDX-License-Identifier: Apache-2.0
package pack

import (
	"strings"
	"testing"
)

const validYAML = `
apiVersion: packs.flightdeck.dev/v1
kind: ScenarioPack
metadata:
  name: acme/kafka-drills
  version: 1.4.2
  publisher: acme
  license: Apache-2.0
  tier: community
spec:
  engine:
    scenarioApiVersions: ["scenario.flightdeck.dev/v2"]
    minLabctlVersion: "1.0.0"
  requires:
    platform: [data/kafka]
  provides:
    scenarios: [kafka-broker-outage]
`

func TestParseAndValidate_OK(t *testing.T) {
	m, err := Parse([]byte(validYAML))
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("expected valid, got: %v", err)
	}
	if m.Tier() != "community" {
		t.Errorf("tier = %q", m.Tier())
	}
}

func TestValidate_ReportsProblems(t *testing.T) {
	m := &Manifest{APIVersion: "wrong", Kind: "Nope",
		Metadata: Metadata{Name: "Bad Name", Version: "notsemver", Tier: "gold"}}
	err := m.Validate()
	if err == nil {
		t.Fatal("expected errors")
	}
	for _, want := range []string{"apiVersion", "kind", "metadata.name", "not valid semver", "tier", "license is required"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("missing %q in: %v", want, err)
		}
	}
}

func TestCheckEngineCompat(t *testing.T) {
	m, _ := Parse([]byte(validYAML))
	supported := []string{"scenario.flightdeck.dev/v2"}

	// dev build skips the version check; apiVersion supported => OK
	if err := m.CheckEngineCompat("dev", supported); err != nil {
		t.Fatalf("dev should be compatible: %v", err)
	}
	// older engine version => incompatible
	if err := m.CheckEngineCompat("0.9.0", supported); err == nil || !strings.Contains(err.Error(), "requires labctl >= 1.0.0") {
		t.Fatalf("expected version incompatibility, got: %v", err)
	}
	// unsupported scenario apiVersion => incompatible
	if err := m.CheckEngineCompat("2.0.0", []string{"scenario.flightdeck.dev/v1"}); err == nil || !strings.Contains(err.Error(), "scenario apiVersion") {
		t.Fatalf("expected scenario apiVersion incompatibility, got: %v", err)
	}
}

func TestLoad_MissingIsNil(t *testing.T) {
	m, err := Load(t.TempDir())
	if err != nil || m != nil {
		t.Fatalf("missing pack.yaml should be (nil,nil), got (%v,%v)", m, err)
	}
}
