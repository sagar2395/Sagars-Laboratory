// SPDX-License-Identifier: Apache-2.0

// Package pack defines the Flightdeck scenario-pack manifest (pack.yaml) and its
// validation. A pack is a versioned, self-contained unit of content (scenarios
// and optionally platform modules / incidents / learning material) that can be
// installed independently from git or an OCI registry. This package is part of
// the public SDK (docs/authoring/sdk-stability-policy.md) and must not import
// anything under internal/.
package pack

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Manifest schema identity.
const (
	APIVersion   = "packs.flightdeck.dev/v1"
	Kind         = "ScenarioPack"
	ManifestFile = "pack.yaml"
)

// Tiers. The OSS engine treats all tiers identically; entitlement is enforced at
// the distribution layer (registry auth) + the entitlement interface (task 070),
// never by the engine. Tier is metadata for discovery/display.
const (
	TierCommunity  = "community"
	TierPremium    = "premium"
	TierEnterprise = "enterprise"
)

var validTiers = map[string]bool{TierCommunity: true, TierPremium: true, TierEnterprise: true}

// nameRe allows an optional "publisher/" prefix then the pack name.
var nameRe = regexp.MustCompile(`^([a-z0-9][a-z0-9-]{0,38}/)?[a-z0-9][a-z0-9-]{0,62}$`)

// Manifest is a parsed pack.yaml.
type Manifest struct {
	APIVersion string   `yaml:"apiVersion" json:"apiVersion"`
	Kind       string   `yaml:"kind" json:"kind"`
	Metadata   Metadata `yaml:"metadata" json:"metadata"`
	Spec       Spec     `yaml:"spec" json:"spec"`
}

// Metadata identifies and describes the pack.
type Metadata struct {
	Name        string   `yaml:"name" json:"name"`
	Version     string   `yaml:"version" json:"version"` // SemVer
	Publisher   string   `yaml:"publisher,omitempty" json:"publisher,omitempty"`
	DisplayName string   `yaml:"displayName,omitempty" json:"displayName,omitempty"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
	Homepage    string   `yaml:"homepage,omitempty" json:"homepage,omitempty"`
	License     string   `yaml:"license,omitempty" json:"license,omitempty"`
	Tier        string   `yaml:"tier,omitempty" json:"tier,omitempty"` // community|premium|enterprise
	Keywords    []string `yaml:"keywords,omitempty" json:"keywords,omitempty"`
	Categories  []string `yaml:"categories,omitempty" json:"categories,omitempty"`
}

// Spec carries dependencies, engine constraints, and provenance.
type Spec struct {
	Engine    Engine   `yaml:"engine,omitempty" json:"engine,omitempty"`
	Requires  Requires `yaml:"requires,omitempty" json:"requires,omitempty"`
	Provides  Provides `yaml:"provides,omitempty" json:"provides,omitempty"`
	Checksum  string   `yaml:"checksum,omitempty" json:"checksum,omitempty"`
	Signature string   `yaml:"signature,omitempty" json:"signature,omitempty"`
}

// Engine declares what the runtime must support to run this pack.
type Engine struct {
	ScenarioAPIVersions []string `yaml:"scenarioApiVersions,omitempty" json:"scenarioApiVersions,omitempty"`
	MinLabctlVersion    string   `yaml:"minLabctlVersion,omitempty" json:"minLabctlVersion,omitempty"`
}

// Requires are hard dependencies the pack needs present.
type Requires struct {
	Platform []string  `yaml:"platform,omitempty" json:"platform,omitempty"` // platform categories
	Packs    []PackDep `yaml:"packs,omitempty" json:"packs,omitempty"`
}

// PackDep is a dependency on another pack (semver range).
type PackDep struct {
	Name    string `yaml:"name" json:"name"`
	Version string `yaml:"version" json:"version"` // e.g. ">=1.0.0 <2.0.0"
}

// Provides advertises what the pack contributes (informational).
type Provides struct {
	Scenarios []string `yaml:"scenarios,omitempty" json:"scenarios,omitempty"`
	Incidents []string `yaml:"incidents,omitempty" json:"incidents,omitempty"`
}

// Tier returns the pack tier, defaulting to community.
func (m *Manifest) Tier() string {
	if m.Metadata.Tier == "" {
		return TierCommunity
	}
	return m.Metadata.Tier
}

// Parse reads a pack.yaml from bytes.
func Parse(data []byte) (*Manifest, error) {
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", ManifestFile, err)
	}
	return &m, nil
}

// Load reads pack.yaml from a pack directory. Returns (nil, nil) if absent so
// callers can treat manifest-less (legacy) packs as a warning, not an error.
func Load(dir string) (*Manifest, error) {
	path := filepath.Join(dir, ManifestFile)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return Parse(data)
}

// Validate checks the manifest for schema correctness, reporting all problems.
func (m *Manifest) Validate() error {
	var errs []string
	add := func(f string, a ...interface{}) { errs = append(errs, fmt.Sprintf(f, a...)) }

	if m.APIVersion != APIVersion {
		add("apiVersion must be %q (got %q)", APIVersion, m.APIVersion)
	}
	if m.Kind != Kind {
		add("kind must be %q (got %q)", Kind, m.Kind)
	}
	if !nameRe.MatchString(m.Metadata.Name) {
		add("metadata.name %q is invalid (lowercase, optionally publisher/name)", m.Metadata.Name)
	}
	if _, err := parseSemver(m.Metadata.Version); err != nil {
		add("metadata.version %q is not valid semver", m.Metadata.Version)
	}
	if m.Metadata.Tier != "" && !validTiers[m.Metadata.Tier] {
		add("metadata.tier %q is invalid (community|premium|enterprise)", m.Metadata.Tier)
	}
	if m.Metadata.License == "" {
		add("metadata.license is required (e.g. Apache-2.0 or LicenseRef-...)")
	}
	for _, d := range m.Spec.Requires.Packs {
		if d.Name == "" {
			add("requires.packs: a dependency is missing a name")
		}
		if d.Version == "" {
			add("requires.packs[%s]: a version constraint is required", d.Name)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("invalid %s for pack %q:\n  - %s", ManifestFile, m.Metadata.Name, strings.Join(errs, "\n  - "))
	}
	return nil
}

// CheckEngineCompat verifies the running engine can execute the pack: every
// scenario apiVersion the pack declares must be supported, and the engine
// version must be >= minLabctlVersion. A labctlVersion of "" or "dev" skips the
// version check (development builds run everything).
func (m *Manifest) CheckEngineCompat(labctlVersion string, supportedScenarioAPIVersions []string) error {
	supported := map[string]bool{}
	for _, v := range supportedScenarioAPIVersions {
		supported[v] = true
	}
	var errs []string
	for _, v := range m.Spec.Engine.ScenarioAPIVersions {
		if !supported[v] {
			errs = append(errs, fmt.Sprintf("requires scenario apiVersion %q (engine supports: %s)",
				v, strings.Join(supportedScenarioAPIVersions, ", ")))
		}
	}
	if min := m.Spec.Engine.MinLabctlVersion; min != "" && labctlVersion != "" && labctlVersion != "dev" {
		cmp, err := compareSemver(labctlVersion, min)
		if err == nil && cmp < 0 {
			errs = append(errs, fmt.Sprintf("requires labctl >= %s (running %s)", min, labctlVersion))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("pack %q is not compatible with this engine:\n  - %s",
			m.Metadata.Name, strings.Join(errs, "\n  - "))
	}
	return nil
}

// --- minimal semver (major.minor.patch; ignores pre-release/build) ---

func parseSemver(v string) ([3]int, error) {
	var out [3]int
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		v = v[:i]
	}
	parts := strings.Split(v, ".")
	if len(parts) == 0 || len(parts) > 3 || v == "" {
		return out, fmt.Errorf("not semver: %q", v)
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return out, fmt.Errorf("not semver: %q", v)
		}
		out[i] = n
	}
	return out, nil
}

func compareSemver(a, b string) (int, error) {
	pa, err := parseSemver(a)
	if err != nil {
		return 0, err
	}
	pb, err := parseSemver(b)
	if err != nil {
		return 0, err
	}
	for i := 0; i < 3; i++ {
		if pa[i] != pb[i] {
			if pa[i] < pb[i] {
				return -1, nil
			}
			return 1, nil
		}
	}
	return 0, nil
}
