# Exchange CLI Reference

Use the `exchange` CLI for all Exchange Online operations. Output is always JSON; never add `-o table` when parsing results programmatically.

## Mail

- `mail list [--folder <name>] [--sender <alias-or-addr>] [--subject <text>] [--start YYYY-MM-DD] [--end YYYY-MM-DD] [--limit N] [--read] [--unread] [--has-attachments]`
- `mail show --ids <id,...>` full body; OR `mail show --batch N [filter flags]` show N filtered emails with body (mutually exclusive)
- `mail send --to <email> [--to ...] [--cc <email>] [--subject <str>] [--body <str>] [--body-type text|html] [--importance low|normal|high] [--attach <file>]`
- `mail archive --ids <id,...>` | `mail delete --ids <id,...>`
- `mail draft create` same flags as send, no whitelist check
- `mail draft send <id>` positional arg, whitelist validated
- `mail draft attach <id> --attach <file> [--attach ...]`
- `mail folder list`
- `mail attachment list --message-id <id> [--include-content] [--name <substr>] [--no-inline]` -- `--include-content` returns decoded text inline
- `mail attachment download --message-id <id> [--id <att-id>] [--dir <path>] [--name <substr>]`

## Calendar

- `calendar event list [--calendar <id>] [--start YYYY-MM-DD] [--end YYYY-MM-DD] [--limit N]`
- `calendar event show <id>` positional arg
- `calendar availability check --emails <email,...> [--start/--end YYYY-MM-DDTHH:MM:SS | --timespan Nh|Nd|Nw] [--interval N] [--business-hours HH:MM-HH:MM] [--include-weekends]`
- `calendar list` list all calendars

## Syntax Notes

- `--sender` accepts alias (`cfavero`) or address, never display names
- Repeatable flags (`--to`, `--cc`, `--attach`): specify multiple times
- Comma-separated: `--ids`, `--emails`
- Boolean flags take no value: `--read`, `--unread`, `--has-attachments`, `--include-content`, `--no-inline`, `--include-weekends`
- Dates: `YYYY-MM-DD` for mail/calendar filters, `YYYY-MM-DDTHH:MM:SS` for availability start/end
