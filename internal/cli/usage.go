package cli

import (
	"fmt"
	"io"
)

func PrintUsage(w io.Writer) {
	fmt.Fprint(w, `zocli - unofficial Zomato order tracker

Usage:
  zocli <command> [options]

Commands:
  auth       Save your Zomato web cookie (or run auth login)
  sync       Fetch orders and store locally
  orders     List stored orders
  stats      Show totals and grouping summaries
  config     Show config and data paths
  version    Print version
  help       Show usage

Examples:
  zocli auth login
  zocli auth import --browser chrome
  zocli auth logout
  zocli auth status
  zocli auth --cookie "<cookie header>"
  zocli sync --mock
  zocli orders
  zocli stats --group month
`)
}
