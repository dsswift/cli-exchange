package config

import (
	"strings"
	"testing"
)

func TestParseBusinessHours_Valid(t *testing.T) {
	tests := []struct {
		input string
		start string
		end   string
	}{
		{"08:00-17:00", "08:00", "17:00"},
		{"07:00-18:00", "07:00", "18:00"},
		{"00:00-23:59", "00:00", "23:59"},
		{"09:30-16:45", "09:30", "16:45"},
	}
	for _, tt := range tests {
		bh, err := ParseBusinessHours(tt.input)
		if err != nil {
			t.Errorf("ParseBusinessHours(%q): unexpected error: %v", tt.input, err)
			continue
		}
		if bh.Start != tt.start || bh.End != tt.end {
			t.Errorf("ParseBusinessHours(%q) = {%s, %s}, want {%s, %s}", tt.input, bh.Start, bh.End, tt.start, tt.end)
		}
	}
}

func TestParseBusinessHours_Invalid(t *testing.T) {
	tests := []string{
		"",
		"08:00",
		"17:00-08:00",
		"8:00-17:00",
		"08:00-17:0",
		"25:00-17:00",
		"08:60-17:00",
		"abc-def",
		"08:00-08:00",
	}
	for _, input := range tests {
		_, err := ParseBusinessHours(input)
		if err == nil {
			t.Errorf("ParseBusinessHours(%q): expected error", input)
		}
	}
}

func TestBusinessHours_StartHourMinute(t *testing.T) {
	bh := &BusinessHours{Start: "09:30", End: "17:00"}
	h, m := bh.StartHourMinute()
	if h != 9 || m != 30 {
		t.Errorf("StartHourMinute: got %d:%d, want 9:30", h, m)
	}
}

func TestBusinessHours_EndHourMinute(t *testing.T) {
	bh := &BusinessHours{Start: "08:00", End: "16:45"}
	h, m := bh.EndHourMinute()
	if h != 16 || m != 45 {
		t.Errorf("EndHourMinute: got %d:%d, want 16:45", h, m)
	}
}

func TestNormalizeEmailWithAliases_NoAliases(t *testing.T) {
	cfg := &ExchangeConfig{}
	got := cfg.NormalizeEmailWithAliases("User@Example.COM")
	if got != "user@example.com" {
		t.Errorf("got %q, want %q", got, "user@example.com")
	}
}

func TestNormalizeEmailWithAliases_ResolvesAlias(t *testing.T) {
	cfg := &ExchangeConfig{
		DomainAliases: map[string][]string{
			"dcim.com": {"dcim.com", "dciartform.com"},
		},
	}
	got := cfg.NormalizeEmailWithAliases("cfavero@dciartform.com")
	if got != "cfavero@dcim.com" {
		t.Errorf("got %q, want %q", got, "cfavero@dcim.com")
	}
}

func TestNormalizeEmailWithAliases_PrimaryUnchanged(t *testing.T) {
	cfg := &ExchangeConfig{
		DomainAliases: map[string][]string{
			"dcim.com": {"dcim.com", "dciartform.com"},
		},
	}
	got := cfg.NormalizeEmailWithAliases("cfavero@dcim.com")
	if got != "cfavero@dcim.com" {
		t.Errorf("got %q, want %q", got, "cfavero@dcim.com")
	}
}

func TestNormalizeEmailWithAliases_NoAtSign(t *testing.T) {
	cfg := &ExchangeConfig{}
	got := cfg.NormalizeEmailWithAliases("noatsign")
	if got != "noatsign" {
		t.Errorf("got %q, want %q", got, "noatsign")
	}
}

func TestValidateSendRecipients_SelfAlwaysAllowed(t *testing.T) {
	cfg := &ExchangeConfig{
		UserEmail: "josh@example.com",
	}
	err := cfg.ValidateSendRecipients([]string{"josh@example.com"})
	if err != nil {
		t.Errorf("self should be allowed: %v", err)
	}
}

func TestValidateSendRecipients_SelfWithAlias(t *testing.T) {
	cfg := &ExchangeConfig{
		UserEmail: "josh@dciartform.com",
		DomainAliases: map[string][]string{
			"dcim.com": {"dcim.com", "dciartform.com"},
		},
	}
	// Self address resolves through alias to dcim.com
	err := cfg.ValidateSendRecipients([]string{"josh@dcim.com"})
	if err != nil {
		t.Errorf("self via alias should be allowed: %v", err)
	}
}

func TestValidateSendRecipients_WhitelistedAllowed(t *testing.T) {
	cfg := &ExchangeConfig{
		AllowSendToRecipients: []string{"alice@example.com"},
	}
	err := cfg.ValidateSendRecipients([]string{"alice@example.com"})
	if err != nil {
		t.Errorf("whitelisted recipient should be allowed: %v", err)
	}
}

func TestValidateSendRecipients_Blocked(t *testing.T) {
	cfg := &ExchangeConfig{
		AllowSendToRecipients: []string{"alice@example.com"},
	}
	err := cfg.ValidateSendRecipients([]string{"bob@example.com"})
	if err == nil {
		t.Fatal("expected error for blocked recipient")
	}
	if !strings.Contains(err.Error(), "bob@example.com") {
		t.Errorf("error should mention blocked recipient: %v", err)
	}
}

func TestValidateSendRecipients_PipeSyntax(t *testing.T) {
	cfg := &ExchangeConfig{
		AllowSendToRecipients: []string{"cfavero@dcim.com|dciartform.com"},
	}
	// Both expanded addresses should be allowed
	if err := cfg.ValidateSendRecipients([]string{"cfavero@dcim.com"}); err != nil {
		t.Errorf("pipe-expanded address should be allowed: %v", err)
	}
	if err := cfg.ValidateSendRecipients([]string{"cfavero@dciartform.com"}); err != nil {
		t.Errorf("pipe-expanded address should be allowed: %v", err)
	}
}

func TestValidateSendRecipients_AliasResolution(t *testing.T) {
	cfg := &ExchangeConfig{
		AllowSendToRecipients: []string{"cfavero@dcim.com"},
		DomainAliases: map[string][]string{
			"dcim.com": {"dcim.com", "dciartform.com"},
		},
	}
	// Sending to alias domain should resolve and match
	err := cfg.ValidateSendRecipients([]string{"cfavero@dciartform.com"})
	if err != nil {
		t.Errorf("alias-resolved recipient should be allowed: %v", err)
	}
}

func TestValidateSendRecipients_EmptyUserEmail(t *testing.T) {
	cfg := &ExchangeConfig{
		AllowSendToRecipients: []string{"alice@example.com"},
	}
	// Should not block when UserEmail is empty, just skip self-check
	if err := cfg.ValidateSendRecipients([]string{"alice@example.com"}); err != nil {
		t.Errorf("should still allow whitelisted when no user email: %v", err)
	}
	if err := cfg.ValidateSendRecipients([]string{"bob@example.com"}); err == nil {
		t.Error("should block non-whitelisted even with no user email")
	}
}

func TestValidateSendRecipients_MixedAllowedAndBlocked(t *testing.T) {
	cfg := &ExchangeConfig{
		UserEmail:             "josh@example.com",
		AllowSendToRecipients: []string{"alice@example.com"},
	}
	err := cfg.ValidateSendRecipients([]string{"josh@example.com", "alice@example.com", "bob@example.com"})
	if err == nil {
		t.Fatal("expected error for mixed recipients")
	}
	if !strings.Contains(err.Error(), "bob@example.com") {
		t.Errorf("error should list blocked recipient: %v", err)
	}
	if strings.Contains(err.Error(), "alice@example.com") {
		t.Errorf("error should not list allowed recipient: %v", err)
	}
}

func TestValidateSendRecipients_CaseInsensitive(t *testing.T) {
	cfg := &ExchangeConfig{
		AllowSendToRecipients: []string{"Alice@Example.COM"},
	}
	err := cfg.ValidateSendRecipients([]string{"alice@example.com"})
	if err != nil {
		t.Errorf("validation should be case-insensitive: %v", err)
	}
}

func TestValidateSendRecipients_EmptyList(t *testing.T) {
	cfg := &ExchangeConfig{}
	err := cfg.ValidateSendRecipients([]string{})
	if err != nil {
		t.Errorf("empty recipients should not error: %v", err)
	}
}
