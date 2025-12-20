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

type Tab int

const (
	TabSummary Tab = iota
	TabOrders
	TabInflation
)

type Model struct {
	activeTab Tab
	width     int
	height    int
	
	// Data
	summary stats.Summary
	orders  []zomato.Order

	// Components
	orderTable table.Model
	
	// Styles
	styles Styles
}

func NewModel(orders []zomato.Order) Model {
	summary := stats.ComputeSummary(orders)
	
	m := Model{
		activeTab: TabSummary,
		orders:    orders,
		summary:   summary,
		styles:    DefaultStyles(),
	}
	
	m.initOrderTable()
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

	return m, nil
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
	return m.styles.Footer.Render(fmt.Sprintf("%d orders • %s total • q: quit • tab: switch view", m.summary.Count, itemsString(m.summary.Total)))
}

func itemsString(val float64) string {
	return fmt.Sprintf("₹%.2f", val)
}

func (m Model) viewSummary() string {
	return lipgloss.Place(m.width, m.height-5, lipgloss.Center, lipgloss.Center, 
		m.styles.Title.Render("Summary Dashboard Coming Soon"),
	)
}

func (m Model) viewInflation() string {
	return lipgloss.Place(m.width, m.height-5, lipgloss.Center, lipgloss.Center,
		m.styles.Title.Render("Inflation Tracker Coming Soon"),
	)
}
