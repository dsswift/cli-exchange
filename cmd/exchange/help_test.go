package main

import (
	"encoding/json"
	"io"
	"os"
	"testing"
)

func TestParseFlags_HelpOnCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		command string
		help    bool
	}{
		{"mail list --help", []string{"mail", "list", "--help"}, "mail-list", true},
		{"mail --help", []string{"mail", "--help"}, "mail", true},
		{"calendar event --help", []string{"calendar", "event", "--help"}, "calendar-event", true},
		{"calendar --help", []string{"calendar", "--help"}, "calendar", true},
		{"calendar availability --help", []string{"calendar", "availability", "--help"}, "calendar-availability", true},
		{"config --help", []string{"config", "--help"}, "config", true},
		{"config alias --help", []string{"config", "alias", "--help"}, "config-alias", true},
		{"mail draft --help", []string{"mail", "draft", "--help"}, "mail-draft", true},
		{"mail folder --help", []string{"mail", "folder", "--help"}, "mail-folder", true},
		{"--help (root)", []string{"--help"}, "", true},
		{"mail list --sender foo --help stops parsing", []string{"mail", "list", "--sender", "foo", "--help"}, "mail-list", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := parseFlags(tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if f.command != tt.command {
				t.Errorf("expected command %q, got %q", tt.command, f.command)
			}
			if f.help != tt.help {
				t.Errorf("expected help=%v, got %v", tt.help, f.help)
			}
		})
	}
}

func TestParseFlags_HelpStopsParsing(t *testing.T) {
	// --help should stop flag parsing; sender should not be fully parsed after --help
	f, err := parseFlags([]string{"mail", "list", "--sender", "foo", "--help"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f.help {
		t.Error("expected help=true")
	}
	// sender was parsed before --help, that's fine
	if f.sender != "foo" {
		t.Errorf("expected sender 'foo', got %q", f.sender)
	}
}

func TestParseFlags_HelpCommand(t *testing.T) {
	f, err := parseFlags([]string{"help"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.command != "help" {
		t.Errorf("expected command 'help', got %q", f.command)
	}
}

func TestParseFlags_HelpCommandWithJSON(t *testing.T) {
	f, err := parseFlags([]string{"help", "-o", "json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.command != "help" {
		t.Errorf("expected command 'help', got %q", f.command)
	}
	if f.output != "json" {
		t.Errorf("expected output 'json', got %q", f.output)
	}
}

func TestHandleHelp_KnownCommand(t *testing.T) {
	// Redirect stdout to discard output
	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = old }()

	f := flags{command: "mail-list", help: true}
	code := handleHelp(f, "test")
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestHandleHelp_KnownGroup(t *testing.T) {
	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = old }()

	f := flags{command: "mail", help: true}
	code := handleHelp(f, "test")
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestHandleHelp_RootHelp(t *testing.T) {
	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = old }()

	f := flags{command: "", help: true}
	code := handleHelp(f, "test")
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestHandleHelp_JSONOutput(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f := flags{command: "help", output: "json"}
	code := handleHelp(f, "1.0.0")

	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	var result map[string]any
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("invalid JSON output: %v\nOutput: %s", err, string(out))
	}

	if result["name"] != "exchange" {
		t.Errorf("expected name 'exchange', got %v", result["name"])
	}
	if result["version"] != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %v", result["version"])
	}

	cmds, ok := result["commands"].([]any)
	if !ok {
		t.Fatal("expected commands array in JSON output")
	}
	if len(cmds) != 23 {
		t.Errorf("expected 23 commands, got %d", len(cmds))
	}
}
