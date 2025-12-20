package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) initOrderTable() {
	columns := []table.Column{
		{Title: "Date", Width: 10},
		{Title: "Restaurant", Width: 30},
		{Title: "Items", Width: 40},
		{Title: "Total", Width: 10},
		{Title: "Status", Width: 15},
	}

	rows := make([]table.Row, len(m.orders))
	for i, order := range m.orders {
		items := order.Items[0].Name
		if len(order.Items) > 1 {
			items = fmt.Sprintf("%s +%d", items, len(order.Items)-1)
		}
		
		rows[i] = table.Row{
			order.PlacedAt.Format("2006-01-02"),
			order.Restaurant,
			items,
			order.Total,
			order.Status,
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

	m.orderTable = t
}

func (m *Model) updateTableSize() {
	m.orderTable.SetWidth(m.width - 2) // Account for borders
	m.orderTable.SetHeight(m.height - 10) // Leave room for tabs and footer
}

func (m Model) renderTabs() string {
	var tabs []string
	
	labels := []string{"Summary", "Orders", "Inflation"}
	for i, label := range labels {
		if m.activeTab == Tab(i) {
			tabs = append(tabs, m.styles.ActiveTab.Render(label))
		} else {
			tabs = append(tabs, m.styles.Tab.Render(label))
		}
	}
	
	row := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	return lipgloss.NewStyle().Padding(1, 0).Render(row)
}

// Helpers for model.go fixes
// We need to fix NewModel to accept []zomato.Order
