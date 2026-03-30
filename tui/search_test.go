package tui

import (
	"testing"
)

func TestSearchModelInitialState(t *testing.T) {
	m := NewSearchModel("wallhaven", 10)
	if m.step != stepQuery {
		t.Errorf("expected stepQuery, got %v", m.step)
	}
	if m.source != "wallhaven" {
		t.Errorf("expected source=wallhaven, got %s", m.source)
	}
}

func TestSearchResult(t *testing.T) {
	m := NewSearchModel("wallhaven", 10)
	m.queryInput.SetValue("cyberpunk")
	r := m.Result()
	if r.Query != "cyberpunk" {
		t.Errorf("Query = %q, want cyberpunk", r.Query)
	}
	if r.Count != 10 {
		t.Errorf("Count = %d, want 10", r.Count)
	}
}

func TestSearchColorSelection(t *testing.T) {
	m := NewSearchModel("wallhaven", 10)
	m.step = stepColors
	m.colorSel[0] = true // white
	m.colorSel[1] = true // black
	r := m.Result()
	if len(r.Colors) != 2 {
		t.Errorf("expected 2 colors, got %d", len(r.Colors))
	}
	if r.Colors[0] != "ffffff" {
		t.Errorf("expected ffffff, got %s", r.Colors[0])
	}
}

func TestSearchCustomCount(t *testing.T) {
	m := NewSearchModel("wallhaven", 10)
	m.countIdx = len(countPresets) // custom
	m.countInput.SetValue("42")
	r := m.Result()
	if r.Count != 42 {
		t.Errorf("Count = %d, want 42", r.Count)
	}
}

func TestSearchResolution(t *testing.T) {
	m := NewSearchModel("wallhaven", 10)
	m.resIdx = 2 // 2560x1440
	r := m.Result()
	if r.MinRes != "2560x1440" {
		t.Errorf("MinRes = %s, want 2560x1440", r.MinRes)
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{500, "500"},
		{1500, "1.5k"},
		{25000, "25.0k"},
		{1500000, "1.5M"},
	}
	for _, tt := range tests {
		got := formatNumber(tt.n)
		if got != tt.want {
			t.Errorf("formatNumber(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}
