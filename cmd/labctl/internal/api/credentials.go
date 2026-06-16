// SPDX-License-Identifier: Apache-2.0
package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"go.flightdeck.dev/labctl/pkg/credential"
)

// --- Credential API endpoints (task 078) ---
//
// These endpoints expose the certification/training framework over REST.
// The OSS engine handles record format, signing, and storage. Certification
// content and exam delivery live in premium packs (private repo).

// handleCredentialList handles GET /api/credentials
func (s *Server) handleCredentialList(w http.ResponseWriter, r *http.Request) {
	store := credential.NewStore(s.credentialStorePath())
	certs, err := store.All()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "store_error", err.Error())
		return
	}
	if certs == nil {
		certs = []*credential.Certificate{}
	}
	respondJSON(w, http.StatusOK, certs)
}

// IssueCredentialRequest is the POST /api/credentials/issue body.
type IssueCredentialRequest struct {
	Subject     string                `json:"subject"`
	Achievement string                `json:"achievement"`
	Score       int                   `json:"score"`
	Issuer      string                `json:"issuer,omitempty"`
	Evidence    []credential.Evidence `json:"evidence,omitempty"`
}

// handleCredentialIssue handles POST /api/credentials/issue
func (s *Server) handleCredentialIssue(w http.ResponseWriter, r *http.Request) {
	var req IssueCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_body", "request body must be JSON")
		return
	}
	if req.Achievement == "" {
		respondError(w, http.StatusBadRequest, "missing_achievement", "achievement is required")
		return
	}
	if req.Score < 0 || req.Score > 100 {
		respondError(w, http.StatusBadRequest, "invalid_score", "score must be 0-100")
		return
	}

	subject := req.Subject
	if subject == "" {
		subject = actingUser(r)
	}

	grade := credential.GradeForScore(req.Score)
	issuer := req.Issuer
	if issuer == "" {
		issuer = "Flightdeck"
	}

	cert, err := credential.Issue(credential.IssueRequest{
		Subject:     subject,
		Achievement: req.Achievement,
		Grade:       grade,
		Score:       req.Score,
		Issuer:      issuer,
		Evidence:    req.Evidence,
		SigningKey:  s.credentialSigningKey(),
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "issue_failed", err.Error())
		return
	}

	store := credential.NewStore(s.credentialStorePath())
	if err := store.Save(cert); err != nil {
		respondError(w, http.StatusInternalServerError, "save_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, cert)
}

// handleCredentialGet handles GET /api/credentials/{id}
func (s *Server) handleCredentialGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing_id", "credential ID is required")
		return
	}

	store := credential.NewStore(s.credentialStorePath())
	cert, err := store.Get(id)
	if err != nil {
		respondError(w, http.StatusNotFound, "not_found", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, cert)
}

// handleCredentialVerify handles POST /api/credentials/{id}/verify
func (s *Server) handleCredentialVerify(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing_id", "credential ID is required")
		return
	}

	store := credential.NewStore(s.credentialStorePath())
	cert, err := store.Get(id)
	if err != nil {
		respondError(w, http.StatusNotFound, "not_found", err.Error())
		return
	}

	key := s.credentialSigningKey()
	verifyErr := cert.Verify(key)
	if verifyErr != nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"valid":  false,
			"reason": verifyErr.Error(),
			"cert":   cert,
		})
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"valid": true,
		"cert":  cert,
	})
}

// --- helpers ---

func (s *Server) credentialStorePath() string {
	return filepath.Join(s.cfg.ProjectRoot, ".labctl", "credentials")
}

func (s *Server) credentialSigningKey() []byte {
	hex := os.Getenv("LABCTL_SIGNING_KEY")
	if hex == "" {
		return nil
	}
	b, err := hexDecodeKey(hex)
	if err != nil {
		return nil
	}
	return b
}

// hexDecodeKey converts a hex string to bytes.
func hexDecodeKey(s string) ([]byte, error) {
	if len(s)%2 != 0 {
		return nil, nil
	}
	b := make([]byte, len(s)/2)
	for i := 0; i < len(s); i += 2 {
		hi := hexNibble(s[i])
		lo := hexNibble(s[i+1])
		if hi == 255 || lo == 255 {
			return nil, nil
		}
		b[i/2] = hi<<4 | lo
	}
	return b, nil
}

func hexNibble(c byte) byte {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10
	}
	return 255
}
