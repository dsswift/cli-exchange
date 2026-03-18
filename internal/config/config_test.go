package config

import (
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
