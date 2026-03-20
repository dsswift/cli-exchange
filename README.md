# exchange

A CLI for Microsoft Exchange Online operations (mail, calendar) via the Microsoft Graph API.

## Install

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/dsswift/cli-exchange/main/install.sh | sh
```

Detects your OS and architecture, downloads the latest release binary, verifies the SHA256 checksum, and installs to `~/.local/bin/exchange`.

### Windows

```powershell
irm https://raw.githubusercontent.com/dsswift/cli-exchange/main/install.ps1 | iex
```

Downloads the latest release binary, verifies the SHA256 checksum, installs to `%LOCALAPPDATA%\Programs\exchange\exchange.exe`, and adds it to your user PATH.

### From source

```bash
go install github.com/dsswift/cli-exchange/cmd/exchange@latest
```

Or clone and build:

```bash
git clone https://github.com/dsswift/cli-exchange.git
cd cli-exchange
make install
```

## Configuration

Run `exchange login` to get started. The CLI will prompt for your Azure AD client ID and tenant ID, then save them to `~/.config/exchange/config.json`.

You can also configure via environment variables or CLI flags. Priority: CLI flags > env vars > config file > defaults.

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `EXCHANGE_CLIENT_ID` | Yes | -- | Azure AD app registration client ID |
| `EXCHANGE_TENANT_ID` | No | `common` | Azure AD tenant ID |
| `EXCHANGE_TIMEZONE` | No | `UTC` | Display timezone (e.g. `America/New_York`) |
| `EXCHANGE_OUTPUT` | No | `json` | Default output format (`json` or `table`) |
| `EXCHANGE_TIMEOUT` | No | `30` | HTTP timeout in seconds |
| `EXCHANGE_TOKEN_CACHE` | No | `~/.exchange-cli-token-cache.json` | Token cache file path |

### Config commands

Manage persistent settings in `~/.config/exchange/config.json`:

```bash
exchange config show                        # display resolved config
exchange config set output table            # change default output format
exchange config set timezone America/Chicago
```

### Domain aliases

Domain aliases expand `--sender` filters to match multiple domains for the same user. Useful when an organization uses proxy or alias domains (e.g., `dcim.com` and `dciartform.com` both route to the same mailbox).

```bash
exchange config alias list
exchange config alias add dcim.com 'dcim.com|dciartform.com'
exchange config alias delete dcim.com
```

With the alias above, `--sender cfavero@dcim.com` automatically expands to match both `cfavero@dcim.com` and `cfavero@dciartform.com`.

## Global Options

These options apply to all commands:

| Option | Description |
|--------|-------------|
| `-o, --output <format>` | Output format: `json` or `table` (default: json) |
| `--timezone <tz>` | Override timezone for this invocation |
| `--client-id <id>` | Override Azure AD client ID |
| `--tenant-id <id>` | Override Azure AD tenant ID |
| `--version` | Show version |

JSON is the default output format. Use `-o table` for human-readable table output.

## Commands

### Session

#### `exchange login`

Authenticate using device code flow. Opens a browser prompt for Microsoft login. If no client ID is configured, prompts for it interactively and saves to `~/.config/exchange/config.json`.

```bash
exchange login
exchange login --client-id <id> --tenant-id <id>
```

#### `exchange logout`

Clear cached authentication tokens.

```bash
exchange logout
```

#### `exchange status`

Show current configuration and connection info.

```bash
exchange status
exchange status -o table
```

### Mail

#### `exchange mail list`

List or search emails.

| Option | Description |
|--------|-------------|
| `--folder <name>` | Filter by folder (inbox, archive, drafts, sentitems, deleteditems, junkemail, or folder name) |
| `--sender <value>` | Filter by sender (see sender filtering below) |
| `--subject <text>` | Filter by subject (substring match) |
| `--start <date>` | Filter messages received on or after date (YYYY-MM-DD) |
| `--end <date>` | Filter messages received on or before date (YYYY-MM-DD) |
| `--limit <n>` | Max results, 1-100 (default: 25) |
| `--read` | Show only read messages |
| `--unread` | Show only unread messages |
| `--has-attachments` | Show only messages with attachments |

**Sender filtering** uses smart detection based on input format:

| Input | Strategy | Example |
|-------|----------|---------|
| No `@` (name only) | `$search` with `from:` query | `--sender cfavero` |
| With `@`, single domain | `$filter` with exact match | `--sender cfavero@dcim.com` |
| With `@`, pipe syntax | `$filter` with multiple exact matches | `--sender 'cfavero@dcim.com\|dciartform.com'` |
| With `@`, domain alias configured | `$filter` with alias-expanded matches | `--sender cfavero@dcim.com` (auto-expands) |

```bash
exchange mail list
exchange mail list --sender cfavero --limit 5
exchange mail list --sender cfavero@dcim.com
exchange mail list --sender 'cfavero@dcim.com|dciartform.com'
exchange mail list --folder inbox --unread --start 2026-03-01
exchange mail list --subject "quarterly report" -o table
```

#### `exchange mail show <id>`

Show full email details including body.

```bash
exchange mail show AAMkAGI2...
exchange mail show AAMkAGI2... -o table
```

#### `exchange mail archive <id>`

Move an email to the Archive folder.

```bash
exchange mail archive AAMkAGI2...
```

#### `exchange mail delete <id>`

Delete an email (moves to Deleted Items).

```bash
exchange mail delete AAMkAGI2...
```

#### `exchange mail draft create`

Create a draft email.

| Option | Description |
|--------|-------------|
| `--to <email>` | Add recipient (repeatable) |
| `--cc <email>` | Add CC recipient (repeatable) |
| `--subject <text>` | Email subject |
| `--body <text>` | Email body content |
| `--body-type <type>` | Body type: `text` or `html` (default: text) |
| `--importance <level>` | Importance: `low`, `normal`, `high` (default: normal) |

```bash
exchange mail draft create --to user@example.com --subject "Hello" --body "Message body"
exchange mail draft create --to a@example.com --to b@example.com --cc c@example.com --importance high
```

#### `exchange mail folder list`

List all mail folders.

```bash
exchange mail folder list
exchange mail folder list -o table
```

### Calendar

#### `exchange calendar list`

List all calendars.

```bash
exchange calendar list
exchange calendar list -o table
```

#### `exchange calendar event list`

List calendar events.

| Option | Description |
|--------|-------------|
| `--calendar <id>` | Calendar ID (default: primary calendar) |
| `--start <date>` | Filter events starting on or after date (YYYY-MM-DD) |
| `--end <date>` | Filter events ending on or before date (YYYY-MM-DD) |
| `--limit <n>` | Max results, 1-100 (default: 25) |

```bash
exchange calendar event list
exchange calendar event list --start 2026-03-01 --end 2026-03-14
exchange calendar event list --calendar AAMkAGI2... --limit 10 -o table
```

#### `exchange calendar event show <id>`

Show full event details.

```bash
exchange calendar event show AAMkAGI2...
exchange calendar event show AAMkAGI2... -o table
```

#### `exchange calendar availability check`

Check free/busy schedule for one or more users.

| Option | Description |
|--------|-------------|
| `--emails <email>` | Email address to check (repeatable, max 20) |
| `--start <datetime>` | Start time (YYYY-MM-DDTHH:MM:SS) |
| `--end <datetime>` | End time (YYYY-MM-DDTHH:MM:SS) |
| `--interval <minutes>` | Availability interval in minutes (default: 30) |

```bash
exchange calendar availability check --emails user@example.com --start 2026-03-13T09:00:00 --end 2026-03-13T17:00:00
exchange calendar availability check --emails a@example.com --emails b@example.com --interval 15 --start 2026-03-13T08:00:00 --end 2026-03-13T18:00:00 -o table
```

### Config

#### `exchange config show`

Show resolved configuration (merged from flags, env vars, and config file).

```bash
exchange config show
exchange config show -o table
```

#### `exchange config set <key> <value>`

Set a persistent config value. Valid keys: `output`, `timezone`.

```bash
exchange config set output table
exchange config set timezone America/New_York
```

#### `exchange config alias list`

List configured domain aliases.

```bash
exchange config alias list
```

#### `exchange config alias add <domain> <aliases>`

Add a domain alias mapping. Aliases are pipe-separated.

```bash
exchange config alias add dcim.com 'dcim.com|dciartform.com'
```

#### `exchange config alias delete <domain>`

Delete a domain alias mapping.

```bash
exchange config alias delete dcim.com
```

## Agent Instructions

Copy the following into your AI agent's system prompt or instructions file for optimal use of the `exchange` CLI.

---

Use the `exchange` CLI for all Exchange Online operations (mail, calendar, availability). Default output is JSON. Run `exchange` with no arguments to discover commands.

**Quick reference:**
- `--sender` takes email alias or address (`cfavero`, `wwoller`), not full names
- All filters combine freely: `--sender`, `--subject`, `--unread`, `--read`, `--has-attachments`, `--start`, `--end`
- Full email body: `mail show --ids <id>`
- Download attachments: `mail show --ids <id>` to get attachment IDs, then `mail attachment download --message-id <id> --id <att-id> --dir /tmp`

## Development

```bash
make help          # show available targets
make build         # build for current platform
make test          # run tests with race detection
make lint          # run linter
make build-all     # build for all platforms
make tidy          # tidy module dependencies
```
