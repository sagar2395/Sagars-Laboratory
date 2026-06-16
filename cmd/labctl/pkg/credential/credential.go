// SPDX-License-Identifier: Apache-2.0

// Package credential defines the OSS engine side of the certification and
// training framework (task 078). It provides schema types and the verifiable
// certificate record format. Certificate CONTENT and exam delivery live in the
// private premium repo; the engine only defines the interfaces and schemas.
//
// This package is part of the public SDK; it must not import internal/.
package credential

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// SchemaVersion is the apiVersion for certificate records.
const SchemaVersion = "credential.flightdeck.dev/v1"

// Grade represents the grade awarded on a certificate.
type Grade string

const (
	GradePass        Grade = "pass"
	GradeMerit       Grade = "merit"
	GradeDistinction Grade = "distinction"
)

// Evidence is one completed lab activity that contributed to the certificate.
type Evidence struct {
	Kind      string `json:"kind"` // challenge | learn | scenario | incident
	Name      string `json:"name"`
	Score     int    `json:"score"`
	Completed string `json:"completed"` // RFC3339
}

// Certificate is a verifiable achievement record. It is self-contained: the
// payload (everything except Signature) is HMAC-SHA256 signed so it can be
// shared and independently verified. The signing key is set by the issuer (a
// maintainer-held secret in the private repo). In the OSS engine the key is
// empty and certificates are issued unsigned.
type Certificate struct {
	Schema      string     `json:"schema"`      // SchemaVersion
	ID          string     `json:"id"`          // UUID v4 or ULID
	Subject     string     `json:"subject"`     // who earned it
	Achievement string     `json:"achievement"` // track name, e.g. "kubernetes-foundations"
	Grade       Grade      `json:"grade"`
	Score       int        `json:"score"` // 0-100
	IssuedAt    time.Time  `json:"issuedAt"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
	Issuer      string     `json:"issuer"` // e.g. "Flightdeck"
	Evidence    []Evidence `json:"evidence,omitempty"`
	Signature   string     `json:"signature,omitempty"` // base64 HMAC-SHA256
}

// Payload returns the JSON bytes that are signed (everything except Signature).
func (c Certificate) Payload() ([]byte, error) {
	type payload Certificate
	p := payload(c)
	p.Signature = ""
	return json.Marshal(p)
}

// Sign sets Signature to the HMAC-SHA256 of Payload() using key.
// If key is empty the Signature is cleared (unsigned, but still valid for OSS use).
func (c *Certificate) Sign(key []byte) error {
	if len(key) == 0 {
		c.Signature = ""
		return nil
	}
	payload, err := c.Payload()
	if err != nil {
		return fmt.Errorf("serialising certificate payload: %w", err)
	}
	mac := hmac.New(sha256.New, key)
	mac.Write(payload)
	c.Signature = base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return nil
}

// Verify checks the certificate's HMAC-SHA256 signature. It returns nil when
// the signature is valid or when both the certificate and key are unsigned
// (Signature=="", key==nil). Returns ErrUnsigned when the certificate has no
// signature but a key is provided, and ErrInvalidSignature when they do not match.
func (c *Certificate) Verify(key []byte) error {
	if c.Signature == "" {
		if len(key) == 0 {
			return nil // unsigned issued + no key: consistent, ok for OSS
		}
		return ErrUnsigned
	}
	if len(key) == 0 {
		return ErrUnsigned // signature present but no key to verify with
	}
	payload, err := c.Payload()
	if err != nil {
		return fmt.Errorf("serialising certificate payload: %w", err)
	}
	mac := hmac.New(sha256.New, key)
	mac.Write(payload)
	want := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(c.Signature), []byte(want)) {
		return ErrInvalidSignature
	}
	return nil
}

// GradeForScore converts a numeric score (0-100) into a Grade.
//
//	>= 90 → Distinction
//	>= 75 → Merit
//	   < 75 → Pass
//
// Note: callers are responsible for deciding whether to issue a certificate;
// this function returns Pass for any score below Merit.
func GradeForScore(score int) Grade {
	switch {
	case score >= 90:
		return GradeDistinction
	case score >= 75:
		return GradeMerit
	default:
		return GradePass
	}
}

// Errors
var (
	ErrUnsigned         = fmt.Errorf("certificate is unsigned")
	ErrInvalidSignature = fmt.Errorf("certificate signature is invalid")
)
