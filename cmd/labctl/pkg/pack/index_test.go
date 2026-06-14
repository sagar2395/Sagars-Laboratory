// SPDX-License-Identifier: Apache-2.0

package pack

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const sampleIndex = `{
  "apiVersion": "packs.flightdeck.dev/v1",
  "kind": "PackIndex",
  "generated": "2026-06-14T00:00:00Z",
  "entries": [
    {"name": "flightdeck/hello-pack", "version": "0.1.0", "ref": "oci://ghcr.io/snowops/hello-pack:0.1.0", "publisher": "flightdeck", "displayName": "Hello Pack", "description": "minimal example", "license": "Apache-2.0", "tier": "community", "keywords": ["example"], "verified": true},
    {"name": "acme/kafka-drills", "version": "1.4.2", "ref": "oci://ghcr.io/acme/kafka-drills:1.4.2", "tier": "community", "license": "Apache-2.0", "keywords": ["kafka", "incident"]},
    {"name": "acme/kafka-drills", "version": "1.5.0", "ref": "git+https://github.com/acme/kafka-drills@v1.5.0", "tier": "community", "license": "Apache-2.0", "keywords": ["kafka", "incident"]}
  ]
}`

func TestParseIndexAndValidate(t *testing.T) {
	idx, err := ParseIndex([]byte(sampleIndex))
	if err != nil {
		t.Fatal(err)
	}
	if err := idx.Validate(); err != nil {
		t.Fatalf("expected valid index: %v", err)
	}
}

func TestIndexValidate_ReportsProblems(t *testing.T) {
	bad := `{"apiVersion":"x","kind":"y","entries":[
	  {"name":"Bad Name","version":"nope","ref":"ftp://x","tier":"gold"}]}`
	idx, err := ParseIndex([]byte(bad))
	if err != nil {
		t.Fatal(err)
	}
	err = idx.Validate()
	if err == nil {
		t.Fatal("expected errors")
	}
	for _, want := range []string{"apiVersion", "kind", "name", "semver", "oci:// or git+http", "tier"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("missing %q in: %v", want, err)
		}
	}
}

func TestIndexFind_HighestVersionAndBareName(t *testing.T) {
	idx, _ := ParseIndex([]byte(sampleIndex))
	// bare name resolves, picking the highest version (1.5.0, git+).
	e, ok := idx.Find("kafka-drills")
	if !ok {
		t.Fatal("expected to find kafka-drills")
	}
	if e.Version != "1.5.0" {
		t.Errorf("Find returned version %q, want 1.5.0 (highest)", e.Version)
	}
	if got := e.ResolveRef(); got != "https://github.com/acme/kafka-drills@v1.5.0" {
		t.Errorf("ResolveRef stripped git+ wrong: %q", got)
	}
	// full name also works.
	if _, ok := idx.Find("flightdeck/hello-pack"); !ok {
		t.Error("expected to find flightdeck/hello-pack by full name")
	}
	if _, ok := idx.Find("nonexistent"); ok {
		t.Error("did not expect to find nonexistent")
	}
}

func TestIndexSearch_DedupesToLatest(t *testing.T) {
	idx, _ := ParseIndex([]byte(sampleIndex))
	res := idx.Search("kafka")
	if len(res) != 1 {
		t.Fatalf("expected 1 deduped kafka result, got %d", len(res))
	}
	if res[0].Version != "1.5.0" {
		t.Errorf("search returned %q, want latest 1.5.0", res[0].Version)
	}
	// empty term lists everything (deduped): hello-pack + kafka-drills.
	if all := idx.Search(""); len(all) != 2 {
		t.Errorf("empty search returned %d, want 2", len(all))
	}
	// keyword match on description.
	if r := idx.Search("minimal"); len(r) != 1 || r[0].Name != "flightdeck/hello-pack" {
		t.Errorf("description search failed: %+v", r)
	}
}

func TestIndexClient_FetchCacheTTL(t *testing.T) {
	cache := filepath.Join(t.TempDir(), "registry-index.json")
	var fetches int
	get := func(_ context.Context, url string) ([]byte, error) {
		fetches++
		return []byte(sampleIndex), nil
	}
	c := &IndexClient{
		URL: "https://example/index.json", CachePath: cache,
		TTL: time.Hour, Get: get,
	}
	// First load fetches and caches.
	if _, stale, err := c.Load(context.Background()); err != nil || stale {
		t.Fatalf("first load: stale=%v err=%v", stale, err)
	}
	if fetches != 1 {
		t.Fatalf("expected 1 fetch, got %d", fetches)
	}
	// Second load within TTL hits the cache (no fetch).
	if _, _, err := c.Load(context.Background()); err != nil {
		t.Fatal(err)
	}
	if fetches != 1 {
		t.Errorf("expected cache hit (still 1 fetch), got %d", fetches)
	}
	// Advance the clock past the TTL → refetch.
	c.now = func() time.Time { return time.Now().Add(2 * time.Hour) }
	if _, _, err := c.Load(context.Background()); err != nil {
		t.Fatal(err)
	}
	if fetches != 2 {
		t.Errorf("expected refetch after TTL, got %d fetches", fetches)
	}
}

func TestIndexClient_StaleFallbackOnFetchError(t *testing.T) {
	cache := filepath.Join(t.TempDir(), "registry-index.json")
	if err := os.WriteFile(cache, []byte(sampleIndex), 0o644); err != nil {
		t.Fatal(err)
	}
	// Make the cache look stale and the network down.
	c := &IndexClient{
		URL: "https://example/index.json", CachePath: cache, TTL: time.Nanosecond,
		now: func() time.Time { return time.Now().Add(time.Hour) },
		Get: func(context.Context, string) ([]byte, error) { return nil, fmt.Errorf("network down") },
	}
	idx, stale, err := c.Load(context.Background())
	if err != nil {
		t.Fatalf("expected stale fallback, got err: %v", err)
	}
	if !stale {
		t.Error("expected stale=true on cache fallback")
	}
	if idx == nil || len(idx.Entries) != 3 {
		t.Errorf("stale index not loaded: %+v", idx)
	}
}

func TestIndexClient_VerifiesSignature(t *testing.T) {
	cache := filepath.Join(t.TempDir(), "registry-index.json")
	var verified bool
	get := func(_ context.Context, url string) ([]byte, error) {
		if strings.HasSuffix(url, ".sig") {
			return []byte("SIGNATURE"), nil
		}
		return []byte(sampleIndex), nil
	}
	cosign := func(_ context.Context, _ string, args ...string) error {
		if args[0] != "verify-blob" {
			t.Errorf("expected verify-blob, got %v", args)
		}
		verified = true
		return nil
	}
	c := &IndexClient{
		URL: "https://example/index.json", CachePath: cache,
		VerifyKey: "registry.pub", Get: get, Cosign: cosign,
	}
	if _, _, err := c.Load(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !verified {
		t.Error("signature was not verified before trusting the index")
	}
}

func TestIndexClient_FailsClosedOnBadSignature(t *testing.T) {
	c := &IndexClient{
		URL:       "https://example/index.json",
		CachePath: filepath.Join(t.TempDir(), "i.json"),
		VerifyKey: "registry.pub",
		Get:       func(context.Context, string) ([]byte, error) { return []byte(sampleIndex), nil },
		Cosign:    func(context.Context, string, ...string) error { return fmt.Errorf("invalid signature") },
	}
	if _, _, err := c.Load(context.Background()); err == nil {
		t.Fatal("expected a signature failure to fail closed")
	}
}
