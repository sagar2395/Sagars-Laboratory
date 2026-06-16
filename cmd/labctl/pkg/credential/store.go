// SPDX-License-Identifier: Apache-2.0
package credential

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

// IssueRequest holds the inputs for Issue().
type IssueRequest struct {
	Subject     string
	Achievement string
	Grade       Grade
	Score       int
	Issuer      string
	Evidence    []Evidence
	ExpiresAt   *time.Time
	SigningKey  []byte // nil => unsigned
}

// Issue creates, optionally signs, and returns a new Certificate.
func Issue(req IssueRequest) (*Certificate, error) {
	issuer := req.Issuer
	if issuer == "" {
		issuer = "Flightdeck"
	}
	cert := &Certificate{
		Schema:      SchemaVersion,
		ID:          newID(),
		Subject:     req.Subject,
		Achievement: req.Achievement,
		Grade:       req.Grade,
		Score:       req.Score,
		IssuedAt:    time.Now().UTC().Truncate(time.Second),
		ExpiresAt:   req.ExpiresAt,
		Issuer:      issuer,
		Evidence:    req.Evidence,
	}
	if err := cert.Sign(req.SigningKey); err != nil {
		return nil, err
	}
	return cert, nil
}

// Store is a credential store: one JSON file per certificate in a directory.
type Store struct {
	dir string
}

// NewStore creates a Store backed by the given directory.
func NewStore(dir string) *Store {
	return &Store{dir: dir}
}

// Save persists a certificate as <id>.json inside the store directory.
func (s *Store) Save(cert *Certificate) error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cert, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling certificate: %w", err)
	}
	path := filepath.Join(s.dir, cert.ID+".json")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// All returns all certificates in the store, sorted by IssuedAt ascending.
func (s *Store) All() ([]*Certificate, error) {
	entries, err := os.ReadDir(s.dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading credential store: %w", err)
	}
	var certs []*Certificate
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.dir, e.Name()))
		if err != nil {
			continue
		}
		var c Certificate
		if err := json.Unmarshal(data, &c); err != nil {
			continue
		}
		certs = append(certs, &c)
	}
	return certs, nil
}

// Get retrieves a single certificate by ID (prefix match supported).
func (s *Store) Get(id string) (*Certificate, error) {
	// Try exact match first.
	path := filepath.Join(s.dir, id+".json")
	if data, err := os.ReadFile(path); err == nil {
		var c Certificate
		if err := json.Unmarshal(data, &c); err != nil {
			return nil, err
		}
		return &c, nil
	}
	// Prefix match.
	entries, err := os.ReadDir(s.dir)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("credential %q not found", id)
	}
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		name := e.Name()[:len(e.Name())-5] // strip .json
		if len(name) >= len(id) && name[:len(id)] == id {
			data, err := os.ReadFile(filepath.Join(s.dir, e.Name()))
			if err != nil {
				return nil, err
			}
			var c Certificate
			if err := json.Unmarshal(data, &c); err != nil {
				return nil, err
			}
			return &c, nil
		}
	}
	return nil, fmt.Errorf("credential %q not found", id)
}

// newID generates a simple random hex ID (16 bytes = 32 hex chars).
func newID() string {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	b := make([]byte, 16)
	for i := range b {
		b[i] = byte(r.Intn(256))
	}
	return fmt.Sprintf("%x", b)
}
