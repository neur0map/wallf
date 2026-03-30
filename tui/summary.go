package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SummaryModel displays a final report.
type SummaryModel struct {
	records []DownloadRecord
	dir     string
}

func NewSummaryModel(records []DownloadRecord, dir string) SummaryModel {
	return SummaryModel{records: records, dir: dir}
}

func (m SummaryModel) Counts() (int, int, int) {
	var dl, sk, er int
	for _, r := range m.records {
		switch {
		case r.Downloaded:
			dl++
		case r.Skipped:
			sk++
		case r.Error != "":
			er++
		}
	}
	return dl, sk, er
}

func (m SummaryModel) Init() tea.Cmd { return nil }

func (m SummaryModel) Update(msg tea.Msg) (SummaryModel, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "q", "enter":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m SummaryModel) View() string {
	var sb strings.Builder

	dl, sk, er := m.Counts()
	total := dl + sk + er

	// Header with large count
	sb.WriteString(styleTitle.Render("  wallf") + styleDim.Render("  complete") + "\n")
	sb.WriteString(styleDim.Render("  ─────────────────────────────────────") + "\n\n")

	// Big stats
	countStyle := lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	sb.WriteString("  " + countStyle.Render(fmt.Sprintf("%d", dl)) + styleDim.Render(" wallpapers saved"))
	if er > 0 {
		sb.WriteString("  " + styleError.Render(fmt.Sprintf("%d failed", er)))
	}
	if sk > 0 {
		sb.WriteString("  " + styleDim.Render(fmt.Sprintf("%d skipped", sk)))
	}
	sb.WriteString("\n\n")

	// File list (show up to 12)
	sb.WriteString(styleDim.Render("  ─────────────────────────────────────") + "\n")
	shown := 0
	for _, r := range m.records {
		if shown >= 12 {
			remaining := total - shown
			if remaining > 0 {
				sb.WriteString(styleDim.Render(fmt.Sprintf("  … and %d more", remaining)) + "\n")
			}
			break
		}
		name := r.Result.Filename()
		if r.Path != "" {
			name = filepath.Base(r.Path)
		}
		res := r.Result.Resolution

		switch {
		case r.Downloaded:
			sb.WriteString(styleSuccess.Render("  ✓ "))
			sb.WriteString(lipgloss.NewStyle().Foreground(colorText).Render(name))
			sb.WriteString(styleDim.Render("  " + res) + "\n")
		case r.Error != "":
			sb.WriteString(styleError.Render("  ✗ "))
			sb.WriteString(styleDim.Render(name + "  " + r.Error) + "\n")
		case r.Skipped:
			sb.WriteString(styleDim.Render("  – " + name + "  skipped") + "\n")
		}
		shown++
	}
	sb.WriteString(styleDim.Render("  ─────────────────────────────────────") + "\n\n")

	// Save location
	if m.dir != "" {
		sb.WriteString(styleDim.Render("  saved to ") + lipgloss.NewStyle().Foreground(colorMauve).Render(m.dir) + "\n\n")
	}

	sb.WriteString(styleDim.Render("  q / enter to exit"))

	return stylePanel.Render(sb.String())
}
