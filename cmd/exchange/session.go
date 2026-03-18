package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/dsswift/cli-exchange/internal/auth"
	"github.com/dsswift/cli-exchange/internal/config"
	"github.com/dsswift/cli-exchange/internal/output"
)

func cmdLogin(f flags) int {
	cfg := config.LoadConfigPartial(configOverrides(f))

	if cfg.ClientID == "" {
		v := prompt("Client ID (Azure AD app registration): ")
		if v == "" {
			fmt.Fprintln(os.Stderr, "Error: client ID is required")
			return 1
		}
		cfg.ClientID = v
		cfg.Authority = fmt.Sprintf("https://login.microsoftonline.com/%s", cfg.TenantID)
	}

	if cfg.TenantID == "common" {
		v := prompt("Tenant ID [common]: ")
		if v != "" {
			cfg.TenantID = v
			cfg.Authority = fmt.Sprintf("https://login.microsoftonline.com/%s", cfg.TenantID)
		}
	}

	if err := config.SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %s\n", err)
		return 1
	}

	authenticator, err := auth.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	token, err := authenticator.GetAccessToken(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	if f.output == "table" {
		fmt.Println("Authenticated successfully.")
	} else {
		output.PrintJSON(map[string]any{
			"authenticated": true,
			"tokenLength":   len(token),
		})
	}
	return 0
}

func cmdLogout(f flags) int {
	cfg, err := config.LoadConfig(configOverrides(f))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	authenticator, err := auth.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	if err := authenticator.ClearCache(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	if f.output == "table" {
		fmt.Println("Token cache cleared.")
	} else {
		output.PrintJSON(map[string]any{"loggedOut": true})
	}
	return 0
}

func cmdStatus(f flags) int {
	cfg := config.LoadConfigPartial(configOverrides(f))

	cacheExists := false
	if _, err := os.Stat(cfg.TokenCachePath); err == nil {
		cacheExists = true
	}

	if f.output == "table" {
		clientID := cfg.ClientID
		if clientID == "" {
			clientID = "(not configured)"
		}
		fmt.Printf("Client ID:        %s\n", clientID)
		fmt.Printf("Tenant ID:        %s\n", cfg.TenantID)
		fmt.Printf("Authority:        %s\n", cfg.Authority)
		fmt.Printf("Timezone:         %s\n", cfg.Timezone)
		fmt.Printf("Timeout:          %s\n", cfg.Timeout)
		fmt.Printf("Token Cache:      %s\n", cfg.TokenCachePath)
		if cacheExists {
			fmt.Printf("Token Status:     cached\n")
		} else {
			fmt.Printf("Token Status:     not found\n")
		}
	} else {
		output.PrintJSON(map[string]any{
			"clientId":         cfg.ClientID,
			"tenantId":         cfg.TenantID,
			"authority":        cfg.Authority,
			"timezone":         cfg.Timezone,
			"timeout":          cfg.Timeout.String(),
			"tokenCachePath":   cfg.TokenCachePath,
			"tokenCacheExists": cacheExists,
		})
	}
	return 0
}

func prompt(label string) string {
	fmt.Fprint(os.Stderr, label)
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}
