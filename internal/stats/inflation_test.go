package stats

import (
	"testing"
	"time"

	"github.com/maheshrijal/zocli/internal/zomato"
)

func TestCalculateInflation(t *testing.T) {
	t1 := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2023, 2, 1, 10, 0, 0, 0, time.UTC)
	t3 := time.Date(2023, 3, 1, 10, 0, 0, 0, time.UTC)

	orders := []zomato.Order{
		// Match 1: Pizza, Price 100
		{ID: "1", PlacedAt: t1, Status: "Delivered", Total: "₹100", Items: []zomato.OrderItem{{Name: "Cheese Pizza", Quantity: 1}}},
		
		// Multi-item (Should be ignored)
		{ID: "2", PlacedAt: t2, Status: "Delivered", Total: "₹250", Items: []zomato.OrderItem{{Name: "Cheese Pizza", Quantity: 1}, {Name: "Coke", Quantity: 1}}},
		
		// Match 2: Pizza, Price 120 (Inflation +20%) to check rounding logic
		{ID: "3", PlacedAt: t3, Status: "Delivered", Total: "₹240", Items: []zomato.OrderItem{{Name: "Cheese Pizza", Quantity: 2}}},
	
		// No Match
		{ID: "4", PlacedAt: t3, Status: "Delivered", Total: "₹50", Items: []zomato.OrderItem{{Name: "Garlic Bread", Quantity: 1}}},
	}

	points, err := CalculateInflation(orders, "Pizza")
	if err != nil {
		t.Fatalf("CalculateInflation failed: %v", err)
	}

	if len(points) != 2 {
		t.Fatalf("Got %d points, want 2", len(points))
	}

	p1 := points[0]
	if p1.UnitPrice != 100.0 {
		t.Errorf("Point 1 unit price = %f, want 100.0", p1.UnitPrice)
	}
	if p1.Change != 0 {
		t.Errorf("Point 1 change = %f, want 0", p1.Change)
	}

	p2 := points[1]
	if p2.UnitPrice != 120.0 {
		t.Errorf("Point 2 unit price = %f, want 120.0 (240/2)", p2.UnitPrice)
	}
	if p2.Change != 20.0 {
		t.Errorf("Point 2 change = %f, want 20.0", p2.Change)
	}
}
