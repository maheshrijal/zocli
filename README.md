# zocli

Unofficial CLI for tracking Zomato orders (personal use, community tinkering). This project is intentionally small and hackable so other developers can plug in their own API discovery work.

## What this is
- A Go CLI scaffold with a **config store**, **local cache**, and **table output**
- A **mock data path** so contributors can build UI and storage without live calls
- A **browser-based auth flow** that captures cookies without copy/paste
- A **placeholder Zomato client** where real endpoints can be wired in

## What this is not
- An official Zomato integration
- A stable API client
- A guarantee of ToS compliance

## Quick start

```bash
go build ./cmd/zocli

./zocli help
./zocli sync --mock
./zocli orders
```

## Auth without copy/paste (recommended)

```bash
./zocli auth login
```

This opens a real Chrome window. Log in to Zomato and the CLI will detect it and store cookies automatically.

## Using your own cookie (manual, unofficial)
1. Open Zomato in your browser and log in.
2. Open DevTools → Network.
3. Copy the `Cookie` header from any authenticated request.
4. Save it:

```bash
./zocli auth --cookie "<cookie header>"
```

Then sync:

```bash
./zocli sync
```

## Commands

- `auth` — Save your Zomato cookie for future requests
- `auth status` — Check whether your saved cookie is still valid
- `sync` — Fetch orders and store locally (or `--mock`)
- `orders` — List stored orders
- `stats` — Summarize total spend with optional grouping
- `config` — Show config and cache paths

Example:

```bash
./zocli stats --group year
```

## Project layout

```
cmd/zocli          # CLI entrypoint
internal/auth          # Browser-based login
internal/cli           # Usage and help text
internal/config        # Config file handling
internal/store         # Local cache for orders
internal/zomato        # Zomato client + models
internal/stats         # Spending summaries
internal/sample        # Embedded mock orders
```

## Roadmap ideas
- Add richer order details (line items, receipts)
- Track live order status
- Add `watch` command with refresh interval
- Add `export` (JSON/CSV)
- Add `swiggy` client behind the same interface

## Disclaimer
This is a **personal, unofficial** tool. Using cookies or reverse-engineered endpoints may violate terms of service. Use at your own risk.

## Releasing (GoReleaser + Homebrew)
This repo uses GoReleaser to publish binaries and update the Homebrew tap at `maheshrijal/homebrew-tap`.

One-time setup:
1. Create the tap repo: `github.com/maheshrijal/homebrew-tap`
2. Add a repo-scoped token with access to the tap as a GitHub secret on this repo:
   - `HOMEBREW_TAP_GITHUB_TOKEN`

Release flow:
```bash
git tag v0.1.0
git push origin v0.1.0
```

Users can install with:
```bash
brew install maheshrijal/tap/zocli
```

## Contributing
See `CONTRIBUTING.md`.
