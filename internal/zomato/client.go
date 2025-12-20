package zomato

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	HTTPClient *http.Client
	BaseURL    string
	Cookie     string
}

func NewClient(cookie string) *Client {
	return &Client{
		HTTPClient: http.DefaultClient,
		BaseURL:    "https://www.zomato.com",
		Cookie:     cookie,
	}
}

func (c *Client) FetchOrders(ctx context.Context) ([]Order, error) {
	var all []Order
	page := 1
	seen := map[string]struct{}{}
	const maxPages = 50

	for {
		resp, err := c.fetchOrdersPage(ctx, page)
		if err != nil {
			return nil, err
		}

		orders := ordersFromResponse(resp)
		newCount := 0
		for _, order := range orders {
			if order.ID == "" {
				continue
			}
			if _, ok := seen[order.ID]; ok {
				continue
			}
			seen[order.ID] = struct{}{}
			all = append(all, order)
			newCount++
		}

		totalPages := resp.Sections.OrderHistory.TotalPages
		if newCount == 0 {
			break
		}
		if totalPages == 0 || page >= totalPages {
			break
		}
		if page >= maxPages {
			break
		}
		time.Sleep(500 * time.Millisecond)
		page++
	}

	return all, nil
}

func (c *Client) CheckAuth(ctx context.Context) (bool, error) {
	endpoint := c.BaseURL + "/webroutes/user/address"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return false, err
	}
	c.decorateRequest(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusUnauthorized, http.StatusForbidden:
		return false, nil
	default:
		return false, fmt.Errorf("auth status request failed: %s", resp.Status)
	}
}

type ordersResponse struct {
	Sections struct {
		OrderHistory orderHistorySection `json:"SECTION_USER_ORDER_HISTORY"`
	} `json:"sections"`
	Entities struct {
		Order map[string]orderEntity `json:"ORDER"`
	} `json:"entities"`
}

type orderHistorySection struct {
	Count       int               `json:"count"`
	CurrentPage int               `json:"currentPage"`
	TotalPages  int               `json:"totalPages"`
	Entities    []orderEntityList `json:"entities"`
}

type orderEntityList struct {
	EntityType string  `json:"entity_type"`
	EntityIDs  []int64 `json:"entity_ids"`
}

type orderEntity struct {
	OrderID         int64  `json:"orderId"`
	TotalCost       string `json:"totalCost"`
	OrderDate       string `json:"orderDate"`
	DishString      string `json:"dishString"`
	HashID          string `json:"hashId"`
	Status          int    `json:"status"`
	PaymentStatus   int    `json:"paymentStatus"`
	DeliveryDetails struct {
		DeliveryLabel   string `json:"deliveryLabel"`
		DeliveryMessage string `json:"deliveryMessage"`
	} `json:"deliveryDetails"`
	ResInfo struct {
		Name string `json:"name"`
	} `json:"resInfo"`
}

func (c *Client) fetchOrdersPage(ctx context.Context, page int) (ordersResponse, error) {
	var out ordersResponse

	endpoint, err := url.Parse(c.BaseURL + "/webroutes/user/orders")
	if err != nil {
		return out, err
	}
	query := endpoint.Query()
	if page > 1 {
		query.Set("page", strconv.Itoa(page))
	}
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return out, err
	}
	c.decorateRequest(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return out, fmt.Errorf("zomato orders request failed: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return out, err
	}

	return out, nil
}

func (c *Client) decorateRequest(req *http.Request) {
	req.Header.Set("Cookie", c.Cookie)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)")
}

func ordersFromResponse(resp ordersResponse) []Order {
	if len(resp.Sections.OrderHistory.Entities) == 0 {
		return nil
	}

	var orders []Order
	for _, entity := range resp.Sections.OrderHistory.Entities {
		if entity.EntityType != "ORDER" {
			continue
		}
		for _, id := range entity.EntityIDs {
			raw, ok := resp.Entities.Order[strconv.FormatInt(id, 10)]
			if !ok {
				continue
			}
			orders = append(orders, normalizeOrder(raw))
		}
	}
	return orders
}

func normalizeOrder(raw orderEntity) Order {
	status := strings.TrimSpace(raw.DeliveryDetails.DeliveryLabel)
	if status == "" {
		status = strings.TrimSpace(raw.DeliveryDetails.DeliveryMessage)
	}
	if status == "" && raw.Status != 0 {
		status = fmt.Sprintf("Status %d", raw.Status)
	}

	return Order{
		ID:         strconv.FormatInt(raw.OrderID, 10),
		Restaurant: strings.TrimSpace(raw.ResInfo.Name),
		Status:     status,
		PlacedAt:   parseOrderDate(raw.OrderDate),
		Total:      strings.TrimSpace(raw.TotalCost),
		Items:      parseItems(raw.DishString),
	}
}

func parseOrderDate(input string) time.Time {
	input = strings.TrimSpace(input)
	if input == "" {
		return time.Time{}
	}
	layouts := []string{
		"January 2, 2006 at 03:04 PM",
		"January 2, 2006 03:04 PM",
		"Jan 2, 2006 at 03:04 PM",
		"Jan 2, 2006 03:04 PM",
		"02 Jan 2006 at 03:04 PM",
		"02 Jan 2006 03:04 PM",
		"2006-01-02 15:04",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, input, time.Local); err == nil {
			return parsed
		}
	}
	return time.Time{}
}

var dishQtyPattern = regexp.MustCompile(`^\\s*(\\d+)\\s*x\\s*(.+)\\s*$`)

func parseItems(dishString string) []OrderItem {
	dishString = strings.TrimSpace(dishString)
	if dishString == "" {
		return nil
	}

	parts := strings.Split(dishString, ",")
	items := make([]OrderItem, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if matches := dishQtyPattern.FindStringSubmatch(part); len(matches) == 3 {
			qty, err := strconv.Atoi(matches[1])
			if err != nil {
				qty = 1
			}
			name := strings.TrimSpace(matches[2])
			if name == "" {
				name = part
			}
			items = append(items, OrderItem{Name: name, Quantity: qty})
			continue
		}
		items = append(items, OrderItem{Name: part, Quantity: 1})
	}
	return items
}
