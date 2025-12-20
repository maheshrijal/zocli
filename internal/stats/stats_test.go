package stats

import (
	"testing"
	"time"

	"github.com/maheshrijal/zocli/internal/zomato"
)

func TestComputeSummary(t *testing.T) {
	orders := []zomato.Order{
		{Total: "₹100", Status: "Delivered", PlacedAt: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)},
		{Total: "₹200", Status: "Delivered", PlacedAt: time.Date(2023, 1, 2, 10, 0, 0, 0, time.UTC)},
	}

	got := ComputeSummary(orders)

	if got.Count != 2 {
		t.Errorf("Count = %d, want 2", got.Count)
	}
	if got.Total != 300.0 {
		t.Errorf("Total = %f, want 300.0", got.Total)
	}
	if got.Currency != "₹" {
		t.Errorf("Currency = %q, want ₹", got.Currency)
	}
	if got.Earliest.IsZero() || got.Latest.IsZero() {
		t.Error("Earliest/Latest should not be zero")
	}
}

func TestGroupOrders(t *testing.T) {
	orders := []zomato.Order{
		{Total: "₹100", PlacedAt: time.Date(2023, 1, 15, 10, 0, 0, 0, time.UTC)}, // Jan
		{Total: "₹200", PlacedAt: time.Date(2023, 1, 20, 10, 0, 0, 0, time.UTC)}, // Jan
		{Total: "₹300", PlacedAt: time.Date(2023, 2, 10, 10, 0, 0, 0, time.UTC)}, // Feb
	}

	groups, err := GroupOrders(orders, "month")
	if err != nil {
		t.Fatalf("GroupOrders failed: %v", err)
	}

	if len(groups) != 2 {
		t.Fatalf("Got %d groups, want 2", len(groups))
	}

	// Assuming sorted by date
	jan := groups[0]
	if jan.Key != "Jan 2023" {
		t.Errorf("First group key = %q, want Jan 2023", jan.Key)
	}
	if jan.Count != 2 {
		t.Errorf("Jan count = %d, want 2", jan.Count)
	}
	if jan.Total != 300.0 {
		t.Errorf("Jan total = %f, want 300.0", jan.Total)
	}

	feb := groups[1]
	if feb.Key != "Feb 2023" {
		t.Errorf("Second group key = %q, want Feb 2023", feb.Key)
	}
}

func TestTopRestaurants(t *testing.T) {
	orders := []zomato.Order{
		{Restaurant: "Pizza Hut"},
		{Restaurant: "Pizza Hut"},
		{Restaurant: "Dominos"},
	}

	got := TopRestaurants(orders, 5)
	if len(got) != 2 {
		t.Fatalf("Got %d items, want 2", len(got))
	}

	if got[0].Key != "Pizza Hut" || got[0].Count != 2 {
		t.Errorf("Top 1 = %v, want Pizza Hut (2)", got[0])
	}
	if got[1].Key != "Dominos" || got[1].Count != 1 {
		t.Errorf("Top 2 = %v, want Dominos (1)", got[1])
	}
}

func TestParseAmount(t *testing.T) {
	tests := []struct {
		input    string
		wantVal  float64
		wantCur  string
	}{
		{"₹123.45", 123.45, "₹"},
		{"Rs. 100", 100.0, "₹"},
		{"1,234.56", 1234.56, ""},
		{"$50", 50.0, "$"},
		{"invalid", 0.0, ""},
	}

	for _, tt := range tests {
		val, cur := parseAmount(tt.input)
		if val != tt.wantVal || cur != tt.wantCur {
			t.Errorf("parseAmount(%q) = (%v, %q), want (%v, %q)", tt.input, val, cur, tt.wantVal, tt.wantCur)
		}
	}
}
