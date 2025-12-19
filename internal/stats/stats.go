package stats

import (
	"errors"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/maheshrijal/zocli/internal/zomato"
)

type Summary struct {
	Count    int
	Total    float64
	Average  float64
	Currency string
	Earliest time.Time
	Latest   time.Time
}

type Group struct {
	Key     string
	Count   int
	Total   float64
	Average float64
}

func ComputeSummary(orders []zomato.Order) Summary {
	var total float64
	var currency string
	var earliest time.Time
	var latest time.Time

	for _, order := range orders {
		amount, cur := parseAmount(order.Total)
		total += amount
		if currency == "" && cur != "" {
			currency = cur
		}
		if !order.PlacedAt.IsZero() {
			if earliest.IsZero() || order.PlacedAt.Before(earliest) {
				earliest = order.PlacedAt
			}
			if latest.IsZero() || order.PlacedAt.After(latest) {
				latest = order.PlacedAt
			}
		}
	}

	var avg float64
	if len(orders) > 0 {
		avg = total / float64(len(orders))
	}

	return Summary{
		Count:    len(orders),
		Total:    total,
		Average:  avg,
		Currency: currency,
		Earliest: earliest,
		Latest:   latest,
	}
}

func GroupOrders(orders []zomato.Order, groupBy string) ([]Group, error) {
	groupBy = strings.ToLower(strings.TrimSpace(groupBy))
	if groupBy == "" {
		groupBy = "none"
	}

	if groupBy != "none" && groupBy != "month" && groupBy != "year" {
		return nil, errors.New("group must be one of: none, month, year")
	}

	if len(orders) == 0 {
		return []Group{}, nil
	}

	groups := map[string]*Group{}
	for _, order := range orders {
		key := groupKey(order.PlacedAt, groupBy)
		amount, _ := parseAmount(order.Total)
		entry, ok := groups[key]
		if !ok {
			entry = &Group{Key: key}
			groups[key] = entry
		}
		entry.Count++
		entry.Total += amount
	}

	out := make([]Group, 0, len(groups))
	for _, entry := range groups {
		if entry.Count > 0 {
			entry.Average = entry.Total / float64(entry.Count)
		}
		out = append(out, *entry)
	}

	sortGroups(out, groupBy)
	return out, nil
}

func groupKey(ts time.Time, groupBy string) string {
	if groupBy == "none" {
		return "all"
	}
	if ts.IsZero() {
		return "unknown"
	}
	switch groupBy {
	case "year":
		return ts.Format("2006")
	default:
		return ts.Format("Jan 2006")
	}
}

func sortGroups(groups []Group, groupBy string) {
	if groupBy == "none" {
		return
	}

	layout := "Jan 2006"
	if groupBy == "year" {
		layout = "2006"
	}

	sort.Slice(groups, func(i, j int) bool {
		ai, bi := groups[i], groups[j]
		if ai.Key == "unknown" {
			return false
		}
		if bi.Key == "unknown" {
			return true
		}
		at, aerr := time.Parse(layout, ai.Key)
		bt, berr := time.Parse(layout, bi.Key)
		if aerr == nil && berr == nil {
			return at.Before(bt)
		}
		return ai.Key < bi.Key
	})
}

var amountPattern = regexp.MustCompile(`[0-9]+(?:[.,][0-9]+)?`)

func parseAmount(input string) (float64, string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return 0, ""
	}

	currency := ""
	for _, r := range input {
		if unicode.IsDigit(r) || r == '.' || r == ',' || unicode.IsSpace(r) {
			continue
		}
		currency = string(r)
		break
	}

	match := amountPattern.FindString(input)
	if match == "" {
		return 0, currency
	}
	match = strings.ReplaceAll(match, ",", "")
	value, err := strconv.ParseFloat(match, 64)
	if err != nil {
		return 0, currency
	}
	return value, currency
}
