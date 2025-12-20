package format

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/maheshrijal/zocli/internal/stats"
)

type InflationSummary struct {
	ItemName    string
	FirstSeen   string
	FirstPrice  float64
	LastPrice   float64
	TotalChange float64
}

func InflationTable(w io.Writer, points []stats.ItemPricePoint) {
	if len(points) == 0 {
		fmt.Fprintln(w, "No matching orders found (or no single-item orders for accurate pricing).")
		return
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "DATE\tRESTAURANT\tITEM\tUNIT PRICE\tCHANGE %")
	fmt.Fprintln(tw, "----\t----------\t----\t----------\t--------")

	for _, p := range points {
		changeStr := "-"
		if p.Change > 0 {
			changeStr = fmt.Sprintf("+%.1f%% 🔺", p.Change)
		} else if p.Change < 0 {
			changeStr = fmt.Sprintf("%.1f%% 🔻", p.Change)
		} else if p.Change == 0 {
			changeStr = "0%"
		}

		// Truncate item name if too long
		item := p.ItemName
		if len(item) > 30 {
			item = item[:27] + "..."
		}

		fmt.Fprintf(tw, "%s\t%s\t%s\t%.2f\t%s\n",
			p.Date.Format("2006-01-02"),
			p.Restaurant,
			item,
			p.UnitPrice,
			changeStr,
		)
	}
	tw.Flush()
}

func InflationSummaryTable(w io.Writer, summaries []InflationSummary) {
	if len(summaries) == 0 {
		fmt.Fprintln(w, "No data available for top items.")
		return
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ITEM\tFIRST SEEN\tFIRST PRICE\tLAST PRICE\tCHANGE")
	fmt.Fprintln(tw, "----\t----------\t-----------\t----------\t------")

	for _, s := range summaries {
		changeStr := "0%"
		if s.TotalChange > 0 {
			changeStr = fmt.Sprintf("+%.1f%% 🔺", s.TotalChange)
		} else if s.TotalChange < 0 {
			changeStr = fmt.Sprintf("%.1f%% 🔻", s.TotalChange)
		}

		item := s.ItemName
		if len(item) > 30 {
			item = item[:27] + "..."
		}

		fmt.Fprintf(tw, "%s\t%s\t%.2f\t%.2f\t%s\n",
			item,
			s.FirstSeen,
			s.FirstPrice,
			s.LastPrice,
			changeStr,
		)
	}
	tw.Flush()
}
