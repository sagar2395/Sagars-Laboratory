// SPDX-License-Identifier: Apache-2.0
package edition

import (
	"testing"
)

func TestCurrent_DefaultCE(t *testing.T) {
	t.Setenv("LABCTL_EDITION", "")
	if got := Current(); got != CE {
		t.Errorf("Current() = %q, want CE (%q)", got, CE)
	}
}

func TestCurrent_UnknownValueDefaultsCE(t *testing.T) {
	t.Setenv("LABCTL_EDITION", "unicorn")
	if got := Current(); got != CE {
		t.Errorf("Current() = %q for unknown value, want CE (%q)", got, CE)
	}
}

func TestCurrent_EditionFromEnv(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		want   Edition
	}{
		{"professional", string(Pro), Pro},
		{"enterprise", string(Enterprise), Enterprise},
		{"community explicit", string(CE), CE},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("LABCTL_EDITION", tc.envVal)
			if got := Current(); got != tc.want {
				t.Errorf("Current() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestEdition_String(t *testing.T) {
	tests := []struct {
		edition Edition
		want    string
	}{
		{CE, "community"},
		{Pro, "professional"},
		{Enterprise, "enterprise"},
	}
	for _, tc := range tests {
		t.Run(string(tc.edition), func(t *testing.T) {
			if got := tc.edition.String(); got != tc.want {
				t.Errorf("String() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestEdition_DisplayName(t *testing.T) {
	tests := []struct {
		edition Edition
		want    string
	}{
		{CE, "Community Edition (CE)"},
		{Pro, "Professional"},
		{Enterprise, "Enterprise"},
	}
	for _, tc := range tests {
		t.Run(string(tc.edition), func(t *testing.T) {
			if got := tc.edition.DisplayName(); got != tc.want {
				t.Errorf("DisplayName() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestEdition_Summary_NonEmpty(t *testing.T) {
	for _, e := range []Edition{CE, Pro, Enterprise} {
		t.Run(string(e), func(t *testing.T) {
			s := e.Summary()
			if s == "" {
				t.Errorf("Summary() returned empty string for edition %q", e)
			}
		})
	}
}

func TestEdition_FeatureLine_CoreFeaturesAlwaysTrue(t *testing.T) {
	coreFeatures := []string{
		"scenario-engine",
		"incident-engine",
		"learning-paths",
		"challenge-mode",
		"results-api",
		"leaderboard",
		"auth-rbac",
		"team-mode",
		"pack-install",
		"oci-registry",
		"community-packs",
		"multi-env",
		"day2-drills",
	}
	for _, edition := range []Edition{CE, Pro, Enterprise} {
		fl := edition.FeatureLine()
		for _, feat := range coreFeatures {
			if !fl[feat] {
				t.Errorf("edition %q: core feature %q should be true but is false", edition, feat)
			}
		}
	}
}

func TestEdition_FeatureLine_PremiumPacks(t *testing.T) {
	tests := []struct {
		edition Edition
		feature string
		want    bool
	}{
		{CE, "premium-packs", false},
		{Pro, "premium-packs", true},
		{Enterprise, "premium-packs", true},
		{CE, "enterprise-packs", false},
		{Pro, "enterprise-packs", false},
		{Enterprise, "enterprise-packs", true},
		{CE, "air-gapped-mirror", false},
		{Pro, "air-gapped-mirror", false},
		{Enterprise, "air-gapped-mirror", true},
		{CE, "compliance-packs", false},
		{Pro, "compliance-packs", false},
		{Enterprise, "compliance-packs", true},
		{CE, "certification-tracks", false},
		{Pro, "certification-tracks", true},
		{Enterprise, "certification-tracks", true},
		{CE, "commercial-support", false},
		{Pro, "commercial-support", true},
		{Enterprise, "commercial-support", true},
	}
	for _, tc := range tests {
		t.Run(string(tc.edition)+"/"+tc.feature, func(t *testing.T) {
			fl := tc.edition.FeatureLine()
			if got := fl[tc.feature]; got != tc.want {
				t.Errorf("edition %q: feature %q = %v, want %v", tc.edition, tc.feature, got, tc.want)
			}
		})
	}
}
