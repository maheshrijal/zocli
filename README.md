# zocli

A tiny CLI to view and analyze your Zomato order history.

## What this is
- A Go CLI for Zomato to check order history, spent amount, tracking etc. 

## What this is not
- An official Zomato integration
- A stable API client
- A guarantee of ToS compliance

## Quick start

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


## Disclaimer
This is an **unofficial** project and is **not affiliated with Zomato**. Please don't sue me 🙏  
Using cookies or reverse-engineered endpoints may violate terms of service. Use at your own risk.

Install with:
```bash
brew install maheshrijal/tap/zocli
```
