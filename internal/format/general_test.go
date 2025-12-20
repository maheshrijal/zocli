package format

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/maheshrijal/zocli/internal/stats"
	"github.com/maheshrijal/zocli/internal/zomato"
)

func TestStatsTable(t *testing.T) {
	buf := new(bytes.Buffer)
	
	summary := stats.Summary{
		Count:    10,
		Total:    1000.50,
		Average:  100.05,
		Currency: "₹",
		Earliest: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		Latest:   time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
	}

	StatsSummary(buf, summary)
	output := buf.String()

	if !strings.Contains(output, "Orders") || !strings.Contains(output, "10") {
		t.Error("Missing count")
	}
	if !strings.Contains(output, "Total Spent") || !strings.Contains(output, "₹1000.50") {
		t.Error("Missing total spent")
	}
	if !strings.Contains(output, "Average Order") || !strings.Contains(output, "₹100.05") {
		t.Error("Missing average")
	}
}

func TestOrdersTable(t *testing.T) {
	buf := new(bytes.Buffer)

	orders := []zomato.Order{
		{
			ID:         "123",
			Restaurant: "Test Rest",
			Status:     "Delivered",
			PlacedAt:   time.Date(2023, 5, 20, 14, 30, 0, 0, time.Local),
			Total:      "₹500",
			Items:      []zomato.OrderItem{{Name: "Food", Quantity: 1}},
		},
	}

	OrdersTable(buf, orders)
	output := buf.String()

	if !strings.Contains(output, "ID") || !strings.Contains(output, "Restaurant") {
		t.Error("Missing headers")
	}
	if !strings.Contains(output, "Test Rest") {
		t.Error("Missing restaurant name")
	}
	if !strings.Contains(output, "Delivered") {
		t.Error("Missing status")
	}
	// Verify date format "2006-01-02 15:04"
	if !strings.Contains(output, "2023-05-20 14:30") {
		t.Error("Missing date")
	}
}
