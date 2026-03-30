package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/neur0map/wallf/source"
)

// --- Messages ---

// SwipeDoneMsg signals completion of the download view.
type SwipeDoneMsg struct {
	Records []DownloadRecord
}

// StartDownloadMsg triggers download of a single result.
type StartDownloadMsg struct {
	Index  int
	Result source.WallpaperResult
}

// DownloadProgressMsg is sent as each item finishes.
type DownloadProgressMsg struct {
	Index int
	Path  string
	Err   error
}

// tickMsg drives animation.
type tickMsg time.Time

// --- Model ---

// SwipeModel is the bulk download view with animated progress.
type SwipeModel struct {
	results  []source.WallpaperResult
	records  []DownloadRecord
	status   string
	active   int  // index currently downloading
	done     bool // all downloads complete
	spinner  spinner.Model
	progress progress.Model
	width    int
}

func NewSwipeModel(results []source.WallpaperResult) SwipeModel {
	records := make([]DownloadRecord, len(results))
	for i, r := range results {
		records[i] = DownloadRecord{Result: r}
	}

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = lipgloss.NewStyle().Foreground(colorAccent)

	prog := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)

	return SwipeModel{
		results:  results,
		records:  records,
		active:   0,
		spinner:  sp,
		progress: prog,
	}
}

func (m SwipeModel) Total() int { return len(m.results) }

func (m SwipeModel) Records() []DownloadRecord {
	out := make([]DownloadRecord, len(m.records))
	copy(out, m.records)
	return out
}

func (m SwipeModel) Current() source.WallpaperResult {
	if m.active < len(m.results) {
		return m.results[m.active]
	}
	return source.WallpaperResult{}
}

// Remaining returns unprocessed results.
func (m SwipeModel) Remaining() []source.WallpaperResult {
	var out []source.WallpaperResult
	for i := range m.results {
		if !m.records[i].Downloaded && !m.records[i].Skipped && m.records[i].Error == "" {
			out = append(out, m.results[i])
		}
	}
	return out
}

func (m SwipeModel) completedCount() int {
	n := 0
	for _, r := range m.records {
		if r.Downloaded || r.Skipped || r.Error != "" {
			n++
		}
	}
	return n
}

func (m SwipeModel) downloadedCount() int {
	n := 0
	for _, r := range m.records {
		if r.Downloaded {
			n++
		}
	}
	return n
}

// Init starts the spinner and kicks off the first download.
func (m SwipeModel) Init() tea.Cmd {
	if len(m.results) == 0 {
		return nil
	}
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			return StartDownloadMsg{Index: 0, Result: m.results[0]}
		},
	)
}

func (m SwipeModel) Update(msg tea.Msg) (SwipeModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			records := m.Records()
			return m, func() tea.Msg { return SwipeDoneMsg{Records: records} }
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		model, cmd := m.progress.Update(msg)
		m.progress = model.(progress.Model)
		return m, cmd

	case DownloadProgressMsg:
		if msg.Index < len(m.records) {
			if msg.Err != nil {
				m.records[msg.Index].Error = msg.Err.Error()
			} else {
				m.records[msg.Index].Downloaded = true
				m.records[msg.Index].Path = msg.Path
			}
		}

		// Move to next
		next := msg.Index + 1
		m.active = next

		if next >= len(m.results) {
			// All done
			m.done = true
			m.status = styleSuccess.Render(fmt.Sprintf("  %d wallpapers downloaded", m.downloadedCount()))
			return m, m.progress.SetPercent(1.0)
		}

		pct := float64(next) / float64(len(m.results))
		return m, tea.Batch(
			m.progress.SetPercent(pct),
			func() tea.Msg {
				return StartDownloadMsg{Index: next, Result: m.results[next]}
			},
		)
	}
	return m, nil
}

func (m SwipeModel) View() string {
	if len(m.results) == 0 {
		s := styleDim.Render("  No results found. Press q to go back.")
		return stylePanel.Render(s)
	}

	var sb strings.Builder
	total := len(m.results)
	completed := m.completedCount()
	downloaded := m.downloadedCount()

	// Header
	sb.WriteString(styleTitle.Render("  wallf"))
	sb.WriteString(styleDim.Render("  fetching wallpapers"))
	sb.WriteString("\n\n")

	// Progress bar
	sb.WriteString("  " + m.progress.View() + "\n")
	sb.WriteString(styleDim.Render(fmt.Sprintf("  %d / %d", completed, total)))
	sb.WriteString("\n\n")

	// Live feed — show last ~8 completed items + current
	sb.WriteString(styleDim.Render("  ─────────────────────────────────────") + "\n")

	start := 0
	if completed > 8 {
		start = completed - 8
	}

	for i := start; i < len(m.records); i++ {
		r := m.records[i]
		filename := r.Result.Filename()
		res := r.Result.Resolution

		switch {
		case r.Downloaded:
			sb.WriteString(styleSuccess.Render("  ✓ "))
			sb.WriteString(lipgloss.NewStyle().Foreground(colorText).Render(filename))
			sb.WriteString(styleDim.Render("  " + res))
			sb.WriteString("\n")
		case r.Error != "":
			sb.WriteString(styleError.Render("  ✗ "))
			sb.WriteString(styleDim.Render(filename + "  " + r.Error))
			sb.WriteString("\n")
		case r.Skipped:
			sb.WriteString(styleDim.Render("  – " + filename + "  skipped"))
			sb.WriteString("\n")
		case i == m.active && !m.done:
			// Currently downloading
			sb.WriteString("  " + m.spinner.View() + " ")
			sb.WriteString(lipgloss.NewStyle().Foreground(colorAccent).Render(filename))
			sb.WriteString(styleDim.Render("  " + res))
			sb.WriteString("\n")
		default:
			// Pending
			if i <= m.active+2 {
				sb.WriteString(styleDim.Render("  · " + filename))
				sb.WriteString("\n")
			}
		}

		// Limit visible lines
		if i > start+10 && i < m.active-1 {
			continue
		}
	}

	sb.WriteString(styleDim.Render("  ─────────────────────────────────────") + "\n\n")

	// Status / stats
	if m.done {
		sb.WriteString(m.status + "\n\n")
		failed := 0
		for _, r := range m.records {
			if r.Error != "" {
				failed++
			}
		}
		sb.WriteString(styleDim.Render(fmt.Sprintf("  %d downloaded", downloaded)))
		if failed > 0 {
			sb.WriteString(styleError.Render(fmt.Sprintf("  %d failed", failed)))
		}
		sb.WriteString("\n\n")
		sb.WriteString(styleDim.Render("  press q to continue"))
	} else {
		cur := m.Current()
		sb.WriteString("  " + m.spinner.View() + " ")
		sb.WriteString(styleDim.Render("downloading "))
		sb.WriteString(lipgloss.NewStyle().Foreground(colorMauve).Bold(true).Render(filepath.Base(cur.FullURL)))
		sb.WriteString("\n\n")
		sb.WriteString(styleDim.Render("  press q to stop"))
	}

	return stylePanel.Render(sb.String())
}
