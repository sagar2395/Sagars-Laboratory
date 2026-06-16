// SPDX-License-Identifier: Apache-2.0
package credential

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

// baseCert returns a minimal valid Certificate for use in tests.
func baseCert() Certificate {
	return Certificate{
		Schema:      SchemaVersion,
		ID:          "01HVTEST000000000000000001",
		Subject:     "alice@example.com",
		Achievement: "kubernetes-foundations",
		Grade:       GradePass,
		Score:       70,
		IssuedAt:    time.Date(2026, 6, 16, 0, 0, 0, 0, time.UTC),
		Issuer:      "Flightdeck",
	}
}

func TestPayload_ExcludesSignature(t *testing.T) {
	cert := baseCert()
	cert.Signature = "some-sig"

	payload, err := cert.Payload()
	if err != nil {
		t.Fatalf("Payload(): %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(payload, &m); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if _, ok := m["signature"]; ok {
		t.Error("Payload() should not include 'signature' field")
	}
	if _, ok := m["schema"]; !ok {
		t.Error("Payload() should include 'schema' field")
	}
	if _, ok := m["subject"]; !ok {
		t.Error("Payload() should include 'subject' field")
	}
}

func TestSign_EmptyKey_ClearsSignature(t *testing.T) {
	cert := baseCert()
	cert.Signature = "pre-existing-sig"

	if err := cert.Sign(nil); err != nil {
		t.Fatalf("Sign(nil): %v", err)
	}
	if cert.Signature != "" {
		t.Errorf("Sign(nil) should clear Signature, got %q", cert.Signature)
	}
}

func TestSign_EmptySlice_ClearsSignature(t *testing.T) {
	cert := baseCert()
	if err := cert.Sign([]byte{}); err != nil {
		t.Fatalf("Sign([]byte{}): %v", err)
	}
	if cert.Signature != "" {
		t.Errorf("Sign([]byte{}) should clear Signature, got %q", cert.Signature)
	}
}

func TestSign_WithKey_SetsNonEmptySignature(t *testing.T) {
	cert := baseCert()
	key := []byte("supersecretkey")

	if err := cert.Sign(key); err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if cert.Signature == "" {
		t.Error("Sign with key should set a non-empty Signature")
	}
}

func TestSign_DifferentKeysDifferentSignatures(t *testing.T) {
	cert1 := baseCert()
	cert2 := baseCert()

	if err := cert1.Sign([]byte("key-one")); err != nil {
		t.Fatalf("Sign cert1: %v", err)
	}
	if err := cert2.Sign([]byte("key-two")); err != nil {
		t.Fatalf("Sign cert2: %v", err)
	}
	if cert1.Signature == cert2.Signature {
		t.Error("different keys should produce different signatures")
	}
}

func TestVerify_Cases(t *testing.T) {
	key := []byte("testkey")

	// Build a validly signed cert.
	signed := baseCert()
	if err := signed.Sign(key); err != nil {
		t.Fatalf("setup Sign: %v", err)
	}

	// Build an unsigned cert (no key).
	unsigned := baseCert()

	tests := []struct {
		name    string
		cert    Certificate
		key     []byte
		wantErr error
	}{
		{
			name:    "unsigned + no key → nil (OSS consistent)",
			cert:    unsigned,
			key:     nil,
			wantErr: nil,
		},
		{
			name:    "unsigned + key → ErrUnsigned",
			cert:    unsigned,
			key:     key,
			wantErr: ErrUnsigned,
		},
		{
			name:    "signed + no key → ErrUnsigned",
			cert:    signed,
			key:     nil,
			wantErr: ErrUnsigned,
		},
		{
			name:    "signed + wrong key → ErrInvalidSignature",
			cert:    signed,
			key:     []byte("wrongkey"),
			wantErr: ErrInvalidSignature,
		},
		{
			name:    "signed + correct key → nil",
			cert:    signed,
			key:     key,
			wantErr: nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cert := tc.cert // copy
			err := cert.Verify(tc.key)
			if tc.wantErr == nil {
				if err != nil {
					t.Errorf("Verify() = %v, want nil", err)
				}
			} else {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("Verify() = %v, want %v", err, tc.wantErr)
				}
			}
		})
	}
}

func TestVerify_TamperedCert_Fails(t *testing.T) {
	key := []byte("signingkey")
	cert := baseCert()
	if err := cert.Sign(key); err != nil {
		t.Fatalf("Sign: %v", err)
	}
	// Tamper with the score after signing.
	cert.Score = 99

	err := cert.Verify(key)
	if !errors.Is(err, ErrInvalidSignature) {
		t.Errorf("tampered cert should fail verification, got: %v", err)
	}
}

func TestGradeForScore(t *testing.T) {
	tests := []struct {
		score int
		want  Grade
	}{
		// Distinction boundary (>= 90)
		{100, GradeDistinction},
		{90, GradeDistinction},
		// Merit boundary (>= 75, < 90)
		{89, GradeMerit},
		{75, GradeMerit},
		// Pass (< 75)
		{74, GradePass},
		{60, GradePass},
		{1, GradePass},
		{0, GradePass},
		// Negative (still returns Pass; callers decide to issue or not)
		{-1, GradePass},
	}
	for _, tc := range tests {
		t.Run("", func(t *testing.T) {
			if got := GradeForScore(tc.score); got != tc.want {
				t.Errorf("GradeForScore(%d) = %q, want %q", tc.score, got, tc.want)
			}
		})
	}
}

func TestCertificate_Evidence(t *testing.T) {
	cert := baseCert()
	cert.Evidence = []Evidence{
		{Kind: "challenge", Name: "pod-crash-loop", Score: 95, Completed: "2026-06-10T12:00:00Z"},
		{Kind: "scenario", Name: "autoscaling-under-load", Score: 80, Completed: "2026-06-11T09:00:00Z"},
	}
	key := []byte("evidencekey")
	if err := cert.Sign(key); err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if err := cert.Verify(key); err != nil {
		t.Fatalf("Verify after signing with evidence: %v", err)
	}
	// Tamper with evidence.
	cert.Evidence[0].Score = 1
	if err := cert.Verify(key); !errors.Is(err, ErrInvalidSignature) {
		t.Errorf("tampered evidence should fail, got: %v", err)
	}
}

func TestCertificate_ExpiresAt(t *testing.T) {
	cert := baseCert()
	exp := time.Date(2027, 6, 16, 0, 0, 0, 0, time.UTC)
	cert.ExpiresAt = &exp

	key := []byte("expirykey")
	if err := cert.Sign(key); err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if err := cert.Verify(key); err != nil {
		t.Fatalf("Verify: %v", err)
	}

	// Tamper with expiry.
	tampered := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)
	cert.ExpiresAt = &tampered
	if err := cert.Verify(key); !errors.Is(err, ErrInvalidSignature) {
		t.Errorf("tampered ExpiresAt should fail verification, got: %v", err)
	}
}
