package tui

import "github.com/charmbracelet/lipgloss"

// Catppuccin-inspired adaptive color palette
var (
	colorAccent  = lipgloss.AdaptiveColor{Light: "#7287fd", Dark: "#b4befe"}
	colorMauve   = lipgloss.AdaptiveColor{Light: "#8839ef", Dark: "#cba6f7"}
	colorText    = lipgloss.AdaptiveColor{Light: "#5c5f77", Dark: "#a6adc8"}
	colorDim     = lipgloss.AdaptiveColor{Light: "#9ca0b0", Dark: "#6c7086"}
	colorBorder  = lipgloss.AdaptiveColor{Light: "#bcc0cc", Dark: "#45475a"}
	colorSuccess = lipgloss.AdaptiveColor{Light: "#40a02b", Dark: "#a6e3a1"}
	colorError   = lipgloss.AdaptiveColor{Light: "#d20f39", Dark: "#f38ba8"}
	colorBase    = lipgloss.AdaptiveColor{Light: "#eff1f5", Dark: "#1e1e2e"}
)

var (
	stylePanel     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colorBorder).Padding(1, 2)
	styleTitle     = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	styleSubtitle  = lipgloss.NewStyle().Foreground(colorDim)
	styleLabel     = lipgloss.NewStyle().Foreground(colorText)
	styleSelected  = lipgloss.NewStyle().Foreground(colorMauve).Bold(true)
	styleStatusBar = lipgloss.NewStyle().Foreground(colorDim)
	styleSuccess   = lipgloss.NewStyle().Foreground(colorSuccess)
	styleError     = lipgloss.NewStyle().Foreground(colorError)
	styleDim       = lipgloss.NewStyle().Foreground(colorDim)
	styleDotFilled = lipgloss.NewStyle().Foreground(colorAccent)
	styleDotEmpty  = lipgloss.NewStyle().Foreground(colorDim)
)

func ProgressDots(current, total int) string {
	s := ""
	for i := 0; i < total; i++ {
		if i < current {
			s += styleDotFilled.Render("●")
		} else {
			s += styleDotEmpty.Render("○")
		}
	}
	return s
}
