# Contributing

Thanks for helping improve `zocli`!

## Ground rules
- Keep changes small and focused.
- Avoid committing personal cookies or credentials.
- Document reverse-engineered endpoints in the PR description.
- Prefer Go standard library unless there is a clear win.

## Getting started

```bash
# build

go build ./cmd/zocli

# run with sample data
./zocli sync --mock
./zocli orders
```

## Suggested areas to tackle
- Implement `internal/zomato/client.go` for real order fetch
- Add better error handling for auth/session expiry
- Add tests around config/store and order parsing

