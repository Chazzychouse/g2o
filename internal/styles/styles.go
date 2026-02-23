package styles

import "github.com/charmbracelet/lipgloss"

var (
	Title    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "#0066cc", Dark: "#6699ff"})
	Label    = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#555555", Dark: "#999999"})
	Value    = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#111111", Dark: "#eeeeee"})
	Error    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "#cc2200", Dark: "#ff6655"})
	Success  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#226600", Dark: "#55cc44"})
	Prompt   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "#7700aa", Dark: "#cc77ff"})
	HelpCmd  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "#006677", Dark: "#44ccdd"})
	HelpDesc = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#444444", Dark: "#aaaaaa"})
	Banner   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "#880077", Dark: "#dd88cc"})
)
