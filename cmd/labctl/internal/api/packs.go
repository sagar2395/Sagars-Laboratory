// SPDX-License-Identifier: Apache-2.0
package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"go.flightdeck.dev/labctl/pkg/pack"
)

// --- Pack catalog API endpoints (task 073 / 075) ---
//
// These endpoints expose pack discovery and management over REST so the
// Marketplace UI can browse, search, install, and remove packs without the
// user needing to open a terminal.
//
// The backend uses the same two-tier resolution as the CLI:
//   1. Hosted catalog API (LABCTL_CATALOG_URL) — augmented metadata, download
//      counts, verified-publisher badges.
//   2. Static registry index (PACK_REGISTRY_INDEX) — offline-capable fallback.

// PackSearchResult is the API shape returned by GET /api/packs.
type PackSearchResult struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Publisher   string   `json:"publisher,omitempty"`
	DisplayName string   `json:"displayName,omitempty"`
	Description string   `json:"description,omitempty"`
	License     string   `json:"license,omitempty"`
	Tier        string   `json:"tier"`
	Verified    bool     `json:"verified"`
	Ref         string   `json:"ref"`
	Digest      string   `json:"digest,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
	Categories  []string `json:"categories,omitempty"`
	Downloads   int64    `json:"downloads,omitempty"`
	Rating      float64  `json:"rating,omitempty"`
	RatingCount int      `json:"ratingCount,omitempty"`
	Installed   bool     `json:"installed"`
}

// handlePackSearch handles GET /api/packs?q=<term>
// Searches the registry (hosted catalog preferred, static index as fallback).
func (s *Server) handlePackSearch(w http.ResponseWriter, r *http.Request) {
	term := r.URL.Query().Get("q")

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	results, err := s.searchPacks(ctx, term)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "search_failed", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, results)
}

// handlePackInfo handles GET /api/packs/{name}
func (s *Server) handlePackInfo(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if name == "" {
		respondError(w, http.StatusBadRequest, "missing_name", "pack name is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	results, err := s.searchPacks(ctx, name)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "lookup_failed", err.Error())
		return
	}
	for _, e := range results {
		if e.Name == name || barePack(e.Name) == name {
			respondJSON(w, http.StatusOK, e)
			return
		}
	}
	respondError(w, http.StatusNotFound, "not_found", "pack not found: "+name)
}

// handlePackInstall handles POST /api/packs/{name}/install
func (s *Server) handlePackInstall(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if name == "" {
		respondError(w, http.StatusBadRequest, "missing_name", "pack name is required")
		return
	}

	// Resolve name to a ref via the registry index.
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	ref, err := s.resolvePackRef(ctx, name)
	if err != nil {
		respondError(w, http.StatusBadRequest, "resolve_failed", err.Error())
		return
	}

	jobID := s.exec.NextActionID()
	label := fmt.Sprintf("Install pack: %s", name)
	s.exec.BroadcastStart(jobID, label)
	go func() {
		_, err := s.scenes.InstallPack(ref, name, false, nil)
		s.exec.BroadcastEnd(jobID, label, err)
	}()
	respondJSON(w, http.StatusAccepted, map[string]string{"jobId": jobID, "status": "accepted"})
}

// handlePackRemove handles DELETE /api/packs/{name}
func (s *Server) handlePackRemove(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if name == "" {
		respondError(w, http.StatusBadRequest, "missing_name", "pack name is required")
		return
	}

	jobID := s.exec.NextActionID()
	label := fmt.Sprintf("Remove pack: %s", name)
	s.exec.BroadcastStart(jobID, label)
	go func() {
		s.exec.BroadcastEnd(jobID, label, s.scenes.UninstallPack(name))
	}()
	respondJSON(w, http.StatusAccepted, map[string]string{"jobId": jobID, "status": "accepted"})
}

// handlePackListInstalled handles GET /api/packs/installed
func (s *Server) handlePackListInstalled(w http.ResponseWriter, r *http.Request) {
	packs := s.scenes.Packs()
	results := make([]PackSearchResult, 0, len(packs))
	for _, p := range packs {
		tier := "community"
		displayName := p.Name
		if p.Manifest != nil {
			if p.Manifest.Metadata.Tier != "" {
				tier = p.Manifest.Metadata.Tier
			}
			if p.Manifest.Metadata.DisplayName != "" {
				displayName = p.Manifest.Metadata.DisplayName
			}
		}
		results = append(results, PackSearchResult{
			Name:        p.Name,
			DisplayName: displayName,
			Tier:        tier,
			Installed:   true,
		})
	}
	respondJSON(w, http.StatusOK, results)
}

// --- helpers ---

// searchPacks queries the hosted catalog (when configured), then falls back to
// the static registry index, and marks which packs are already installed.
func (s *Server) searchPacks(ctx context.Context, term string) ([]PackSearchResult, error) {
	// Try hosted catalog first.
	catalogURL := os.Getenv("LABCTL_CATALOG_URL")
	if catalogURL != "" {
		cat := &pack.CatalogClient{URL: catalogURL}
		if entries, err := cat.Search(ctx, term); err == nil {
			installed := s.installedPackNames()
			results := make([]PackSearchResult, 0, len(entries))
			for _, e := range entries {
				t := e.Tier
				if t == "" {
					t = "community"
				}
				results = append(results, PackSearchResult{
					Name:        e.Name,
					Version:     e.Version,
					Publisher:   e.Publisher,
					DisplayName: e.DisplayName,
					Description: e.Description,
					License:     e.License,
					Tier:        t,
					Verified:    e.Verified,
					Ref:         e.Ref,
					Digest:      e.Digest,
					Keywords:    e.Keywords,
					Categories:  e.Categories,
					Downloads:   e.Downloads,
					Rating:      e.Rating,
					RatingCount: e.RatingCount,
					Installed:   installed[e.Name] || installed[barePack(e.Name)],
				})
			}
			return results, nil
		}
		// Fall through to static index on catalog error.
	}

	// Static registry index.
	indexURL := s.cfg.RegistryIndexURL
	if indexURL == "" {
		return nil, nil
	}
	client := &pack.IndexClient{
		URL:       indexURL,
		VerifyKey: s.cfg.RegistryIndexKey,
		CachePath: s.cfg.ProjectRoot + "/.labctl/cache/registry-index.json",
	}
	idx, _, err := client.Load(ctx)
	if err != nil {
		return nil, err
	}
	entries := idx.Search(term)
	installed := s.installedPackNames()
	results := make([]PackSearchResult, 0, len(entries))
	for _, e := range entries {
		t := e.Tier
		if t == "" {
			t = "community"
		}
		results = append(results, PackSearchResult{
			Name:        e.Name,
			Version:     e.Version,
			Publisher:   e.Publisher,
			DisplayName: e.DisplayName,
			Description: e.Description,
			License:     e.License,
			Tier:        t,
			Verified:    e.Verified,
			Ref:         e.Ref,
			Digest:      e.Digest,
			Keywords:    e.Keywords,
			Categories:  e.Categories,
			Installed:   installed[e.Name] || installed[barePack(e.Name)],
		})
	}
	return results, nil
}

func (s *Server) resolvePackRef(ctx context.Context, name string) (string, error) {
	indexURL := s.cfg.RegistryIndexURL
	if indexURL == "" {
		return "", nil
	}
	client := &pack.IndexClient{
		URL:       indexURL,
		VerifyKey: s.cfg.RegistryIndexKey,
		CachePath: s.cfg.ProjectRoot + "/.labctl/cache/registry-index.json",
	}
	idx, _, err := client.Load(ctx)
	if err != nil {
		return name, nil // let the engine try the raw name
	}
	if entry, ok := idx.Find(name); ok {
		return entry.ResolveRef(), nil
	}
	return name, nil
}

func (s *Server) installedPackNames() map[string]bool {
	m := make(map[string]bool)
	for _, p := range s.scenes.Packs() {
		m[p.Name] = true
		m[barePack(p.Name)] = true
	}
	return m
}

func barePack(name string) string {
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '/' {
			return name[i+1:]
		}
	}
	return name
}
