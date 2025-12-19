package sample

import (
	"embed"
	"encoding/json"

	"github.com/maheshrijal/zocli/internal/zomato"
)

//go:embed orders.json
var ordersFS embed.FS

func Orders() ([]zomato.Order, error) {
	data, err := ordersFS.ReadFile("orders.json")
	if err != nil {
		return nil, err
	}
	var orders []zomato.Order
	if err := json.Unmarshal(data, &orders); err != nil {
		return nil, err
	}
	return orders, nil
}
