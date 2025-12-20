package tui

import (
	"fmt" // Import fmt
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	
	"github.com/maheshrijal/zocli/internal/stats"
)

func (m *Model) initInflationTable() {
	columns := []table.Column{
		{Title: "Item", Width: 35},
		{Title: "Restaurant", Width: 25},
		{Title: "Old", Width: 10},
		{Title: "New", Width: 10},
		{Title: "Change", Width: 12},
	}

	// Calculate trends
	trends := stats.FindTopInflationTrends(m.orders, 50)
	
	rows := make([]table.Row, len(trends))
	for i, t := range trends {
		changeStr := fmt.Sprintf("%.1f%%", t.TotalChange) // Use fmt.Sprintf
		if t.TotalChange > 0 {
			changeStr = "🔺 " + changeStr
		} else if t.TotalChange < 0 {
			changeStr = "🔻 " + changeStr
		}

		rows[i] = table.Row{
			t.ItemName,
			t.Restaurant,
			fmt.Sprintf("%.0f", t.FirstPrice), // Use fmt.Sprintf
			fmt.Sprintf("%.0f", t.LastPrice),  // Use fmt.Sprintf
			changeStr,
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(20),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m.inflationTable = t
}

func (m Model) viewInflation() string {
	return m.styles.TableContainer.Render(m.inflationTable.View())
}
