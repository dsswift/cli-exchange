package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dsswift/cli-exchange/internal/config"
	"github.com/dsswift/cli-exchange/internal/output"
)

func cmdConfigShow(f flags) int {
	cfg := config.LoadConfigPartial(configOverrides(f))

	if f.output == "table" {
		fmt.Printf("%-20s %s\n", "client-id:", cfg.ClientID)
		fmt.Printf("%-20s %s\n", "tenant-id:", cfg.TenantID)
		fmt.Printf("%-20s %s\n", "timezone:", cfg.Timezone)
		fmt.Printf("%-20s %s\n", "output:", cfg.Output)
		if cfg.BusinessHours != nil {
			fmt.Printf("%-20s %s-%s\n", "business-hours:", cfg.BusinessHours.Start, cfg.BusinessHours.End)
		}
		if cfg.IncludeWeekends != nil {
			fmt.Printf("%-20s %t\n", "include-weekends:", *cfg.IncludeWeekends)
		}
		if len(cfg.DomainAliases) > 0 {
			fmt.Printf("%-20s\n", "domain-aliases:")
			for domain, aliases := range cfg.DomainAliases {
				fmt.Printf("  %-18s %s\n", domain+":", strings.Join(aliases, ", "))
			}
		}
		if cfg.UserEmail != "" {
			fmt.Printf("%-20s %s\n", "user-email:", cfg.UserEmail)
		}
		if len(cfg.AllowSendToRecipients) > 0 {
			fmt.Printf("%-20s\n", "allow-send-to:")
			for _, entry := range cfg.AllowSendToRecipients {
				fmt.Printf("  %s\n", entry)
			}
		}
	} else {
		data := map[string]any{
			"clientId":              cfg.ClientID,
			"tenantId":             cfg.TenantID,
			"timezone":             cfg.Timezone,
			"output":               cfg.Output,
			"domainAliases":        cfg.DomainAliases,
			"allowSendToRecipients": cfg.AllowSendToRecipients,
		}
		if cfg.UserEmail != "" {
			data["userEmail"] = cfg.UserEmail
		}
		if cfg.BusinessHours != nil {
			data["businessHours"] = cfg.BusinessHours
		}
		if cfg.IncludeWeekends != nil {
			data["includeWeekends"] = *cfg.IncludeWeekends
		}
		output.PrintJSON(data)
	}
	return 0
}

var validConfigKeys = map[string]bool{
	"output":           true,
	"timezone":         true,
	"business-hours":   true,
	"include-weekends": true,
}

func cmdConfigSet(f flags) int {
	if f.id == "" {
		fmt.Fprintf(os.Stderr, "Error: config key required (output, timezone, business-hours, include-weekends)\n")
		return 1
	}
	if f.value == "" {
		fmt.Fprintf(os.Stderr, "Error: config value required\n")
		return 1
	}

	key := f.id
	val := f.value

	if !validConfigKeys[key] {
		fmt.Fprintf(os.Stderr, "Error: unknown config key %q (valid: output, timezone, business-hours, include-weekends)\n", key)
		return 1
	}

	if key == "output" && val != "json" && val != "table" {
		fmt.Fprintf(os.Stderr, "Error: output must be 'json' or 'table'\n")
		return 1
	}

	cfg := config.LoadConfigFile()

	switch key {
	case "output":
		cfg.Output = val
	case "timezone":
		cfg.Timezone = val
	case "business-hours":
		if val == "off" || val == "" {
			cfg.BusinessHours = nil
		} else {
			bh, err := config.ParseBusinessHours(val)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %s\n", err)
				return 1
			}
			cfg.BusinessHours = bh
		}
	case "include-weekends":
		switch strings.ToLower(val) {
		case "true":
			b := true
			cfg.IncludeWeekends = &b
		case "false":
			b := false
			cfg.IncludeWeekends = &b
		default:
			fmt.Fprintf(os.Stderr, "Error: include-weekends must be 'true' or 'false'\n")
			return 1
		}
	}

	if err := config.SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	if f.output == "table" {
		fmt.Printf("Set %s = %s\n", key, val)
	} else {
		output.PrintJSON(map[string]any{
			"key":   key,
			"value": val,
		})
	}
	return 0
}

func cmdConfigAliasList(f flags) int {
	cfg := config.LoadConfigFile()

	if f.output == "table" {
		if len(cfg.DomainAliases) == 0 {
			fmt.Println("No domain aliases configured.")
			return 0
		}
		for domain, aliases := range cfg.DomainAliases {
			fmt.Printf("%-30s %s\n", domain, strings.Join(aliases, ", "))
		}
	} else {
		aliases := cfg.DomainAliases
		if aliases == nil {
			aliases = map[string][]string{}
		}
		output.PrintJSON(map[string]any{
			"domainAliases": aliases,
		})
	}
	return 0
}

func cmdConfigAliasAdd(f flags) int {
	if f.id == "" {
		fmt.Fprintf(os.Stderr, "Error: domain required\n")
		return 1
	}
	if f.value == "" {
		fmt.Fprintf(os.Stderr, "Error: aliases required (pipe-separated domains)\n")
		return 1
	}

	domain := f.id
	aliases := strings.Split(f.value, "|")

	cfg := config.LoadConfigFile()
	if cfg.DomainAliases == nil {
		cfg.DomainAliases = map[string][]string{}
	}
	cfg.DomainAliases[domain] = aliases

	if err := config.SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	if f.output == "table" {
		fmt.Printf("Added alias: %s -> %s\n", domain, strings.Join(aliases, ", "))
	} else {
		output.PrintJSON(map[string]any{
			"domain":  domain,
			"aliases": aliases,
		})
	}
	return 0
}

func cmdConfigAliasDelete(f flags) int {
	if f.id == "" {
		fmt.Fprintf(os.Stderr, "Error: domain required\n")
		return 1
	}

	domain := f.id

	cfg := config.LoadConfigFile()
	if cfg.DomainAliases == nil {
		fmt.Fprintf(os.Stderr, "Error: no domain aliases configured\n")
		return 1
	}
	if _, ok := cfg.DomainAliases[domain]; !ok {
		fmt.Fprintf(os.Stderr, "Error: no alias for domain %q\n", domain)
		return 1
	}

	delete(cfg.DomainAliases, domain)

	if err := config.SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	if f.output == "table" {
		fmt.Printf("Deleted alias: %s\n", domain)
	} else {
		output.PrintJSON(map[string]any{
			"deleted": true,
			"domain":  domain,
		})
	}
	return 0
}

func cmdConfigAllowSenderList(f flags) int {
	cfg := config.LoadConfigFile()

	if f.output == "table" {
		if len(cfg.AllowSendToRecipients) == 0 {
			fmt.Println("No allowed senders configured.")
			return 0
		}
		for _, entry := range cfg.AllowSendToRecipients {
			fmt.Println(entry)
		}
	} else {
		list := cfg.AllowSendToRecipients
		if list == nil {
			list = []string{}
		}
		output.PrintJSON(map[string]any{
			"allowSendToRecipients": list,
		})
	}
	return 0
}

func cmdConfigAllowSenderAdd(f flags) int {
	if f.id == "" {
		fmt.Fprintf(os.Stderr, "Error: email address required\n")
		return 1
	}

	entry := strings.ToLower(strings.TrimSpace(f.id))
	if !strings.Contains(entry, "@") {
		fmt.Fprintf(os.Stderr, "Error: entry must contain @ (e.g. user@example.com)\n")
		return 1
	}

	cfg := config.LoadConfigFile()

	// Dedup by exact string
	for _, existing := range cfg.AllowSendToRecipients {
		if existing == entry {
			if f.output == "table" {
				fmt.Printf("Already in allow list: %s\n", entry)
			} else {
				output.PrintJSON(map[string]any{
					"added": false,
					"entry": entry,
				})
			}
			return 0
		}
	}

	cfg.AllowSendToRecipients = append(cfg.AllowSendToRecipients, entry)

	if err := config.SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	if f.output == "table" {
		fmt.Printf("Added to allow list: %s\n", entry)
	} else {
		output.PrintJSON(map[string]any{
			"added": true,
			"entry": entry,
		})
	}
	return 0
}

func cmdConfigAllowSenderDelete(f flags) int {
	if f.id == "" {
		fmt.Fprintf(os.Stderr, "Error: email address required\n")
		return 1
	}

	entry := strings.ToLower(strings.TrimSpace(f.id))

	cfg := config.LoadConfigFile()

	found := false
	var updated []string
	for _, existing := range cfg.AllowSendToRecipients {
		if existing == entry {
			found = true
			continue
		}
		updated = append(updated, existing)
	}

	if !found {
		fmt.Fprintf(os.Stderr, "Error: %q not in allow list\n", entry)
		return 1
	}

	cfg.AllowSendToRecipients = updated

	if err := config.SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	if f.output == "table" {
		fmt.Printf("Removed from allow list: %s\n", entry)
	} else {
		output.PrintJSON(map[string]any{
			"deleted": true,
			"entry":   entry,
		})
	}
	return 0
}
