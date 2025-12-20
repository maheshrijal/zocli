package stats

import (
	"math"
	"sort"
	"strings"
	"time"

	"github.com/maheshrijal/zocli/internal/zomato"
)

type ItemPricePoint struct {
	Date       time.Time
	OrderId    string
	Restaurant string
	ItemName   string
	UnitPrice  float64
	Quantity   int
	OrderTotal float64
	Change     float64 // Percentage change from previous point (from same restaurant)
}

type InflationTrend struct {
	Key         string // Restaurant + Item
	ItemName    string
	Restaurant  string
	FirstSeen   time.Time
	FirstPrice  float64
	LastPrice   float64
	TotalChange float64
	Count       int
	Points      []ItemPricePoint
}

// FindTopInflationTrends identifies distinct Restaurant+Item pairs with significant history.
func FindTopInflationTrends(orders []zomato.Order, limit int) []InflationTrend {
	// 1. Group all single-item orders by "Restaurant|ItemName"
	groups := make(map[string][]ItemPricePoint)

	// Sort orders oldest first
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].PlacedAt.Before(orders[j].PlacedAt)
	})

	for _, order := range orders {
		if order.Status != "Delivered" || len(order.Items) != 1 {
			continue
		}
		total, _ := parseAmount(order.Total)
		if total <= 0 {
			continue
		}
		item := order.Items[0]
		key := order.Restaurant + "|" + item.Name
		
		unitPrice := total / float64(max(item.Quantity, 1))
		unitPrice = math.Round(unitPrice*100) / 100

		groups[key] = append(groups[key], ItemPricePoint{
			Date:       order.PlacedAt,
			Restaurant: order.Restaurant,
			ItemName:   item.Name,
			UnitPrice:  unitPrice,
		})
	}

	// 2. Convert valid groups (>= 2 points) to trends
	var trends []InflationTrend
	for _, points := range groups {
		if len(points) < 2 {
			continue
		}
		first := points[0]
		last := points[len(points)-1]
		
		change := 0.0
		if first.UnitPrice > 0 {
			change = ((last.UnitPrice - first.UnitPrice) / first.UnitPrice) * 100
		}

		trends = append(trends, InflationTrend{
			Key:         points[0].Restaurant + " - " + points[0].ItemName,
			ItemName:    points[0].ItemName,
			Restaurant:  points[0].Restaurant,
			FirstSeen:   first.Date,
			FirstPrice:  first.UnitPrice,
			LastPrice:   last.UnitPrice,
			TotalChange: math.Round(change*100) / 100,
			Count:       len(points),
			Points:      points,
		})
	}

	// 3. Sort by popularity (Count) descending
	sort.Slice(trends, func(i, j int) bool {
		return trends[i].Count > trends[j].Count
	})

	// 4. Return top N
	if len(trends) > limit {
		return trends[:limit]
	}
	return trends
}

func CalculateInflation(orders []zomato.Order, query string) ([]ItemPricePoint, error) {
	query = strings.ToLower(strings.TrimSpace(query))
	var points []ItemPricePoint

	sort.Slice(orders, func(i, j int) bool {
		return orders[i].PlacedAt.Before(orders[j].PlacedAt)
	})

	for _, order := range orders {
		if order.Status != "Delivered" || len(order.Items) != 1 {
			continue
		}

		item := order.Items[0]
		itemName := strings.ToLower(item.Name)

		if query != "" && !strings.Contains(itemName, query) {
			continue
		}

		total, _ := parseAmount(order.Total)
		if total <= 0 {
			continue
		}
		
		unitPrice := total / float64(max(item.Quantity, 1))
		unitPrice = math.Round(unitPrice*100) / 100

		points = append(points, ItemPricePoint{
			Date:       order.PlacedAt,
			OrderId:    order.ID,
			Restaurant: order.Restaurant,
			ItemName:   item.Name,
			UnitPrice:  unitPrice,
			Quantity:   item.Quantity,
			OrderTotal: total,
		})
	}

	// Calculate changes ONLY against same restaurant
	lastPriceByRest := make(map[string]float64)

	for i := range points {
		rest := points[i].Restaurant
		prevPrice := lastPriceByRest[rest]
		
		if prevPrice > 0 {
			change := ((points[i].UnitPrice - prevPrice) / prevPrice) * 100
			points[i].Change = math.Round(change*100) / 100
		}
		
		lastPriceByRest[rest] = points[i].UnitPrice
	}

	return points, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
