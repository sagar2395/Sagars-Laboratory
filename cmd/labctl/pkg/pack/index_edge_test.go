// SPDX-License-Identifier: Apache-2.0
package pack

import (
	"strings"
	"testing"
)

// TestParseIndex_EmptyObject verifies that an empty JSON object produces a
// zero-value Index with no entries.
func TestParseIndex_EmptyObject(t *testing.T) {
	idx, err := ParseIndex([]byte(`{}`))
	if err != nil {
		t.Fatalf("ParseIndex({}) error: %v", err)
	}
	if len(idx.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(idx.Entries))
	}
}

// TestParseIndex_InvalidJSON verifies that malformed JSON returns an error.
func TestParseIndex_InvalidJSON(t *testing.T) {
	_, err := ParseIndex([]byte(`{not valid json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

// TestIndex_Validate_EmptyEntries verifies that an Index with no apiVersion/kind
// (e.g. just `{}`) fails Validate() with actionable messages.
func TestIndex_Validate_EmptyEntries(t *testing.T) {
	idx := &Index{} // zero value: apiVersion="" kind="" entries=nil
	err := idx.Validate()
	if err == nil {
		t.Fatal("expected validation error for empty index, got nil")
	}
	if !strings.Contains(err.Error(), "apiVersion") {
		t.Errorf("error should mention 'apiVersion': %v", err)
	}
}

// TestIndex_Find_NoEntries verifies that Find returns (nil, false) when
// the index has no entries.
func TestIndex_Find_NoEntries(t *testing.T) {
	idx := &Index{}
	entry, ok := idx.Find("anything")
	if ok {
		t.Error("expected ok=false for empty index")
	}
	if entry != nil {
		t.Errorf("expected nil entry, got %+v", entry)
	}
}

// TestIndex_Search_EmptyQueryReturnsAll verifies that an empty search term
// returns all entries (newest version per pack name).
func TestIndex_Search_EmptyQueryReturnsAll(t *testing.T) {
	idx := &Index{
		Entries: []IndexEntry{
			{Name: "acme/pack-a", Version: "1.0.0", Ref: "oci://registry/pack-a:1.0.0"},
			{Name: "acme/pack-b", Version: "2.0.0", Ref: "oci://registry/pack-b:2.0.0"},
		},
	}
	results := idx.Search("")
	if len(results) != 2 {
		t.Errorf("empty query should return all %d entries, got %d", len(idx.Entries), len(results))
	}
}

// TestIndex_Search_NoMatch verifies that a query matching no entries returns
// an empty slice (not nil panic).
func TestIndex_Search_NoMatch(t *testing.T) {
	idx := &Index{
		Entries: []IndexEntry{
			{Name: "acme/pack-a", Version: "1.0.0", Ref: "oci://registry/pack-a:1.0.0"},
		},
	}
	results := idx.Search("zzz-no-match-zzz")
	if len(results) != 0 {
		t.Errorf("expected empty results for non-matching query, got %d", len(results))
	}
}

// TestIndexEntry_ResolveRef_GitPlusStripped verifies that ResolveRef strips the
// "git+" prefix from git source refs.
func TestIndexEntry_ResolveRef_GitPlusStripped(t *testing.T) {
	e := IndexEntry{Ref: "git+https://github.com/example/pack@v1.0.0"}
	got := e.ResolveRef()
	want := "https://github.com/example/pack@v1.0.0"
	if got != want {
		t.Errorf("ResolveRef() = %q, want %q", got, want)
	}
}

// TestIndexEntry_ResolveRef_OCIPassThrough verifies that oci:// refs are
// returned unchanged.
func TestIndexEntry_ResolveRef_OCIPassThrough(t *testing.T) {
	e := IndexEntry{Ref: "oci://registry.example.com/pack:v1.2.3"}
	got := e.ResolveRef()
	if got != e.Ref {
		t.Errorf("ResolveRef() = %q, want %q (unchanged)", got, e.Ref)
	}
}

// TestIndex_Find_ReturnsHighestVersion verifies that when multiple entries share
// the same name, Find returns the highest-version one.
func TestIndex_Find_ReturnsHighestVersion(t *testing.T) {
	idx := &Index{
		Entries: []IndexEntry{
			{Name: "acme/pack-a", Version: "1.0.0", Ref: "oci://registry/pack-a:1.0.0"},
			{Name: "acme/pack-a", Version: "2.0.0", Ref: "oci://registry/pack-a:2.0.0"},
		},
	}
	entry, ok := idx.Find("acme/pack-a")
	if !ok {
		t.Fatal("expected to find entry, got ok=false")
	}
	if entry.Version != "2.0.0" {
		t.Errorf("expected highest version 2.0.0, got %s", entry.Version)
	}
}
