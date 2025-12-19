package zomato

import "time"

type Order struct {
	ID         string      `json:"id"`
	Restaurant string      `json:"restaurant"`
	Status     string      `json:"status"`
	PlacedAt   time.Time   `json:"placed_at"`
	Total      string      `json:"total"`
	Items      []OrderItem `json:"items"`
}

type OrderItem struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}
