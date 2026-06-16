// SPDX-License-Identifier: Apache-2.0
package pack

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// sampleCatalogEntries is the JSON array the mock server returns for list endpoints.
var sampleCatalogEntries = []CatalogEntry{
	{
		IndexEntry: IndexEntry{
			Name:        "flightdeck/hello-pack",
			Version:     "0.1.0",
			Ref:         "oci://ghcr.io/snowops/hello-pack:0.1.0",
			Publisher:   "flightdeck",
			DisplayName: "Hello Pack",
			Description: "minimal example",
			License:     "Apache-2.0",
			Tier:        "community",
			Keywords:    []string{"example"},
			Verified:    true,
		},
		Downloads:   1234,
		Rating:      4.5,
		RatingCount: 42,
	},
	{
		IndexEntry: IndexEntry{
			Name:     "acme/kafka-drills",
			Version:  "1.5.0",
			Ref:      "oci://ghcr.io/acme/kafka-drills:1.5.0",
			Tier:     "community",
			License:  "Apache-2.0",
			Keywords: []string{"kafka", "incident"},
		},
		Downloads:   500,
		Rating:      3.8,
		RatingCount: 10,
	},
}

func mustMarshal(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return b
}

// newCatalogServer starts a mock catalog API server.
// searchHandler handles GET /v1/packs (with optional ?q=)
// findHandler handles GET /v1/packs/<name>
func newCatalogServer(t *testing.T, searchEntries []CatalogEntry, findEntry *CatalogEntry, findStatus int) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/packs/", func(w http.ResponseWriter, r *http.Request) {
		// single-item lookup
		if findStatus != 0 {
			w.WriteHeader(findStatus)
			return
		}
		if findEntry == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(mustMarshal(t, findEntry)) //nolint:errcheck
	})
	mux.HandleFunc("/v1/packs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(mustMarshal(t, searchEntries)) //nolint:errcheck
	})
	return httptest.NewServer(mux)
}

func TestCatalogClient_Enabled(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		enabled bool
	}{
		{"empty URL disabled", "", false},
		{"non-empty URL enabled", "https://catalog.flightdeck.dev", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &CatalogClient{URL: tc.url}
			if got := c.Enabled(); got != tc.enabled {
				t.Errorf("Enabled() = %v, want %v", got, tc.enabled)
			}
		})
	}
}

func TestCatalogClient_Search_HappyPath(t *testing.T) {
	srv := newCatalogServer(t, sampleCatalogEntries, nil, 0)
	defer srv.Close()

	c := &CatalogClient{URL: srv.URL}
	entries, err := c.Search(context.Background(), "kafka")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Downloads != 1234 {
		t.Errorf("expected Downloads=1234, got %d", entries[0].Downloads)
	}
}

func TestCatalogClient_Search_EmptyTerm(t *testing.T) {
	srv := newCatalogServer(t, sampleCatalogEntries, nil, 0)
	defer srv.Close()

	c := &CatalogClient{URL: srv.URL}
	entries, err := c.Search(context.Background(), "")
	if err != nil {
		t.Fatalf("Search (empty): %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries from empty search, got %d", len(entries))
	}
}

func TestCatalogClient_Search_CacheHit(t *testing.T) {
	srv := newCatalogServer(t, sampleCatalogEntries, nil, 0)
	defer srv.Close()

	var callCount int
	// Intercept requests to count them.
	proxy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		// Proxy to the real server.
		resp, err := http.Get(srv.URL + r.URL.RequestURI())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		buf := make([]byte, 1<<16)
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				w.Write(buf[:n]) //nolint:errcheck
			}
			if err != nil {
				break
			}
		}
	})
	proxySrv := httptest.NewServer(proxy)
	defer proxySrv.Close()

	now := time.Now()
	c := &CatalogClient{
		URL: proxySrv.URL,
		TTL: time.Hour,
		now: func() time.Time { return now },
	}

	// First call should hit the server.
	if _, err := c.Search(context.Background(), "kafka"); err != nil {
		t.Fatalf("first Search: %v", err)
	}
	if callCount != 1 {
		t.Fatalf("expected 1 server call, got %d", callCount)
	}

	// Second call within TTL should be served from cache.
	if _, err := c.Search(context.Background(), "kafka"); err != nil {
		t.Fatalf("second Search: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected cache hit (still 1 call), got %d calls", callCount)
	}
}

func TestCatalogClient_Search_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := &CatalogClient{URL: srv.URL}
	_, err := c.Search(context.Background(), "kafka")
	if err == nil {
		t.Fatal("expected error on server 500, got nil")
	}
}

func TestCatalogClient_Search_Disabled(t *testing.T) {
	c := &CatalogClient{URL: ""}
	_, err := c.Search(context.Background(), "kafka")
	if err == nil {
		t.Fatal("expected error when URL is empty")
	}
}

func TestCatalogClient_Find_HappyPath(t *testing.T) {
	entry := &sampleCatalogEntries[0]
	srv := newCatalogServer(t, nil, entry, 0)
	defer srv.Close()

	c := &CatalogClient{URL: srv.URL}
	got, err := c.Find(context.Background(), "flightdeck/hello-pack")
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil entry")
	}
	if got.Name != "flightdeck/hello-pack" {
		t.Errorf("Name = %q, want flightdeck/hello-pack", got.Name)
	}
	if got.Downloads != 1234 {
		t.Errorf("Downloads = %d, want 1234", got.Downloads)
	}
}

func TestCatalogClient_Find_NotFound(t *testing.T) {
	srv := newCatalogServer(t, nil, nil, 0)
	defer srv.Close()

	c := &CatalogClient{URL: srv.URL}
	got, err := c.Find(context.Background(), "nonexistent/pack")
	if err != nil {
		t.Fatalf("Find returned error for 404, want nil,nil: %v", err)
	}
	if got != nil {
		t.Errorf("Find returned entry for 404, want nil: %+v", got)
	}
}

func TestCatalogClient_Find_ServerError(t *testing.T) {
	srv := newCatalogServer(t, nil, nil, http.StatusInternalServerError)
	defer srv.Close()

	c := &CatalogClient{URL: srv.URL}
	_, err := c.Find(context.Background(), "some/pack")
	if err == nil {
		t.Fatal("expected error on server 500, got nil")
	}
}

func TestCatalogClient_Find_CacheHit(t *testing.T) {
	var callCount int
	entry := sampleCatalogEntries[0]
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(mustMarshal(t, entry)) //nolint:errcheck
	}))
	defer srv.Close()

	now := time.Now()
	c := &CatalogClient{
		URL: srv.URL,
		TTL: time.Hour,
		now: func() time.Time { return now },
	}

	// First call hits the server.
	if _, err := c.Find(context.Background(), "flightdeck/hello-pack"); err != nil {
		t.Fatalf("first Find: %v", err)
	}
	if callCount != 1 {
		t.Fatalf("expected 1 call, got %d", callCount)
	}

	// Second call is cache hit.
	if _, err := c.Find(context.Background(), "flightdeck/hello-pack"); err != nil {
		t.Fatalf("second Find: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected cache hit (still 1 call), got %d", callCount)
	}
}

func TestCatalogClient_Find_Disabled(t *testing.T) {
	c := &CatalogClient{URL: ""}
	_, err := c.Find(context.Background(), "some/pack")
	if err == nil {
		t.Fatal("expected error when URL is empty")
	}
}

func TestToCatalogEntry(t *testing.T) {
	idx := IndexEntry{
		Name:    "test/pack",
		Version: "1.0.0",
		Tier:    "community",
	}
	ce := ToCatalogEntry(idx)
	if ce.Name != idx.Name {
		t.Errorf("Name = %q, want %q", ce.Name, idx.Name)
	}
	if ce.Downloads != 0 {
		t.Errorf("Downloads should be 0 for static index conversion, got %d", ce.Downloads)
	}
	if ce.Rating != 0 {
		t.Errorf("Rating should be 0 for static index conversion, got %f", ce.Rating)
	}
}

func TestCatalogClient_Search_NetworkFailure(t *testing.T) {
	// Use a server that immediately closes the connection.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Hijack and close to simulate network failure.
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c := &CatalogClient{URL: srv.URL}
	_, err := c.Search(context.Background(), "term")
	if err == nil {
		t.Fatal("expected error on network failure / 503, got nil")
	}
}

func TestCatalogClient_CacheExpiry(t *testing.T) {
	var callCount int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(mustMarshal(t, sampleCatalogEntries)) //nolint:errcheck
	}))
	defer srv.Close()

	now := time.Now()
	c := &CatalogClient{
		URL: srv.URL,
		TTL: time.Second,
		now: func() time.Time { return now },
	}

	// First call populates the cache.
	if _, err := c.Search(context.Background(), "hello"); err != nil {
		t.Fatalf("first Search: %v", err)
	}

	// Advance time past the TTL.
	now = now.Add(2 * time.Second)

	// Second call should re-fetch because cache expired.
	if _, err := c.Search(context.Background(), "hello"); err != nil {
		t.Fatalf("second Search: %v", err)
	}
	if callCount != 2 {
		t.Errorf("expected 2 server calls after cache expiry, got %d", callCount)
	}
}
