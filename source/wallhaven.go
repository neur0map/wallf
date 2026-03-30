package source

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const wallhavenDefaultBase = "https://wallhaven.cc/api/v1"

// Wallhaven is a Source adapter for the Wallhaven API.
type Wallhaven struct {
	baseURL string
	client  *http.Client
	apiKey  string
}

// NewWallhaven creates a Wallhaven adapter. apiKey may be empty for SFW-only searches.
func NewWallhaven(apiKey string) *Wallhaven {
	return &Wallhaven{
		baseURL: wallhavenDefaultBase,
		client:  &http.Client{},
		apiKey:  apiKey,
	}
}

// NewWallhavenWithBase creates a Wallhaven adapter with a custom base URL and
// HTTP client, intended for testing with httptest servers.
func NewWallhavenWithBase(baseURL, apiKey string) *Wallhaven {
	return &Wallhaven{
		baseURL: baseURL,
		client:  &http.Client{},
		apiKey:  apiKey,
	}
}

// CategoryBits is the exported wrapper around categoryBits for testing.
func CategoryBits(cats string) string { return categoryBits(cats) }

// Name returns the unique slug for this source.
func (w *Wallhaven) Name() string { return "wallhaven" }

// NeedsKey reports whether an API key is required. Wallhaven does not require
// one for SFW searches.
func (w *Wallhaven) NeedsKey() bool { return false }

// Categories returns the categories supported by Wallhaven.
func (w *Wallhaven) Categories() []string {
	return []string{"general", "anime", "people"}
}

// Search fetches wallpapers from Wallhaven matching opts.
func (w *Wallhaven) Search(opts SearchOpts) ([]WallpaperResult, error) {
	params := url.Values{}
	params.Set("purity", "100")

	if opts.Sort != "" {
		params.Set("sorting", opts.Sort)
	}
	if opts.MinRes != "" {
		params.Set("atleast", opts.MinRes)
	}
	if opts.TimeRange != "" {
		params.Set("topRange", opts.TimeRange)
	}
	if opts.Page > 0 {
		params.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.Count > 0 {
		params.Set("per_page", strconv.Itoa(opts.Count))
	}

	cats := categoryBits(strings.Join(opts.Categories, ","))
	params.Set("categories", cats)

	if w.apiKey != "" {
		params.Set("apikey", w.apiKey)
	}

	if opts.Query != "" {
		params.Set("q", opts.Query)
	}

	if len(opts.Colors) > 0 {
		params.Set("colors", strings.Join(opts.Colors, ","))
	}

	// Default to landscape ratios to avoid portrait/mobile images
	if opts.Ratios != "" {
		params.Set("ratios", opts.Ratios)
	} else {
		params.Set("ratios", "landscape")
	}

	reqURL := w.baseURL + "/search?" + params.Encode()

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("wallhaven: build request: %w", err)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("wallhaven: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("wallhaven: rate limited (429)")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wallhaven: unexpected status %d", resp.StatusCode)
	}

	var apiResp wallhavenAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("wallhaven: decode response: %w", err)
	}

	results := make([]WallpaperResult, 0, len(apiResp.Data))
	for _, img := range apiResp.Data {
		tags := make([]string, 0, len(img.Tags))
		for _, t := range img.Tags {
			tags = append(tags, t.Name)
		}
		results = append(results, WallpaperResult{
			ID:         img.ID,
			Source:     w.Name(),
			PreviewURL: img.Thumbs.Large,
			FullURL:    img.Path,
			Resolution: img.Resolution,
			Views:      img.Views,
			Favorites:  img.Favorites,
			Tags:       tags,
		})
	}
	return results, nil
}

// SearchTotal returns the total number of matching wallpapers on the server.
func (w *Wallhaven) SearchTotal(opts SearchOpts) (int, error) {
	// Reuse Search logic but only fetch 1 result to get the meta.total
	params := url.Values{}
	params.Set("purity", "100")
	params.Set("per_page", "1")
	if opts.Query != "" {
		params.Set("q", opts.Query)
	}
	if opts.MinRes != "" {
		params.Set("atleast", opts.MinRes)
	}
	if len(opts.Colors) > 0 {
		params.Set("colors", strings.Join(opts.Colors, ","))
	}
	cats := categoryBits(strings.Join(opts.Categories, ","))
	params.Set("categories", cats)
	params.Set("ratios", "landscape")
	if w.apiKey != "" {
		params.Set("apikey", w.apiKey)
	}

	reqURL := w.baseURL + "/search?" + params.Encode()
	resp, err := w.client.Get(reqURL)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("wallhaven: HTTP %d", resp.StatusCode)
	}
	var apiResp wallhavenAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return -1, err
	}
	return apiResp.Meta.Total, nil
}

// categoryBits converts a comma-separated category list to a 3-bit string
// (general, anime, people). Empty or unrecognised input defaults to "110".
func categoryBits(cats string) string {
	if cats == "" {
		return "110"
	}
	parts := strings.Split(cats, ",")
	set := make(map[string]bool)
	for _, p := range parts {
		set[strings.TrimSpace(p)] = true
	}
	general, anime, people := "0", "0", "0"
	if set["general"] {
		general = "1"
	}
	if set["anime"] {
		anime = "1"
	}
	if set["people"] {
		people = "1"
	}
	result := general + anime + people
	if result == "000" {
		return "110"
	}
	return result
}

// --- API response types ---

type wallhavenAPIResponse struct {
	Data []wallhavenImage `json:"data"`
	Meta wallhavenMeta    `json:"meta"`
}

type wallhavenImage struct {
	ID         string         `json:"id"`
	Path       string         `json:"path"`
	Resolution string         `json:"resolution"`
	Views      int            `json:"views"`
	Favorites  int            `json:"favorites"`
	Thumbs     wallhavenThumbs `json:"thumbs"`
	Tags       []wallhavenTag  `json:"tags"`
}

type wallhavenThumbs struct {
	Large    string `json:"large"`
	Original string `json:"original"`
	Small    string `json:"small"`
}

type wallhavenTag struct {
	Name string `json:"name"`
}

type wallhavenMeta struct {
	Total       int `json:"total"`
	PerPage     int `json:"per_page"`
	CurrentPage int `json:"current_page"`
	LastPage    int `json:"last_page"`
}
