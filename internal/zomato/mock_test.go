package zomato

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_FetchOrders(t *testing.T) {
	// 1. Setup Mock Server
	mockResponse := `
{
  "sections": {
    "SECTION_USER_ORDER_HISTORY": {
      "count": 1,
      "currentPage": 1,
      "totalPages": 1,
      "entities": [
        { "entity_type": "ORDER", "entity_ids": [12345] }
      ]
    }
  },
  "entities": {
    "ORDER": {
      "12345": {
        "orderId": 12345,
        "totalCost": "₹150",
        "orderDate": "January 1, 2024 12:00 PM",
        "dishString": "1 x Burger",
        "status": 6,
        "deliveryDetails": { "deliveryLabel": "Delivered" },
        "resInfo": { "name": "Burger King" }
      }
    }
  }
}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Headers
		if r.Header.Get("Cookie") != "test-cookie" {
			t.Errorf("Expected cookie 'test-cookie', got %q", r.Header.Get("Cookie"))
		}
		if r.URL.Path != "/webroutes/user/orders" {
			t.Errorf("Expected path /webroutes/user/orders, got %s", r.URL.Path)
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockResponse))
	}))
	defer ts.Close()

	// 2. Setup Client with Mock URL
	client := NewClient("test-cookie")
	client.BaseURL = ts.URL // Override BaseURL

	// 3. Run Fetch
	orders, err := client.FetchOrders(context.Background())
	if err != nil {
		t.Fatalf("FetchOrders failed: %v", err)
	}

	// 4. Verify Results
	if len(orders) != 1 {
		t.Fatalf("Expected 1 order, got %d", len(orders))
	}
	got := orders[0]
	if got.Restaurant != "Burger King" {
		t.Errorf("Restaurant = %q, want Burger King", got.Restaurant)
	}
	if got.Total != "₹150" {
		t.Errorf("Total = %q, want ₹150", got.Total)
	}
}

func TestClient_FetchOrders_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := NewClient("test-cookie")
	client.BaseURL = ts.URL

	_, err := client.FetchOrders(context.Background())
	if err == nil {
		t.Error("Expected error on 500 response, got nil")
	}
}

func TestClient_CheckAuth(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"OK", http.StatusOK, true},
		{"Unauthorized", http.StatusUnauthorized, false},
		{"Forbidden", http.StatusForbidden, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/webroutes/user/address" {
					t.Errorf("Path = %s, want /webroutes/user/address", r.URL.Path)
				}
				w.WriteHeader(tt.statusCode)
			}))
			defer ts.Close()

			client := NewClient("test")
			client.BaseURL = ts.URL

			got, err := client.CheckAuth(context.Background())
			if err != nil {
				t.Fatalf("CheckAuth error: %v", err)
			}
			if got != tt.want {
				t.Errorf("CheckAuth = %v, want %v", got, tt.want)
			}
		})
	}
}
