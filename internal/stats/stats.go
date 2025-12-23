package stats

import (
	"errors"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/maheshrijal/zocli/internal/zomato"
)

// SuggestRestaurant recommends a restaurant and a dish based on frequency.
func SuggestRestaurant(orders []zomato.Order) (string, string) {
	if len(orders) == 0 {
		return "No order history to suggest from!", ""
	}

	// 1. Build frequency map for restaurants
	resCounts := make(map[string]int)
	// Map restaurant -> item -> count
	resItems := make(map[string]map[string]int)

	for _, o := range orders {
		resName := strings.TrimSpace(o.Restaurant)
		if resName == "" {
			continue
		}
		resCounts[resName]++
		
		if resItems[resName] == nil {
			resItems[resName] = make(map[string]int)
		}
		for _, item := range o.Items {
			itemName := strings.TrimSpace(item.Name)
			if itemName != "" {
				resItems[resName][itemName] += item.Quantity
			}
		}
	}

	if len(resCounts) == 0 {
		return "No valid restaurants found.", ""
	}

	// 2. Weighted random selection for Restaurant
	var choices []string
	for name, count := range resCounts {
		for i := 0; i < count; i++ {
			choices = append(choices, name)
		}
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	chosenRes := choices[rng.Intn(len(choices))]

	// 3. Pick best item from that restaurant
	items := resItems[chosenRes]
	var bestItem string
	var maxCount int
	
	// Create weighted list for items too, to add variety?
	// Or just pick the favorite. Let's pick the favorite for now, maybe with a fallback.
	var itemChoices []string
	for name, count := range items {
		if count > maxCount {
			maxCount = count
			bestItem = name
		}
		// Also build weight list just in case we want to be random later
		for i := 0; i < count; i++ {
			itemChoices = append(itemChoices, name)
		}
	}
	
	// If we have items, strictly picking the top one is "safe", 
	// but random from top items might be more fun. 
	// Let's stick to the "Most Ordered" item for that restaurant to be helpful.
	if bestItem == "" && len(items) > 0 {
		// Fallback to random key if counts are all 0/weird
		for k := range items {
			bestItem = k
			break
		}
	}

	return chosenRes, bestItem
}

// FilterOrdersByDate filters orders within a specific date range (inclusive).
func FilterOrdersByDate(orders []zomato.Order, start, end time.Time) []zomato.Order {
	var filtered []zomato.Order
	for _, o := range orders {
		if o.PlacedAt.IsZero() {
			continue
		}
		// Check if order is after or at start AND before or at end
		if (o.PlacedAt.Equal(start) || o.PlacedAt.After(start)) && 
		   (o.PlacedAt.Equal(end) || o.PlacedAt.Before(end)) {
			filtered = append(filtered, o)
		}
	}
	return filtered
}

func FindMostExpensiveOrder(orders []zomato.Order) (zomato.Order, float64) {
	var maxOrder zomato.Order
	var maxAmount float64
	
	for _, o := range orders {
		amount, _ := parseAmount(o.Total)
		if amount > maxAmount {
			maxAmount = amount
			maxOrder = o
		}
	}
	return maxOrder, maxAmount
}



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

type Bucket struct {
	Key     string
	Count   int
	Percent float64
}

type SpendBucket struct {
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

func OrdersByWeekday(orders []zomato.Order) []Bucket {
	counts := make(map[time.Weekday]int)
	total := 0
	for _, order := range orders {
		if order.PlacedAt.IsZero() {
			continue
		}
		counts[order.PlacedAt.Weekday()]++
		total++
	}
	out := make([]Bucket, 0, 7)
	for _, day := range []time.Weekday{
		time.Monday,
		time.Tuesday,
		time.Wednesday,
		time.Thursday,
		time.Friday,
		time.Saturday,
		time.Sunday,
	} {
		count := counts[day]
		out = append(out, Bucket{
			Key:     day.String(),
			Count:   count,
			Percent: percent(count, total),
		})
	}
	return out
}

func OrdersByTimeWindow(orders []zomato.Order) []Bucket {
	type window struct {
		Label string
		Start int
		End   int
	}
	windows := []window{
		{Label: "Late night (00-05)", Start: 0, End: 6},
		{Label: "Morning (06-11)", Start: 6, End: 12},
		{Label: "Afternoon (12-17)", Start: 12, End: 18},
		{Label: "Evening (18-23)", Start: 18, End: 24},
	}
	counts := make([]int, len(windows))
	total := 0
	for _, order := range orders {
		if order.PlacedAt.IsZero() {
			continue
		}
		hour := order.PlacedAt.Hour()
		for i, win := range windows {
			if hour >= win.Start && hour < win.End {
				counts[i]++
				total++
				break
			}
		}
	}
	out := make([]Bucket, 0, len(windows))
	for i, win := range windows {
		out = append(out, Bucket{
			Key:     win.Label,
			Count:   counts[i],
			Percent: percent(counts[i], total),
		})
	}
	return out
}

func SpendByWeekday(orders []zomato.Order) []SpendBucket {
	buckets := make(map[time.Weekday]*SpendBucket)
	for _, order := range orders {
		if order.PlacedAt.IsZero() {
			continue
		}
		amount, _ := parseAmount(order.Total)
		day := order.PlacedAt.Weekday()
		entry, ok := buckets[day]
		if !ok {
			entry = &SpendBucket{Key: day.String()}
			buckets[day] = entry
		}
		entry.Count++
		entry.Total += amount
	}

	out := make([]SpendBucket, 0, 7)
	for _, day := range []time.Weekday{
		time.Monday,
		time.Tuesday,
		time.Wednesday,
		time.Thursday,
		time.Friday,
		time.Saturday,
		time.Sunday,
	} {
		entry := buckets[day]
		if entry == nil {
			entry = &SpendBucket{Key: day.String()}
		}
		if entry.Count > 0 {
			entry.Average = entry.Total / float64(entry.Count)
		}
		out = append(out, *entry)
	}
	return out
}

func TopRestaurants(orders []zomato.Order, limit int) []Bucket {
	if limit <= 0 {
		limit = 5
	}
	counts := map[string]int{}
	total := 0
	for _, order := range orders {
		name := strings.TrimSpace(order.Restaurant)
		if name == "" {
			name = "Unknown"
		}
		counts[name]++
		total++
	}
	out := make([]Bucket, 0, len(counts))
	for name, count := range counts {
		out = append(out, Bucket{
			Key:     name,
			Count:   count,
			Percent: percent(count, total),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count == out[j].Count {
			return out[i].Key < out[j].Key
		}
		return out[i].Count > out[j].Count
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

func TopItems(orders []zomato.Order, limit int) []Bucket {
	if limit <= 0 {
		limit = 5
	}
	counts := map[string]int{}
	total := 0
	for _, order := range orders {
		for _, item := range order.Items {
			name := strings.TrimSpace(item.Name)
			if name == "" {
				continue
			}
			qty := item.Quantity
			if qty <= 0 {
				qty = 1
			}
			counts[name] += qty
			total += qty
		}
	}
	out := make([]Bucket, 0, len(counts))
	for name, count := range counts {
		out = append(out, Bucket{
			Key:     name,
			Count:   count,
			Percent: percent(count, total),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count == out[j].Count {
			return out[i].Key < out[j].Key
		}
		return out[i].Count > out[j].Count
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out
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

func percent(value, total int) float64 {
	if total == 0 {
		return 0
	}
	return (float64(value) / float64(total)) * 100
}

var (
	amountPattern = regexp.MustCompile(`[0-9]+(?:[.,][0-9]+)?`)
	moneyPattern  = regexp.MustCompile(`^\s*([^\d\s]+)?\s*([\d.,]+)\s*([^\d\s]+)?\s*$`)
)

func parseAmount(input string) (float64, string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return 0, ""
	}

	currency := ""
	matches := moneyPattern.FindStringSubmatch(input)
	if len(matches) == 4 {
		prefix := strings.TrimSpace(matches[1])
		suffix := strings.TrimSpace(matches[3])
		if prefix != "" {
			currency = prefix
		} else if suffix != "" {
			currency = suffix
		}
		value, cur := parseAmountValue(matches[2], currency)
		return value, cur
	}

	var curBuilder strings.Builder
	for _, r := range input {
		if unicode.IsDigit(r) || r == '.' || r == ',' || unicode.IsSpace(r) {
			continue
		}
		curBuilder.WriteRune(r)
	}
	currency = curBuilder.String()
	amount := amountPattern.FindString(input)
	if amount == "" {
		return 0, ""
	}
	value, cur := parseAmountValue(amount, currency)
	return value, cur
}

func parseAmountValue(amountRaw, currency string) (float64, string) {
	amountRaw = strings.ReplaceAll(amountRaw, ",", "")
	value, err := strconv.ParseFloat(amountRaw, 64)
	if err != nil {
		return 0, normalizeCurrency(currency)
	}
	return value, normalizeCurrency(currency)
}

func normalizeCurrency(cur string) string {
	cur = strings.TrimSpace(cur)
	if cur == "" {
		return ""
	}
	if strings.Contains(cur, "₹") {
		return "₹"
	}
	normalized := strings.ToLower(cur)
	normalized = strings.Trim(normalized, ".")
	switch normalized {
	case "rs", "rs.", "inr":
		return "₹"
	}
	return cur
}
