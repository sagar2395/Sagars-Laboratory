// SPDX-License-Identifier: Apache-2.0

// Package entitlement defines the seam that lets premium and hosted builds gate
// access to packs and features WITHOUT forking the engine (task 070). The OSS
// core ships only the interface and an open, allow-everything default, so the
// open engine behaves identically while leaving a clean extension point.
//
// Anti-lock-in rule: no premium or business logic lives here — only the
// interface, the open default, and the license-token verifier *seam*. Concrete
// gating implementations live in a private repo and are injected at
// construction; they are never compiled into the OSS binary.
//
// This package is part of the public SDK (docs/authoring/sdk-stability-policy.md)
// and must not import anything under internal/.
package entitlement

import (
	"context"
	"fmt"
)

// Request describes a pack (or feature) whose use is being authorized.
type Request struct {
	Name string // pack name, e.g. "acme/kafka-drills"
	Tier string // community | premium | enterprise ("" => community)
	Ref  string // source ref (oci://…, git+https://…, local path)
}

// Entitlement authorizes use of a pack. Authorize returns nil when the request
// is allowed, or an error (ideally a *DeniedError) when it is not. It is
// consulted on the pack-install path; the OSS default allows everything.
type Entitlement interface {
	Authorize(ctx context.Context, req Request) error
}

// DeniedError reports that a request is not entitled. Callers can use errors.As
// to distinguish a policy denial from an operational failure.
type DeniedError struct {
	Name   string
	Reason string
}

func (e *DeniedError) Error() string {
	if e.Reason == "" {
		return fmt.Sprintf("pack %q is not entitled", e.Name)
	}
	return fmt.Sprintf("pack %q is not entitled: %s", e.Name, e.Reason)
}

// Denied is a convenience constructor for a *DeniedError.
func Denied(name, reason string) error { return &DeniedError{Name: name, Reason: reason} }

// AllowAll is the OSS default entitlement: every request is permitted. The open
// engine runs every tier identically — gating is the job of an injected premium
// implementation, never the engine.
type AllowAll struct{}

// Authorize always allows.
func (AllowAll) Authorize(context.Context, Request) error { return nil }

// Default returns the open OSS entitlement (allow everything).
func Default() Entitlement { return AllowAll{} }

// Claims is the result of verifying a license token: who it is for and what
// tiers it entitles.
type Claims struct {
	Subject string   // who the token was issued to
	Tiers   []string // entitled tiers (e.g. ["community","premium"])
	Expires int64    // unix seconds; 0 => no expiry
}

// TokenVerifier is the license-token verification seam. A premium/hosted build
// implements it (e.g. verifying a signed JWT) and wraps the result in an
// Entitlement. The OSS core ships no concrete verifier — only this interface —
// so license logic never enters the open binary.
type TokenVerifier interface {
	Verify(ctx context.Context, token string) (*Claims, error)
}
