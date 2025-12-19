package format

import (
	"fmt"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/maheshrijal/zocli/internal/stats"
)

func StatsSummary(w io.Writer, summary stats.Summary) {
	headers := []string{"Orders", "Total", "Average", "Earliest", "Latest"}
	row := []string{
		fmt.Sprintf("%d", summary.Count),
		formatCurrency(summary.Currency, summary.Total),
		formatCurrency(summary.Currency, summary.Average),
		formatDate(summary.Earliest),
		formatDate(summary.Latest),
	}
	alignRight := []bool{true, true, true, false, false}
	writeBoxTable(w, headers, [][]string{row}, alignRight)
}

func StatsGroups(w io.Writer, groups []stats.Group, currency string) {
	if len(groups) == 0 {
		fmt.Fprintln(w, "No groups to display.")
		return
	}
	headers := []string{"Period", "Orders", "Total", "Average"}
	rows := make([][]string, 0, len(groups))
	for _, group := range groups {
		rows = append(rows, []string{
			group.Key,
			fmt.Sprintf("%d", group.Count),
			formatCurrency(currency, group.Total),
			formatCurrency(currency, group.Average),
		})
	}
	alignRight := []bool{false, true, true, true}
	writeBoxTable(w, headers, rows, alignRight)
}

func writeBoxTable(w io.Writer, headers []string, rows [][]string, alignRight []bool) {
	columns := len(headers)
	widths := make([]int, columns)
	for i, header := range headers {
		widths[i] = runeLen(header)
	}
	for _, row := range rows {
		for i := 0; i < columns && i < len(row); i++ {
			if n := runeLen(row[i]); n > widths[i] {
				widths[i] = n
			}
		}
	}

	border := buildBorder(widths)
	fmt.Fprintln(w, border)
	fmt.Fprintln(w, buildRow(headers, widths, alignRight))
	fmt.Fprintln(w, border)
	for _, row := range rows {
		fmt.Fprintln(w, buildRow(row, widths, alignRight))
	}
	fmt.Fprintln(w, border)
}

func buildBorder(widths []int) string {
	var b strings.Builder
	b.WriteString("+")
	for _, width := range widths {
		b.WriteString(strings.Repeat("-", width+2))
		b.WriteString("+")
	}
	return b.String()
}

func buildRow(row []string, widths []int, alignRight []bool) string {
	var b strings.Builder
	b.WriteString("|")
	for i, width := range widths {
		cell := ""
		if i < len(row) {
			cell = row[i]
		}
		right := false
		if i < len(alignRight) {
			right = alignRight[i]
		}
		b.WriteString(" ")
		b.WriteString(pad(cell, width, right))
		b.WriteString(" |")
	}
	return b.String()
}

func pad(value string, width int, right bool) string {
	length := runeLen(value)
	if length >= width {
		return value
	}
	padding := strings.Repeat(" ", width-length)
	if right {
		return padding + value
	}
	return value + padding
}

func runeLen(value string) int {
	return utf8.RuneCountInString(value)
}

func formatCurrency(currency string, value float64) string {
	if currency == "" {
		return fmt.Sprintf("%.2f", value)
	}
	return fmt.Sprintf("%s%.2f", currency, value)
}

func formatDate(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("2006-01-02")
}
