package auth

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/storage"
	"github.com/chromedp/chromedp"
	"github.com/maheshrijal/zocli/internal/config"
)

func LoginAndSaveCookie(ctx context.Context, cfgPath string, headless bool) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", headless),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	browserCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	if err := chromedp.Run(browserCtx,
		chromedp.Navigate("https://www.zomato.com/restaurants"),
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
	fmt.Println("If the login modal isn't open, click “Log in” on the page.")
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

	var cookies []*network.Cookie
	if err := chromedp.Run(browserCtx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		cookies, err = storage.GetCookies().Do(ctx)
		return err
	})); err != nil {
		return err
	}

	cookieHeader := buildCookieHeader(cookies)
	if cookieHeader == "" {
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
