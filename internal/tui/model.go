package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	
	"github.com/maheshrijal/zocli/internal/stats"
	"github.com/maheshrijal/zocli/internal/zomato"
)

type tickMsg time.Time

type Filter int

const (
	FilterAll Filter = iota
	FilterYear
	FilterMonth
)

type Tab int

const (
	TabSummary Tab = iota
	TabOrders
	TabInflation
)

type Model struct {
	activeTab Tab
	activeFilter Filter
	width     int
	height    int
	
	// Data
	allOrders []zomato.Order // Source of truth
	orders    []zomato.Order // Filtered view
	summary   stats.Summary

	// Components
	orderTable     table.Model
	inflationTable table.Model
	
	// Styles
	styles Styles
}

func NewModel(orders []zomato.Order) Model {
	summary := stats.ComputeSummary(orders)
	
	m := Model{
		activeTab:    TabSummary,
		activeFilter: FilterAll,
		allOrders:    orders,
		orders:       orders,
		summary:      summary,
		styles:       DefaultStyles(),
	}
	
	m.initOrderTable()
	m.initInflationTable()
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			m.activeTab = (m.activeTab + 1) % 3
		case "shift+tab":
			m.activeTab--
			if m.activeTab < 0 {
				m.activeTab = 2
			}
		case "a":
			m.setFilter(FilterAll)
		case "y":
			m.setFilter(FilterYear)
		case "m":
			m.setFilter(FilterMonth)
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateTableSize()
	}

	// Update components based on active tab
	if m.activeTab == TabOrders {
		m.orderTable, cmd = m.orderTable.Update(msg)
		return m, cmd
	}
	if m.activeTab == TabInflation {
		m.inflationTable, cmd = m.inflationTable.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) setFilter(f Filter) {
	m.activeFilter = f
	now := time.Now()
	
	switch f {
	case FilterYear:
		start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.Local)
		end := start.AddDate(1, 0, -1)
		m.orders = stats.FilterOrdersByDate(m.allOrders, start, end)
	case FilterMonth:
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
		end := start.AddDate(0, 1, -1)
		m.orders = stats.FilterOrdersByDate(m.allOrders, start, end)
	default:
		m.orders = m.allOrders
	}
	
	m.summary = stats.ComputeSummary(m.orders)
	// Re-init tables with new data
	m.initOrderTable()
	m.initInflationTable()
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.renderTabs(),
		m.renderContent(),
		m.renderFooter(),
	)
}

func (m Model) renderContent() string {
	switch m.activeTab {
	case TabSummary:
		return m.viewSummary()
	case TabOrders:
		return m.styles.TableContainer.Render(m.orderTable.View())
	case TabInflation:
		return m.viewInflation()
	}
	return "Unknown Tab"
}

func (m Model) renderFooter() string {
	filterStatus := "All Time"
	switch m.activeFilter {
	case FilterYear:
		filterStatus = "This Year"
	case FilterMonth:
		filterStatus = "This Month"
	}

	help := fmt.Sprintf("%s • %d orders • %s total • y: year • m: month • a: all • q: quit", 
		filterStatus, m.summary.Count, itemsString(m.summary.Total))
	return m.styles.Footer.Render(help)
}

func itemsString(val float64) string {
	return fmt.Sprintf("₹%.2f", val)
}
