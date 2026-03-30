package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestAppInitialState(t *testing.T) {
	app := NewApp(AppOpts{NeedWizard: false})
	if app.state != stateSearch {
		t.Errorf("expected stateSearch, got %v", app.state)
	}
}

func TestAppInitialStateWithWizard(t *testing.T) {
	app := NewApp(AppOpts{NeedWizard: true})
	if app.state != stateWizard {
		t.Errorf("expected stateWizard, got %v", app.state)
	}
}

func TestAppQuitFromSearch(t *testing.T) {
	app := NewApp(AppOpts{NeedWizard: false})
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("expected a quit cmd, got nil")
	}
	msg := cmd()
	if msg != tea.Quit() {
		t.Errorf("expected tea.Quit message, got %v", msg)
	}
}

func TestAppStateTransition(t *testing.T) {
	app := NewApp(AppOpts{NeedWizard: false})
	app.state = stateDownload
	if app.state != stateDownload {
		t.Errorf("expected stateDownload, got %v", app.state)
	}
}
