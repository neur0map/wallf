package config

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type wizardStep int

const (
	wizardStepDir wizardStep = iota
	wizardStepRes
	wizardStepConfirm
)

var resolutions = []string{"2560x1440", "1920x1080", "3840x2160"}

// WizardModel is the first-run setup wizard bubbletea component.
type WizardModel struct {
	step     wizardStep
	dirInput textinput.Model
	resIndex int
	done     bool
	config   Config
}

// WizardDoneMsg is emitted when the wizard completes.
type WizardDoneMsg struct{ Config Config }

// NewWizard creates a WizardModel pre-populated with defaults.
func NewWizard() WizardModel {
	ti := textinput.New()
	ti.Placeholder = "~/Pictures/Wallpapers"
	ti.SetValue("~/Pictures/Wallpapers")
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50
	return WizardModel{step: wizardStepDir, dirInput: ti, config: Default()}
}

func (w WizardModel) Config() Config { return w.config }
func (w WizardModel) Done() bool     { return w.done }
func (w WizardModel) Init() tea.Cmd  { return textinput.Blink }

func (w WizardModel) Update(msg tea.Msg) (WizardModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return w.advance()
		case tea.KeyUp:
			if w.step == wizardStepRes && w.resIndex > 0 {
				w.resIndex--
			}
			return w, nil
		case tea.KeyDown:
			if w.step == wizardStepRes && w.resIndex < len(resolutions)-1 {
				w.resIndex++
			}
			return w, nil
		}
	}
	if w.step == wizardStepDir {
		var cmd tea.Cmd
		w.dirInput, cmd = w.dirInput.Update(msg)
		return w, cmd
	}
	return w, nil
}

func (w WizardModel) advance() (WizardModel, tea.Cmd) {
	switch w.step {
	case wizardStepDir:
		if val := w.dirInput.Value(); val != "" {
			w.config.General.DownloadDir = val
		}
		w.step = wizardStepRes
	case wizardStepRes:
		w.config.General.MinResolution = resolutions[w.resIndex]
		w.step = wizardStepConfirm
	case wizardStepConfirm:
		w.done = true
		return w, func() tea.Msg { return WizardDoneMsg{Config: w.config} }
	}
	return w, nil
}

func (w WizardModel) View() string {
	accent := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#7287fd", Dark: "#b4befe"}).Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#9ca0b0", Dark: "#6c7086"})
	panel := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.AdaptiveColor{Light: "#bcc0cc", Dark: "#45475a"}).Padding(1, 2).Width(56)

	s := accent.Render("Welcome to wallf") + "\n\n"
	switch w.step {
	case wizardStepDir:
		s += "Download folder:\n" + w.dirInput.View() + "\n"
	case wizardStepRes:
		s += "Minimum resolution:\n"
		for i, r := range resolutions {
			if i == w.resIndex {
				s += accent.Render("● "+r) + "\n"
			} else {
				s += dim.Render("○ "+r) + "\n"
			}
		}
	case wizardStepConfirm:
		s += fmt.Sprintf("Download to: %s\n", w.config.General.DownloadDir)
		s += fmt.Sprintf("Min resolution: %s\n", w.config.General.MinResolution)
		s += "\n" + accent.Render("3 sources ready: Wallhaven · Reddit · Bing") + "\n\n"
		s += dim.Render("Press enter to continue")
	}
	return panel.Render(s)
}
