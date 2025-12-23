package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/maheshrijal/zocli/internal/zomato"
)

// ToCSV writes orders as CSV to the provided writer.
func ToCSV(orders []zomato.Order, w io.Writer) error {
	cw := csv.NewWriter(w)

	// Header
	if err := cw.Write([]string{
		"Order ID",
		"Restaurant",
		"Date",
		"Status",
		"Total",
		"Items",
	}); err != nil {
		return err
	}

	for _, o := range orders {
		var items []string
		for _, i := range o.Items {
			items = append(items, fmt.Sprintf("%dx %s", i.Quantity, i.Name))
		}
		
		record := []string{
			o.ID,
			o.Restaurant,
			o.PlacedAt.Format("2006-01-02 15:04:05"),
			o.Status,
			o.Total,
			strings.Join(items, "; "),
		}
		if err := cw.Write(record); err != nil {
			return err
		}
	}
	
	cw.Flush()
	return cw.Error()
}

// ToJSON writes orders as pretty-printed JSON to the provided writer.
func ToJSON(orders []zomato.Order, w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(orders)
}

// Helper to convert currency string to float for external usage if needed
func parseCurrency(s string) float64 {
	// Removes non-numeric (except dot) and parses
	s = strings.ReplaceAll(s, "₹", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimSpace(s)
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
