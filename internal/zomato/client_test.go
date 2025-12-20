package zomato

import (
	"reflect"
	"testing"
	"time"
)

func TestParseOrderDate(t *testing.T) {
	tests := []struct {
		input    string
		wantTime string // RFC3339 for easy comparison
		wantErr  bool
	}{
		{
			input:    "January 2, 2023 at 03:04 PM",
			wantTime: "2023-01-02T15:04:00Z",
		},
		{
			input:    "Jan 2, 2023 03:04 PM",
			wantTime: "2023-01-02T15:04:00Z",
		},
		{
			input:    "2023-01-02 15:04",
			wantTime: "2023-01-02T15:04:00Z",
		},
		{
			input:   "",
			wantErr: true,
		},
		{
			input:   "invalid date",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseOrderDate(tt.input)
			if tt.wantErr {
				if !got.IsZero() {
					t.Errorf("parseOrderDate(%q) = %v, want zero time", tt.input, got)
				}
				return
			}

			// Construct expected time in Local timezone to match implementation
			wantTime, err := time.Parse(time.RFC3339, tt.wantTime)
			if err != nil {
				t.Fatalf("bad test case time: %v", err)
			}
			// Convert wantTime (UTC) to Local to match what parseOrderDate returns (Local)
			// effectively we want the same wall clock time if the input didn't specify zone
			// But wait, the input strings don't have zone. So parseOrderDate interprets them as Local.
			// The validation string "2023-01-02T15:04:00Z" is UTC.
			// We should construct the expected time using Date() in Local.
			
			want := time.Date(wantTime.Year(), wantTime.Month(), wantTime.Day(), wantTime.Hour(), wantTime.Minute(), wantTime.Second(), 0, time.Local)
			
			if !got.Equal(want) {
				t.Errorf("parseOrderDate(%q) = %v, want %v", tt.input, got, want)
			}
		})
	}
}

func TestParseItems(t *testing.T) {
	tests := []struct {
		input string
		want  []OrderItem
	}{
		{
			input: "2 x Burger, 1 x Fries",
			want: []OrderItem{
				{Name: "Burger", Quantity: 2},
				{Name: "Fries", Quantity: 1},
			},
		},
		{
			input: "Burger, Fries",
			want: []OrderItem{
				{Name: "Burger", Quantity: 1},
				{Name: "Fries", Quantity: 1},
			},
		},
		{
			input: "  3   x   Pizza  ",
			want: []OrderItem{
				{Name: "Pizza", Quantity: 3},
			},
		},
		{
			input: "1 x Chicken Donne Biryani [Regular, Serves 1, 2 Pieces], 1 x Thumsup",
			want: []OrderItem{
				{Name: "Chicken Donne Biryani [Regular, Serves 1, 2 Pieces]", Quantity: 1},
				{Name: "Thumsup", Quantity: 1},
			},
		},
		{
			input: "1 x Cheese Volcano Double Chicken",
			want: []OrderItem{
				{Name: "Cheese Volcano Double Chicken", Quantity: 1},
			},
		},
		{
			input: "",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseItems(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseItems(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
