package source

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const bingDefaultBase = "https://www.bing.com"
const bingMaxCount = 8

// Bing is a Source adapter that fetches the Bing daily wallpapers.
type Bing struct {
	baseURL string
	client  *http.Client
}

// NewBing creates a Bing adapter.
func NewBing() *Bing {
	return &Bing{
		baseURL: bingDefaultBase,
		client:  &http.Client{},
	}
}

// NewBingWithBase creates a Bing adapter with a custom base URL for testing.
func NewBingWithBase(baseURL string) *Bing {
	return &Bing{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// Name returns the unique slug for this source.
func (b *Bing) Name() string { return "bing" }

// NeedsKey reports whether an API key is required. Bing does not.
func (b *Bing) NeedsKey() bool { return false }

// Categories returns the categories supported by Bing. Bing has no categories.
func (b *Bing) Categories() []string { return nil }

// SearchTotal returns 8 — Bing always has up to 8 daily wallpapers.
func (b *Bing) SearchTotal(_ SearchOpts) (int, error) { return 8, nil }

// Search fetches daily wallpapers from Bing. The query in opts is intentionally
// ignored — Bing always returns its curated daily images.
func (b *Bing) Search(opts SearchOpts) ([]WallpaperResult, error) {
	n := opts.Count
	if n <= 0 || n > bingMaxCount {
		n = bingMaxCount
	}

	reqURL := fmt.Sprintf("%s/HPImageArchive.aspx?format=js&idx=0&n=%d", b.baseURL, n)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("bing: build request: %w", err)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bing: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bing: unexpected status %d", resp.StatusCode)
	}

	var apiResp bingAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("bing: decode response: %w", err)
	}

	results := make([]WallpaperResult, 0, len(apiResp.Images))
	for _, img := range apiResp.Images {
		id := extractBingID(img.URLBase)
		fullURL := b.baseURL + img.URLBase + "_UHD.jpg"
		results = append(results, WallpaperResult{
			ID:         id,
			Source:     b.Name(),
			PreviewURL: b.baseURL + img.URL,
			FullURL:    fullURL,
			Resolution: "3840x2160",
			Views:      0,
			Favorites:  0,
			Tags:       []string{img.Title},
		})
	}
	return results, nil
}

// extractBingID parses an ID from the URLBase field. The URLBase looks like
// "/th?id=OHR.SomeName_EN-US1234567890" — we take everything after "id=".
func extractBingID(urlbase string) string {
	lower := strings.ToLower(urlbase)
	idx := strings.Index(lower, "id=")
	if idx == -1 {
		// Fallback: use the last path segment.
		parts := strings.Split(urlbase, "/")
		return parts[len(parts)-1]
	}
	return urlbase[idx+3:]
}

// --- API response types ---

type bingAPIResponse struct {
	Images []bingImage `json:"images"`
}

type bingImage struct {
	URL       string `json:"url"`
	URLBase   string `json:"urlbase"`
	Title     string `json:"title"`
	Copyright string `json:"copyright"`
}
