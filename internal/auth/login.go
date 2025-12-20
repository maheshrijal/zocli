package auth

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/storage"
	"github.com/chromedp/chromedp"
	"github.com/maheshrijal/zocli/internal/config"
)

type LoginOptions struct {
	Headless    bool
	Browser     string
	BrowserPath string
	UserDataDir string
	ProfileDir  string
	SkipWait    bool
}

func LoginAndSaveCookie(ctx context.Context, cfgPath string, headless bool) error {
	return LoginAndSaveCookieWithOptions(ctx, cfgPath, LoginOptions{
		Headless: headless,
		Browser:  "chrome",
	})
}

func LoginAndSaveCookieWithOptions(ctx context.Context, cfgPath string, opts LoginOptions) error {
	return captureCookie(ctx, cfgPath, opts)
}

func ImportFromBrowser(ctx context.Context, cfgPath string, opts LoginOptions) error {
	opts.SkipWait = true
	if opts.Browser == "" {
		opts.Browser = "chrome"
	}
	if opts.ProfileDir == "" {
		opts.ProfileDir = "Default"
	}
	if opts.UserDataDir == "" {
		if path, err := defaultUserDataDir(opts.Browser); err == nil {
			opts.UserDataDir = path
		}
	}
	return captureCookie(ctx, cfgPath, opts)
}

func captureCookie(ctx context.Context, cfgPath string, opts LoginOptions) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	userDataDir := opts.UserDataDir
	if userDataDir == "" && opts.ProfileDir != "" {
		path, err := defaultUserDataDir(opts.Browser)
		if err != nil {
			return err
		}
		userDataDir = path
	}

	if userDataDir != "" {
		resolved, err := resolveProfile(userDataDir, opts.ProfileDir)
		if err != nil {
			return err
		}
		opts.ProfileDir = resolved
	}

	browserPath := opts.BrowserPath
	if browserPath == "" {
		if path, ok := defaultBrowserPath(opts.Browser); ok {
			browserPath = path
		}
	}
	if browserPath == "" && strings.EqualFold(opts.Browser, "helium") {
		fmt.Println("Helium app not found; using default Chrome. If import fails, pass --browser-path.")
	}
	if browserPath == "" && !strings.EqualFold(opts.Browser, "chrome") && !strings.EqualFold(opts.Browser, "google-chrome") {
		fmt.Printf("Browser %q not found at default path; using Chrome. If import fails, pass --browser-path.\n", opts.Browser)
	}

	execOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", opts.Headless),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
	)
	if browserPath != "" {
		execOpts = append(execOpts, chromedp.ExecPath(browserPath))
	}
	if userDataDir != "" {
		execOpts = append(execOpts, chromedp.UserDataDir(userDataDir))
	}
	if opts.ProfileDir != "" {
		execOpts = append(execOpts, chromedp.Flag("profile-directory", opts.ProfileDir))
	}

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, execOpts...)
	defer cancel()

	browserCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	if opts.SkipWait {
		if userDataDir != "" {
			fmt.Printf("Reading cookies from %s (%s)...\n", userDataDir, opts.ProfileDir)
		} else {
			fmt.Println("Reading cookies from the selected browser profile...")
		}
		if browserPath != "" {
			fmt.Printf("Launching %s to read cookies...\n", filepath.Base(browserPath))
		}
		if err := ensureProfileHasCookies(userDataDir, opts.ProfileDir); err != nil {
			return err
		}
	} else {
		if browserPath != "" {
			fmt.Printf("Launching %s...\n", filepath.Base(browserPath))
		}
		if err := runChromedp(browserCtx,
			chromedp.Navigate("https://www.zomato.com/restaurants"),
		); err != nil {
			return err
		}

		if err := runChromedp(browserCtx,
			chromedp.Evaluate(`(() => {
				const el = Array.from(document.querySelectorAll('a, button, div, span'))
					.find(e => (e.textContent || '').trim().toLowerCase() === 'log in');
				if (!el) return false;
				el.click();
				return true;
			})()`, nil),
			chromedp.Sleep(500*time.Millisecond),
		); err != nil {
			return err
		}

		fmt.Println("A browser window should be open.")
		fmt.Println("If the login modal isn't open, click \"Log in\" on the page.")
		fmt.Println("Log in to Zomato there. This will continue once login is detected.")

		if err := waitForLogin(browserCtx); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Println("Login timed out. Press Enter to capture cookies anyway, or Ctrl+C to cancel.")
				if _, err := bufio.NewReader(os.Stdin).ReadString('\n'); err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	if opts.SkipWait {
		if err := runChromedp(browserCtx, chromedp.Navigate("about:blank")); err != nil {
			return err
		}
	}

	var cookies []*network.Cookie
	if err := runChromedp(browserCtx,
		network.Enable(),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			cookies, err = network.GetCookies().WithURLs([]string{
				"https://www.zomato.com",
				"https://zomato.com",
			}).Do(ctx)
			return err
		}),
	); err != nil {
		return err
	}
	if len(cookies) == 0 {
		if err := runChromedp(browserCtx, chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			cookies, err = storage.GetCookies().Do(ctx)
			return err
		})); err != nil {
			return err
		}
	}

	cookieHeader := buildCookieHeader(cookies)
	if cookieHeader == "" {
		if opts.SkipWait {
			return fmt.Errorf("no zomato cookies found; make sure the selected profile is logged in and %s is closed. If this browser doesn't persist cookies, use `zocli auth login --browser %s` instead", displayBrowserName(opts.Browser), opts.Browser)
		}
		return errors.New("no zomato cookies found; make sure you logged in in the opened browser")
	}

	if err := config.Save(cfgPath, config.Config{Cookie: cookieHeader}); err != nil {
		return err
	}

	fmt.Printf("Saved cookie to %s\n", cfgPath)
	return nil
}

func waitForLogin(ctx context.Context) error {
	ticker := time.NewTicker(800 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			ok, err := isLoggedIn(ctx)
			if err != nil {
				return err
			}
			if ok {
				return nil
			}
		}
	}
}

func isLoggedIn(ctx context.Context) (bool, error) {
	var loggedIn bool
	err := chromedp.Run(ctx, chromedp.Evaluate(`(() => {
		const norm = (s) => (s || '').trim().toLowerCase();
		const hasLogin = Array.from(document.querySelectorAll('a, button, div, span'))
			.some(e => norm(e.textContent) === 'log in');
		return !hasLogin;
	})()`, &loggedIn))
	return loggedIn, err
}

func buildCookieHeader(cookies []*network.Cookie) string {
	pairs := make([]string, 0, len(cookies))
	for _, cookie := range cookies {
		if cookie == nil {
			continue
		}
		domain := strings.ToLower(cookie.Domain)
		if !strings.Contains(domain, "zomato") {
			continue
		}
		if cookie.Name == "" {
			continue
		}
		pairs = append(pairs, fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
	}
	if len(pairs) == 0 {
		return ""
	}
	sort.Strings(pairs)
	return strings.Join(pairs, "; ")
}

func defaultUserDataDir(browser string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	name := strings.ToLower(strings.TrimSpace(browser))
	if name == "" || name == "chrome" || name == "google-chrome" {
		name = "chrome"
	}

	switch runtime.GOOS {
	case "darwin":
		switch name {
		case "chrome":
			return filepath.Join(home, "Library/Application Support/Google/Chrome"), nil
		case "chromium":
			return filepath.Join(home, "Library/Application Support/Chromium"), nil
		case "helium":
			preferred := filepath.Join(home, "Library/Application Support/net.imput.helium")
			if dirExists(preferred) {
				return preferred, nil
			}
			return filepath.Join(home, "Library/Application Support/Helium"), nil
		case "brave":
			return filepath.Join(home, "Library/Application Support/BraveSoftware/Brave-Browser"), nil
		case "edge":
			return filepath.Join(home, "Library/Application Support/Microsoft Edge"), nil
		case "vivaldi":
			return filepath.Join(home, "Library/Application Support/Vivaldi"), nil
		}
	case "linux":
		base := filepath.Join(home, ".config")
		switch name {
		case "chrome":
			return filepath.Join(base, "google-chrome"), nil
		case "chromium":
			return filepath.Join(base, "chromium"), nil
		case "brave":
			return filepath.Join(base, "BraveSoftware/Brave-Browser"), nil
		case "edge":
			return filepath.Join(base, "microsoft-edge"), nil
		}
	case "windows":
		local := os.Getenv("LOCALAPPDATA")
		if local == "" {
			return "", errors.New("LOCALAPPDATA not set")
		}
		switch name {
		case "chrome":
			return filepath.Join(local, "Google", "Chrome", "User Data"), nil
		case "chromium":
			return filepath.Join(local, "Chromium", "User Data"), nil
		case "brave":
			return filepath.Join(local, "BraveSoftware", "Brave-Browser", "User Data"), nil
		case "edge":
			return filepath.Join(local, "Microsoft", "Edge", "User Data"), nil
		}
	}

	return "", fmt.Errorf("unknown browser %q; use --user-data-dir", browser)
}

func runChromedp(ctx context.Context, actions ...chromedp.Action) error {
	if err := chromedp.Run(ctx, actions...); err != nil {
		if isProfileLockedError(err) {
			return errors.New("browser profile is in use. Close the browser and retry")
		}
		return err
	}
	return nil
}

func isProfileLockedError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "SingletonLock") || strings.Contains(msg, "ProcessSingleton")
}

func displayBrowserName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "the browser"
	}
	return name
}

func defaultBrowserPath(browser string) (string, bool) {
	if runtime.GOOS != "darwin" {
		return "", false
	}
	name := strings.ToLower(strings.TrimSpace(browser))
	if name == "" || name == "chrome" || name == "google-chrome" {
		return "", false
	}
	var path string
	switch name {
	case "helium":
		path = "/Applications/Helium.app/Contents/MacOS/Helium"
	case "brave":
		path = "/Applications/Brave Browser.app/Contents/MacOS/Brave Browser"
	case "chromium":
		path = "/Applications/Chromium.app/Contents/MacOS/Chromium"
	case "edge":
		path = "/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge"
	case "vivaldi":
		path = "/Applications/Vivaldi.app/Contents/MacOS/Vivaldi"
	}
	if path == "" {
		return "", false
	}
	if fileExists(path) {
		return path, true
	}
	return "", false
}

func ensureProfileHasCookies(userDataDir, profile string) error {
	if userDataDir == "" || profile == "" {
		return nil
	}
	if _, ok := findCookieDB(userDataDir, profile); ok {
		return nil
	}
	profiles := listProfiles(userDataDir)
	if len(profiles) == 0 {
		return fmt.Errorf("no browser profiles found at %s", userDataDir)
	}
	return fmt.Errorf("no cookies found for profile %q in %s. Available profiles: %s", profile, userDataDir, strings.Join(profiles, ", "))
}

func resolveProfile(userDataDir, profile string) (string, error) {
	if profile == "" {
		return "Default", nil
	}
	if userDataDir == "" {
		return profile, nil
	}
	if _, ok := findCookieDB(userDataDir, profile); ok {
		return profile, nil
	}
	profiles := listProfiles(userDataDir)
	if len(profiles) == 1 {
		fmt.Printf("Profile %q not found; using %q.\n", profile, profiles[0])
		return profiles[0], nil
	}
	if len(profiles) == 0 {
		return profile, nil
	}
	return "", fmt.Errorf("profile %q not found in %s. Available profiles: %s", profile, userDataDir, strings.Join(profiles, ", "))
}

func findCookieDB(userDataDir, profile string) (string, bool) {
	candidates := []string{
		filepath.Join(userDataDir, profile, "Cookies"),
		filepath.Join(userDataDir, profile, "Network", "Cookies"),
	}
	for _, path := range candidates {
		if fileExists(path) {
			return path, true
		}
	}
	return "", false
}

func listProfiles(userDataDir string) []string {
	entries, err := os.ReadDir(userDataDir)
	if err != nil {
		return nil
	}
	var profiles []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if _, ok := findCookieDB(userDataDir, name); ok {
			profiles = append(profiles, name)
		}
	}
	sort.Strings(profiles)
	return profiles
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
