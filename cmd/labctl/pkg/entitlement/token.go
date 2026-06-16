// SPDX-License-Identifier: Apache-2.0
package entitlement

import (
	"context"
	"fmt"
	"time"
)

// TokenEntitlement is an Entitlement that gates premium/enterprise packs behind
// a license token. It uses the injected TokenVerifier seam; the OSS engine never
// ships a concrete verifier. Community (tier=="") packs are always allowed even
// when the token is invalid.
//
// Usage (premium build, not OSS):
//
//	ent := entitlement.NewTokenEntitlement(os.Getenv("LABCTL_TOKEN"), myVerifier{})
//	engine.SetEntitlement(ent)
type TokenEntitlement struct {
	token    string
	verifier TokenVerifier
}

// NewTokenEntitlement creates an Entitlement backed by a token + verifier.
// If token is empty or verifier is nil the entitlement falls back to AllowAll.
func NewTokenEntitlement(token string, verifier TokenVerifier) Entitlement {
	if token == "" || verifier == nil {
		return AllowAll{}
	}
	return &TokenEntitlement{token: token, verifier: verifier}
}

// Authorize checks that the license token entitles the caller to the requested tier.
// Community-tier packs are always allowed. Premium/enterprise packs require a valid
// token with the matching tier claim.
func (t *TokenEntitlement) Authorize(ctx context.Context, req Request) error {
	tier := req.Tier
	if tier == "" || tier == "community" {
		return nil // community packs are always free
	}

	claims, err := t.verifier.Verify(ctx, t.token)
	if err != nil {
		return Denied(req.Name, fmt.Sprintf("token verification failed: %v", err))
	}
	if claims == nil {
		return Denied(req.Name, "token produced no claims")
	}
	if claims.Expires > 0 && time.Now().Unix() > claims.Expires {
		return Denied(req.Name, "license token has expired")
	}
	for _, t := range claims.Tiers {
		if t == tier || t == "enterprise" { // enterprise entitles all tiers
			return nil
		}
	}
	return Denied(req.Name, fmt.Sprintf("license does not include %q tier (entitled: %v)", tier, claims.Tiers))
}
