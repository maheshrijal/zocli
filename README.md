# zocli 🍕

A tiny, powerful CLI to view and analyze your Zomato order history.


## Features

- **📈 Inflation Tracker**: Monitor how prices for your favorite items change over time.
- **💰 Spending Analytics**: Deep dive into spending by weekday, time of day, and top restaurants.
- **🎁 Zomato Wrapped**: A fun, yearly retrospective of your ordering habits.
- **🔒 Privacy First**: Your data stays on your machine. Cookies are stored locally.

## Install

```bash
brew install maheshrijal/tap/zocli
```

## Quick Start

1. **Login** (saves cookie locally):
   ```bash
   zocli auth login
   ```

2. **Sync Orders**:
   ```bash
   zocli sync
   ```


## Showcase

<table>
  <tr>
    <td width="50%" valign="middle">
      <img src="assets/Patterns.png" alt="Spending Patterns" width="100%" />
    </td>
    <td width="50%" valign="middle">
      <img src="assets/Stats.png" alt="General Stats" width="100%" />
    </td>
  </tr>
</table>
<p align="center">
  <img src="assets/Inflation.png" alt="Inflation Tracker" width="100%" />
</p>

## Commands

### `stats`
Analyze your spending habits.
```bash
zocli stats --view patterns   # See when you order the most
zocli stats --view spend      # See spending by weekday
zocli stats --view personal   # Top restaurants and items
```

### `dash`
Interactive dashboard to explore your data.
- **Navigation**: Use `Tab` / `Shift+Tab` to switch tabs.
- **Filters**: Press `y` for **This Year**, `m` for **This Month**, `a` for **All Time**.

### `wrapped`
Generate a Spotify-Wrapped style slideshow of your food journey.
```bash
zocli wrapped
```

### `suggest`
Can't decide what to eat? Let zocli pick a restaurant and dish from your favorites.
```bash
zocli suggest
# Output: How about ordering from: ✨ Pizza Hut ✨
```

### `export`
Export your data for external analysis.
```bash
zocli export --format csv > orders.csv
zocli export --format json > orders.json
```

### `inflation`
Track how much item prices have risen.
```bash
zocli inflation              # Summary of top risers
zocli inflation "Biryani"    # Track specifics
```

### `orders`
List your raw order history.
```bash
zocli orders --limit 50
```

## Project Layout

```
cmd/zocli          # Entrypoint
internal/tui       # Bubble Tea Dashboard components
internal/stats     # Analysis logic
internal/zomato    # API Client
internal/store     # Local JSON storage
```

## Disclaimer
This is an **unofficial** project and is **not affiliated with Zomato**. Use at your own risk.
