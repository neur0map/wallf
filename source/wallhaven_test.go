package source_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/neur0map/wallf/source"
)

func TestWallhavenName(t *testing.T) {
	wh := source.NewWallhaven("")
	if got := wh.Name(); got != "wallhaven" {
		t.Errorf("Name() = %q, want %q", got, "wallhaven")
	}
}

func TestWallhavenNeedsKey(t *testing.T) {
	wh := source.NewWallhaven("")
	if wh.NeedsKey() {
		t.Error("NeedsKey() = true, want false")
	}
}

func TestWallhavenCategories(t *testing.T) {
	wh := source.NewWallhaven("")
	cats := wh.Categories()
	if len(cats) != 3 {
		t.Errorf("Categories() returned %d items, want 3", len(cats))
	}
}

func TestWallhavenSearch(t *testing.T) {
	fixture := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"id":         "abc123",
				"path":       "https://w.wallhaven.cc/full/ab/wallhaven-abc123.jpg",
				"resolution": "3840x2160",
				"views":      1234,
				"favorites":  56,
				"thumbs": map[string]string{
					"large":    "https://th.wallhaven.cc/lg/ab/abc123.jpg",
					"original": "https://th.wallhaven.cc/orig/ab/abc123.jpg",
					"small":    "https://th.wallhaven.cc/sm/ab/abc123.jpg",
				},
				"tags": []map[string]interface{}{
					{"name": "cyberpunk"},
					{"name": "cityscape"},
				},
			},
		},
		"meta": map[string]interface{}{
			"total":        1,
			"per_page":     24,
			"current_page": 1,
			"last_page":    1,
		},
	}

	var capturedQuery url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(fixture)
	}))
	defer srv.Close()

	wh := source.NewWallhavenWithBase(srv.URL, "")
	results, err := wh.Search(source.SearchOpts{
		Query: "cyberpunk",
	})
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}

	// Verify query params
	if q := capturedQuery.Get("q"); q != "cyberpunk" {
		t.Errorf("q param = %q, want %q", q, "cyberpunk")
	}
	if purity := capturedQuery.Get("purity"); purity != "100" {
		t.Errorf("purity param = %q, want %q", purity, "100")
	}

	// Verify result mapping
	if len(results) != 1 {
		t.Fatalf("Search() returned %d results, want 1", len(results))
	}
	r := results[0]
	if r.ID != "abc123" {
		t.Errorf("ID = %q, want %q", r.ID, "abc123")
	}
	if r.Source != "wallhaven" {
		t.Errorf("Source = %q, want %q", r.Source, "wallhaven")
	}
	if r.Resolution != "3840x2160" {
		t.Errorf("Resolution = %q, want %q", r.Resolution, "3840x2160")
	}
	if r.Views != 1234 {
		t.Errorf("Views = %d, want 1234", r.Views)
	}
	if len(r.Tags) != 2 || r.Tags[0] != "cyberpunk" {
		t.Errorf("Tags = %v, want [cyberpunk cityscape]", r.Tags)
	}
}

func TestWallhavenCategoryBits(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"general", "100"},
		{"anime", "010"},
		{"people", "001"},
		{"general,anime", "110"},
		{"", "110"},
	}
	for _, tt := range tests {
		got := source.CategoryBits(tt.input)
		if got != tt.want {
			t.Errorf("CategoryBits(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestWallhavenSearchZeroResults(t *testing.T) {
	fixture := map[string]interface{}{
		"data": []interface{}{},
		"meta": map[string]interface{}{
			"total":        0,
			"per_page":     24,
			"current_page": 1,
			"last_page":    1,
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(fixture)
	}))
	defer srv.Close()

	wh := source.NewWallhavenWithBase(srv.URL, "")
	results, err := wh.Search(source.SearchOpts{})
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Search() returned %d results, want 0", len(results))
	}
}
