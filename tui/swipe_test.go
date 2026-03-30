package tui

import (
	"testing"

	"github.com/neur0map/wallf/source"
)

func makeResults(n int) []source.WallpaperResult {
	results := make([]source.WallpaperResult, n)
	for i := 0; i < n; i++ {
		results[i] = source.WallpaperResult{
			ID:         "id" + string(rune('0'+i+1)),
			Source:     "wallhaven",
			Resolution: "3840x2160",
			Views:      100,
			Favorites:  10,
			FullURL:    "https://example.com/wall.jpg",
		}
	}
	return results
}

func TestSwipeInitialState(t *testing.T) {
	m := NewSwipeModel(makeResults(3))
	if m.Total() != 3 {
		t.Errorf("expected Total()=3, got %d", m.Total())
	}
}

func TestSwipeRecords(t *testing.T) {
	m := NewSwipeModel(makeResults(3))
	m.records[0].Downloaded = true
	m.records[0].Path = "/tmp/a.jpg"
	m.records[1].Error = "failed"

	recs := m.Records()
	if len(recs) != 3 {
		t.Fatalf("expected 3 records, got %d", len(recs))
	}
	if !recs[0].Downloaded {
		t.Error("record[0]: expected Downloaded=true")
	}
	if recs[1].Error == "" {
		t.Error("record[1]: expected error")
	}
}

func TestSwipeRemaining(t *testing.T) {
	m := NewSwipeModel(makeResults(3))
	m.records[0].Downloaded = true
	remaining := m.Remaining()
	if len(remaining) != 2 {
		t.Errorf("expected 2 remaining, got %d", len(remaining))
	}
}

func TestSwipeDownloadProgress(t *testing.T) {
	m := NewSwipeModel(makeResults(3))

	// Simulate first item completing
	m, _ = m.Update(DownloadProgressMsg{Index: 0, Path: "/tmp/wall0.jpg"})
	if !m.records[0].Downloaded {
		t.Error("expected record[0] downloaded after progress msg")
	}
	if m.active != 1 {
		t.Errorf("expected active=1 after first download, got %d", m.active)
	}
}

func TestSwipeViewEmpty(t *testing.T) {
	m := NewSwipeModel(nil)
	v := m.View()
	if v == "" {
		t.Error("empty results should still produce a view")
	}
}

func TestSwipeCounts(t *testing.T) {
	m := NewSwipeModel(makeResults(5))
	m.records[0].Downloaded = true
	m.records[1].Downloaded = true
	m.records[2].Error = "fail"

	if m.downloadedCount() != 2 {
		t.Errorf("expected 2 downloaded, got %d", m.downloadedCount())
	}
	if m.completedCount() != 3 {
		t.Errorf("expected 3 completed, got %d", m.completedCount())
	}
}
