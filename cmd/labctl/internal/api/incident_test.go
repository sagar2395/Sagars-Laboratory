package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sagars-lab/labctl/internal/config"
	"github.com/sagars-lab/labctl/internal/executor"
	"github.com/sagars-lab/labctl/internal/incident"
)

const apiTestFaultYAML = `name: api-fault
displayName: "API Test Fault"
description: "test"
category: workload
severity: low
target:
  namespace: go-api
  workload: go-api
prerequisites:
  apps: [go-api]
detection:
  name: marker-gone
  type: script
  script: checks/resolved.sh
  timeoutSeconds: 5
`

func newIncidentServer(t *testing.T) (*Server, string) {
	t.Helper()
	root := t.TempDir()

	appDir := filepath.Join(root, "apps", "go-api")
	os.MkdirAll(appDir, 0755)
	os.WriteFile(filepath.Join(appDir, "app.env"), []byte("APP_NAME=go-api\n"), 0644)

	dir := filepath.Join(root, "incidents", "api-fault")
	os.MkdirAll(filepath.Join(dir, "checks"), 0755)
	files := map[string]string{
		"fault.yaml":         apiTestFaultYAML,
		"inject.sh":          "#!/usr/bin/env bash\ntouch \"$(dirname \"$0\")/BROKEN\"\n",
		"resolve.sh":         "#!/usr/bin/env bash\nrm -f \"$(dirname \"$0\")/BROKEN\"\n",
		"checks/resolved.sh": "#!/usr/bin/env bash\n[ ! -f \"$(dirname \"$0\")/../BROKEN\" ]\n",
		"hints.md":           "## Hint 1\nx\n## Hint 2\ny\n## Hint 3\nz\n",
		"solution.md":        "# Solution\n",
	}
	for f, content := range files {
		os.WriteFile(filepath.Join(dir, f), []byte(content), 0755)
	}

	return &Server{
		cfg:       &config.Config{ProjectRoot: root, DomainSuffix: "k3d.local"},
		exec:      executor.New(root),
		incidents: incident.NewEngine(root, "k3d.local"),
	}, root
}

func TestHandleIncidentInject_FullLoop(t *testing.T) {
	s, root := newIncidentServer(t)

	// Inject.
	req := httptest.NewRequest(http.MethodPost, "/api/incidents/api-fault/inject", nil)
	req = setVars(req, map[string]string{"name": "api-fault"})
	w := httptest.NewRecorder()
	s.handleIncidentInject(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("inject: got %d (%s)", w.Code, w.Body.String())
	}

	// Second inject without force → 409.
	w = httptest.NewRecorder()
	s.handleIncidentInject(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("second inject: got %d, want 409", w.Code)
	}

	// Status: not resolved.
	w = httptest.NewRecorder()
	s.handleIncidentStatus(w, httptest.NewRequest(http.MethodGet, "/api/incidents/status", nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"resolved":false`) {
		t.Fatalf("status: %d %s", w.Code, w.Body.String())
	}

	// Fix manually, status flips to resolved.
	os.Remove(filepath.Join(root, "incidents", "api-fault", "BROKEN"))
	w = httptest.NewRecorder()
	s.handleIncidentStatus(w, httptest.NewRequest(http.MethodGet, "/api/incidents/status", nil))
	if !strings.Contains(w.Body.String(), `"resolved":true`) {
		t.Fatalf("expected resolved, got %s", w.Body.String())
	}

	// No active incident anymore.
	w = httptest.NewRecorder()
	s.handleIncidentStatus(w, httptest.NewRequest(http.MethodGet, "/api/incidents/status", nil))
	if !strings.Contains(w.Body.String(), `"active":null`) {
		t.Fatalf("expected no active incident, got %s", w.Body.String())
	}
}

func TestHandleIncidentInject_Errors(t *testing.T) {
	s, _ := newIncidentServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/incidents/x/inject", nil)
	req = setVars(req, map[string]string{"name": "../evil"})
	w := httptest.NewRecorder()
	s.handleIncidentInject(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid name: got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/incidents/ghost/inject", nil)
	req = setVars(req, map[string]string{"name": "ghost"})
	w = httptest.NewRecorder()
	s.handleIncidentInject(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("unknown fault: got %d", w.Code)
	}
}

func TestHandleIncidentInject_SilentHidesFault(t *testing.T) {
	s, _ := newIncidentServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/incidents/api-fault/inject?silent=true", nil)
	req = setVars(req, map[string]string{"name": "api-fault"})
	w := httptest.NewRecorder()
	s.handleIncidentInject(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("inject: %d", w.Code)
	}
	if strings.Contains(w.Body.String(), "api-fault") {
		t.Fatalf("silent inject must not reveal the fault: %s", w.Body.String())
	}

	// Status in silent mode hides the identity while unresolved.
	w = httptest.NewRecorder()
	s.handleIncidentStatus(w, httptest.NewRequest(http.MethodGet, "/api/incidents/status", nil))
	if strings.Contains(w.Body.String(), "api-fault") {
		t.Fatalf("silent status must not reveal the fault: %s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "(hidden)") {
		t.Fatalf("expected hidden marker: %s", w.Body.String())
	}
}

func TestHandleIncidentResolve(t *testing.T) {
	s, _ := newIncidentServer(t)

	// Resolve with nothing active → 409.
	w := httptest.NewRecorder()
	s.handleIncidentResolve(w, httptest.NewRequest(http.MethodPost, "/api/incidents/resolve", nil))
	if w.Code != http.StatusConflict {
		t.Fatalf("resolve with nothing active: got %d", w.Code)
	}

	// Inject then resolve.
	req := httptest.NewRequest(http.MethodPost, "/api/incidents/api-fault/inject", nil)
	req = setVars(req, map[string]string{"name": "api-fault"})
	s.handleIncidentInject(httptest.NewRecorder(), req)

	w = httptest.NewRecorder()
	s.handleIncidentResolve(w, httptest.NewRequest(http.MethodPost, "/api/incidents/resolve", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("resolve: got %d (%s)", w.Code, w.Body.String())
	}
}

func TestHandleListIncidents(t *testing.T) {
	s, _ := newIncidentServer(t)
	w := httptest.NewRecorder()
	s.handleListIncidents(w, httptest.NewRequest(http.MethodGet, "/api/incidents", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("list: %d", w.Code)
	}
	var resp struct {
		Faults []json.RawMessage `json:"faults"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil || len(resp.Faults) != 1 {
		t.Fatalf("list response: %v %s", err, w.Body.String())
	}
}
