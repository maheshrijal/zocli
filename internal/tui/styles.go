package tui

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	Title          lipgloss.Style
	Tab            lipgloss.Style
	ActiveTab      lipgloss.Style
	TableContainer lipgloss.Style
	Footer         lipgloss.Style
	StatsBox       lipgloss.Style
}

func DefaultStyles() Styles {
	return Styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1),
		Tab: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(0, 1),
		ActiveTab: lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(lipgloss.Color("205")).
			Padding(0, 1),
		TableContainer: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")),
		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(1, 0),
		StatsBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2).
			Margin(1),
	}
}
