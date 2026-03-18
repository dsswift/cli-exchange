package main

import (
	"testing"
)

func TestParseFlags_NoArgs(t *testing.T) {
	f, err := parseFlags(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.command != "" {
		t.Errorf("expected empty command, got %q", f.command)
	}
}

func TestParseFlags_Version(t *testing.T) {
	f, err := parseFlags([]string{"--version"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f.version {
		t.Error("expected version flag to be true")
	}
}

func TestParseFlags_TopLevelCommands(t *testing.T) {
	for _, cmd := range []string{"login", "logout", "status"} {
		f, err := parseFlags([]string{cmd})
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", cmd, err)
		}
		if f.command != cmd {
			t.Errorf("expected command %q, got %q", cmd, f.command)
		}
	}
}

func TestParseFlags_MailCommands(t *testing.T) {
	tests := []struct {
		args    []string
		command string
	}{
		{[]string{"mail", "list"}, "mail-list"},
		{[]string{"mail", "show"}, "mail-show"},
		{[]string{"mail", "send"}, "mail-send"},
		{[]string{"mail", "archive"}, "mail-archive"},
		{[]string{"mail", "delete"}, "mail-delete"},
		{[]string{"mail", "draft", "create"}, "mail-draft-create"},
		{[]string{"mail", "draft", "send"}, "mail-draft-send"},
		{[]string{"mail", "draft", "attach"}, "mail-draft-attach"},
		{[]string{"mail", "folder", "list"}, "mail-folder-list"},
	}

	for _, tt := range tests {
		f, err := parseFlags(tt.args)
		if err != nil {
			t.Fatalf("unexpected error for %v: %v", tt.args, err)
		}
		if f.command != tt.command {
			t.Errorf("expected command %q, got %q", tt.command, f.command)
		}
	}
}

func TestParseFlags_CalendarCommands(t *testing.T) {
	tests := []struct {
		args    []string
		command string
	}{
		{[]string{"calendar", "list"}, "calendar-list"},
		{[]string{"calendar", "event", "list"}, "calendar-event-list"},
		{[]string{"calendar", "event", "show"}, "calendar-event-show"},
		{[]string{"calendar", "availability", "check"}, "calendar-availability-check"},
	}

	for _, tt := range tests {
		f, err := parseFlags(tt.args)
		if err != nil {
			t.Fatalf("unexpected error for %v: %v", tt.args, err)
		}
		if f.command != tt.command {
			t.Errorf("expected command %q, got %q", tt.command, f.command)
		}
	}
}

func TestParseFlags_ConfigCommands(t *testing.T) {
	tests := []struct {
		args    []string
		command string
	}{
		{[]string{"config", "show"}, "config-show"},
		{[]string{"config", "set", "output", "table"}, "config-set"},
		{[]string{"config", "alias", "list"}, "config-alias-list"},
		{[]string{"config", "alias", "add", "dcim.com", "dcim.com|dciartform.com"}, "config-alias-add"},
		{[]string{"config", "alias", "delete", "dcim.com"}, "config-alias-delete"},
	}

	for _, tt := range tests {
		f, err := parseFlags(tt.args)
		if err != nil {
			t.Fatalf("unexpected error for %v: %v", tt.args, err)
		}
		if f.command != tt.command {
			t.Errorf("expected command %q, got %q", tt.command, f.command)
		}
	}
}

func TestParseFlags_ConfigSetPositional(t *testing.T) {
	f, err := parseFlags([]string{"config", "set", "output", "table"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.id != "output" {
		t.Errorf("expected id 'output', got %q", f.id)
	}
	if f.value != "table" {
		t.Errorf("expected value 'table', got %q", f.value)
	}
}

func TestParseFlags_OutputFlag(t *testing.T) {
	tests := []struct {
		args   []string
		output string
	}{
		{[]string{"mail", "list", "-o", "json"}, "json"},
		{[]string{"mail", "list", "-o", "table"}, "table"},
		{[]string{"mail", "list", "--output", "json"}, "json"},
		{[]string{"mail", "list", "--output=table"}, "table"},
	}

	for _, tt := range tests {
		f, err := parseFlags(tt.args)
		if err != nil {
			t.Fatalf("unexpected error for %v: %v", tt.args, err)
		}
		if f.output != tt.output {
			t.Errorf("expected output %q, got %q", tt.output, f.output)
		}
	}
}

func TestParseFlags_OutputFlagInvalid(t *testing.T) {
	_, err := parseFlags([]string{"mail", "list", "-o", "xml"})
	if err == nil {
		t.Error("expected error for invalid output format")
	}
}

func TestParseFlags_GlobalFlags(t *testing.T) {
	f, err := parseFlags([]string{"mail", "list", "-o", "table", "--timezone", "America/New_York"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.output != "table" {
		t.Errorf("expected output table, got %q", f.output)
	}
	if f.timezone != "America/New_York" {
		t.Errorf("expected timezone America/New_York, got %q", f.timezone)
	}
}

func TestParseFlags_MailListFilters(t *testing.T) {
	f, err := parseFlags([]string{
		"mail", "list",
		"--sender", "test@example.com",
		"--subject", "hello",
		"--folder", "inbox",
		"--start", "2026-01-01",
		"--end", "2026-01-31",
		"--limit", "10",
		"--unread",
		"--has-attachments",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.sender != "test@example.com" {
		t.Errorf("expected sender test@example.com, got %q", f.sender)
	}
	if f.subject != "hello" {
		t.Errorf("expected subject hello, got %q", f.subject)
	}
	if f.folder != "inbox" {
		t.Errorf("expected folder inbox, got %q", f.folder)
	}
	if f.start != "2026-01-01" {
		t.Errorf("expected start 2026-01-01, got %q", f.start)
	}
	if f.end != "2026-01-31" {
		t.Errorf("expected end 2026-01-31, got %q", f.end)
	}
	if f.limit != 10 {
		t.Errorf("expected limit 10, got %d", f.limit)
	}
	if f.isRead == nil || *f.isRead != false {
		t.Error("expected isRead to be false (--unread)")
	}
	if f.hasAttachments == nil || *f.hasAttachments != true {
		t.Error("expected hasAttachments to be true")
	}
}

func TestParseFlags_PositionalIDRejectedForMail(t *testing.T) {
	_, err := parseFlags([]string{"mail", "show", "ABC123"})
	if err == nil {
		t.Error("expected error for positional arg on mail show")
	}
}

func TestParseFlags_PositionalIDAllowedForCalendar(t *testing.T) {
	f, err := parseFlags([]string{"calendar", "event", "show", "ABC123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.id != "ABC123" {
		t.Errorf("expected id ABC123, got %q", f.id)
	}
}

func TestParseFlags_IDs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		ids  []string
	}{
		{"single", []string{"mail", "show", "--ids", "ABC"}, []string{"ABC"}},
		{"multiple", []string{"mail", "show", "--ids", "ABC,DEF,GHI"}, []string{"ABC", "DEF", "GHI"}},
		{"equals syntax", []string{"mail", "show", "--ids=ABC,DEF"}, []string{"ABC", "DEF"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := parseFlags(tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(f.ids) != len(tt.ids) {
				t.Fatalf("expected %d ids, got %d", len(tt.ids), len(f.ids))
			}
			for i, id := range tt.ids {
				if f.ids[i] != id {
					t.Errorf("ids[%d]: expected %q, got %q", i, id, f.ids[i])
				}
			}
		})
	}
}

func TestParseFlags_Batch(t *testing.T) {
	f, err := parseFlags([]string{"mail", "show", "--batch", "5"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.batch != 5 {
		t.Errorf("expected batch 5, got %d", f.batch)
	}
}

func TestParseFlags_BatchEqualsSyntax(t *testing.T) {
	f, err := parseFlags([]string{"mail", "show", "--batch=10"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.batch != 10 {
		t.Errorf("expected batch 10, got %d", f.batch)
	}
}

func TestParseFlags_BatchInvalid(t *testing.T) {
	_, err := parseFlags([]string{"mail", "show", "--batch", "abc"})
	if err == nil {
		t.Error("expected error for non-numeric --batch")
	}
}

func TestParseFlags_DraftOptions(t *testing.T) {
	f, err := parseFlags([]string{
		"mail", "draft", "create",
		"--to", "a@example.com",
		"--to", "b@example.com",
		"--cc", "c@example.com",
		"--subject", "Test",
		"--body", "Hello",
		"--body-type", "html",
		"--importance", "high",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.to) != 2 {
		t.Errorf("expected 2 to recipients, got %d", len(f.to))
	}
	if len(f.cc) != 1 {
		t.Errorf("expected 1 cc recipient, got %d", len(f.cc))
	}
	if f.body != "Hello" {
		t.Errorf("expected body Hello, got %q", f.body)
	}
	if f.bodyType != "html" {
		t.Errorf("expected body-type html, got %q", f.bodyType)
	}
	if f.importance != "high" {
		t.Errorf("expected importance high, got %q", f.importance)
	}
}

func TestParseFlags_EqualsSyntax(t *testing.T) {
	f, err := parseFlags([]string{"mail", "list", "--sender=test@example.com", "--limit=5"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.sender != "test@example.com" {
		t.Errorf("expected sender test@example.com, got %q", f.sender)
	}
	if f.limit != 5 {
		t.Errorf("expected limit 5, got %d", f.limit)
	}
}

func TestParseFlags_AvailabilityOptions(t *testing.T) {
	f, err := parseFlags([]string{
		"calendar", "availability", "check",
		"--emails", "a@example.com,b@example.com",
		"--start", "2026-03-13T09:00:00",
		"--end", "2026-03-13T17:00:00",
		"--interval", "15",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.emails) != 2 {
		t.Errorf("expected 2 emails, got %d", len(f.emails))
	}
	if f.interval != 15 {
		t.Errorf("expected interval 15, got %d", f.interval)
	}
}

func TestParseFlags_EmailsCommaSeparated(t *testing.T) {
	f, err := parseFlags([]string{
		"calendar", "availability", "check",
		"--emails", "a@example.com,b@example.com,c@example.com",
		"--timespan", "1d",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.emails) != 3 {
		t.Fatalf("expected 3 emails, got %d", len(f.emails))
	}
	expected := []string{"a@example.com", "b@example.com", "c@example.com"}
	for i, e := range expected {
		if f.emails[i] != e {
			t.Errorf("emails[%d]: expected %q, got %q", i, e, f.emails[i])
		}
	}
}

func TestParseFlags_MissingSubcommand(t *testing.T) {
	for _, cmd := range []string{"mail", "calendar", "config"} {
		_, err := parseFlags([]string{cmd})
		if err == nil {
			t.Errorf("expected error for missing subcommand on %s", cmd)
		}
	}
}

func TestParseFlags_MissingSubResourceVerb(t *testing.T) {
	tests := [][]string{
		{"mail", "draft"},
		{"mail", "folder"},
		{"calendar", "event"},
		{"calendar", "availability"},
		{"config", "alias"},
	}
	for _, args := range tests {
		_, err := parseFlags(args)
		if err == nil {
			t.Errorf("expected error for missing verb on %v", args)
		}
	}
}

func TestParseFlags_UnknownFlag(t *testing.T) {
	_, err := parseFlags([]string{"mail", "list", "--bogus"})
	if err == nil {
		t.Error("expected error for unknown flag")
	}
}

func TestParseFlags_UnknownCommand(t *testing.T) {
	_, err := parseFlags([]string{"bogus"})
	if err == nil {
		t.Error("expected error for unknown command")
	}
}

func TestParseFlags_ConfigOverrides(t *testing.T) {
	f, err := parseFlags([]string{"login", "--client-id", "abc-123", "--tenant-id", "my-tenant"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.clientID != "abc-123" {
		t.Errorf("expected clientID abc-123, got %q", f.clientID)
	}
	if f.tenantID != "my-tenant" {
		t.Errorf("expected tenantID my-tenant, got %q", f.tenantID)
	}
}

func TestParseFlags_ConfigOverridesEqualsSyntax(t *testing.T) {
	f, err := parseFlags([]string{"status", "--client-id=abc-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.clientID != "abc-123" {
		t.Errorf("expected clientID abc-123, got %q", f.clientID)
	}
}

func TestExpandSender_NoAt(t *testing.T) {
	result := expandSender("cfavero", nil)
	if result != nil {
		t.Errorf("expected nil for no-@ sender, got %v", result)
	}
}

func TestExpandSender_SingleAddress(t *testing.T) {
	result := expandSender("cfavero@dcim.com", nil)
	if len(result) != 1 || result[0] != "cfavero@dcim.com" {
		t.Errorf("expected [cfavero@dcim.com], got %v", result)
	}
}

func TestExpandSender_PipeSyntax(t *testing.T) {
	result := expandSender("cfavero@dcim.com|dciartform.com", nil)
	if len(result) != 2 {
		t.Fatalf("expected 2 addresses, got %d", len(result))
	}
	if result[0] != "cfavero@dcim.com" {
		t.Errorf("expected cfavero@dcim.com, got %s", result[0])
	}
	if result[1] != "cfavero@dciartform.com" {
		t.Errorf("expected cfavero@dciartform.com, got %s", result[1])
	}
}

func TestExpandSender_DomainAlias(t *testing.T) {
	aliases := map[string][]string{
		"dcim.com": {"dcim.com", "dciartform.com"},
	}
	result := expandSender("cfavero@dcim.com", aliases)
	if len(result) != 2 {
		t.Fatalf("expected 2 addresses, got %d", len(result))
	}
	if result[0] != "cfavero@dcim.com" {
		t.Errorf("expected cfavero@dcim.com, got %s", result[0])
	}
	if result[1] != "cfavero@dciartform.com" {
		t.Errorf("expected cfavero@dciartform.com, got %s", result[1])
	}
}

func TestParseFlags_Timespan(t *testing.T) {
	f, err := parseFlags([]string{
		"calendar", "availability", "check",
		"--emails", "a@example.com",
		"--timespan", "3d",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.timespan != "3d" {
		t.Errorf("expected timespan 3d, got %q", f.timespan)
	}
}

func TestParseFlags_BusinessHoursFlag(t *testing.T) {
	f, err := parseFlags([]string{
		"calendar", "availability", "check",
		"--emails", "a@example.com",
		"--timespan", "1d",
		"--business-hours", "07:00-18:00",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.businessHours != "07:00-18:00" {
		t.Errorf("expected businessHours 07:00-18:00, got %q", f.businessHours)
	}
}

func TestParseFlags_IncludeWeekends(t *testing.T) {
	f, err := parseFlags([]string{
		"calendar", "availability", "check",
		"--emails", "a@example.com",
		"--timespan", "1w",
		"--include-weekends",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f.includeWeekends {
		t.Error("expected includeWeekends to be true")
	}
}

func TestExpandSender_NoMatchingAlias(t *testing.T) {
	aliases := map[string][]string{
		"other.com": {"other.com", "other2.com"},
	}
	result := expandSender("cfavero@dcim.com", aliases)
	if len(result) != 1 || result[0] != "cfavero@dcim.com" {
		t.Errorf("expected [cfavero@dcim.com], got %v", result)
	}
}

func TestParseFlags_NoSaveToSentItems(t *testing.T) {
	f, err := parseFlags([]string{"mail", "send", "--to", "a@example.com", "--no-save-to-sent-items"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.saveToSentItems == nil {
		t.Fatal("expected saveToSentItems to be set")
	}
	if *f.saveToSentItems != false {
		t.Error("expected saveToSentItems to be false")
	}
}

func TestParseFlags_AttachRepeatable(t *testing.T) {
	f, err := parseFlags([]string{
		"mail", "send",
		"--to", "a@example.com",
		"--attach", "file1.pdf",
		"--attach", "file2.csv",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.attachFiles) != 2 {
		t.Fatalf("expected 2 attach files, got %d", len(f.attachFiles))
	}
	if f.attachFiles[0] != "file1.pdf" {
		t.Errorf("expected file1.pdf, got %q", f.attachFiles[0])
	}
	if f.attachFiles[1] != "file2.csv" {
		t.Errorf("expected file2.csv, got %q", f.attachFiles[1])
	}
}

func TestParseFlags_MailDraftSendPositionalID(t *testing.T) {
	f, err := parseFlags([]string{"mail", "draft", "send", "ABC123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.command != "mail-draft-send" {
		t.Errorf("expected command mail-draft-send, got %q", f.command)
	}
	if f.id != "ABC123" {
		t.Errorf("expected id ABC123, got %q", f.id)
	}
}

func TestParseFlags_MailDraftAttachWithFiles(t *testing.T) {
	f, err := parseFlags([]string{
		"mail", "draft", "attach", "ABC123",
		"--attach", "file.pdf",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.command != "mail-draft-attach" {
		t.Errorf("expected command mail-draft-attach, got %q", f.command)
	}
	if f.id != "ABC123" {
		t.Errorf("expected id ABC123, got %q", f.id)
	}
	if len(f.attachFiles) != 1 || f.attachFiles[0] != "file.pdf" {
		t.Errorf("expected [file.pdf], got %v", f.attachFiles)
	}
}

func TestParseFlags_MailAttachmentCommands(t *testing.T) {
	tests := []struct {
		args    []string
		command string
	}{
		{[]string{"mail", "attachment", "list"}, "mail-attachment-list"},
		{[]string{"mail", "attachment", "download"}, "mail-attachment-download"},
	}

	for _, tt := range tests {
		f, err := parseFlags(tt.args)
		if err != nil {
			t.Fatalf("unexpected error for %v: %v", tt.args, err)
		}
		if f.command != tt.command {
			t.Errorf("expected command %q, got %q", tt.command, f.command)
		}
	}
}

func TestParseFlags_AttachmentListFlags(t *testing.T) {
	f, err := parseFlags([]string{
		"mail", "attachment", "list",
		"--message-id", "MSG123",
		"--include-content",
		"--name", "report",
		"--no-inline",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.command != "mail-attachment-list" {
		t.Errorf("expected command mail-attachment-list, got %q", f.command)
	}
	if f.messageID != "MSG123" {
		t.Errorf("expected messageID MSG123, got %q", f.messageID)
	}
	if !f.includeContent {
		t.Error("expected includeContent to be true")
	}
	if f.name != "report" {
		t.Errorf("expected name report, got %q", f.name)
	}
	if !f.noInline {
		t.Error("expected noInline to be true")
	}
}

func TestParseFlags_AttachmentListSingleByID(t *testing.T) {
	f, err := parseFlags([]string{
		"mail", "attachment", "list",
		"--message-id", "MSG123",
		"--id", "ATT456",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.messageID != "MSG123" {
		t.Errorf("expected messageID MSG123, got %q", f.messageID)
	}
	if f.id != "ATT456" {
		t.Errorf("expected id ATT456, got %q", f.id)
	}
}

func TestParseFlags_AttachmentDownloadFlags(t *testing.T) {
	f, err := parseFlags([]string{
		"mail", "attachment", "download",
		"--message-id", "MSG123",
		"--dir", "/tmp/downloads",
		"--name", "invoice",
		"--no-inline",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.command != "mail-attachment-download" {
		t.Errorf("expected command mail-attachment-download, got %q", f.command)
	}
	if f.messageID != "MSG123" {
		t.Errorf("expected messageID MSG123, got %q", f.messageID)
	}
	if f.dir != "/tmp/downloads" {
		t.Errorf("expected dir /tmp/downloads, got %q", f.dir)
	}
	if f.name != "invoice" {
		t.Errorf("expected name invoice, got %q", f.name)
	}
	if !f.noInline {
		t.Error("expected noInline to be true")
	}
}

func TestParseFlags_AttachmentDownloadSingleByID(t *testing.T) {
	f, err := parseFlags([]string{
		"mail", "attachment", "download",
		"--message-id", "MSG123",
		"--id", "ATT456",
		"--dir", "~/Downloads",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.messageID != "MSG123" {
		t.Errorf("expected messageID MSG123, got %q", f.messageID)
	}
	if f.id != "ATT456" {
		t.Errorf("expected id ATT456, got %q", f.id)
	}
	if f.dir != "~/Downloads" {
		t.Errorf("expected dir ~/Downloads, got %q", f.dir)
	}
}

func TestParseFlags_AttachmentMissingVerb(t *testing.T) {
	_, err := parseFlags([]string{"mail", "attachment"})
	if err == nil {
		t.Error("expected error for missing verb on mail attachment")
	}
}
