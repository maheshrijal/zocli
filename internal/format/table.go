package format

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/maheshrijal/zocli/internal/zomato"
)

func OrdersTable(w io.Writer, orders []zomato.Order) {
	fmt.Fprintln(w, "ID\tRestaurant\tStatus\tPlaced\tTotal\tItems")
	for _, order := range orders {
		items := make([]string, 0, len(order.Items))
		for _, item := range order.Items {
			if item.Quantity > 1 {
				items = append(items, fmt.Sprintf("%s x%d", item.Name, item.Quantity))
				continue
			}
			items = append(items, item.Name)
		}
		placed := formatTime(order.PlacedAt)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			order.ID,
			order.Restaurant,
			order.Status,
			placed,
			order.Total,
			strings.Join(items, ", "),
		)
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("2006-01-02 15:04")
}
