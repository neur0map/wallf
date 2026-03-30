package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/neur0map/wallf/source"
)

// SearchResult carries choices from the search form.
type SearchResult struct {
	Source    string
	Query     string
	Sort      string
	Count     int
	MinRes    string
	Colors    []string
	Subreddit string
}

// SearchDoneMsg is emitted when the form is submitted.
type SearchDoneMsg struct {
	Result SearchResult
}

// totalFetchedMsg carries the server-side total for the query.
type totalFetchedMsg struct {
	total int
	err   error
}

type searchStep int

const (
	stepQuery searchStep = iota
	stepCount
	stepResolution
	stepColors
)

var (
	countPresets = []int{5, 10, 20, 50}
	resOptions   = []string{"any", "1920x1080", "2560x1440", "3840x2160"}
	colorOptions = []struct {
		name  string
		hex   string
		style lipgloss.Style
	}{
		{"white", "ffffff", lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff"))},
		{"black", "000000", lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))},
		{"red", "cc0000", lipgloss.NewStyle().Foreground(lipgloss.Color("#cc0000"))},
		{"blue", "0066cc", lipgloss.NewStyle().Foreground(lipgloss.Color("#0066cc"))},
		{"green", "00cc00", lipgloss.NewStyle().Foreground(lipgloss.Color("#00cc00"))},
		{"orange", "ff6600", lipgloss.NewStyle().Foreground(lipgloss.Color("#ff6600"))},
		{"purple", "663399", lipgloss.NewStyle().Foreground(lipgloss.Color("#663399"))},
		{"teal", "009999", lipgloss.NewStyle().Foreground(lipgloss.Color("#009999"))},
	}
)

// SearchModel is the multi-step search form.
type SearchModel struct {
	step        searchStep
	queryInput  textinput.Model
	countInput  textinput.Model // custom count input
	countIdx    int             // -1 means custom input is active
	resIdx      int
	colorSel    []bool
	cursor      int
	done        bool
	source      string
	serverTotal int // -1 = unknown, 0 = fetching
	sources     map[string]source.Source
}

func NewSearchModel(defaultSource string, count int) SearchModel {
	ti := textinput.New()
	ti.Placeholder = "describe what you're looking for..."
	ti.Focus()
	ti.CharLimit = 128
	ti.Width = 44

	ci := textinput.New()
	ci.Placeholder = "enter number..."
	ci.CharLimit = 5
	ci.Width = 10

	src := "wallhaven"
	if defaultSource != "" {
		src = defaultSource
	}

	return SearchModel{
		step:        stepQuery,
		queryInput:  ti,
		countInput:  ci,
		countIdx:    1, // default: 10
		resIdx:      2, // default: 2560x1440
		colorSel:    make([]bool, len(colorOptions)),
		source:      src,
		serverTotal: -1,
	}
}

// SetSources gives the model access to sources for total fetching.
func (m *SearchModel) SetSources(s map[string]source.Source) {
	m.sources = s
}

func (m SearchModel) Result() SearchResult {
	count := 10
	if m.countIdx >= 0 && m.countIdx < len(countPresets) {
		count = countPresets[m.countIdx]
	} else if m.countIdx == len(countPresets) {
		// custom
		if n, err := strconv.Atoi(m.countInput.Value()); err == nil && n > 0 {
			count = n
		}
	}

	res := ""
	if m.resIdx > 0 {
		res = resOptions[m.resIdx]
	}

	var colors []string
	for i, sel := range m.colorSel {
		if sel {
			colors = append(colors, colorOptions[i].hex)
		}
	}

	return SearchResult{
		Source: m.source,
		Query:  strings.TrimSpace(m.queryInput.Value()),
		Sort:   "relevance",
		Count:  count,
		MinRes: res,
		Colors: colors,
	}
}

func (m SearchModel) Done() bool { return m.done }

func (m SearchModel) Init() tea.Cmd { return textinput.Blink }

func (m SearchModel) Update(msg tea.Msg) (SearchModel, tea.Cmd) {
	switch msg := msg.(type) {
	case totalFetchedMsg:
		if msg.err == nil {
			m.serverTotal = msg.total
		} else {
			m.serverTotal = -1
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return m.advance()
		case tea.KeyEsc:
			if m.step > stepQuery {
				m.step--
				m.cursor = 0
				if m.step == stepQuery {
					m.queryInput.Focus()
					m.countInput.Blur()
				}
				return m, nil
			}
		case tea.KeyUp, tea.KeyShiftTab:
			if m.step != stepQuery && !m.isCustomCountActive() {
				m.moveCursor(-1)
			}
			return m, nil
		case tea.KeyDown, tea.KeyTab:
			if m.step != stepQuery && !m.isCustomCountActive() {
				m.moveCursor(1)
			}
			return m, nil
		}

		switch msg.String() {
		case "k":
			if m.step != stepQuery && !m.isCustomCountActive() {
				m.moveCursor(-1)
				return m, nil
			}
		case "j":
			if m.step != stepQuery && !m.isCustomCountActive() {
				m.moveCursor(1)
				return m, nil
			}
		case " ":
			if m.step == stepColors {
				m.colorSel[m.cursor] = !m.colorSel[m.cursor]
				return m, nil
			}
		}
	}

	// Update active text input
	if m.step == stepQuery {
		var cmd tea.Cmd
		m.queryInput, cmd = m.queryInput.Update(msg)
		return m, cmd
	}
	if m.isCustomCountActive() {
		var cmd tea.Cmd
		m.countInput, cmd = m.countInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m SearchModel) isCustomCountActive() bool {
	return m.step == stepCount && m.cursor == len(countPresets)
}

func (m *SearchModel) moveCursor(delta int) {
	max := m.maxCursor()
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = max - 1
	}
	if m.cursor >= max {
		m.cursor = 0
	}
	// Toggle custom input focus
	if m.step == stepCount {
		if m.cursor == len(countPresets) {
			m.countInput.Focus()
		} else {
			m.countInput.Blur()
		}
	}
}

func (m SearchModel) maxCursor() int {
	switch m.step {
	case stepCount:
		return len(countPresets) + 1 // +1 for custom
	case stepResolution:
		return len(resOptions)
	case stepColors:
		return len(colorOptions)
	}
	return 1
}

func (m SearchModel) advance() (SearchModel, tea.Cmd) {
	switch m.step {
	case stepQuery:
		if m.queryInput.Value() == "" {
			return m, nil
		}
		m.queryInput.Blur()
		m.step = stepCount
		m.cursor = m.countIdx
		m.serverTotal = 0 // fetching
		// Kick off total count fetch
		return m, m.fetchTotal()
	case stepCount:
		if m.isCustomCountActive() {
			if _, err := strconv.Atoi(m.countInput.Value()); err != nil || m.countInput.Value() == "" {
				return m, nil // invalid custom number
			}
		}
		m.countIdx = m.cursor
		m.countInput.Blur()
		m.step = stepResolution
		m.cursor = m.resIdx
	case stepResolution:
		m.resIdx = m.cursor
		m.step = stepColors
		m.cursor = 0
	case stepColors:
		m.done = true
		result := m.Result()
		return m, func() tea.Msg { return SearchDoneMsg{Result: result} }
	}
	return m, nil
}

func (m SearchModel) fetchTotal() tea.Cmd {
	src, ok := m.sources[m.source]
	if !ok {
		return nil
	}
	query := strings.TrimSpace(m.queryInput.Value())
	return func() tea.Msg {
		total, err := src.SearchTotal(source.SearchOpts{Query: query})
		return totalFetchedMsg{total: total, err: err}
	}
}

func (m SearchModel) View() string {
	var sb strings.Builder
	highlight := lipgloss.NewStyle().Foreground(colorAccent).Bold(true)

	sb.WriteString(styleTitle.Render("  wallf") + "\n")
	sb.WriteString(styleDim.Render("  ─────────────────────────────────────────") + "\n\n")

	// Step indicator
	steps := []string{"search", "count", "size", "color"}
	var stepLine []string
	for i, s := range steps {
		if searchStep(i) == m.step {
			stepLine = append(stepLine, highlight.Render(s))
		} else if searchStep(i) < m.step {
			stepLine = append(stepLine, styleSuccess.Render(s))
		} else {
			stepLine = append(stepLine, styleDim.Render(s))
		}
	}
	sb.WriteString("  " + strings.Join(stepLine, styleDim.Render(" → ")) + "\n\n")

	switch m.step {
	case stepQuery:
		sb.WriteString(styleLabel.Render("  What are you looking for?") + "\n\n")
		sb.WriteString("  " + m.queryInput.View() + "\n\n")
		sb.WriteString(styleDim.Render("  enter to continue"))

	case stepCount:
		sb.WriteString(styleLabel.Render("  How many wallpapers?"))
		// Show server total if available
		switch {
		case m.serverTotal > 0:
			totalStr := formatNumber(m.serverTotal)
			sb.WriteString(styleDim.Render("  (") + lipgloss.NewStyle().Foreground(colorMauve).Render(totalStr+" available") + styleDim.Render(")"))
		case m.serverTotal == 0:
			sb.WriteString(styleDim.Render("  (checking...)"))
		}
		sb.WriteString("\n\n")
		for i, n := range countPresets {
			label := strconv.Itoa(n)
			if i == m.cursor {
				sb.WriteString(highlight.Render("  ● " + label) + "\n")
			} else {
				sb.WriteString(styleDim.Render("  ○ " + label) + "\n")
			}
		}
		// Custom option
		if m.cursor == len(countPresets) {
			sb.WriteString(highlight.Render("  ● custom: ") + m.countInput.View() + "\n")
		} else {
			sb.WriteString(styleDim.Render("  ○ custom") + "\n")
		}
		sb.WriteString("\n" + styleDim.Render("  ↑/↓ select · enter to continue · esc back"))

	case stepResolution:
		sb.WriteString(styleLabel.Render("  Minimum resolution?") + "\n\n")
		for i, opt := range resOptions {
			label := opt
			if opt == "any" {
				label = "any size"
			}
			if i == m.cursor {
				sb.WriteString(highlight.Render("  ● " + label) + "\n")
			} else {
				sb.WriteString(styleDim.Render("  ○ " + label) + "\n")
			}
		}
		sb.WriteString("\n" + styleDim.Render("  ↑/↓ select · enter to continue · esc back"))

	case stepColors:
		sb.WriteString(styleLabel.Render("  Filter by dominant color?") + "\n")
		sb.WriteString(styleDim.Render("  (optional — space to toggle, enter to search)") + "\n\n")
		for i, opt := range colorOptions {
			check := "○"
			if m.colorSel[i] {
				check = styleSuccess.Render("✓")
			}
			name := opt.style.Render(opt.name)
			pointer := "  "
			if i == m.cursor {
				pointer = highlight.Render("▸ ")
			}
			sb.WriteString("  " + pointer + check + " " + name + "\n")
		}
		selected := m.selectedColorNames()
		if len(selected) > 0 {
			sb.WriteString("\n" + styleDim.Render("  selected: ") + styleLabel.Render(strings.Join(selected, ", ")) + "\n")
		}
		sb.WriteString("\n" + styleDim.Render("  space toggle · enter to search · esc back"))
	}

	return stylePanel.Render(sb.String())
}

func (m SearchModel) selectedColorNames() []string {
	var names []string
	for i, sel := range m.colorSel {
		if sel {
			names = append(names, colorOptions[i].name)
		}
	}
	return names
}

func formatNumber(n int) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return strconv.Itoa(n)
}
