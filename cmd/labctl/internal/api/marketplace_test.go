// SPDX-License-Identifier: Apache-2.0
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.flightdeck.dev/labctl/internal/config"
)

func TestHandleEditionInfo(t *testing.T) {
	s := &Server{cfg: &config.Config{}}
	req := httptest.NewRequest(http.MethodGet, "/api/edition", nil)
	w := httptest.NewRecorder()
	s.handleEditionInfo(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("handleEditionInfo: got %d, want 200", w.Code)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := resp["edition"]; !ok {
		t.Error("response missing 'edition' field")
	}
	if _, ok := resp["features"]; !ok {
		t.Error("response missing 'features' field")
	}
}

func TestHandlePackSearch_NoRegistry(t *testing.T) {
	s := &Server{cfg: &config.Config{RegistryIndexURL: ""}}
	req := httptest.NewRequest(http.MethodGet, "/api/packs", nil)
	w := httptest.NewRecorder()
	s.handlePackSearch(w, req)

	// No index configured → should return 200 with empty array (not error).
	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", w.Code)
	}
}

func TestHandlePackInfo_MissingName(t *testing.T) {
	s := &Server{cfg: &config.Config{}}
	req := httptest.NewRequest(http.MethodGet, "/api/packs/", nil)
	req = setVars(req, map[string]string{"name": ""})
	w := httptest.NewRecorder()
	s.handlePackInfo(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", w.Code)
	}
}

func TestHandlePackInstall_MissingName(t *testing.T) {
	s := &Server{cfg: &config.Config{}}
	req := httptest.NewRequest(http.MethodPost, "/api/packs//install", nil)
	req = setVars(req, map[string]string{"name": ""})
	w := httptest.NewRecorder()
	s.handlePackInstall(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", w.Code)
	}
}

func TestHandlePackRemove_MissingName(t *testing.T) {
	s := &Server{cfg: &config.Config{}}
	req := httptest.NewRequest(http.MethodDelete, "/api/packs/", nil)
	req = setVars(req, map[string]string{"name": ""})
	w := httptest.NewRecorder()
	s.handlePackRemove(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", w.Code)
	}
}

func TestHandleCredentialList_Empty(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: &config.Config{ProjectRoot: dir}}
	req := httptest.NewRequest(http.MethodGet, "/api/credentials", nil)
	w := httptest.NewRecorder()
	s.handleCredentialList(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", w.Code)
	}
	var certs []interface{}
	if err := json.NewDecoder(w.Body).Decode(&certs); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(certs) != 0 {
		t.Errorf("expected empty array, got %d", len(certs))
	}
}

func TestHandleCredentialIssue_MissingAchievement(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: &config.Config{ProjectRoot: dir}}
	body := `{"subject":"alice","score":80}`
	req := httptest.NewRequest(http.MethodPost, "/api/credentials/issue", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleCredentialIssue(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", w.Code)
	}
}

func TestHandleCredentialIssue_InvalidScore(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: &config.Config{ProjectRoot: dir}}
	body := `{"achievement":"k8s","score":150}`
	req := httptest.NewRequest(http.MethodPost, "/api/credentials/issue", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleCredentialIssue(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", w.Code)
	}
}

func TestHandleCredentialIssueAndGet(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: &config.Config{ProjectRoot: dir}}

	// Issue a credential
	body := `{"achievement":"kubernetes-foundations","subject":"alice","score":88}`
	req := httptest.NewRequest(http.MethodPost, "/api/credentials/issue", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handleCredentialIssue(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("issue: got %d, want 201 — body: %s", w.Code, w.Body.String())
	}

	var cert map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&cert); err != nil {
		t.Fatalf("decode cert: %v", err)
	}
	id, _ := cert["id"].(string)
	if id == "" {
		t.Fatal("expected non-empty id in issued credential")
	}

	// Get it by ID
	req2 := httptest.NewRequest(http.MethodGet, "/api/credentials/"+id, nil)
	req2 = setVars(req2, map[string]string{"id": id})
	w2 := httptest.NewRecorder()
	s.handleCredentialGet(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("get: got %d, want 200", w2.Code)
	}

	// Verify it (unsigned — no signing key)
	req3 := httptest.NewRequest(http.MethodPost, "/api/credentials/"+id+"/verify", nil)
	req3 = setVars(req3, map[string]string{"id": id})
	w3 := httptest.NewRecorder()
	s.handleCredentialVerify(w3, req3)
	if w3.Code != http.StatusOK {
		t.Fatalf("verify: got %d, want 200", w3.Code)
	}
	var result map[string]interface{}
	if err := json.NewDecoder(w3.Body).Decode(&result); err != nil {
		t.Fatalf("decode verify: %v", err)
	}
	if result["valid"] != true {
		t.Errorf("expected valid=true for unsigned cert with no key, got %v", result["valid"])
	}
}

func TestHandleCredentialGet_NotFound(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: &config.Config{ProjectRoot: dir}}
	req := httptest.NewRequest(http.MethodGet, "/api/credentials/doesnotexist", nil)
	req = setVars(req, map[string]string{"id": "doesnotexist"})
	w := httptest.NewRecorder()
	s.handleCredentialGet(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("got %d, want 404", w.Code)
	}
}
