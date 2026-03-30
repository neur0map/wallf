package source_test

import (
	"testing"

	"github.com/neur0map/wallf/source"
)

func TestWallpaperResultFilename(t *testing.T) {
	r := source.WallpaperResult{
		ID:      "abc123",
		Source:  "wallhaven",
		FullURL: "https://w.wallhaven.cc/full/abc/wallhaven-abc123.jpg",
	}
	got := r.Filename()
	want := "wallhaven-abc123.jpg"
	if got != want {
		t.Errorf("Filename() = %q, want %q", got, want)
	}
}

func TestWallpaperResultFilenameReddit(t *testing.T) {
	r := source.WallpaperResult{
		ID:      "t3_xyz789",
		Source:  "reddit",
		FullURL: "https://i.redd.it/someimage.png?s=abc123",
	}
	got := r.Filename()
	want := "reddit-t3_xyz789.png"
	if got != want {
		t.Errorf("Filename() = %q, want %q", got, want)
	}
}

func TestParseResolution(t *testing.T) {
	tests := []struct {
		input  string
		wantW  int
		wantH  int
		wantOK bool
	}{
		{"3840x2160", 3840, 2160, true},
		{"", 0, 0, false},
		{"invalid", 0, 0, false},
		{"1920x", 0, 0, false},
		{"x1080", 0, 0, false},
		{"abcxdef", 0, 0, false},
	}

	for _, tt := range tests {
		w, h, ok := source.ParseResolution(tt.input)
		if ok != tt.wantOK || w != tt.wantW || h != tt.wantH {
			t.Errorf("ParseResolution(%q) = (%d, %d, %v), want (%d, %d, %v)",
				tt.input, w, h, ok, tt.wantW, tt.wantH, tt.wantOK)
		}
	}
}

func TestMeetsMinResolution(t *testing.T) {
	tests := []struct {
		result source.WallpaperResult
		minRes string
		want   bool
	}{
		{
			result: source.WallpaperResult{Resolution: "3840x2160"},
			minRes: "2560x1440",
			want:   true,
		},
		{
			result: source.WallpaperResult{Resolution: "3840x2160"},
			minRes: "7680x4320",
			want:   false,
		},
		{
			// Unknown resolution passes
			result: source.WallpaperResult{Resolution: ""},
			minRes: "2560x1440",
			want:   true,
		},
		{
			// Invalid minRes passes
			result: source.WallpaperResult{Resolution: "1920x1080"},
			minRes: "",
			want:   true,
		},
	}

	for _, tt := range tests {
		got := tt.result.MeetsMinResolution(tt.minRes)
		if got != tt.want {
			t.Errorf("MeetsMinResolution(%q) on %q = %v, want %v",
				tt.minRes, tt.result.Resolution, got, tt.want)
		}
	}
}
