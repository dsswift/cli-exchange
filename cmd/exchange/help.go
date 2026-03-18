package main

import (
	"fmt"

	"github.com/dsswift/cli-exchange/internal/output"
)

// FlagHelp describes a single flag for help output.
type FlagHelp struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Default     string   `json:"default,omitempty"`
	Required    bool     `json:"required,omitempty"`
	Repeatable  bool     `json:"repeatable,omitempty"`
	Description string   `json:"description"`
	Values      []string `json:"values,omitempty"`
}

// CommandHelp describes a single command for help output.
type CommandHelp struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Flags       []FlagHelp `json:"flags,omitempty"`
	Positional  []FlagHelp `json:"positional,omitempty"`
	Examples    []string   `json:"examples,omitempty"`
	Notes       []string   `json:"notes,omitempty"`
}

// GroupHelp describes a resource group (e.g. "mail", "calendar").
type GroupHelp struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Commands    []string `json:"commands"`
}

var globalFlags = []FlagHelp{
	{Name: "-o, --output", Type: "string", Default: "json", Description: "Output format", Values: []string{"json", "table"}},
	{Name: "--timezone", Type: "string", Description: "Override timezone (e.g. America/New_York)"},
	{Name: "--client-id", Type: "string", Description: "Override Azure AD client ID"},
	{Name: "--tenant-id", Type: "string", Description: "Override Azure AD tenant ID"},
	{Name: "--version", Type: "bool", Description: "Show version"},
}

var mailListFilterFlags = []FlagHelp{
	{Name: "--folder", Type: "string", Description: "Filter by folder name"},
	{Name: "--sender", Type: "string", Description: "Filter by sender address (supports aliases and pipe syntax)"},
	{Name: "--subject", Type: "string", Description: "Filter by subject (contains match)"},
	{Name: "--start", Type: "string", Description: "From date (YYYY-MM-DD)"},
	{Name: "--end", Type: "string", Description: "To date (YYYY-MM-DD)"},
	{Name: "--limit", Type: "int", Default: "25", Description: "Max results to return"},
	{Name: "--read", Type: "bool", Description: "Read messages only"},
	{Name: "--unread", Type: "bool", Description: "Unread messages only"},
	{Name: "--has-attachments", Type: "bool", Description: "Messages with attachments only"},
}

var commands = map[string]CommandHelp{
	"login": {
		Name:        "login",
		Description: "Authenticate via device code flow",
		Flags: []FlagHelp{
			{Name: "--client-id", Type: "string", Description: "Azure AD client ID (prompted if missing)"},
			{Name: "--tenant-id", Type: "string", Default: "common", Description: "Azure AD tenant ID"},
		},
		Examples: []string{
			"exchange login",
			"exchange login --client-id abc-123 --tenant-id my-tenant",
		},
	},
	"logout": {
		Name:        "logout",
		Description: "Clear cached authentication tokens",
		Examples:    []string{"exchange logout"},
	},
	"status": {
		Name:        "status",
		Description: "Show auth config and connection info",
		Examples:    []string{"exchange status", "exchange status -o table"},
	},
	"mail-list": {
		Name:        "mail list",
		Description: "List and search emails",
		Flags:       mailListFilterFlags,
		Examples: []string{
			"exchange mail list",
			"exchange mail list --sender user@example.com --unread",
			"exchange mail list --folder Inbox --start 2026-01-01 --end 2026-01-31",
			"exchange mail list --has-attachments --limit 10",
		},
	},
	"mail-show": {
		Name:        "mail show",
		Description: "Show email details with full body",
		Flags: []FlagHelp{
			{Name: "--ids", Type: "string", Description: "Message IDs (comma-separated)"},
			{Name: "--batch", Type: "int", Description: "Show N filtered emails with body (use with filter flags)"},
			{Name: "--folder", Type: "string", Description: "Filter by folder name (with --batch)"},
			{Name: "--sender", Type: "string", Description: "Filter by sender (with --batch)"},
			{Name: "--subject", Type: "string", Description: "Filter by subject (with --batch)"},
			{Name: "--start", Type: "string", Description: "From date (with --batch)"},
			{Name: "--end", Type: "string", Description: "To date (with --batch)"},
			{Name: "--read", Type: "bool", Description: "Read messages only (with --batch)"},
			{Name: "--unread", Type: "bool", Description: "Unread messages only (with --batch)"},
			{Name: "--has-attachments", Type: "bool", Description: "With attachments only (with --batch)"},
		},
		Notes: []string{
			"--ids and --batch are mutually exclusive; one is required",
			"Filter flags (--folder, --sender, etc.) only apply with --batch",
		},
		Examples: []string{
			"exchange mail show --ids ABC123",
			"exchange mail show --ids ABC123,DEF456",
			"exchange mail show --batch 5 --unread",
			"exchange mail show --batch 3 --sender user@example.com",
		},
	},
	"mail-archive": {
		Name:        "mail archive",
		Description: "Archive emails (move to Archive folder)",
		Flags: []FlagHelp{
			{Name: "--ids", Type: "string", Required: true, Description: "Message IDs to archive (comma-separated)"},
		},
		Examples: []string{
			"exchange mail archive --ids ABC123",
			"exchange mail archive --ids ABC123,DEF456",
		},
	},
	"mail-delete": {
		Name:        "mail delete",
		Description: "Delete emails (move to Deleted Items)",
		Flags: []FlagHelp{
			{Name: "--ids", Type: "string", Required: true, Description: "Message IDs to delete (comma-separated)"},
		},
		Examples: []string{
			"exchange mail delete --ids ABC123",
		},
	},
	"mail-send": {
		Name:        "mail send",
		Description: "Compose and send an email",
		Flags: []FlagHelp{
			{Name: "--to", Type: "string", Required: true, Repeatable: true, Description: "Recipient email address (repeatable)"},
			{Name: "--cc", Type: "string", Repeatable: true, Description: "CC recipient email address (repeatable)"},
			{Name: "--subject", Type: "string", Description: "Email subject"},
			{Name: "--body", Type: "string", Description: "Email body content"},
			{Name: "--body-type", Type: "string", Default: "text", Description: "Body content type", Values: []string{"text", "html"}},
			{Name: "--importance", Type: "string", Default: "normal", Description: "Message importance", Values: []string{"low", "normal", "high"}},
			{Name: "--attach", Type: "string", Repeatable: true, Description: "File to attach (repeatable, max 3MB each)"},
			{Name: "--no-save-to-sent-items", Type: "bool", Description: "Do not save to Sent Items"},
		},
		Examples: []string{
			"exchange mail send --to user@example.com --subject \"Hello\" --body \"Hi there\"",
			"exchange mail send --to user@example.com --subject \"Report\" --attach report.pdf",
		},
	},
	"mail-draft-create": {
		Name:        "mail draft create",
		Description: "Create a draft email",
		Flags: []FlagHelp{
			{Name: "--to", Type: "string", Repeatable: true, Description: "Recipient email address (repeatable)"},
			{Name: "--cc", Type: "string", Repeatable: true, Description: "CC recipient email address (repeatable)"},
			{Name: "--subject", Type: "string", Description: "Email subject"},
			{Name: "--body", Type: "string", Description: "Email body content"},
			{Name: "--body-type", Type: "string", Default: "text", Description: "Body content type", Values: []string{"text", "html"}},
			{Name: "--importance", Type: "string", Default: "normal", Description: "Message importance", Values: []string{"low", "normal", "high"}},
			{Name: "--attach", Type: "string", Repeatable: true, Description: "File to attach (repeatable, max 3MB each)"},
		},
		Examples: []string{
			"exchange mail draft create --to user@example.com --subject \"Hello\" --body \"Hi there\"",
			"exchange mail draft create --to a@example.com --to b@example.com --cc c@example.com --subject \"Team update\"",
			"exchange mail draft create --to user@example.com --subject \"Report\" --attach report.pdf",
		},
	},
	"mail-draft-send": {
		Name:        "mail draft send",
		Description: "Send an existing draft email",
		Positional: []FlagHelp{
			{Name: "<id>", Type: "string", Required: true, Description: "Draft message ID"},
		},
		Examples: []string{
			"exchange mail draft send ABC123",
		},
	},
	"mail-draft-attach": {
		Name:        "mail draft attach",
		Description: "Attach files to an existing draft",
		Positional: []FlagHelp{
			{Name: "<id>", Type: "string", Required: true, Description: "Draft message ID"},
		},
		Flags: []FlagHelp{
			{Name: "--attach", Type: "string", Required: true, Repeatable: true, Description: "File to attach (repeatable, max 3MB each)"},
		},
		Examples: []string{
			"exchange mail draft attach ABC123 --attach report.pdf",
			"exchange mail draft attach ABC123 --attach file1.pdf --attach file2.csv",
		},
	},
	"mail-folder-list": {
		Name:        "mail folder list",
		Description: "List mail folders",
		Examples:    []string{"exchange mail folder list"},
	},
	"mail-attachment-list": {
		Name:        "mail attachment list",
		Description: "List attachment metadata for a message",
		Flags: []FlagHelp{
			{Name: "--message-id", Type: "string", Required: true, Description: "Parent message ID"},
			{Name: "--id", Type: "string", Description: "Attachment ID; if set, fetch only this attachment"},
			{Name: "--include-content", Type: "bool", Description: "Include base64 content bytes (and decoded text for text/* types); ignored when --id is set (content always included)"},
			{Name: "--name", Type: "string", Description: "Filter by filename substring (ignored when --id is set)"},
			{Name: "--no-inline", Type: "bool", Description: "Exclude inline (CID) attachments (ignored when --id is set)"},
		},
		Examples: []string{
			"exchange mail attachment list --message-id MSG123",
			"exchange mail attachment list --message-id MSG123 --include-content",
			"exchange mail attachment list --message-id MSG123 --name report --no-inline",
			"exchange mail attachment list --message-id MSG123 --id ATT456",
		},
	},
	"mail-attachment-download": {
		Name:        "mail attachment download",
		Description: "Save attachment(s) to disk",
		Flags: []FlagHelp{
			{Name: "--message-id", Type: "string", Required: true, Description: "Parent message ID"},
			{Name: "--id", Type: "string", Description: "Attachment ID; if set, download only this attachment"},
			{Name: "--dir", Type: "string", Default: ".", Description: "Output directory"},
			{Name: "--name", Type: "string", Description: "Filter by filename substring (ignored when --id is set)"},
			{Name: "--no-inline", Type: "bool", Description: "Exclude inline attachments (ignored when --id is set)"},
		},
		Notes: []string{
			"Conflict resolution: if report.pdf exists, saves as report (1).pdf",
		},
		Examples: []string{
			"exchange mail attachment download --message-id MSG123",
			"exchange mail attachment download --message-id MSG123 --dir /tmp/attachments",
			"exchange mail attachment download --message-id MSG123 --name report",
			"exchange mail attachment download --message-id MSG123 --id ATT456 --dir ~/Downloads",
		},
	},
	"calendar-list": {
		Name:        "calendar list",
		Description: "List calendars",
		Examples:    []string{"exchange calendar list"},
	},
	"calendar-event-list": {
		Name:        "calendar event list",
		Description: "List calendar events",
		Flags: []FlagHelp{
			{Name: "--calendar", Type: "string", Description: "Calendar ID (default: primary)"},
			{Name: "--start", Type: "string", Description: "From date (YYYY-MM-DD)"},
			{Name: "--end", Type: "string", Description: "To date (YYYY-MM-DD)"},
			{Name: "--limit", Type: "int", Default: "25", Description: "Max results to return"},
		},
		Examples: []string{
			"exchange calendar event list",
			"exchange calendar event list --start 2026-03-01 --end 2026-03-31",
		},
	},
	"calendar-event-show": {
		Name:        "calendar event show",
		Description: "Show event details",
		Positional: []FlagHelp{
			{Name: "<id>", Type: "string", Required: true, Description: "Event ID"},
		},
		Examples: []string{
			"exchange calendar event show ABC123",
		},
	},
	"calendar-availability-check": {
		Name:        "calendar availability check",
		Description: "Check free/busy schedule for one or more users",
		Flags: []FlagHelp{
			{Name: "--emails", Type: "string", Required: true, Description: "Email addresses to check (comma-separated)"},
			{Name: "--start", Type: "string", Description: "Start time (YYYY-MM-DDTHH:MM:SS)"},
			{Name: "--end", Type: "string", Description: "End time (YYYY-MM-DDTHH:MM:SS)"},
			{Name: "--timespan", Type: "string", Description: "Time span: Nh, Nd, Nw (e.g. 3d, 8h, 1w)"},
			{Name: "--interval", Type: "int", Default: "30", Description: "Slot interval in minutes"},
			{Name: "--business-hours", Type: "string", Description: "Override business hours (e.g. 07:00-18:00)"},
			{Name: "--include-weekends", Type: "bool", Description: "Include Saturday and Sunday in results"},
		},
		Notes: []string{
			"Use --start/--end for explicit range, or --timespan for relative range from now",
		},
		Examples: []string{
			"exchange calendar availability check --emails user@example.com --timespan 1d",
			"exchange calendar availability check --emails a@example.com,b@example.com --timespan 1w --interval 15",
			"exchange calendar availability check --emails user@example.com --start 2026-03-13T09:00:00 --end 2026-03-13T17:00:00",
		},
	},
	"config-show": {
		Name:        "config show",
		Description: "Show resolved configuration",
		Examples:    []string{"exchange config show"},
	},
	"config-set": {
		Name:        "config set",
		Description: "Set a configuration value",
		Positional: []FlagHelp{
			{Name: "<key>", Type: "string", Required: true, Description: "Config key", Values: []string{"output", "timezone", "business-hours", "include-weekends"}},
			{Name: "<value>", Type: "string", Required: true, Description: "Config value"},
		},
		Examples: []string{
			"exchange config set output table",
			"exchange config set timezone America/New_York",
			"exchange config set business-hours 07:00-18:00",
			"exchange config set include-weekends true",
		},
	},
	"config-alias-list": {
		Name:        "config alias list",
		Description: "List domain aliases",
		Examples:    []string{"exchange config alias list"},
	},
	"config-alias-add": {
		Name:        "config alias add",
		Description: "Add a domain alias",
		Positional: []FlagHelp{
			{Name: "<domain>", Type: "string", Required: true, Description: "Primary domain"},
			{Name: "<aliases>", Type: "string", Required: true, Description: "Pipe-separated alias domains (e.g. dcim.com|dciartform.com)"},
		},
		Examples: []string{
			"exchange config alias add dcim.com \"dcim.com|dciartform.com\"",
		},
	},
	"config-alias-delete": {
		Name:        "config alias delete",
		Description: "Delete a domain alias",
		Positional: []FlagHelp{
			{Name: "<domain>", Type: "string", Required: true, Description: "Domain to remove"},
		},
		Examples: []string{
			"exchange config alias delete dcim.com",
		},
	},
}

var groups = map[string]GroupHelp{
	"mail": {
		Name:        "mail",
		Description: "Email operations",
		Commands:    []string{"mail-list", "mail-show", "mail-send", "mail-archive", "mail-delete", "mail-draft-create", "mail-draft-send", "mail-draft-attach", "mail-folder-list", "mail-attachment-list", "mail-attachment-download"},
	},
	"mail-draft": {
		Name:        "mail draft",
		Description: "Draft email operations",
		Commands:    []string{"mail-draft-create", "mail-draft-send", "mail-draft-attach"},
	},
	"mail-folder": {
		Name:        "mail folder",
		Description: "Mail folder operations",
		Commands:    []string{"mail-folder-list"},
	},
	"mail-attachment": {
		Name:        "mail attachment",
		Description: "Mail attachment operations",
		Commands:    []string{"mail-attachment-list", "mail-attachment-download"},
	},
	"calendar": {
		Name:        "calendar",
		Description: "Calendar operations",
		Commands:    []string{"calendar-list", "calendar-event-list", "calendar-event-show", "calendar-availability-check"},
	},
	"calendar-event": {
		Name:        "calendar event",
		Description: "Calendar event operations",
		Commands:    []string{"calendar-event-list", "calendar-event-show"},
	},
	"calendar-availability": {
		Name:        "calendar availability",
		Description: "Calendar availability operations",
		Commands:    []string{"calendar-availability-check"},
	},
	"config": {
		Name:        "config",
		Description: "Configuration operations",
		Commands:    []string{"config-show", "config-set", "config-alias-list", "config-alias-add", "config-alias-delete"},
	},
	"config-alias": {
		Name:        "config alias",
		Description: "Domain alias operations",
		Commands:    []string{"config-alias-list", "config-alias-add", "config-alias-delete"},
	},
	"session": {
		Name:        "session",
		Description: "Authentication and session management",
		Commands:    []string{"login", "logout", "status"},
	},
}

func handleHelp(f flags, version string) int {
	if f.output == "json" || (f.command == "help" && f.output == "") {
		// Only print JSON if explicitly requested via -o json on the help command
		if f.output == "json" && f.command == "help" {
			printHelpJSON(version)
			return 0
		}
	}

	// Per-command help
	if cmd, ok := commands[f.command]; ok {
		printCommandHelp(cmd)
		return 0
	}

	// Per-group help
	if group, ok := groups[f.command]; ok {
		printGroupHelp(group)
		return 0
	}

	// Root help (empty command, "help" command, or unknown)
	printFullHelp()
	return 0
}

func printFullHelp() {
	fmt.Print(`Usage: exchange <command> [options]

Session Commands:
  login                              Authenticate (device code flow)
  logout                             Clear cached tokens
  status                             Show auth config and connection info

Mail Commands:
  mail list [options]                List and search emails
    --folder <name>                    Filter by folder
    --sender <address>                 Filter by sender (supports aliases)
    --subject <text>                   Filter by subject (contains)
    --start <date>                     From date (YYYY-MM-DD)
    --end <date>                       To date (YYYY-MM-DD)
    --limit <n>                        Max results (default: 25)
    --read                             Read messages only
    --unread                           Unread messages only
    --has-attachments                  With attachments only

  mail show [options]                Show email details
    --ids <id[,id,...]>                Message IDs
    --batch <n>                        Show N filtered emails with body
    (filter flags from mail list apply with --batch)

  mail archive --ids <id[,id,...]>   Archive email (move to Archive)
  mail delete --ids <id[,id,...]>    Delete email (move to Deleted Items)

  mail send [options]                Compose and send an email
    --to <address>                     Recipient (repeatable, required)
    --cc <address>                     CC recipient (repeatable)
    --subject <text>                   Email subject
    --body <text>                      Email body
    --body-type <type>                 Body type: text, html (default: text)
    --importance <level>               Importance: low, normal, high (default: normal)
    --attach <file>                    File to attach (repeatable, max 3MB each)
    --no-save-to-sent-items            Do not save to Sent Items

  mail draft create [options]        Create draft email
    --to <address>                     Recipient (repeatable)
    --cc <address>                     CC recipient (repeatable)
    --subject <text>                   Email subject
    --body <text>                      Email body
    --body-type <type>                 Body type: text, html (default: text)
    --importance <level>               Importance: low, normal, high (default: normal)
    --attach <file>                    File to attach (repeatable, max 3MB each)

  mail draft send <id>               Send an existing draft
  mail draft attach <id> [options]   Attach files to a draft
    --attach <file>                    File to attach (repeatable, required, max 3MB each)

  mail folder list                   List mail folders

  mail attachment list [options]     List attachment metadata
    --message-id <id>                  Parent message ID (required)
    --id <id>                          Attachment ID (fetch single attachment)
    --include-content                  Include base64 content bytes
    --name <substring>                 Filter by filename substring
    --no-inline                        Exclude inline (CID) attachments

  mail attachment download [options] Save attachment(s) to disk
    --message-id <id>                  Parent message ID (required)
    --id <id>                          Attachment ID (download single attachment)
    --dir <path>                       Output directory (default: .)
    --name <substring>                 Filter by filename substring
    --no-inline                        Exclude inline attachments

Calendar Commands:
  calendar list                      List calendars

  calendar event list [options]      List calendar events
    --calendar <id>                    Calendar ID (default: primary)
    --start <date>                     From date (YYYY-MM-DD)
    --end <date>                       To date (YYYY-MM-DD)
    --limit <n>                        Max results (default: 25)

  calendar event show <id>           Show event details

  calendar availability check [opts] Check free/busy schedule
    --emails <addr[,addr,...]>         Email addresses to check (required)
    --start <datetime>                 Start time (YYYY-MM-DDTHH:MM:SS)
    --end <datetime>                   End time (YYYY-MM-DDTHH:MM:SS)
    --timespan <span>                  Time span: Nh, Nd, Nw (e.g. 3d, 8h, 1w)
    --interval <minutes>               Slot interval (default: 30)
    --business-hours <range>           Override business hours (e.g. 07:00-18:00)
    --include-weekends                 Include Saturday and Sunday

Config Commands:
  config show                        Show resolved configuration
  config set <key> <value>           Set config value (output, timezone, business-hours, include-weekends)

  config alias list                  List domain aliases
  config alias add <domain> <aliases> Add domain alias (pipe-separated)
  config alias delete <domain>       Delete domain alias

Global Options:
  -o, --output <format>              Output format: json, table (default: json)
  --timezone <tz>                    Override timezone (e.g. America/New_York)
  --client-id <id>                   Override Azure AD client ID
  --tenant-id <id>                   Override Azure AD tenant ID
  --version                          Show version

Use 'exchange <command> --help' for detailed command help with examples.
Use 'exchange help -o json' for machine-readable output.
`)
}

func printCommandHelp(cmd CommandHelp) {
	fmt.Printf("Usage: exchange %s", cmd.Name)
	if len(cmd.Flags) > 0 {
		fmt.Print(" [options]")
	}
	for _, p := range cmd.Positional {
		fmt.Printf(" %s", p.Name)
	}
	fmt.Println()
	fmt.Println()

	fmt.Println(cmd.Description)
	fmt.Println()

	if len(cmd.Positional) > 0 {
		fmt.Println("Arguments:")
		for _, p := range cmd.Positional {
			line := fmt.Sprintf("  %-30s %s", p.Name, p.Description)
			if len(p.Values) > 0 {
				line += fmt.Sprintf(" (%s)", joinValues(p.Values))
			}
			fmt.Println(line)
		}
		fmt.Println()
	}

	if len(cmd.Flags) > 0 {
		fmt.Println("Options:")
		for _, fl := range cmd.Flags {
			label := fl.Name
			switch fl.Type {
			case "string":
				label += " <value>"
			case "int":
				label += " <n>"
			}
			desc := fl.Description
			if fl.Default != "" {
				desc += fmt.Sprintf(" (default: %s)", fl.Default)
			}
			if fl.Required {
				desc += " (required)"
			}
			if fl.Repeatable {
				desc += " (repeatable)"
			}
			if len(fl.Values) > 0 {
				desc += fmt.Sprintf(" [%s]", joinValues(fl.Values))
			}
			fmt.Printf("  %-30s %s\n", label, desc)
		}
		fmt.Println()
	}

	if len(cmd.Notes) > 0 {
		fmt.Println("Notes:")
		for _, n := range cmd.Notes {
			fmt.Printf("  %s\n", n)
		}
		fmt.Println()
	}

	if len(cmd.Examples) > 0 {
		fmt.Println("Examples:")
		for _, ex := range cmd.Examples {
			fmt.Printf("  %s\n", ex)
		}
		fmt.Println()
	}
}

func printGroupHelp(group GroupHelp) {
	fmt.Printf("%s - %s\n\n", group.Name, group.Description)
	fmt.Println("Commands:")
	for _, cmdKey := range group.Commands {
		if cmd, ok := commands[cmdKey]; ok {
			fmt.Printf("  %-36s %s\n", cmd.Name, cmd.Description)
		}
	}
	fmt.Println()
	fmt.Println("Use 'exchange <command> --help' for detailed command help.")
	fmt.Println()
}

type helpJSON struct {
	Name        string        `json:"name"`
	Version     string        `json:"version"`
	GlobalFlags []FlagHelp    `json:"globalFlags"`
	Groups      []GroupHelp   `json:"groups"`
	Commands    []CommandHelp `json:"commands"`
}

func printHelpJSON(version string) {
	// Build ordered group list
	groupOrder := []string{"session", "mail", "mail-draft", "mail-folder", "mail-attachment", "calendar", "calendar-event", "calendar-availability", "config", "config-alias"}
	var groupList []GroupHelp
	for _, key := range groupOrder {
		if g, ok := groups[key]; ok {
			groupList = append(groupList, g)
		}
	}

	// Build ordered command list
	cmdOrder := []string{
		"login", "logout", "status",
		"mail-list", "mail-show", "mail-send", "mail-archive", "mail-delete", "mail-draft-create", "mail-draft-send", "mail-draft-attach", "mail-folder-list", "mail-attachment-list", "mail-attachment-download",
		"calendar-list", "calendar-event-list", "calendar-event-show", "calendar-availability-check",
		"config-show", "config-set", "config-alias-list", "config-alias-add", "config-alias-delete",
	}
	var cmdList []CommandHelp
	for _, key := range cmdOrder {
		if cmd, ok := commands[key]; ok {
			cmdList = append(cmdList, cmd)
		}
	}

	h := helpJSON{
		Name:        "exchange",
		Version:     version,
		GlobalFlags: globalFlags,
		Groups:      groupList,
		Commands:    cmdList,
	}
	output.PrintJSON(h)
}

func joinValues(vals []string) string {
	result := ""
	for i, v := range vals {
		if i > 0 {
			result += ", "
		}
		result += v
	}
	return result
}
