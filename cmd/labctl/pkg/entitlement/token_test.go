// SPDX-License-Identifier: Apache-2.0
package entitlement

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

// stubVerifier is a test-only TokenVerifier.
type stubVerifier struct {
	claims *Claims
	err    error
}

func (s stubVerifier) Verify(_ context.Context, _ string) (*Claims, error) {
	return s.claims, s.err
}

func TestNewTokenEntitlement_FallsBackToAllowAll(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		verifier TokenVerifier
	}{
		{"empty token", "", stubVerifier{claims: &Claims{Tiers: []string{"premium"}}}},
		{"nil verifier", "mytoken", nil},
		{"both empty and nil", "", nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ent := NewTokenEntitlement(tc.token, tc.verifier)
			// AllowAll should pass even premium tier.
			for _, tier := range []string{"", "community", "premium", "enterprise"} {
				if err := ent.Authorize(context.Background(), Request{Name: "x/y", Tier: tier}); err != nil {
					t.Errorf("expected AllowAll for tier %q, got: %v", tier, err)
				}
			}
		})
	}
}

func TestTokenEntitlement_CommunityAlwaysPasses(t *testing.T) {
	// Even with a verifier that errors, community tier must pass.
	verifier := stubVerifier{err: fmt.Errorf("bad token")}
	ent := NewTokenEntitlement("some-token", verifier)

	for _, tier := range []string{"", "community"} {
		t.Run("tier="+tier, func(t *testing.T) {
			if err := ent.Authorize(context.Background(), Request{Name: "a/b", Tier: tier}); err != nil {
				t.Errorf("community tier should always pass, got: %v", err)
			}
		})
	}
}

func TestTokenEntitlement_ValidToken_AllowsMatchingTier(t *testing.T) {
	tests := []struct {
		name          string
		claims        *Claims
		requestedTier string
		wantErr       bool
	}{
		{
			name:          "premium token allows premium pack",
			claims:        &Claims{Subject: "user", Tiers: []string{"community", "premium"}},
			requestedTier: "premium",
			wantErr:       false,
		},
		{
			name:          "enterprise token allows enterprise pack",
			claims:        &Claims{Subject: "user", Tiers: []string{"enterprise"}},
			requestedTier: "enterprise",
			wantErr:       false,
		},
		{
			name:          "enterprise token allows premium pack (superset)",
			claims:        &Claims{Subject: "user", Tiers: []string{"enterprise"}},
			requestedTier: "premium",
			wantErr:       false,
		},
		{
			name:          "premium token denies enterprise pack",
			claims:        &Claims{Subject: "user", Tiers: []string{"community", "premium"}},
			requestedTier: "enterprise",
			wantErr:       true,
		},
		{
			name:          "community-only token denies premium pack",
			claims:        &Claims{Subject: "user", Tiers: []string{"community"}},
			requestedTier: "premium",
			wantErr:       true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			verifier := stubVerifier{claims: tc.claims}
			ent := NewTokenEntitlement("valid-token", verifier)
			err := ent.Authorize(context.Background(), Request{Name: "a/b", Tier: tc.requestedTier})
			if tc.wantErr && err == nil {
				t.Errorf("expected denial, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("expected allow, got: %v", err)
			}
			// When denied, must be a *DeniedError.
			if tc.wantErr && err != nil {
				var de *DeniedError
				if !errors.As(err, &de) {
					t.Errorf("expected *DeniedError, got %T: %v", err, err)
				}
			}
		})
	}
}

func TestTokenEntitlement_VerifierError_Denies(t *testing.T) {
	verifier := stubVerifier{err: fmt.Errorf("network timeout")}
	ent := NewTokenEntitlement("bad-token", verifier)
	err := ent.Authorize(context.Background(), Request{Name: "a/b", Tier: "premium"})
	if err == nil {
		t.Fatal("expected denial on verifier error, got nil")
	}
	var de *DeniedError
	if !errors.As(err, &de) {
		t.Errorf("expected *DeniedError, got %T: %v", err, err)
	}
}

func TestTokenEntitlement_NilClaims_Denies(t *testing.T) {
	verifier := stubVerifier{claims: nil, err: nil}
	ent := NewTokenEntitlement("token", verifier)
	err := ent.Authorize(context.Background(), Request{Name: "a/b", Tier: "premium"})
	if err == nil {
		t.Fatal("expected denial when claims are nil, got nil")
	}
	var de *DeniedError
	if !errors.As(err, &de) {
		t.Errorf("expected *DeniedError, got %T: %v", err, err)
	}
}

func TestTokenEntitlement_ExpiredToken_Denies(t *testing.T) {
	pastExpiry := time.Now().Add(-time.Hour).Unix()
	verifier := stubVerifier{
		claims: &Claims{
			Subject: "user",
			Tiers:   []string{"premium"},
			Expires: pastExpiry,
		},
	}
	ent := NewTokenEntitlement("expired-token", verifier)
	err := ent.Authorize(context.Background(), Request{Name: "a/b", Tier: "premium"})
	if err == nil {
		t.Fatal("expected denial on expired token, got nil")
	}
	var de *DeniedError
	if !errors.As(err, &de) {
		t.Errorf("expected *DeniedError, got %T", err)
	}
}

func TestTokenEntitlement_NoExpiry_Passes(t *testing.T) {
	// Expires == 0 means no expiry.
	verifier := stubVerifier{
		claims: &Claims{
			Subject: "user",
			Tiers:   []string{"premium"},
			Expires: 0,
		},
	}
	ent := NewTokenEntitlement("no-expiry-token", verifier)
	if err := ent.Authorize(context.Background(), Request{Name: "a/b", Tier: "premium"}); err != nil {
		t.Errorf("expected allow with no expiry set, got: %v", err)
	}
}

func TestTokenEntitlement_FutureExpiry_Passes(t *testing.T) {
	futureExpiry := time.Now().Add(24 * time.Hour).Unix()
	verifier := stubVerifier{
		claims: &Claims{
			Subject: "user",
			Tiers:   []string{"premium"},
			Expires: futureExpiry,
		},
	}
	ent := NewTokenEntitlement("valid-token", verifier)
	if err := ent.Authorize(context.Background(), Request{Name: "a/b", Tier: "premium"}); err != nil {
		t.Errorf("expected allow with future expiry, got: %v", err)
	}
}
