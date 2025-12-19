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
		must(runConfig())
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
		}
	}

	fs := flag.NewFlagSet("auth", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	cookie := fs.String("cookie", "", "Zomato Cookie header value")
	cookieFile := fs.String("cookie-file", "", "Path to a file containing the Cookie header value")
	if err := fs.Parse(args); err != nil {
		return err
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
	browser := fs.String("browser", "chrome", "Browser profile to use (chrome, chromium, brave, edge)")
	userDataDir := fs.String("user-data-dir", "", "Path to browser user data dir (uses default if profile is set)")
	profile := fs.String("profile", "", "Browser profile directory name (e.g. Default, Profile 1)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfgPath, err := config.DefaultPath()
	if err != nil {
		return err
	}

	return auth.LoginAndSaveCookieWithOptions(context.Background(), cfgPath, auth.LoginOptions{
		Headless:    *headless,
		Browser:     *browser,
		UserDataDir: *userDataDir,
		ProfileDir:  *profile,
	})
}

func runAuthStatus(args []string) error {
	fs := flag.NewFlagSet("auth status", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	offline := fs.Bool("offline", false, "Only check if a cookie is saved; skip network validation")
	if err := fs.Parse(args); err != nil {
		return err
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
	if err := fs.Parse(args); err != nil {
		return err
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
	browser := fs.String("browser", "chrome", "Browser profile to read (chrome, chromium, brave, edge)")
	userDataDir := fs.String("user-data-dir", "", "Path to browser user data dir (default for --browser)")
	profile := fs.String("profile", "Default", "Browser profile directory name")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfgPath, err := config.DefaultPath()
	if err != nil {
		return err
	}

	return auth.ImportFromBrowser(context.Background(), cfgPath, auth.LoginOptions{
		Headless:    *headless,
		Browser:     *browser,
		UserDataDir: *userDataDir,
		ProfileDir:  *profile,
	})
}

func runSync(args []string) error {
	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	mock := fs.Bool("mock", false, "Use sample data instead of hitting Zomato")
	if err := fs.Parse(args); err != nil {
		return err
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
	orders, err := client.FetchOrders(context.Background())
	if err != nil {
		return err
	}
	if err := st.Save(orders); err != nil {
		return err
	}
	fmt.Printf("Stored %d orders in %s\n", len(orders), storePath)
	return nil
}

func runOrders(args []string) error {
	fs := flag.NewFlagSet("orders", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	limit := fs.Int("limit", 20, "Max orders to print")
	if err := fs.Parse(args); err != nil {
		return err
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
	fs := flag.NewFlagSet("stats", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	group := fs.String("group", "month", "Group by: none, month, year")
	if err := fs.Parse(args); err != nil {
		return err
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

	summary := stats.ComputeSummary(orders)
	groups, err := stats.GroupOrders(orders, *group)
	if err != nil {
		return err
	}
	format.StatsGroups(os.Stdout, groups, summary.Currency)
	fmt.Fprintln(os.Stdout)
	format.StatsSummary(os.Stdout, summary)
	return nil
}

func runConfig() error {
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
