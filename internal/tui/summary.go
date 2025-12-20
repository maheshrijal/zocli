package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) viewSummary() string {
	// Top Row: Total Spent & Total Orders
	totalBox := m.styles.StatsBox.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			m.styles.Title.Copy().Background(lipgloss.Color("62")).Render("Total Spent"),
			fmt.Sprintf("\n%s", itemsString(m.summary.Total)),
		),
	)

	countBox := m.styles.StatsBox.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			m.styles.Title.Copy().Background(lipgloss.Color("205")).Render("Total Orders"),
			fmt.Sprintf("\n%d", m.summary.Count),
		),
	)
	
	avgBox := m.styles.StatsBox.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			m.styles.Title.Copy().Background(lipgloss.Color("33")).Render("Average Order"),
			fmt.Sprintf("\n%s", itemsString(m.summary.Average)),
		),
	)

	row1 := lipgloss.JoinHorizontal(lipgloss.Top, totalBox, countBox, avgBox)

	// Bottom Row: Date Range
	dateRange := fmt.Sprintf("From %s to %s", 
		m.summary.Earliest.Format("Jan 02, 2006"),
		m.summary.Latest.Format("Jan 02, 2006"),
	)
	
	content := lipgloss.JoinVertical(lipgloss.Center,
		row1,
		"\n",
		m.styles.Footer.Render(dateRange),
	)

	return lipgloss.Place(m.width, m.height-5, lipgloss.Center, lipgloss.Center, content)
}
