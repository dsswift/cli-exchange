package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type BusinessHours struct {
	Start string `json:"start"` // "HH:MM" 24h format
	End   string `json:"end"`   // "HH:MM" 24h format
}

type ExchangeConfig struct {
	ClientID               string              `json:"clientId"`
	TenantID               string              `json:"tenantId"`
	Authority              string              `json:"-"`
	Timezone               string              `json:"timezone,omitempty"`
	Output                 string              `json:"output,omitempty"`
	UserEmail              string              `json:"userEmail,omitempty"`
	AllowSendToRecipients  []string            `json:"allowSendToRecipients,omitempty"`
	DomainAliases          map[string][]string `json:"domainAliases,omitempty"`
	BusinessHours          *BusinessHours      `json:"businessHours,omitempty"`
	IncludeWeekends        *bool               `json:"includeWeekends,omitempty"`
	Timeout                time.Duration       `json:"-"`
	TokenCachePath         string              `json:"-"`
}

// Overrides holds values from CLI flags that take priority over all other sources.
type Overrides struct {
	ClientID string
	TenantID string
	Timezone string
	Output   string
}

func configDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "exchange")
}

func configPath() string {
	return filepath.Join(configDir(), "config.json")
}

// loadFile reads the persisted config file, returning an empty config if none exists.
func loadFile() *ExchangeConfig {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return &ExchangeConfig{}
	}
	var cfg ExchangeConfig
	if json.Unmarshal(data, &cfg) != nil {
		return &ExchangeConfig{}
	}
	return &cfg
}

// SaveConfig persists the config to ~/.config/exchange/config.json.
func SaveConfig(cfg *ExchangeConfig) error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	if err := os.WriteFile(configPath(), data, 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// LoadConfig resolves configuration from CLI overrides, env vars, and config file.
// Returns an error if client ID is not available from any source.
func LoadConfig(overrides Overrides) (*ExchangeConfig, error) {
	cfg := resolve(overrides)
	if cfg.ClientID == "" {
		return nil, fmt.Errorf("client ID not configured (set EXCHANGE_CLIENT_ID or run 'exchange login')")
	}
	return cfg, nil
}

// LoadConfigPartial resolves config without requiring client ID.
// Used by login to read whatever is available before prompting.
func LoadConfigPartial(overrides Overrides) *ExchangeConfig {
	return resolve(overrides)
}

// LoadConfigFile returns the raw config file contents for display and editing.
func LoadConfigFile() *ExchangeConfig {
	return loadFile()
}

// resolve merges config from all sources. Priority: overrides > env > file > defaults.
func resolve(ov Overrides) *ExchangeConfig {
	file := loadFile()

	clientID := first(ov.ClientID, os.Getenv("EXCHANGE_CLIENT_ID"), file.ClientID)
	tenantID := first(ov.TenantID, os.Getenv("EXCHANGE_TENANT_ID"), file.TenantID, "common")
	tz := first(ov.Timezone, os.Getenv("EXCHANGE_TIMEZONE"), file.Timezone, "UTC")
	output := first(ov.Output, os.Getenv("EXCHANGE_OUTPUT"), file.Output, "json")

	timeoutSec := 30
	if v := os.Getenv("EXCHANGE_TIMEOUT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			timeoutSec = n
		}
	}

	tokenCache := os.Getenv("EXCHANGE_TOKEN_CACHE")
	if tokenCache == "" {
		home, _ := os.UserHomeDir()
		tokenCache = filepath.Join(home, ".exchange-cli-token-cache.json")
	}

	return &ExchangeConfig{
		ClientID:              clientID,
		TenantID:              tenantID,
		Authority:             fmt.Sprintf("https://login.microsoftonline.com/%s", tenantID),
		Timezone:              tz,
		Output:                output,
		UserEmail:             file.UserEmail,
		AllowSendToRecipients: file.AllowSendToRecipients,
		DomainAliases:         file.DomainAliases,
		BusinessHours:         file.BusinessHours,
		IncludeWeekends:       file.IncludeWeekends,
		Timeout:               time.Duration(timeoutSec) * time.Second,
		TokenCachePath:        tokenCache,
	}
}

// ParseBusinessHours parses a "HH:MM-HH:MM" string into a BusinessHours struct.
func ParseBusinessHours(value string) (*BusinessHours, error) {
	parts := strings.SplitN(value, "-", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid business hours %q, expected HH:MM-HH:MM", value)
	}
	start := strings.TrimSpace(parts[0])
	end := strings.TrimSpace(parts[1])
	if err := validateHHMM(start); err != nil {
		return nil, fmt.Errorf("invalid start time: %w", err)
	}
	if err := validateHHMM(end); err != nil {
		return nil, fmt.Errorf("invalid end time: %w", err)
	}
	sh, sm, _ := parseHHMM(start)
	eh, em, _ := parseHHMM(end)
	if sh*60+sm >= eh*60+em {
		return nil, fmt.Errorf("business hours start %s must be before end %s", start, end)
	}
	return &BusinessHours{Start: start, End: end}, nil
}

func validateHHMM(s string) error {
	if len(s) != 5 || s[2] != ':' {
		return fmt.Errorf("%q is not HH:MM format", s)
	}
	h, err := strconv.Atoi(s[:2])
	if err != nil || h < 0 || h > 23 {
		return fmt.Errorf("%q has invalid hour", s)
	}
	m, err := strconv.Atoi(s[3:])
	if err != nil || m < 0 || m > 59 {
		return fmt.Errorf("%q has invalid minute", s)
	}
	return nil
}

func parseHHMM(s string) (int, int, error) {
	h, _ := strconv.Atoi(s[:2])
	m, _ := strconv.Atoi(s[3:])
	return h, m, nil
}

// StartHourMinute returns the hour and minute from the start time.
func (bh *BusinessHours) StartHourMinute() (int, int) {
	h, m, _ := parseHHMM(bh.Start)
	return h, m
}

// EndHourMinute returns the hour and minute from the end time.
func (bh *BusinessHours) EndHourMinute() (int, int) {
	h, m, _ := parseHHMM(bh.End)
	return h, m
}

// NormalizeEmailWithAliases lowercases an email and resolves alias domains
// to the primary domain using DomainAliases configuration.
func (cfg *ExchangeConfig) NormalizeEmailWithAliases(email string) string {
	email = strings.ToLower(strings.TrimSpace(email))
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		return email
	}
	user, domain := parts[0], parts[1]

	// Build reverse lookup: alias domain -> primary domain.
	// Sort keys for deterministic resolution when a domain appears
	// in multiple alias groups.
	keys := make([]string, 0, len(cfg.DomainAliases))
	for k := range cfg.DomainAliases {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, primary := range keys {
		for _, alias := range cfg.DomainAliases[primary] {
			if strings.ToLower(alias) == domain {
				return user + "@" + strings.ToLower(primary)
			}
		}
	}
	return email
}

// expandAllowedRecipients expands all AllowSendToRecipients entries using
// pipe syntax and domain alias resolution, returning a deduplicated set of
// normalized allowed addresses.
func (cfg *ExchangeConfig) expandAllowedRecipients() []string {
	seen := map[string]bool{}
	var result []string

	for _, entry := range cfg.AllowSendToRecipients {
		entry = strings.ToLower(strings.TrimSpace(entry))
		if !strings.Contains(entry, "@") {
			continue
		}

		parts := strings.SplitN(entry, "@", 2)
		user, domainPart := parts[0], parts[1]

		var addrs []string
		if strings.Contains(domainPart, "|") {
			// Pipe syntax: user@domain1|domain2
			domains := strings.Split(domainPart, "|")
			for _, d := range domains {
				addrs = append(addrs, user+"@"+d)
			}
		} else {
			addrs = []string{entry}
		}

		// Normalize each expanded address via alias resolution
		for _, addr := range addrs {
			normalized := cfg.NormalizeEmailWithAliases(addr)
			if !seen[normalized] {
				seen[normalized] = true
				result = append(result, normalized)
			}
		}
	}
	return result
}

// ValidateSendRecipients checks that all recipients are allowed by the
// whitelist or match the authenticated user's own email. Returns an error
// listing any blocked recipients.
func (cfg *ExchangeConfig) ValidateSendRecipients(recipients []string) error {
	allowed := cfg.expandAllowedRecipients()
	allowedSet := map[string]bool{}
	for _, a := range allowed {
		allowedSet[a] = true
	}

	// Self is always allowed
	if cfg.UserEmail != "" {
		selfNorm := cfg.NormalizeEmailWithAliases(cfg.UserEmail)
		allowedSet[selfNorm] = true
	}

	var blocked []string
	for _, r := range recipients {
		normalized := cfg.NormalizeEmailWithAliases(r)
		if !allowedSet[normalized] {
			blocked = append(blocked, r)
		}
	}

	if len(blocked) > 0 {
		return fmt.Errorf("send blocked: recipient(s) not in allow list: %s", strings.Join(blocked, ", "))
	}
	return nil
}

// first returns the first non-empty string.
func first(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
