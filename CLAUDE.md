# cli-exchange

Go CLI for Microsoft Exchange Online (mail, calendar) via Microsoft Graph API.

## Build & Test

```bash
make build          # build bin/exchange
make test           # go test -race -coverprofile=coverage.out ./...
make lint           # golangci-lint run
make install        # build + copy to ~/.local/bin/exchange
```

## Architecture

- `cmd/exchange/` -- CLI entry point, flag parsing, command dispatch
- `internal/config/` -- config file (~/.config/exchange/config.json), env vars, CLI overrides
- `internal/auth/` -- MSAL device code flow, token cache
- `internal/graph/` -- Graph API client, error types, models, mail/calendar operations
- `internal/output/` -- JSON and table output formatters
- `internal/tz/` -- timezone handling

No CLI framework. Hand-rolled flag parsing with `os.Args`. Switch dispatch in `main.go`.

Command structure: `resource [sub-resource] verb` pattern. Sub-resources are singular. Commands parse up to 3 tokens (e.g., `calendar event list` -> `calendar-event-list`).

## Module

`github.com/dsswift/cli-exchange`

## Dependencies

Only external dependency: `github.com/AzureAD/microsoft-authentication-library-for-go`. Everything else is stdlib.

## Configuration

Config resolution priority: CLI flags (`-o`, `--client-id`, `--tenant-id`) > env vars > config file > defaults.

Default output format is JSON. Use `-o table` for human-readable output.

`exchange login` prompts interactively for missing client ID and saves to `~/.config/exchange/config.json`.

### Environment Variables

- `EXCHANGE_CLIENT_ID` -- Azure AD client ID
- `EXCHANGE_TENANT_ID` -- defaults to "common"
- `EXCHANGE_TIMEZONE` -- defaults to "UTC"
- `EXCHANGE_OUTPUT` -- output format, defaults to "json"
- `EXCHANGE_TIMEOUT` -- HTTP timeout in seconds, defaults to 30
- `EXCHANGE_TOKEN_CACHE` -- token cache file path
