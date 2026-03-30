package source

import (
	"fmt"
	"path"
	"strconv"
	"strings"
)

// Source is the interface that all wallpaper sources must implement.
type Source interface {
	// Search fetches wallpapers matching opts.
	Search(opts SearchOpts) ([]WallpaperResult, error)
	// SearchTotal returns how many results are available server-side for the
	// given opts without downloading all of them. Returns -1 if unknown.
	SearchTotal(opts SearchOpts) (int, error)
	// Name returns the unique slug for this source (e.g. "wallhaven").
	Name() string
	// NeedsKey reports whether this source requires an API key.
	NeedsKey() bool
	// Categories returns the list of categories supported by this source.
	Categories() []string
}

// SearchOpts carries all parameters for a search request.
type SearchOpts struct {
	Query      string
	Sort       string
	Count      int
	Page       int
	MinRes     string
	Categories []string
	TimeRange  string
	Colors     []string // hex colors e.g. "000000", "ffffff"
	Ratios     string   // "landscape", "portrait", "16x9", etc.
	// Subreddit is only meaningful for the reddit source.
	Subreddit string
}

// WallpaperResult is a single wallpaper returned by a Source.
type WallpaperResult struct {
	ID         string
	Source     string
	PreviewURL string
	FullURL    string
	Resolution string
	Views      int
	Favorites  int
	Tags       []string
}

// Filename returns a deterministic local filename for the wallpaper in the
// form "{source}-{id}.{ext}". Query parameters are stripped from the
// extension so that URLs like "…/image.png?s=abc" produce ".png".
func (r WallpaperResult) Filename() string {
	// Derive extension from FullURL, stripping any query string first.
	rawURL := r.FullURL
	if idx := strings.Index(rawURL, "?"); idx != -1 {
		rawURL = rawURL[:idx]
	}
	ext := path.Ext(rawURL) // includes the leading dot, e.g. ".jpg"
	return fmt.Sprintf("%s-%s%s", r.Source, r.ID, ext)
}

// ParseResolution parses a "WxH" resolution string and returns the width,
// height and whether the parse succeeded.
func ParseResolution(res string) (int, int, bool) {
	parts := strings.SplitN(res, "x", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return 0, 0, false
	}
	w, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, false
	}
	h, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, false
	}
	return w, h, true
}

// MeetsMinResolution reports whether the wallpaper's resolution is at least
// minRes. If the wallpaper's own resolution is unknown (empty / unparseable),
// or if minRes is empty / unparseable, the function returns true so that the
// wallpaper is not filtered out.
func (r WallpaperResult) MeetsMinResolution(minRes string) bool {
	rw, rh, rOK := ParseResolution(r.Resolution)
	mw, mh, mOK := ParseResolution(minRes)

	// Unknown resolution or unknown minimum → don't filter.
	if !rOK || !mOK {
		return true
	}
	return rw >= mw && rh >= mh
}
