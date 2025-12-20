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
	Change     float64 // Percentage change from previous point (0 for first)
}

func CalculateInflation(orders []zomato.Order, query string) ([]ItemPricePoint, error) {
	query = strings.ToLower(strings.TrimSpace(query))
	var points []ItemPricePoint

	// Sort orders by date ascending (oldest first) to track trend
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].PlacedAt.Before(orders[j].PlacedAt)
	})

	for _, order := range orders {
		if order.Status != "Delivered" { // Only count completed orders
			continue
		}
		
		// Strategy: Only use single-item-type orders for accuracy
		if len(order.Items) != 1 {
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

		qty := item.Quantity
		if qty <= 0 {
			qty = 1
		}

		unitPrice := total / float64(qty)
		
		// Round to 2 decimal places
		unitPrice = math.Round(unitPrice*100) / 100

		points = append(points, ItemPricePoint{
			Date:       order.PlacedAt,
			OrderId:    order.ID,
			Restaurant: order.Restaurant,
			ItemName:   item.Name,
			UnitPrice:  unitPrice,
			Quantity:   qty,
			OrderTotal: total,
		})
	}

	// Calculate changes
	for i := 1; i < len(points); i++ {
		prev := points[i-1].UnitPrice
		curr := points[i].UnitPrice
		if prev > 0 {
			change := ((curr - prev) / prev) * 100
			points[i].Change = math.Round(change*100) / 100
		}
	}

	return points, nil
}
