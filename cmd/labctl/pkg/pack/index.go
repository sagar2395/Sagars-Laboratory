// SPDX-License-Identifier: Apache-2.0

package pack

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Pack registry index (task 069). The marketplace catalog is a static, signed
// JSON artifact — one published file aggregating one entry per pack/version,
// served from a plain repo over Pages/CDN (the Krew / Artifact-Hub pattern: no
// infra, auditable, PR-moderated). The CLI fetches it, caches it with a TTL,
// optionally verifies its cosign signature, and resolves a pack name to an
// installable ref. A hosted API (task 073) is a future superset, never a
// replacement for this portable artifact.

// Index identity.
const (
	IndexAPIVersion = "packs.flightdeck.dev/v1"
	IndexKind       = "PackIndex"
)

// DefaultIndexTTL is how long a cached index is trusted before a refetch.
const DefaultIndexTTL = time.Hour

// Index is a parsed registry index.
type Index struct {
	APIVersion string       `json:"apiVersion"`
	Kind       string       `json:"kind"`
	Generated  string       `json:"generated,omitempty"` // RFC3339 build time
	Entries    []IndexEntry `json:"entries"`
}

// IndexEntry mirrors pack.yaml metadata plus the resolvable ref and publisher
// verification status. It is the unit the CLI searches and resolves.
type IndexEntry struct {
	Name        string   `json:"name"`    // publisher/name or name
	Version     string   `json:"version"` // SemVer
	Ref         string   `json:"ref"`     // oci://… or git+https://…@ver
	Publisher   string   `json:"publisher,omitempty"`
	DisplayName string   `json:"displayName,omitempty"`
	Description string   `json:"description,omitempty"`
	License     string   `json:"license,omitempty"`
	Tier        string   `json:"tier,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
	Categories  []string `json:"categories,omitempty"`
	Verified    bool     `json:"verified,omitempty"` // publisher identity vetted by the registry
	Digest      string   `json:"digest,omitempty"`   // optional oci layer digest to pin
}

// ResolveRef returns the source string `labctl pack add` understands for this
// entry. OCI refs pass through; a "git+" prefix is stripped so the git
// installer sees a plain URL[@ref].
func (e IndexEntry) ResolveRef() string {
	if strings.HasPrefix(e.Ref, "git+") {
		return strings.TrimPrefix(e.Ref, "git+")
	}
	return e.Ref
}

// ParseIndex reads an index from JSON bytes.
func ParseIndex(data []byte) (*Index, error) {
	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("parsing registry index: %w", err)
	}
	return &idx, nil
}

// Validate checks the index for schema correctness, reporting every problem.
func (idx *Index) Validate() error {
	var errs []string
	add := func(f string, a ...interface{}) { errs = append(errs, fmt.Sprintf(f, a...)) }

	if idx.APIVersion != IndexAPIVersion {
		add("apiVersion must be %q (got %q)", IndexAPIVersion, idx.APIVersion)
	}
	if idx.Kind != IndexKind {
		add("kind must be %q (got %q)", IndexKind, idx.Kind)
	}
	seen := map[string]bool{}
	for i, e := range idx.Entries {
		where := fmt.Sprintf("entries[%d]", i)
		if !nameRe.MatchString(e.Name) {
			add("%s: name %q is invalid", where, e.Name)
		}
		if _, err := parseSemver(e.Version); err != nil {
			add("%s (%s): version %q is not valid semver", where, e.Name, e.Version)
		}
		if !IsOCIRef(e.Ref) && !strings.HasPrefix(e.Ref, "git+http") {
			add("%s (%s): ref %q must be oci:// or git+http(s)://", where, e.Name, e.Ref)
		}
		if e.Tier != "" && !validTiers[e.Tier] {
			add("%s (%s): tier %q is invalid", where, e.Name, e.Tier)
		}
		key := e.Name + "@" + e.Version
		if seen[key] {
			add("%s: duplicate entry %s", where, key)
		}
		seen[key] = true
	}
	if len(errs) > 0 {
		return fmt.Errorf("invalid registry index:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return nil
}

// Find returns the highest-version entry whose name matches q. The match is
// exact on the full "publisher/name", or on the bare name when q has no slash
// (so "kafka-drills" finds "acme/kafka-drills" if unambiguous).
func (idx *Index) Find(q string) (*IndexEntry, bool) {
	var matches []IndexEntry
	for _, e := range idx.Entries {
		if e.Name == q || (!strings.Contains(q, "/") && bareName(e.Name) == q) {
			matches = append(matches, e)
		}
	}
	if len(matches) == 0 {
		return nil, false
	}
	sort.SliceStable(matches, func(i, j int) bool {
		c, _ := compareSemver(matches[i].Version, matches[j].Version)
		return c > 0 // highest version first
	})
	best := matches[0]
	return &best, true
}

// Search returns entries matching term (case-insensitive substring over name,
// display name, description, and keywords), newest version per pack, sorted by
// name. An empty term lists everything.
func (idx *Index) Search(term string) []IndexEntry {
	term = strings.ToLower(strings.TrimSpace(term))
	latest := map[string]IndexEntry{}
	for _, e := range idx.Entries {
		if term != "" && !entryMatches(e, term) {
			continue
		}
		cur, ok := latest[e.Name]
		if !ok {
			latest[e.Name] = e
			continue
		}
		if c, _ := compareSemver(e.Version, cur.Version); c > 0 {
			latest[e.Name] = e
		}
	}
	out := make([]IndexEntry, 0, len(latest))
	for _, e := range latest {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func entryMatches(e IndexEntry, term string) bool {
	hay := strings.ToLower(strings.Join(append([]string{
		e.Name, e.DisplayName, e.Description, e.Publisher,
	}, e.Keywords...), " "))
	hay += " " + strings.ToLower(strings.Join(e.Categories, " "))
	return strings.Contains(hay, term)
}

func bareName(name string) string {
	if i := strings.LastIndexByte(name, '/'); i >= 0 {
		return name[i+1:]
	}
	return name
}

// --- fetch / cache / verify ---

// Getter fetches the bytes at url. The default is an HTTP GET; tests inject a
// stub so index loading is hermetic.
type Getter func(ctx context.Context, url string) ([]byte, error)

// DefaultGetter fetches url over HTTP(S), or reads a local file when url uses
// the file:// scheme (handy for local index fixtures and air-gapped mirrors).
func DefaultGetter(ctx context.Context, url string) ([]byte, error) {
	if strings.HasPrefix(url, "file://") {
		return os.ReadFile(strings.TrimPrefix(url, "file://"))
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 16<<20)) // 16 MiB cap
}

// IndexClient loads a registry index with TTL caching and optional cosign
// signature verification.
type IndexClient struct {
	URL       string        // index.json URL
	SigURL    string        // signature URL; defaults to URL + ".sig"
	CachePath string        // local cache file (e.g. .labctl/cache/registry-index.json)
	TTL       time.Duration // cache freshness window (0 => DefaultIndexTTL)
	VerifyKey string        // cosign public key path; "" => skip verification
	Get       Getter        // defaults to DefaultGetter
	Cosign    Runner        // defaults to DefaultCosign (verify-blob)

	now func() time.Time // injectable clock for tests
}

func (c *IndexClient) ttl() time.Duration {
	if c.TTL > 0 {
		return c.TTL
	}
	return DefaultIndexTTL
}

func (c *IndexClient) clock() time.Time {
	if c.now != nil {
		return c.now()
	}
	return time.Now()
}

// Load returns the index, preferring a fresh cache, otherwise fetching (and
// verifying) and refreshing the cache. If a fetch fails but a (stale) cache
// exists, the cache is used so the CLI keeps working offline; the staleness is
// surfaced via the returned bool (stale=true).
func (c *IndexClient) Load(ctx context.Context) (*Index, bool, error) {
	if c.URL == "" {
		return nil, false, fmt.Errorf("no registry index configured (set PACK_REGISTRY_INDEX)")
	}
	if c.cacheFresh() {
		if idx, err := c.readCache(); err == nil {
			return idx, false, nil
		}
	}

	data, ferr := c.fetchAndVerify(ctx)
	if ferr != nil {
		// Fall back to any cached copy so discovery survives a network blip.
		if idx, err := c.readCache(); err == nil {
			return idx, true, nil
		}
		return nil, false, ferr
	}

	idx, err := ParseIndex(data)
	if err != nil {
		return nil, false, err
	}
	if err := idx.Validate(); err != nil {
		return nil, false, err
	}
	if err := c.writeCache(data); err != nil {
		// A cache-write failure is non-fatal; we still have a valid index.
		return idx, false, nil
	}
	return idx, false, nil
}

func (c *IndexClient) cacheFresh() bool {
	if c.CachePath == "" {
		return false
	}
	info, err := os.Stat(c.CachePath)
	if err != nil {
		return false
	}
	return c.clock().Sub(info.ModTime()) < c.ttl()
}

func (c *IndexClient) readCache() (*Index, error) {
	if c.CachePath == "" {
		return nil, fmt.Errorf("no cache configured")
	}
	data, err := os.ReadFile(c.CachePath)
	if err != nil {
		return nil, err
	}
	idx, err := ParseIndex(data)
	if err != nil {
		return nil, err
	}
	if err := idx.Validate(); err != nil {
		return nil, err
	}
	return idx, nil
}

func (c *IndexClient) writeCache(data []byte) error {
	if c.CachePath == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(c.CachePath), 0o755); err != nil {
		return err
	}
	tmp := c.CachePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, c.CachePath)
}

func (c *IndexClient) fetchAndVerify(ctx context.Context) ([]byte, error) {
	get := c.Get
	if get == nil {
		get = DefaultGetter
	}
	data, err := get(ctx, c.URL)
	if err != nil {
		return nil, fmt.Errorf("fetching registry index: %w", err)
	}

	if c.VerifyKey == "" {
		return data, nil // community default: unsigned index, no verification
	}

	sigURL := c.SigURL
	if sigURL == "" {
		sigURL = c.URL + ".sig"
	}
	sig, err := get(ctx, sigURL)
	if err != nil {
		return nil, fmt.Errorf("fetching index signature: %w", err)
	}

	cosign := c.Cosign
	if cosign == nil {
		cosign = DefaultCosign
	}
	stage, err := os.MkdirTemp("", "labctl-index-verify-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(stage)
	blob := filepath.Join(stage, "index.json")
	sigFile := filepath.Join(stage, "index.json.sig")
	if err := os.WriteFile(blob, data, 0o644); err != nil {
		return nil, err
	}
	if err := os.WriteFile(sigFile, sig, 0o644); err != nil {
		return nil, err
	}
	if err := cosign(ctx, "", "verify-blob", "--key", c.VerifyKey, "--signature", sigFile, blob); err != nil {
		return nil, fmt.Errorf("registry index signature verification failed: %w", err)
	}
	return data, nil
}
