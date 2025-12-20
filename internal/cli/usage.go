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
  auth       Log in / import cookies
  sync       Fetch orders and store locally
  dash       Interactive dashboard (TUI)
  orders     List stored orders
  stats      Summarize spend
  inflation  Track unit price history
  config     Show config and data paths
  version    Print version
  help       Help for a command

Try:
  zocli help auth
  zocli auth login
`)
}

func PrintAuthUsage(w io.Writer) {
	fmt.Fprint(w, `auth - save or manage your Zomato session cookie

Usage:
  zocli auth login [--headless] [--browser chrome|chromium|brave|edge|helium|vivaldi] [--profile "Default"] [--user-data-dir PATH] [--browser-path PATH]
  zocli auth import [--browser chrome|chromium|brave|edge|helium|vivaldi] [--profile "Default"] [--user-data-dir PATH] [--browser-path PATH]
  zocli auth logout
  zocli auth status [--offline]
  zocli auth --cookie "<cookie header>"
  zocli auth --cookie-file PATH

Examples:
  zocli auth login
  zocli auth login --browser helium --profile "Default"
  zocli auth import --browser chrome --profile "Default"
  zocli auth logout
`)
}

func PrintAuthLoginUsage(w io.Writer) {
	fmt.Fprint(w, `zocli auth login

Usage:
  zocli auth login [--browser chrome|chromium|brave|edge|helium|vivaldi] [--profile "Default"]

Options:
  --browser       Browser profile to use (default: chrome)
  --profile       Profile dir name (default: Default)
  --user-data-dir Advanced: path to browser user data dir
  --browser-path  Advanced: path to browser executable
  --headless      Advanced: run headless (not recommended for login)

Examples:
  zocli auth login
  zocli auth login --browser brave --profile "Profile 1"
`)
}

func PrintAuthImportUsage(w io.Writer) {
	fmt.Fprint(w, `zocli auth import

Usage:
  zocli auth import [--browser chrome|chromium|brave|edge|helium|vivaldi] [--profile "Default"]

Options:
  --browser       Browser profile to read (default: chrome)
  --profile       Profile dir name (default: Default)
  --user-data-dir Advanced: path to browser user data dir
  --browser-path  Advanced: path to browser executable
  --headless      Advanced: run headless (default: true)

Examples:
  zocli auth import
  zocli auth import --browser helium
`)
}

func PrintAuthStatusUsage(w io.Writer) {
	fmt.Fprint(w, `zocli auth status

Usage:
  zocli auth status [--offline]
`)
}

func PrintAuthLogoutUsage(w io.Writer) {
	fmt.Fprint(w, `zocli auth logout

Usage:
  zocli auth logout
`)
}

func PrintSyncUsage(w io.Writer) {
	fmt.Fprint(w, `zocli sync

Usage:
  zocli sync [--mock]
`)
}

func PrintOrdersUsage(w io.Writer) {
	fmt.Fprint(w, `zocli orders

Usage:
  zocli orders [--limit 20]
`)
}

func PrintStatsUsage(w io.Writer) {
	fmt.Fprint(w, `zocli stats

Usage:
  zocli stats [--group month|year|none] [--view basic|spend|patterns|personal|all] [--top 5]
`)
}

func PrintConfigUsage(w io.Writer) {
	fmt.Fprint(w, `zocli config

Usage:
  zocli config
`)
}

func PrintCommandUsage(w io.Writer, cmd string) bool {
	switch cmd {
	case "auth":
		PrintAuthUsage(w)
	case "sync":
		PrintSyncUsage(w)
	case "orders":
		PrintOrdersUsage(w)
	case "stats":
		PrintStatsUsage(w)
	case "config":
		PrintConfigUsage(w)
	default:
		return false
	}
	return true
}
