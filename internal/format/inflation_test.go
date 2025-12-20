package format

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/maheshrijal/zocli/internal/stats"
)

func TestInflationTable(t *testing.T) {
	buf := new(bytes.Buffer)
	
	t1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)

	points := []stats.ItemPricePoint{
		{
			Date:       t1,
			Restaurant: "RestA",
			ItemName:   "Item1",
			UnitPrice:  100.0,
			Change:     0,
		},
		{
			Date:       t2,
			Restaurant: "RestA",
			ItemName:   "Item1",
			UnitPrice:  120.0,
			Change:     20.0,
		},
	}

	InflationTable(buf, points)
	output := buf.String()

	// Check headers
	if !strings.Contains(output, "DATE") || !strings.Contains(output, "UNIT PRICE") {
		t.Error("Output missing headers")
	}

	// Check data rows
	if !strings.Contains(output, "2024-01-01") || !strings.Contains(output, "100.00") {
		t.Error("Missing first data point")
	}
	if !strings.Contains(output, "2024-02-01") || !strings.Contains(output, "120.00") {
		t.Error("Missing second data point")
	}
	
	// Check visual indicators
	if !strings.Contains(output, "0%") {
		t.Error("Missing 0% indicator")
	}
	if !strings.Contains(output, "+20.0% 🔺") {
		t.Error("Missing inflation indicator")
	}
}

func TestInflationSummaryTable(t *testing.T) {
	buf := new(bytes.Buffer)
	
	summaries := []InflationSummary{
		{
			ItemName:    "RestA - Item1",
			FirstSeen:   "2024-01-01",
			FirstPrice:  100.0,
			LastPrice:   150.0,
			TotalChange: 50.0,
		},
	}

	InflationSummaryTable(buf, summaries)
	output := buf.String()

	// Check headers
	if !strings.Contains(output, "FIRST SEEN") || !strings.Contains(output, "LAST PRICE") {
		t.Error("Output missing headers")
	}

	// Check data
	if !strings.Contains(output, "RestA - Item1") {
		t.Error("Missing item name")
	}
	if !strings.Contains(output, "100.00") || !strings.Contains(output, "150.00") {
		t.Error("Missing prices")
	}
	if !strings.Contains(output, "+50.0% 🔺") {
		t.Error("Missing shift indicator")
	}
}
