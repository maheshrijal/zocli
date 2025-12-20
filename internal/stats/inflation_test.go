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
		// Match 1: Pizza, Price 100 @ Dominos
		{ID: "1", Restaurant: "Dominos", PlacedAt: t1, Status: "Delivered", Total: "₹100", Items: []zomato.OrderItem{{Name: "Cheese Pizza", Quantity: 1}}},

		// Match 2: Pizza, Price 120 @ Dominos (Inflation +20%)
		{ID: "3", Restaurant: "Dominos", PlacedAt: t3, Status: "Delivered", Total: "₹240", Items: []zomato.OrderItem{{Name: "Cheese Pizza", Quantity: 2}}},

		// Match 3: Pizza, Price 500 @ PizzaHut (Should NOT compare with Dominos)
		// 100 -> 120 (Dominos)
		// 500 (PizzaHut, New Chain) -> Change should be 0 because it's the first time seeing PizzaHut
		{ID: "5", Restaurant: "PizzaHut", PlacedAt: t2, Status: "Delivered", Total: "₹500", Items: []zomato.OrderItem{{Name: "Cheese Pizza", Quantity: 1}}},
	}

	points, err := CalculateInflation(orders, "Pizza")
	if err != nil {
		t.Fatalf("CalculateInflation failed: %v", err)
	}

	// Should have 3 points: Dominos (x2), PizzaHut (x1)
	if len(points) != 3 {
		t.Fatalf("Got %d points, want 3", len(points))
	}

	// Point 1: Dominos 100
	if points[0].Restaurant != "Dominos" || points[0].UnitPrice != 100.0 {
		t.Errorf("P1 mismatch: %v", points[0])
	}

	// Point 2: PizzaHut 500 (Date is t2, so it comes second)
	// Change should be 0 because it's first time for PizzaHut
	if points[1].Restaurant != "PizzaHut" || points[1].UnitPrice != 500.0 {
		t.Errorf("P2 mismatch: %v", points[1])
	}
	if points[1].Change != 0 {
		t.Errorf("P2 change = %f, want 0 (different restaurant)", points[1].Change)
	}

	// Point 3: Dominos 120 (Date t3)
	// Change should be calculated against P1 (Dominos 100) -> +20%
	if points[2].Restaurant != "Dominos" || points[2].UnitPrice != 120.0 {
		t.Errorf("P3 mismatch: %v", points[2])
	}
	if points[2].Change != 20.0 {
		t.Errorf("P3 change = %f, want 20.0", points[2].Change)
	}
}

func TestFindTopInflationTrends(t *testing.T) {
	t1 := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2023, 2, 1, 10, 0, 0, 0, time.UTC)

	orders := []zomato.Order{
		// Trend 1: Burger @ McD (2 orders) -> VALID
		{ID: "1", Restaurant: "McD", PlacedAt: t1, Status: "Delivered", Total: "₹50", Items: []zomato.OrderItem{{Name: "Burger", Quantity: 1}}},
		{ID: "2", Restaurant: "McD", PlacedAt: t2, Status: "Delivered", Total: "₹60", Items: []zomato.OrderItem{{Name: "Burger", Quantity: 1}}},

		// Trend 2: Pizza @ Dominos (1 order) -> INVALID (need >=2)
		{ID: "3", Restaurant: "Dominos", PlacedAt: t1, Status: "Delivered", Total: "₹100", Items: []zomato.OrderItem{{Name: "Pizza", Quantity: 1}}},
		
		// Trend 3: Pizza @ PizzaHut (1 order) -> INVALID (need >=2)
		{ID: "4", Restaurant: "PizzaHut", PlacedAt: t1, Status: "Delivered", Total: "₹100", Items: []zomato.OrderItem{{Name: "Pizza", Quantity: 1}}},
	}

	trends := FindTopInflationTrends(orders, 5)

	if len(trends) != 1 {
		t.Fatalf("Got %d trends, want 1 (only McD has history)", len(trends))
	}

	if trends[0].Restaurant != "McD" {
		t.Errorf("Expected McD, got %s", trends[0].Restaurant)
	}
	if trends[0].TotalChange != 20.0 {
		t.Errorf("Expected 20%% change, got %f", trends[0].TotalChange)
	}
}
