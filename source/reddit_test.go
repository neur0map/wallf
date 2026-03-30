package source_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/neur0map/wallf/source"
)

func TestRedditName(t *testing.T) {
	r := source.NewReddit()
	if got := r.Name(); got != "reddit" {
		t.Errorf("Name() = %q, want %q", got, "reddit")
	}
}

func TestRedditNeedsKey(t *testing.T) {
	r := source.NewReddit()
	if r.NeedsKey() {
		t.Error("NeedsKey() = true, want false")
	}
}

func TestRedditCategories(t *testing.T) {
	r := source.NewReddit()
	cats := r.Categories()
	want := []string{"wallpapers", "wallpaper", "unixporn"}
	if len(cats) != len(want) {
		t.Fatalf("Categories() returned %d items, want %d", len(cats), len(want))
	}
	for i, c := range cats {
		if c != want[i] {
			t.Errorf("Categories()[%d] = %q, want %q", i, c, want[i])
		}
	}
}

func TestRedditSearch(t *testing.T) {
	fixture := map[string]interface{}{
		"data": map[string]interface{}{
			"children": []map[string]interface{}{
				{
					"data": map[string]interface{}{
						"id":       "post1",
						"title":    "Cool Wallpaper",
						"url":      "https://i.redd.it/cool.jpg",
						"ups":      9001,
						"is_video": false,
						"preview": map[string]interface{}{
							"images": []map[string]interface{}{
								{
									"source": map[string]interface{}{
										"url":    "https://preview.redd.it/cool.jpg?auto=webp",
										"width":  3840,
										"height": 2160,
									},
								},
							},
						},
					},
				},
				{
					"data": map[string]interface{}{
						"id":       "post2",
						"title":    "A Video Post",
						"url":      "https://v.redd.it/somevideo",
						"ups":      100,
						"is_video": true,
						"preview":  map[string]interface{}{},
					},
				},
			},
		},
	}

	var capturedUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUA = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(fixture)
	}))
	defer srv.Close()

	reddit := source.NewRedditWithBase(srv.URL)
	results, err := reddit.Search(source.SearchOpts{Subreddit: "wallpapers"})
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}

	// Verify User-Agent header
	if capturedUA != "wallf/1.0 (wallpaper fetcher)" {
		t.Errorf("User-Agent = %q, want %q", capturedUA, "wallf/1.0 (wallpaper fetcher)")
	}

	// Video post should be filtered
	if len(results) != 1 {
		t.Fatalf("Search() returned %d results, want 1 (video should be filtered)", len(results))
	}

	r := results[0]
	if r.ID != "t3_post1" {
		t.Errorf("ID = %q, want %q", r.ID, "t3_post1")
	}
	if r.Source != "reddit" {
		t.Errorf("Source = %q, want %q", r.Source, "reddit")
	}
	if r.Resolution != "3840x2160" {
		t.Errorf("Resolution = %q, want %q", r.Resolution, "3840x2160")
	}
	if r.Favorites != 9001 {
		t.Errorf("Favorites (ups) = %d, want 9001", r.Favorites)
	}
}

func TestRedditFiltersByExtension(t *testing.T) {
	makePost := func(id, urlStr string, isVideo bool) map[string]interface{} {
		return map[string]interface{}{
			"data": map[string]interface{}{
				"id":       id,
				"title":    id,
				"url":      urlStr,
				"ups":      0,
				"is_video": isVideo,
				"preview":  map[string]interface{}{},
			},
		}
	}

	fixture := map[string]interface{}{
		"data": map[string]interface{}{
			"children": []map[string]interface{}{
				makePost("a", "https://i.redd.it/img.jpg", false),
				makePost("b", "https://i.redd.it/img.png", false),
				makePost("c", "https://i.redd.it/img.webp", false),
				makePost("d", "https://i.redd.it/img.jpeg", false),
				makePost("e", "https://i.redd.it/anim.gif", false),
				makePost("f", "https://v.redd.it/video", false),
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(fixture)
	}))
	defer srv.Close()

	reddit := source.NewRedditWithBase(srv.URL)
	results, err := reddit.Search(source.SearchOpts{})
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}
	if len(results) != 4 {
		t.Errorf("Search() returned %d results, want 4 (.gif and non-image URL filtered)", len(results))
	}
}

func TestRedditDefaultSubreddit(t *testing.T) {
	var capturedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"children": []interface{}{}},
		})
	}))
	defer srv.Close()

	reddit := source.NewRedditWithBase(srv.URL)

	// Empty subreddit → wallpapers
	_, err := reddit.Search(source.SearchOpts{Subreddit: ""})
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}
	if capturedPath != "/r/wallpapers/hot.json" {
		t.Errorf("path = %q, want /r/wallpapers/hot.json", capturedPath)
	}

	// Explicit unixporn → unixporn
	_, err = reddit.Search(source.SearchOpts{Subreddit: "unixporn"})
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}
	if capturedPath != "/r/unixporn/hot.json" {
		t.Errorf("path = %q, want /r/unixporn/hot.json", capturedPath)
	}
}
