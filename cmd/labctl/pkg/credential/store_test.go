// SPDX-License-Identifier: Apache-2.0
package credential

import (
	"testing"
	"time"
)

func TestIssue(t *testing.T) {
	tests := []struct {
		name    string
		req     IssueRequest
		wantErr bool
		check   func(t *testing.T, c *Certificate)
	}{
		{
			name: "unsigned",
			req:  IssueRequest{Subject: "alice", Achievement: "k8s", Grade: GradePass, Score: 70, Issuer: "Flightdeck"},
			check: func(t *testing.T, c *Certificate) {
				if c.Signature != "" {
					t.Errorf("expected no signature, got %q", c.Signature)
				}
				if c.ID == "" {
					t.Error("expected non-empty ID")
				}
				if c.Schema != SchemaVersion {
					t.Errorf("got schema %q, want %q", c.Schema, SchemaVersion)
				}
				if c.Issuer != "Flightdeck" {
					t.Errorf("got issuer %q, want Flightdeck", c.Issuer)
				}
			},
		},
		{
			name: "signed",
			req:  IssueRequest{Subject: "bob", Achievement: "sre", Grade: GradeMerit, Score: 80, SigningKey: []byte("secret-key-here!")},
			check: func(t *testing.T, c *Certificate) {
				if c.Signature == "" {
					t.Error("expected signature when key provided")
				}
				// Verify it round-trips
				if err := c.Verify([]byte("secret-key-here!")); err != nil {
					t.Errorf("verify failed: %v", err)
				}
			},
		},
		{
			name: "default_issuer",
			req:  IssueRequest{Subject: "carol", Achievement: "net", Grade: GradePass, Score: 65},
			check: func(t *testing.T, c *Certificate) {
				if c.Issuer != "Flightdeck" {
					t.Errorf("expected default issuer Flightdeck, got %q", c.Issuer)
				}
			},
		},
		{
			name: "with_expires",
			req: IssueRequest{
				Subject: "dave", Achievement: "cert", Grade: GradeDistinction, Score: 95,
				ExpiresAt: func() *time.Time { t := time.Now().Add(365 * 24 * time.Hour); return &t }(),
			},
			check: func(t *testing.T, c *Certificate) {
				if c.ExpiresAt == nil {
					t.Error("expected ExpiresAt to be set")
				}
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, err := Issue(tc.req)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Issue() error = %v, wantErr %v", err, tc.wantErr)
			}
			if err == nil && tc.check != nil {
				tc.check(t, c)
			}
		})
	}
}

func TestStoreSaveAll(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	// Empty store
	certs, err := store.All()
	if err != nil {
		t.Fatalf("All() on empty store: %v", err)
	}
	if len(certs) != 0 {
		t.Errorf("expected 0 certs, got %d", len(certs))
	}

	// Save two certs
	c1, _ := Issue(IssueRequest{Subject: "alice", Achievement: "k8s", Grade: GradePass, Score: 70})
	c2, _ := Issue(IssueRequest{Subject: "bob", Achievement: "sre", Grade: GradeMerit, Score: 80})

	if err := store.Save(c1); err != nil {
		t.Fatalf("Save c1: %v", err)
	}
	if err := store.Save(c2); err != nil {
		t.Fatalf("Save c2: %v", err)
	}

	certs, err = store.All()
	if err != nil {
		t.Fatalf("All(): %v", err)
	}
	if len(certs) != 2 {
		t.Errorf("expected 2 certs, got %d", len(certs))
	}
}

func TestStoreGet(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	c, _ := Issue(IssueRequest{Subject: "alice", Achievement: "k8s", Grade: GradePass, Score: 70})
	if err := store.Save(c); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Exact ID
	got, err := store.Get(c.ID)
	if err != nil {
		t.Fatalf("Get exact: %v", err)
	}
	if got.ID != c.ID {
		t.Errorf("got ID %q, want %q", got.ID, c.ID)
	}

	// Prefix match (first 8 chars)
	prefix := c.ID[:8]
	got2, err := store.Get(prefix)
	if err != nil {
		t.Fatalf("Get prefix: %v", err)
	}
	if got2.ID != c.ID {
		t.Errorf("prefix match: got ID %q, want %q", got2.ID, c.ID)
	}

	// Not found
	_, err = store.Get("doesnotexist")
	if err == nil {
		t.Error("expected error for missing ID")
	}
}
