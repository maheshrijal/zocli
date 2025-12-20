package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/maheshrijal/zocli/internal/auth"
	"github.com/maheshrijal/zocli/internal/cli"
	"github.com/maheshrijal/zocli/internal/config"
	"github.com/maheshrijal/zocli/internal/format"
	"github.com/maheshrijal/zocli/internal/sample"
	"github.com/maheshrijal/zocli/internal/stats"
	"github.com/maheshrijal/zocli/internal/store"
	"github.com/maheshrijal/zocli/internal/zomato"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		cli.PrintUsage(os.Stdout)
		os.Exit(1)
	}

	if len(os.Args) > 2 && (os.Args[2] == "help" || os.Args[2] == "-h" || os.Args[2] == "--help") {
		if !cli.PrintCommandUsage(os.Stdout, os.Args[1]) {
			fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
			cli.PrintUsage(os.Stderr)
			os.Exit(1)
		}
		return
	}

	if os.Args[1] == "help" {
		if len(os.Args) > 2 {
			if !cli.PrintCommandUsage(os.Stdout, os.Args[2]) {
				fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[2])
				cli.PrintUsage(os.Stderr)
				os.Exit(1)
			}
			return
		}
		cli.PrintUsage(os.Stdout)
		return
	}

	switch os.Args[1] {
	case "help", "-h", "--help":
		cli.PrintUsage(os.Stdout)
	case "version", "-v", "--version":
		fmt.Println(version)
	case "auth":
		must(runAuth(os.Args[2:]))
	case "sync":
		must(runSync(os.Args[2:]))
	case "orders":
		must(runOrders(os.Args[2:]))
	case "stats":
		must(runStats(os.Args[2:]))
	case "config":
		must(runConfig(os.Args[2:]))
	case "inflation":
		must(runInflation(os.Args[2:]))
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		cli.PrintUsage(os.Stderr)
		os.Exit(1)
	}
}

func runAuth(args []string) error {
	if len(args) > 0 {
		switch args[0] {
		case "login":
			return runAuthLogin(args[1:])
		case "import":
			return runAuthImport(args[1:])
		case "logout":
			return runAuthLogout(args[1:])
		case "status":
			return runAuthStatus(args[1:])
		case "help", "-h", "--help":
			cli.PrintAuthUsage(os.Stdout)
			return nil
		default:
			if !strings.HasPrefix(args[0], "-") {
				cli.PrintAuthUsage(os.Stderr)
				return fmt.Errorf("unknown auth command: %s", args[0])
			}
		}
	}

	if len(args) == 0 {
		cli.PrintAuthUsage(os.Stdout)
		return nil
	}

	fs := flag.NewFlagSet("auth", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	cookie := fs.String("cookie", "", "Zomato Cookie header value")
	cookieFile := fs.String("cookie-file", "", "Path to a file containing the Cookie header value")
	fs.Usage = func() {
		cli.PrintAuthUsage(os.Stderr)
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if extra := fs.Args(); len(extra) > 0 {
		cli.PrintAuthUsage(os.Stderr)
		return fmt.Errorf("unknown arguments: %s", strings.Join(extra, " "))
	}

	value := strings.TrimSpace(*cookie)
	if value == "" && *cookieFile != "" {
		data, err := os.ReadFile(*cookieFile)
		if err != nil {
			return err
		}
		value = strings.TrimSpace(string(data))
	}
	if value == "" {
		return errors.New("cookie value is required; use 'auth login' or --cookie/--cookie-file")
	}

	cfgPath, err := config.DefaultPath()
	if err != nil {
		return err
	}

	if err := config.Save(cfgPath, config.Config{Cookie: value}); err != nil {
		return err
	}

	fmt.Printf("Saved cookie to %s\n", cfgPath)
	return nil
}

func runAuthLogin(args []string) error {
	fs := flag.NewFlagSet("auth login", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	headless := fs.Bool("headless", false, "Run Chrome in headless mode (not recommended for login)")
	browser := fs.String("browser", "chrome", "Browser profile to use (chrome, chromium, brave, edge, helium, vivaldi)")
	browserPath := fs.String("browser-path", "", "Path to browser executable (optional)")
	userDataDir := fs.String("user-data-dir", "", "Path to browser user data dir (uses default if profile is set)")
	profile := fs.String("profile", "", "Browser profile directory name (e.g. Default, Profile 1)")
	fs.Usage = func() {
		cli.PrintAuthLoginUsage(os.Stderr)
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if extra := fs.Args(); len(extra) > 0 {
		cli.PrintAuthLoginUsage(os.Stderr)
		return fmt.Errorf("unknown arguments: %s", strings.Join(extra, " "))
	}

	var browserSet, userDataDirSet, profileSet, browserPathSet bool
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "browser":
			browserSet = true
		case "user-data-dir":
			userDataDirSet = true
		case "profile":
			profileSet = true
		case "browser-path":
			browserPathSet = true
		}
	})
	if *profile == "" && (browserSet || userDataDirSet || profileSet || browserPathSet) {
		*profile = "Default"
	}

	cfgPath, err := config.DefaultPath()
	if err != nil {
		return err
	}

	return auth.LoginAndSaveCookieWithOptions(context.Background(), cfgPath, auth.LoginOptions{
		Headless:    *headless,
		Browser:     *browser,
		BrowserPath: *browserPath,
		UserDataDir: *userDataDir,
		ProfileDir:  *profile,
	})
}

func runAuthStatus(args []string) error {
	fs := flag.NewFlagSet("auth status", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	offline := fs.Bool("offline", false, "Only check if a cookie is saved; skip network validation")
	fs.Usage = func() {
		cli.PrintAuthStatusUsage(os.Stderr)
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if extra := fs.Args(); len(extra) > 0 {
		cli.PrintAuthStatusUsage(os.Stderr)
		return fmt.Errorf("unknown arguments: %s", strings.Join(extra, " "))
	}

	cfgPath, err := config.DefaultPath()
	if err != nil {
		return err
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Not logged in (no saved cookie). Run `zocli auth login`.")
			return nil
		}
		return err
	}
	if strings.TrimSpace(cfg.Cookie) == "" {
		fmt.Println("Not logged in (empty cookie). Run `zocli auth login`.")
		return nil
	}
	if *offline {
		fmt.Println("Saved cookie found. Use `zocli auth status` to validate it.")
		return nil
	}

	client := zomato.NewClient(cfg.Cookie)
	ok, err := client.CheckAuth(context.Background())
	if err != nil {
		return err
	}
	if ok {
		fmt.Println("Logged in.")
		return nil
	}
	fmt.Println("Not logged in (cookie invalid or expired). Run `zocli auth login`.")
	return nil
}

func runAuthLogout(args []string) error {
	fs := flag.NewFlagSet("auth logout", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		cli.PrintAuthLogoutUsage(os.Stderr)
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if extra := fs.Args(); len(extra) > 0 {
		cli.PrintAuthLogoutUsage(os.Stderr)
		return fmt.Errorf("unknown arguments: %s", strings.Join(extra, " "))
	}

	cfgPath, err := config.DefaultPath()
	if err != nil {
		return err
	}

	if err := config.Save(cfgPath, config.Config{Cookie: ""}); err != nil {
		return err
	}
	fmt.Println("Logged out (saved cookie cleared).")
	return nil
}

func runAuthImport(args []string) error {
	fs := flag.NewFlagSet("auth import", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	headless := fs.Bool("headless", true, "Run Chrome in headless mode (default true)")
	browser := fs.String("browser", "chrome", "Browser profile to read (chrome, chromium, brave, edge, helium, vivaldi)")
	browserPath := fs.String("browser-path", "", "Path to browser executable (optional)")
	userDataDir := fs.String("user-data-dir", "", "Path to browser user data dir (default for --browser)")
	profile := fs.String("profile", "Default", "Browser profile directory name")
	fs.Usage = func() {
		cli.PrintAuthImportUsage(os.Stderr)
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if extra := fs.Args(); len(extra) > 0 {
		cli.PrintAuthImportUsage(os.Stderr)
		return fmt.Errorf("unknown arguments: %s", strings.Join(extra, " "))
	}
	if *profile == "" {
		*profile = "Default"
	}

	fmt.Println("If import fails, close the browser and try again.")

	cfgPath, err := config.DefaultPath()
	if err != nil {
		return err
	}

	return auth.ImportFromBrowser(context.Background(), cfgPath, auth.LoginOptions{
		Headless:    *headless,
		Browser:     *browser,
		BrowserPath: *browserPath,
		UserDataDir: *userDataDir,
		ProfileDir:  *profile,
	})
}

func runSync(args []string) error {
	if len(args) > 0 && (args[0] == "help" || args[0] == "-h" || args[0] == "--help") {
		cli.PrintSyncUsage(os.Stdout)
		return nil
	}
	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		cli.PrintSyncUsage(os.Stderr)
	}
	mock := fs.Bool("mock", false, "Use sample data instead of hitting Zomato")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if extra := fs.Args(); len(extra) > 0 {
		cli.PrintSyncUsage(os.Stderr)
		return fmt.Errorf("unknown arguments: %s", strings.Join(extra, " "))
	}

	storePath, err := store.DefaultPath()
	if err != nil {
		return err
	}
	st, err := store.New(storePath)
	if err != nil {
		return err
	}

	if *mock {
		orders, err := sample.Orders()
		if err != nil {
			return err
		}
		if err := st.Save(orders); err != nil {
			return err
		}
		fmt.Printf("Stored %d sample orders in %s\n", len(orders), storePath)
		return nil
	}

	cfgPath, err := config.DefaultPath()
	if err != nil {
		return err
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if strings.TrimSpace(cfg.Cookie) == "" {
		return errors.New("no cookie found; run 'zocli auth login' first")
	}

	client := zomato.NewClient(cfg.Cookie)
	terminal := isTerminal(os.Stdout)
	progress := func(p zomato.FetchProgress) {
		total := "?"
		if p.TotalPages > 0 {
			total = fmt.Sprintf("%d", p.TotalPages)
		}
		if terminal {
			fmt.Fprintf(os.Stdout, "\rFetched page %d/%s (orders: %d)", p.Page, total, p.TotalOrders)
		} else {
			fmt.Fprintf(os.Stdout, "Fetched page %d/%s (orders: %d)\n", p.Page, total, p.TotalOrders)
		}
	}
	orders, err := client.FetchOrdersWithProgress(context.Background(), progress)
	if err != nil {
		return err
	}
	if terminal {
		fmt.Fprintln(os.Stdout)
	}
	if err := st.Save(orders); err != nil {
		return err
	}
	fmt.Printf("Stored %d orders in %s\n", len(orders), storePath)
	return nil
}

func runOrders(args []string) error {
	if len(args) > 0 && (args[0] == "help" || args[0] == "-h" || args[0] == "--help") {
		cli.PrintOrdersUsage(os.Stdout)
		return nil
	}
	fs := flag.NewFlagSet("orders", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	limit := fs.Int("limit", 20, "Max orders to print")
	fs.Usage = func() {
		cli.PrintOrdersUsage(os.Stderr)
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if extra := fs.Args(); len(extra) > 0 {
		cli.PrintOrdersUsage(os.Stderr)
		return fmt.Errorf("unknown arguments: %s", strings.Join(extra, " "))
	}

	storePath, err := store.DefaultPath()
	if err != nil {
		return err
	}
	st, err := store.New(storePath)
	if err != nil {
		return err
	}
	orders, err := st.Load()
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("no stored orders yet; run 'zocli sync' first")
		}
		return err
	}
	if *limit > 0 && len(orders) > *limit {
		orders = orders[:*limit]
	}
	format.OrdersTable(os.Stdout, orders)
	return nil
}

func runStats(args []string) error {
	if len(args) > 0 && (args[0] == "help" || args[0] == "-h" || args[0] == "--help") {
		cli.PrintStatsUsage(os.Stdout)
		return nil
	}
	fs := flag.NewFlagSet("stats", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	group := fs.String("group", "month", "Group by: none, month, year")
	view := fs.String("view", "basic", "View: basic, spend, patterns, personal, all")
	top := fs.Int("top", 5, "Top N restaurants/items")
	fs.Usage = func() {
		cli.PrintStatsUsage(os.Stderr)
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if extra := fs.Args(); len(extra) > 0 {
		cli.PrintStatsUsage(os.Stderr)
		return fmt.Errorf("unknown arguments: %s", strings.Join(extra, " "))
	}

	storePath, err := store.DefaultPath()
	if err != nil {
		return err
	}
	st, err := store.New(storePath)
	if err != nil {
		return err
	}
	orders, err := st.Load()
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("no stored orders yet; run 'zocli sync' first")
		}
		return err
	}

	viewKey := strings.ToLower(strings.TrimSpace(*view))
	if viewKey == "" {
		viewKey = "basic"
	}
	switch viewKey {
	case "basic", "all", "spend", "patterns", "personal":
	default:
		cli.PrintStatsUsage(os.Stderr)
		return fmt.Errorf("unknown view: %s", viewKey)
	}

	summary := stats.ComputeSummary(orders)

	showBasic := viewKey == "basic"
	showSpend := viewKey == "all" || viewKey == "spend"
	showPatterns := viewKey == "all" || viewKey == "patterns"
	showPersonal := viewKey == "all" || viewKey == "personal"

	if showBasic || showSpend {
		groups, err := stats.GroupOrders(orders, *group)
		if err != nil {
			return err
		}
		format.StatsGroups(os.Stdout, groups, summary.Currency)
		fmt.Fprintln(os.Stdout)
		format.StatsSummary(os.Stdout, summary)
	}

	if showBasic {
		fmt.Fprintln(os.Stdout)
		fmt.Fprintln(os.Stdout, "More views: zocli stats --view spend | patterns | personal")
		return nil
	}

	if showSpend {
		fmt.Fprintln(os.Stdout)
		fmt.Fprintln(os.Stdout, "Spend by weekday")
		format.StatsSpendByWeekday(os.Stdout, stats.SpendByWeekday(orders), summary.Currency)
		fmt.Fprintln(os.Stdout)
	}

	if showPatterns {
		fmt.Fprintln(os.Stdout, "Ordering patterns")
		format.StatsWeekdayOrders(os.Stdout, stats.OrdersByWeekday(orders))
		fmt.Fprintln(os.Stdout)
		format.StatsTimeWindows(os.Stdout, stats.OrdersByTimeWindow(orders))
		fmt.Fprintln(os.Stdout)
	}

	if showPersonal {
		fmt.Fprintln(os.Stdout, "Personal stats")
		format.StatsTopList(os.Stdout, "Restaurant", stats.TopRestaurants(orders, *top))
		fmt.Fprintln(os.Stdout)
		items := stats.TopItems(orders, *top)
		if len(items) == 0 {
			fmt.Fprintln(os.Stdout, "No item data to display.")
		} else {
			format.StatsTopList(os.Stdout, "Item", items)
		}
		fmt.Fprintln(os.Stdout)
	}
	return nil
}

func runConfig(args []string) error {
	if len(args) > 0 {
		if args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
			cli.PrintConfigUsage(os.Stdout)
			return nil
		}
		cli.PrintConfigUsage(os.Stderr)
		return fmt.Errorf("unknown arguments: %s", strings.Join(args, " "))
	}
	cfgPath, err := config.DefaultPath()
	if err != nil {
		return err
	}
	storePath, err := store.DefaultPath()
	if err != nil {
		return err
	}

	fmt.Printf("Config: %s\n", cfgPath)
	fmt.Printf("Orders: %s\n", storePath)
	return nil
}

func runInflation(args []string) error {
	storePath, err := store.DefaultPath()
	if err != nil {
		return err
	}
	st, err := store.New(storePath)
	if err != nil {
		return err
	}
	orders, err := st.Load()
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("no stored orders yet; run 'zocli sync' first")
		}
		return err
	}

	// Case 1: Show top 5 items summary if no args
	if len(args) == 0 {
		trends := stats.FindTopInflationTrends(orders, 5)
		var summaries []format.InflationSummary

		for _, t := range trends {
			summaries = append(summaries, format.InflationSummary{
				ItemName:    t.Key, // Shows "Restaurant - Item"
				FirstSeen:   t.FirstSeen.Format("2006-01-02"),
				FirstPrice:  t.FirstPrice,
				LastPrice:   t.LastPrice,
				TotalChange: t.TotalChange,
			})
		}
		
		fmt.Println("Top Inflation Trends (Restaurant specific)")
		format.InflationSummaryTable(os.Stdout, summaries)
		fmt.Println("\nTip: Run 'zocli inflation <item name>' for detailed history.")
		return nil
	}
	
	query := strings.Join(args, " ")
	if strings.HasPrefix(query, "-") {
		if query == "--help" || query == "-h" {
			fmt.Fprintln(os.Stdout, "Usage: zocli inflation [item name]")
			fmt.Fprintln(os.Stdout, "  No args: Shows trend summary for top 5 most ordered items.")
			fmt.Fprintln(os.Stdout, "  Arg: Shows detailed price history for matching items.")
			return nil
		}
	}

	points, err := stats.CalculateInflation(orders, query)
	if err != nil {
		return err
	}

	format.InflationTable(os.Stdout, points)
	return nil
}

func must(err error) {
	if err == nil {
		return
	}
	if errors.Is(err, flag.ErrHelp) {
		os.Exit(0)
	}
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}
