package tui

import (
	"strings"
	"testing"

	"github.com/neur0map/wallf/source"
)

func makeSummaryRecords() []DownloadRecord {
	result := func(id string) source.WallpaperResult {
		return source.WallpaperResult{
			ID:      id,
			Source:  "wallhaven",
			FullURL: "https://example.com/" + id + ".jpg",
		}
	}
	return []DownloadRecord{
		{Result: result("aaa"), Downloaded: true, Path: "/tmp/wallhaven-aaa.jpg"},
		{Result: result("bbb"), Downloaded: true, Path: "/tmp/wallhaven-bbb.jpg"},
		{Result: result("ccc"), Skipped: true},
	}
}

func TestSummaryView(t *testing.T) {
	m := NewSummaryModel(makeSummaryRecords(), "/tmp")
	view := m.View()

	// Should contain the downloaded filename
	if !strings.Contains(view, "wallhaven-aaa.jpg") {
		t.Errorf("expected filename 'wallhaven-aaa.jpg' in view:\n%s", view)
	}
	// Should contain "skipped"
	if !strings.Contains(view, "skipped") {
		t.Errorf("expected 'skipped' in view:\n%s", view)
	}
	// Should contain counts
	if !strings.Contains(view, "2") || !strings.Contains(view, "wallpapers saved") {
		t.Errorf("expected '2 wallpapers saved' in view:\n%s", view)
	}
	if !strings.Contains(view, "skipped") {
		t.Errorf("expected 'skipped' count in view:\n%s", view)
	}
}

func TestSummaryCounts(t *testing.T) {
	records := []DownloadRecord{
		{Downloaded: true},
		{Downloaded: true},
		{Skipped: true},
		{Error: "network timeout"},
	}
	m := NewSummaryModel(records, "/tmp")
	dl, sk, er := m.Counts()
	if dl != 2 {
		t.Errorf("expected 2 downloaded, got %d", dl)
	}
	if sk != 1 {
		t.Errorf("expected 1 skipped, got %d", sk)
	}
	if er != 1 {
		t.Errorf("expected 1 error, got %d", er)
	}
}

func TestSummaryEmpty(t *testing.T) {
	m := NewSummaryModel(nil, "")
	view := m.View()
	if !strings.Contains(view, "0") || !strings.Contains(view, "wallpapers saved") {
		t.Errorf("expected '0 wallpapers saved' in empty summary view:\n%s", view)
	}
}
