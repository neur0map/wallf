package source

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"
)

const redditDefaultBase = "https://www.reddit.com"
const redditUserAgent = "wallf/1.0 (wallpaper fetcher)"

var defaultSubreddits = []string{"wallpapers", "wallpaper", "unixporn"}

var imageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

// Reddit is a Source adapter that fetches wallpapers from Reddit.
type Reddit struct {
	baseURL string
	client  *http.Client
}

// NewReddit creates a Reddit adapter.
func NewReddit() *Reddit {
	return &Reddit{
		baseURL: redditDefaultBase,
		client:  &http.Client{},
	}
}

// NewRedditWithBase creates a Reddit adapter with a custom base URL for testing.
func NewRedditWithBase(baseURL string) *Reddit {
	return &Reddit{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// Name returns the unique slug for this source.
func (r *Reddit) Name() string { return "reddit" }

// NeedsKey reports whether an API key is required.
func (r *Reddit) NeedsKey() bool { return false }

// SearchTotal returns -1 since Reddit doesn't expose total counts.
func (r *Reddit) SearchTotal(_ SearchOpts) (int, error) { return -1, nil }

// Categories returns the default subreddits available as categories.
func (r *Reddit) Categories() []string {
	out := make([]string, len(defaultSubreddits))
	copy(out, defaultSubreddits)
	return out
}

// Search fetches wallpapers from the given subreddit.
func (r *Reddit) Search(opts SearchOpts) ([]WallpaperResult, error) {
	sub := resolveSubreddit(opts.Subreddit)
	sort := opts.Sort
	if sort == "" {
		sort = "hot"
	}

	limit := opts.Count
	if limit <= 0 {
		limit = 100
	}

	reqURL := fmt.Sprintf("%s/r/%s/%s.json?limit=%d&raw_json=1", r.baseURL, sub, sort, limit)
	if sort == "top" && opts.TimeRange != "" {
		reqURL += "&t=" + opts.TimeRange
	}

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("reddit: build request: %w", err)
	}
	req.Header.Set("User-Agent", redditUserAgent)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("reddit: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("reddit: unexpected status %d", resp.StatusCode)
	}

	var listing redditListing
	if err := json.NewDecoder(resp.Body).Decode(&listing); err != nil {
		return nil, fmt.Errorf("reddit: decode response: %w", err)
	}

	results := make([]WallpaperResult, 0, len(listing.Data.Children))
	for _, child := range listing.Data.Children {
		post := child.Data
		if post.IsVideo {
			continue
		}
		if !isImageURL(post.URL) {
			continue
		}

		var res string
		var previewURL string
		if len(post.Preview.Images) > 0 {
			src := post.Preview.Images[0].Source
			if src.Width > 0 && src.Height > 0 {
				res = fmt.Sprintf("%dx%d", src.Width, src.Height)
			}
			previewURL = src.URL
		}

		results = append(results, WallpaperResult{
			ID:         "t3_" + post.ID,
			Source:     r.Name(),
			PreviewURL: previewURL,
			FullURL:    post.URL,
			Resolution: res,
			Views:      0,
			Favorites:  post.Ups,
			Tags:       []string{sub},
		})
	}
	return results, nil
}

// resolveSubreddit returns the appropriate subreddit name from the input.
// It strips any "r/" prefix and falls back to "wallpapers" for empty input.
// If the input matches one of the default subreddits (fuzzy), that is returned.
func resolveSubreddit(input string) string {
	if input == "" {
		return defaultSubreddits[0]
	}
	// Strip leading "r/" prefix.
	s := strings.TrimPrefix(input, "r/")
	s = strings.TrimSpace(s)
	if s == "" {
		return defaultSubreddits[0]
	}
	// Exact or case-insensitive match against defaults.
	lower := strings.ToLower(s)
	for _, d := range defaultSubreddits {
		if strings.ToLower(d) == lower {
			return d
		}
	}
	return s
}

// isImageURL reports whether the URL points to a supported image type.
func isImageURL(rawURL string) bool {
	// Strip query string before checking extension.
	u := rawURL
	if idx := strings.Index(u, "?"); idx != -1 {
		u = u[:idx]
	}
	ext := strings.ToLower(path.Ext(u))
	return imageExtensions[ext]
}

// --- API response types ---

type redditListing struct {
	Data redditListingData `json:"data"`
}

type redditListingData struct {
	Children []redditChild `json:"children"`
}

type redditChild struct {
	Data redditPost `json:"data"`
}

type redditPost struct {
	ID      string        `json:"id"`
	Title   string        `json:"title"`
	URL     string        `json:"url"`
	Ups     int           `json:"ups"`
	IsVideo bool          `json:"is_video"`
	Preview redditPreview `json:"preview"`
}

type redditPreview struct {
	Images []redditPreviewImage `json:"images"`
}

type redditPreviewImage struct {
	Source redditImageSource `json:"source"`
}

type redditImageSource struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}
