// SPDX-License-Identifier: Apache-2.0
package pack

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// CatalogEntry is an IndexEntry augmented with catalog-specific metadata
// (download counts, rating, verified publisher badge). The hosted catalog API
// returns these; the static index returns the same shape with zero augmentations.
type CatalogEntry struct {
	IndexEntry
	Downloads   int64   `json:"downloads"`
	Rating      float64 `json:"rating"`      // 0-5 average user rating
	RatingCount int     `json:"ratingCount"` // number of ratings
}

// ToCatalogEntry converts a static IndexEntry into a CatalogEntry (zeros for hosted-only fields).
func ToCatalogEntry(e IndexEntry) CatalogEntry { return CatalogEntry{IndexEntry: e} }

// CatalogClient fetches pack metadata from the hosted catalog API (task 073).
// It is consulted BEFORE the static index when URL is non-empty. If the hosted
// API is unavailable the caller should fall back to the static IndexClient.
//
// API contract (the server side is external/private; this is the client seam):
//
//	GET <URL>/v1/packs?q=<term>           → []CatalogEntry
//	GET <URL>/v1/packs/<publisher>/<name> → CatalogEntry
//
// Results are cached in-process with a short TTL.
type CatalogClient struct {
	URL        string // e.g. "https://catalog.flightdeck.dev"; empty => disabled
	HTTPClient *http.Client
	TTL        time.Duration // cache TTL (default: 5 minutes)

	mu    sync.Mutex
	cache map[string]catalogCacheEntry // key: search term or pack name
	now   func() time.Time
}

type catalogCacheEntry struct {
	data      []CatalogEntry
	fetchedAt time.Time
}

func (c *CatalogClient) ttl() time.Duration {
	if c.TTL > 0 {
		return c.TTL
	}
	return 5 * time.Minute
}

func (c *CatalogClient) clock() time.Time {
	if c.now != nil {
		return c.now()
	}
	return time.Now()
}

func (c *CatalogClient) client() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: 10 * time.Second}
}

// Enabled reports whether the client is configured (URL is non-empty).
func (c *CatalogClient) Enabled() bool { return c.URL != "" }

// Search queries the catalog for packs matching term (empty = all).
// Returns an error only for non-transient failures; callers should fall back
// to the static index on any error.
func (c *CatalogClient) Search(ctx context.Context, term string) ([]CatalogEntry, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("catalog URL not configured")
	}
	key := "search:" + term
	if cached := c.fromCache(key); cached != nil {
		return cached, nil
	}
	apiURL := strings.TrimRight(c.URL, "/") + "/v1/packs"
	if term != "" {
		apiURL += "?q=" + url.QueryEscape(term)
	}
	entries, err := c.fetch(ctx, apiURL)
	if err != nil {
		return nil, err
	}
	c.toCache(key, entries)
	return entries, nil
}

// Find queries the catalog for a specific pack by name (publisher/name or bare name).
func (c *CatalogClient) Find(ctx context.Context, name string) (*CatalogEntry, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("catalog URL not configured")
	}
	key := "find:" + name
	if cached := c.fromCache(key); cached != nil && len(cached) > 0 {
		e := cached[0]
		return &e, nil
	}
	// Try exact lookup first, then fall back to list search.
	apiURL := strings.TrimRight(c.URL, "/") + "/v1/packs/" + url.PathEscape(name)
	// Try single-entry endpoint.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("catalog GET %s: %s", apiURL, resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	var entry CatalogEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("parsing catalog entry: %w", err)
	}
	c.toCache(key, []CatalogEntry{entry})
	return &entry, nil
}

func (c *CatalogClient) fromCache(key string) []CatalogEntry {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache == nil {
		return nil
	}
	e, ok := c.cache[key]
	if !ok || c.clock().Sub(e.fetchedAt) > c.ttl() {
		return nil
	}
	return e.data
}

func (c *CatalogClient) toCache(key string, data []CatalogEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache == nil {
		c.cache = map[string]catalogCacheEntry{}
	}
	c.cache[key] = catalogCacheEntry{data: data, fetchedAt: c.clock()}
}

func (c *CatalogClient) fetch(ctx context.Context, apiURL string) ([]CatalogEntry, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("catalog GET %s: %s", apiURL, resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, err
	}
	var entries []CatalogEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parsing catalog response: %w", err)
	}
	return entries, nil
}
