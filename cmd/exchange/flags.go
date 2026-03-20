package main

import (
	"fmt"
	"strconv"
	"strings"
)

type flags struct {
	command  string
	output   string
	timezone string
	version  bool
	help     bool

	// Config overrides
	clientID string
	tenantID string

	// Positional argument (message ID, event ID, etc.)
	id string

	// Second positional argument (config set value, alias domains)
	value string

	// Mail list/show filters
	folder         string
	sender         string
	subject        string
	start          string
	end            string
	limit          int
	isRead         *bool
	hasAttachments *bool

	// Mail show batch mode
	batch int
	ids   []string

	// Mail draft options
	to         []string
	cc         []string
	body       string
	bodyType   string
	importance string

	// Attachment and send options
	attachFiles     []string
	saveToSentItems *bool

	// Attachment list/download options
	messageID      string // parent message ID for mail attachment commands
	includeContent bool
	dir            string
	name           string
	noInline       bool

	// Calendar options
	calendarID string
	emails     []string
	interval        int
	timespan        string
	businessHours   string
	includeWeekends bool
}

func parseFlags(args []string) (flags, error) {
	f := flags{
		limit:    25,
		interval: 30,
	}

	if len(args) == 0 {
		return f, nil
	}

	i := 0

	// Parse command: resource [sub-resource] verb
	switch args[0] {
	case "login", "logout", "status":
		f.command = args[0]
		i = 1
	case "mail":
		if len(args) < 2 {
			return f, fmt.Errorf("mail requires a subcommand")
		}
		if args[1] == "--help" {
			f.command = "mail"
			f.help = true
			return f, nil
		}
		// Sub-resources: draft, folder, attachment
		if args[1] == "draft" || args[1] == "folder" || args[1] == "attachment" {
			if len(args) < 3 {
				return f, fmt.Errorf("mail %s requires a verb", args[1])
			}
			if args[2] == "--help" {
				f.command = "mail-" + args[1]
				f.help = true
				return f, nil
			}
			f.command = "mail-" + args[1] + "-" + args[2]
			i = 3
		} else {
			f.command = "mail-" + args[1]
			i = 2
		}
	case "calendar":
		if len(args) < 2 {
			return f, fmt.Errorf("calendar requires a subcommand")
		}
		if args[1] == "--help" {
			f.command = "calendar"
			f.help = true
			return f, nil
		}
		// Sub-resources: event, availability
		if args[1] == "event" || args[1] == "availability" {
			if len(args) < 3 {
				return f, fmt.Errorf("calendar %s requires a verb", args[1])
			}
			if args[2] == "--help" {
				f.command = "calendar-" + args[1]
				f.help = true
				return f, nil
			}
			f.command = "calendar-" + args[1] + "-" + args[2]
			i = 3
		} else {
			f.command = "calendar-" + args[1]
			i = 2
		}
	case "config":
		if len(args) < 2 {
			return f, fmt.Errorf("config requires a subcommand")
		}
		if args[1] == "--help" {
			f.command = "config"
			f.help = true
			return f, nil
		}
		if args[1] == "alias" {
			if len(args) < 3 {
				return f, fmt.Errorf("config alias requires a verb")
			}
			if args[2] == "--help" {
				f.command = "config-alias"
				f.help = true
				return f, nil
			}
			f.command = "config-alias-" + args[2]
			i = 3
		} else if args[1] == "allow-sender" {
			if len(args) < 3 {
				return f, fmt.Errorf("config allow-sender requires a verb")
			}
			if args[2] == "--help" {
				f.command = "config-allow-sender"
				f.help = true
				return f, nil
			}
			f.command = "config-allow-sender-" + args[2]
			i = 3
		} else {
			f.command = "config-" + args[1]
			i = 2
		}
	case "--version":
		f.version = true
		return f, nil
	case "--help":
		f.help = true
		return f, nil
	case "help":
		f.command = "help"
		i = 1
	default:
		return f, fmt.Errorf("unknown command: %s", args[0])
	}

	// Parse remaining flags and positional args
	for i < len(args) {
		arg := args[i]

		// Handle --flag=value form
		key := arg
		var val string
		hasValue := false
		if eqIdx := strings.IndexByte(arg, '='); eqIdx >= 0 {
			key = arg[:eqIdx]
			val = arg[eqIdx+1:]
			hasValue = true
		}

		nextVal := func() (string, error) {
			if hasValue {
				return val, nil
			}
			i++
			if i >= len(args) {
				return "", fmt.Errorf("%s requires a value", key)
			}
			return args[i], nil
		}

		switch key {
		case "-o", "--output":
			v, err := nextVal()
			if err != nil {
				return f, err
			}
			if v != "json" && v != "table" {
				return f, fmt.Errorf("--output must be 'json' or 'table'")
			}
			f.output = v
		case "--help":
			f.help = true
			return f, nil
		case "--version":
			f.version = true
		case "--client-id":
			v, err := nextVal()
			if err != nil {
				return f, err
			}
			f.clientID = v
		case "--tenant-id":
			v, err := nextVal()
			if err != nil {
				return f, err
			}
			f.tenantID = v
		case "--timezone":
			v, err := nextVal()
			if err != nil {
				return f, err
			}
			f.timezone = v
		case "--folder":
			v, err := nextVal()
			if err != nil {
				return f, err
			}
			f.folder = v
		case "--sender":
			v, err := nextVal()
			if err != nil {
				return f, err
			}
			f.sender = v
		case "--subject":
			v, err := nextVal()
			if err != nil {
				return f, err
			}
			f.subject = v
		case "--start":
			v, err := nextVal()
			if err != nil {
				return f, err
			}
			f.start = v
		case "--end":
			v, err := nextVal()
			if err != nil {
				return f, err
			}
			f.end = v
		case "--limit":
			v, err := nextVal()
			if err != nil {
				return f, err
			}
			n, err := strconv.Atoi(v)
			if err != nil {
				return f, fmt.Errorf("--limit must be a number: %s", v)
			}
			f.limit = n
		case "--read":
			b := true
			f.isRead = &b
		case "--unread":
			b := false
			f.isRead = &b
		case "--has-attachments":
			b := true
			f.hasAttachments = &b
		case "--batch":
			v, err := nextVal()
			if err != nil {
				return f, err
			}
			n, err := strconv.Atoi(v)
			if err != nil {
				return f, fmt.Errorf("--batch must be a number: %s", v)
			}
			f.batch = n
		case "--ids":
			v, err := nextVal()
			if err != nil {
				return f, err
			}
			f.ids = splitComma(v)
		case "--to":
			v, err := nextVal()
			if err != nil {
				return f, err
			}
			f.to = append(f.to, v)
		case "--cc":
			v, err := nextVal()
			if err != nil {
				return f, err
			}
			f.cc = append(f.cc, v)
		case "--body":
			v, err := nextVal()
			if err != nil {
				return f, err
			}
			f.body = v
		case "--body-type":
			v, err := nextVal()
			if err != nil {
				return f, err
			}
			f.bodyType = v
		case "--importance":
			v, err := nextVal()
			if err != nil {
				return f, err
			}
			f.importance = v
		case "--attach":
			v, err := nextVal()
			if err != nil {
				return f, err
			}
			f.attachFiles = append(f.attachFiles, v)
		case "--no-save-to-sent-items":
			b := false
			f.saveToSentItems = &b
		case "--id":
			v, err := nextVal()
			if err != nil {
				return f, err
			}
			f.id = v
		case "--message-id":
			v, err := nextVal()
			if err != nil {
				return f, err
			}
			f.messageID = v
		case "--include-content":
			f.includeContent = true
		case "--dir":
			v, err := nextVal()
			if err != nil {
				return f, err
			}
			f.dir = v
		case "--name":
			v, err := nextVal()
			if err != nil {
				return f, err
			}
			f.name = v
		case "--no-inline":
			f.noInline = true
		case "--calendar":
			v, err := nextVal()
			if err != nil {
				return f, err
			}
			f.calendarID = v
		case "--emails":
			v, err := nextVal()
			if err != nil {
				return f, err
			}
			f.emails = splitComma(v)
		case "--interval":
			v, err := nextVal()
			if err != nil {
				return f, err
			}
			n, err := strconv.Atoi(v)
			if err != nil {
				return f, fmt.Errorf("--interval must be a number: %s", v)
			}
			f.interval = n
		case "--timespan":
			v, err := nextVal()
			if err != nil {
				return f, err
			}
			f.timespan = v
		case "--business-hours":
			v, err := nextVal()
			if err != nil {
				return f, err
			}
			f.businessHours = v
		case "--include-weekends":
			f.includeWeekends = true
		default:
			if strings.HasPrefix(arg, "-") {
				return f, fmt.Errorf("unknown flag: %s", arg)
			}
			// Mail commands do not accept positional args
			if strings.HasPrefix(f.command, "mail-") && !strings.HasPrefix(f.command, "mail-draft") && !strings.HasPrefix(f.command, "mail-folder") {
				if strings.HasPrefix(f.command, "mail-attachment") {
					return f, fmt.Errorf("unexpected positional argument: %s (use --message-id for message ID)", arg)
				}
				return f, fmt.Errorf("unexpected positional argument: %s (use --ids for message IDs)", arg)
			}
			// Positional arguments: first is id, second is value
			if f.id == "" {
				f.id = arg
			} else if f.value == "" {
				f.value = arg
			}
		}

		i++
	}

	return f, nil
}

func splitComma(s string) []string {
	var result []string
	for _, v := range strings.Split(s, ",") {
		v = strings.TrimSpace(v)
		if v != "" {
			result = append(result, v)
		}
	}
	return result
}
