package main

import (
	"context"
	"fmt"
	"os"

	"github.com/dsswift/cli-exchange/internal/auth"
	"github.com/dsswift/cli-exchange/internal/config"
	"github.com/dsswift/cli-exchange/internal/graph"
	"github.com/dsswift/cli-exchange/internal/tz"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	os.Exit(run())
}

func run() int {
	f, err := parseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	if f.version {
		fmt.Printf("exchange %s (built %s)\n", Version, BuildTime)
		return 0
	}

	if f.help {
		return handleHelp(f, Version)
	}
	if f.command == "help" {
		return handleHelp(f, Version)
	}

	if f.command == "" {
		printFullHelp()
		return 0
	}

	// Resolve output format from config if not set by flag
	if f.output == "" {
		cfg := config.LoadConfigPartial(configOverrides(f))
		f.output = cfg.Output
	}

	switch f.command {
	case "login":
		return cmdLogin(f)
	case "logout":
		return cmdLogout(f)
	case "status":
		return cmdStatus(f)

	// Mail commands
	case "mail-list":
		return cmdMailList(f)
	case "mail-show":
		return cmdMailShow(f)
	case "mail-archive":
		return cmdMailArchive(f)
	case "mail-delete":
		return cmdMailDelete(f)
	case "mail-send":
		return cmdMailSend(f)
	case "mail-draft-create":
		return cmdMailDraftCreate(f)
	case "mail-draft-send":
		return cmdMailDraftSend(f)
	case "mail-draft-attach":
		return cmdMailDraftAttach(f)
	case "mail-folder-list":
		return cmdMailFolderList(f)
	case "mail-attachment-list":
		return cmdMailAttachmentList(f)
	case "mail-attachment-download":
		return cmdMailAttachmentDownload(f)

	// Calendar commands
	case "calendar-list":
		return cmdCalendarList(f)
	case "calendar-event-list":
		return cmdCalendarEventList(f)
	case "calendar-event-show":
		return cmdCalendarEventShow(f)
	case "calendar-availability-check":
		return cmdCalendarAvailabilityCheck(f)

	// Config commands
	case "config-show":
		return cmdConfigShow(f)
	case "config-set":
		return cmdConfigSet(f)
	case "config-alias-list":
		return cmdConfigAliasList(f)
	case "config-alias-add":
		return cmdConfigAliasAdd(f)
	case "config-alias-delete":
		return cmdConfigAliasDelete(f)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", f.command)
		printFullHelp()
		return 1
	}
}

func configOverrides(f flags) config.Overrides {
	return config.Overrides{
		ClientID: f.clientID,
		TenantID: f.tenantID,
		Timezone: f.timezone,
		Output:   f.output,
	}
}

func newClient(f flags) (*graph.GraphClient, *tz.Service, error) {
	cfg, err := config.LoadConfig(configOverrides(f))
	if err != nil {
		return nil, nil, err
	}

	authenticator, err := auth.New(cfg)
	if err != nil {
		return nil, nil, err
	}

	tzSvc, err := tz.NewService(cfg.Timezone)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()
	tokenFn := func() (string, error) {
		return authenticator.GetAccessToken(ctx)
	}

	client := graph.NewClient("", cfg.Timeout, tokenFn)
	return client, tzSvc, nil
}

