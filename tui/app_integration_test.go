package tui

import (
	"fmt"
	"testing"

	"github.com/neur0map/wallf/source"
)

func TestSearchDoneMsgTriggersSearch(t *testing.T) {
	sources := map[string]source.Source{
		"wallhaven": source.NewWallhaven(""),
	}
	app := NewApp(AppOpts{
		Sources: sources,
		Count:   10,
	})
	app.state = stateSearch

	sr := SearchResult{Source: "wallhaven", Query: "test", Sort: "relevance", Count: 10}
	result, cmd := app.Update(SearchDoneMsg{Result: sr})
	a := result.(App)

	if !a.loading {
		t.Error("expected loading=true after SearchDoneMsg")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd for search")
	}
}

func TestSearchResultsMsgTransitionsToDownload(t *testing.T) {
	app := NewApp(AppOpts{Count: 10})
	app.state = stateSearch

	results := makeResults(3)
	result, _ := app.Update(searchResultsMsg{results: results})
	a := result.(App)

	if a.state != stateDownload {
		t.Errorf("expected stateDownload, got %v", a.state)
	}
	if a.swipe.Total() != 3 {
		t.Errorf("expected swipe.Total()=3, got %d", a.swipe.Total())
	}
	if a.loading {
		t.Error("expected loading=false after searchResultsMsg")
	}
}

func TestSearchErrorSurfaced(t *testing.T) {
	app := NewApp(AppOpts{Count: 10})
	app.state = stateSearch

	result, _ := app.Update(searchResultsMsg{err: fmt.Errorf("network timeout")})
	a := result.(App)

	if a.state != stateDownload {
		t.Errorf("expected stateDownload on error, got %v", a.state)
	}
	if a.swipe.status == "" {
		t.Error("expected non-empty status on error")
	}
}

func TestSwipeDoneNoResultsReturnsToSearch(t *testing.T) {
	app := NewApp(AppOpts{Count: 10})
	app.state = stateDownload
	app.swipe = NewSwipeModel(nil)

	result, _ := app.Update(SwipeDoneMsg{})
	a := result.(App)

	if a.state != stateSearch {
		t.Errorf("expected stateSearch when no results, got %v", a.state)
	}
}

func TestSwipeDoneWithResultsGoesToSummary(t *testing.T) {
	results := makeResults(3)
	app := NewApp(AppOpts{Count: 10})
	app.state = stateDownload
	app.swipe = NewSwipeModel(results)

	records := app.swipe.Records()
	result, _ := app.Update(SwipeDoneMsg{Records: records})
	a := result.(App)

	if a.state != stateSummary {
		t.Errorf("expected stateSummary, got %v", a.state)
	}
}

func TestDownloadProgressUpdatesRecords(t *testing.T) {
	results := makeResults(3)
	app := NewApp(AppOpts{Count: 10})
	app.state = stateDownload
	app.swipe = NewSwipeModel(results)

	// Simulate first download completing
	result, _ := app.Update(DownloadProgressMsg{Index: 0, Path: "/tmp/wall0.jpg"})
	a := result.(App)

	if !a.swipe.records[0].Downloaded {
		t.Error("expected record[0] downloaded")
	}
	if a.swipe.records[0].Path != "/tmp/wall0.jpg" {
		t.Errorf("expected path=/tmp/wall0.jpg, got %s", a.swipe.records[0].Path)
	}

	// Simulate error on second
	result2, _ := a.Update(DownloadProgressMsg{Index: 1, Err: fmt.Errorf("network error")})
	a2 := result2.(App)

	if a2.swipe.records[1].Error == "" {
		t.Error("expected record[1] to have error")
	}
}
