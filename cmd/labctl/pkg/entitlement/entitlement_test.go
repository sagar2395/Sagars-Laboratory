// SPDX-License-Identifier: Apache-2.0

package entitlement

import (
	"context"
	"errors"
	"testing"
)

func TestAllowAll_AllowsEverything(t *testing.T) {
	e := Default()
	for _, tier := range []string{"", "community", "premium", "enterprise"} {
		if err := e.Authorize(context.Background(), Request{Name: "x/y", Tier: tier}); err != nil {
			t.Errorf("AllowAll denied tier %q: %v", tier, err)
		}
	}
}

func TestDeniedError(t *testing.T) {
	err := Denied("acme/premium-pack", "no license token")
	var de *DeniedError
	if !errors.As(err, &de) {
		t.Fatalf("Denied should produce a *DeniedError, got %T", err)
	}
	if de.Name != "acme/premium-pack" {
		t.Errorf("name = %q", de.Name)
	}
	if got := err.Error(); got == "" || got == "pack \"acme/premium-pack\" is not entitled" {
		// reason should be included
		t.Errorf("error missing reason: %q", got)
	}
}

// denyTier is the documented example of injecting a custom entitlement: a
// premium-style policy that allows community packs but denies anything else
// unless a token was presented. It lives in the test (and docs), never in the
// OSS engine.
type denyTier struct {
	hasToken bool
}

func (d denyTier) Authorize(_ context.Context, req Request) error {
	if req.Tier == "" || req.Tier == "community" {
		return nil
	}
	if d.hasToken {
		return nil
	}
	return Denied(req.Name, "tier "+req.Tier+" requires a license token")
}

func TestCustomEntitlement_GatesPremium(t *testing.T) {
	e := denyTier{hasToken: false}
	if err := e.Authorize(context.Background(), Request{Name: "a/b", Tier: "community"}); err != nil {
		t.Errorf("community should be allowed: %v", err)
	}
	err := e.Authorize(context.Background(), Request{Name: "a/b", Tier: "premium"})
	var de *DeniedError
	if !errors.As(err, &de) {
		t.Fatalf("premium without token should be denied, got %v", err)
	}
	// With a token, premium is allowed.
	if err := (denyTier{hasToken: true}).Authorize(context.Background(), Request{Name: "a/b", Tier: "premium"}); err != nil {
		t.Errorf("premium with token should be allowed: %v", err)
	}
}
