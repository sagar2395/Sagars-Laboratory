// SPDX-License-Identifier: Apache-2.0

// Package edition describes which product edition is running. Editions differ
// ONLY in entitled content and optional services injected through the public
// seams — the OSS engine binary is identical for all editions (task 076).
//
// The community edition (CE) is the OSS default. Pro and Enterprise editions are
// assembled by combining the OSS engine binary with premium packs and optional
// closed services from the private repo. The engine is never forked.
//
// This package is part of the public SDK; it must not import internal/.
package edition

import "os"

// Edition identifies the product tier.
type Edition string

const (
	// CE is the open-source community edition (default). All core engine
	// features are available; no premium packs are included.
	CE Edition = "community"

	// Pro is the professional edition: CE plus curated premium pack bundles
	// and commercial support SLAs. No source changes vs CE.
	Pro Edition = "professional"

	// Enterprise is the full enterprise edition: Pro plus SSO/RBAC
	// (task 062), team mode (task 063), air-gapped registry mirror,
	// compliance and audit packs, private hosting. No source changes vs CE.
	Enterprise Edition = "enterprise"
)

// Current returns the active edition. It reads the LABCTL_EDITION environment
// variable (set at deployment time); defaults to CE when unset or unknown.
func Current() Edition {
	switch os.Getenv("LABCTL_EDITION") {
	case string(Pro):
		return Pro
	case string(Enterprise):
		return Enterprise
	default:
		return CE
	}
}

// String implements fmt.Stringer.
func (e Edition) String() string { return string(e) }

// DisplayName returns a human-readable label.
func (e Edition) DisplayName() string {
	switch e {
	case Pro:
		return "Professional"
	case Enterprise:
		return "Enterprise"
	default:
		return "Community Edition (CE)"
	}
}

// FeatureLine documents the CE-vs-paid boundary. It returns a map of feature
// name → whether it is available in this edition. The OSS engine always returns
// all features enabled; premium content is gated by the entitlement seam, not
// by missing code paths.
//
// Anti-lock-in invariant: a feature available in CE today MUST remain in CE
// forever. Only additive features (new premium packs / services) can be gated.
func (e Edition) FeatureLine() map[string]bool {
	// All engine-level features are available in every edition. The CE feature
	// line is the complete list of what ships in the OSS engine; Pro/Enterprise
	// add entitled content on top, not different engine behaviour.
	base := map[string]bool{
		"scenario-engine": true,
		"incident-engine": true,
		"learning-paths":  true,
		"challenge-mode":  true,
		"results-api":     true,
		"leaderboard":     true,
		"auth-rbac":       true,
		"team-mode":       true,
		"pack-install":    true,
		"oci-registry":    true,
		"community-packs": true,
		"multi-env":       true,
		"day2-drills":     true,
	}
	// Premium content (not engine features — content lives in private repo)
	premiumContent := map[string]bool{
		"premium-packs":        e == Pro || e == Enterprise,
		"enterprise-packs":     e == Enterprise,
		"air-gapped-mirror":    e == Enterprise,
		"compliance-packs":     e == Enterprise,
		"certification-tracks": e == Pro || e == Enterprise,
		"commercial-support":   e == Pro || e == Enterprise,
	}
	for k, v := range premiumContent {
		base[k] = v
	}
	return base
}

// Summary returns a one-line human description of the edition.
func (e Edition) Summary() string {
	switch e {
	case Pro:
		return "Professional edition: CE + premium pack bundles + commercial support"
	case Enterprise:
		return "Enterprise edition: Pro + SSO/RBAC + air-gapped mirror + compliance packs + private hosting"
	default:
		return "Community Edition (CE): full open-source engine + community packs (Apache-2.0)"
	}
}
