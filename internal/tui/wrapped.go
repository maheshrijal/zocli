package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/maheshrijal/zocli/internal/stats"
	"github.com/maheshrijal/zocli/internal/zomato"
)

type WrappedModel struct {
	orders []zomato.Order
	width  int
	height int
	slide  int
	
	// Stats
	year           int
	totalSpent     float64
	orderCount     int
	topRestaurant  string
	topItem        string
	mostExpensive  zomato.Order
	maxAmount      float64
	busiestWeekday string
	busiestTime    string
	currency       string
}

func NewWrappedModel(orders []zomato.Order, year int) WrappedModel {
	summary := stats.ComputeSummary(orders)
	topRes := stats.TopRestaurants(orders, 1)
	topItems := stats.TopItems(orders, 1)
	maxOrder, maxAmt := stats.FindMostExpensiveOrder(orders)
	weekdays := stats.OrdersByWeekday(orders)
	times := stats.OrdersByTimeWindow(orders)
	
	m := WrappedModel{
		orders:        orders,
		year:          year,
		totalSpent:    summary.Total,
		orderCount:    summary.Count,
		mostExpensive: maxOrder,
		maxAmount:     maxAmt,
		currency:      summary.Currency,
	}
	
	if len(topRes) > 0 {
		m.topRestaurant = topRes[0].Key
	}
	if len(topItems) > 0 {
		m.topItem = topItems[0].Key
	}

	// Find max weekday
	var maxDayCount int
	for _, w := range weekdays {
		if w.Count > maxDayCount {
			maxDayCount = w.Count
			m.busiestWeekday = w.Key
		}
	}

	// Find max time
	var maxTimeCount int
	for _, t := range times {
		if t.Count > maxTimeCount {
			maxTimeCount = t.Count
			m.busiestTime = t.Key
		}
	}

	return m
}

func (m WrappedModel) Init() tea.Cmd {
	return nil
}

func (m WrappedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "tab", "right", "l", "space", "enter":
			if m.slide < 4 {
				m.slide++
			} else {
				return m, tea.Quit
			}
		case "shift+tab", "left", "h":
			if m.slide > 0 {
				m.slide--
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m WrappedModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var content string
	switch m.slide {
	case 0:
		content = m.slideIntro()
	case 1:
		content = m.slideTotals()
	case 2:
		content = m.slideFavorites()
	case 3:
		content = m.slideHabits()
	case 4:
		content = m.slideExpensive()
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// Slides

func (m WrappedModel) slideIntro() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#E23744")). // Zomato Red
		Background(lipgloss.Color("#FFFFFF")).
		Padding(1, 3).
		Render(fmt.Sprintf("Zomato Wrapped %d", m.year))
		
	sub := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("Your year in food 🍕")
		
	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("238")).
		MarginTop(2).
		Render("Press Tab to start")

	return lipgloss.JoinVertical(lipgloss.Center, title, sub, hint)
}

func (m WrappedModel) slideTotals() string {
	bigNum := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render(fmt.Sprintf("%d", m.orderCount))
		
	text := "Orders placed"
	
	bigSpend := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("42")).
		Render(fmt.Sprintf("%s%.2f", m.currency, m.totalSpent))
		
	spendText := "Total value of food consumed"

	return lipgloss.JoinVertical(lipgloss.Center, 
		bigNum, text, 
		"\n\n",
		bigSpend, spendText,
	)
}

func (m WrappedModel) slideFavorites() string {
	topRes := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("220")). // Gold
		Render("🏆 " + m.topRestaurant)
		
	topItem := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("220")).
		Render("🍽️ " + m.topItem)
		
	return lipgloss.JoinVertical(lipgloss.Center,
		"You kept going back to...",
		"\n",
		topRes,
		"\n\n",
		"And you couldn't get enough of...",
		"\n",
		topItem,
	)
}

func (m WrappedModel) slideHabits() string {
	day := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render(m.busiestWeekday)
	time := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render(m.busiestTime)

	return lipgloss.JoinVertical(lipgloss.Center,
		"You loved ordering on...",
		"\n",
		day,
		"\n\n",
		"Especially during...",
		"\n",
		time,
	)
}

func (m WrappedModel) slideExpensive() string {
	amt := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color("#FFD700")).
		Foreground(lipgloss.Color("#000000")).
		Padding(0, 1).
		Render(fmt.Sprintf("%s%.2f", m.currency, m.maxAmount))

	rest := lipgloss.NewStyle().Italic(true).Render(m.mostExpensive.Restaurant)
	
	var items []string
	for _, i := range m.mostExpensive.Items {
		items = append(items, fmt.Sprintf("%dx %s", i.Quantity, i.Name))
	}
	itemList := strings.Join(items, ", ")
	if len(itemList) > 50 {
		itemList = itemList[:47] + "..."
	}

	return lipgloss.JoinVertical(lipgloss.Center,
		"Do you remember this feast? 💸",
		"\n",
		amt,
		"\n",
		rest,
		"\n",
		itemList,
		"\n\n",
		lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Render("(Press Tab to finish)"),
	)
}
