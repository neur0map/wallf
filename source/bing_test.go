package source_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/neur0map/wallf/source"
)

func TestBingName(t *testing.T) {
	b := source.NewBing()
	if got := b.Name(); got != "bing" {
		t.Errorf("Name() = %q, want %q", got, "bing")
	}
}

func TestBingNeedsKey(t *testing.T) {
	b := source.NewBing()
	if b.NeedsKey() {
		t.Error("NeedsKey() = true, want false")
	}
}

func TestBingCategories(t *testing.T) {
	b := source.NewBing()
	cats := b.Categories()
	if len(cats) != 0 {
		t.Errorf("Categories() returned %d items, want 0/nil", len(cats))
	}
}

func TestBingSearch(t *testing.T) {
	fixture := map[string]interface{}{
		"images": []map[string]interface{}{
			{
				"url":       "/th?id=OHR.SomeScene_EN-US1234567890&rf=LaDigue_1920x1080.jpg",
				"urlbase":   "/th?id=OHR.SomeScene_EN-US1234567890",
				"title":     "Some Scenic View",
				"copyright": "Some Scenic View © Photographer",
			},
		},
	}

	var capturedQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(fixture)
	}))
	defer srv.Close()

	b := source.NewBingWithBase(srv.URL)
	results, err := b.Search(source.SearchOpts{Count: 1})
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}

	// Verify format=js param is present
	if !strings.Contains(capturedQuery, "format=js") {
		t.Errorf("query %q missing format=js", capturedQuery)
	}

	if len(results) != 1 {
		t.Fatalf("Search() returned %d results, want 1", len(results))
	}

	r := results[0]

	// Verify UHD URL construction: {base}{urlbase}_UHD.jpg
	wantFullURL := srv.URL + "/th?id=OHR.SomeScene_EN-US1234567890_UHD.jpg"
	if r.FullURL != wantFullURL {
		t.Errorf("FullURL = %q, want %q", r.FullURL, wantFullURL)
	}

	// Verify Tags contains title
	if len(r.Tags) == 0 || r.Tags[0] != "Some Scenic View" {
		t.Errorf("Tags = %v, want [Some Scenic View]", r.Tags)
	}

	if r.Source != "bing" {
		t.Errorf("Source = %q, want %q", r.Source, "bing")
	}
	if r.Resolution != "3840x2160" {
		t.Errorf("Resolution = %q, want %q", r.Resolution, "3840x2160")
	}
}

func TestBingIgnoresQuery(t *testing.T) {
	fixture := map[string]interface{}{
		"images": []map[string]interface{}{
			{
				"url":       "/th?id=OHR.Mountain_EN-US111&rf=LaDigue_1920x1080.jpg",
				"urlbase":   "/th?id=OHR.Mountain_EN-US111",
				"title":     "Mountain View",
				"copyright": "Mountain View © Photo",
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(fixture)
	}))
	defer srv.Close()

	b := source.NewBingWithBase(srv.URL)

	// Passing a query should not cause an error — Bing ignores it.
	results, err := b.Search(source.SearchOpts{Sort: "mountains"})
	if err != nil {
		t.Fatalf("Search() with query error: %v", err)
	}
	if len(results) == 0 {
		t.Error("Search() returned no results, want at least 1")
	}
}
