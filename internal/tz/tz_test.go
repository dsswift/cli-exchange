package tz

import (
	"testing"
	"time"
)

func TestParseTimespan_Valid(t *testing.T) {
	tests := []struct {
		input string
		value int
		unit  rune
	}{
		{"3h", 3, 'h'},
		{"1d", 1, 'd'},
		{"2w", 2, 'w'},
		{"24h", 24, 'h'},
		{"14d", 14, 'd'},
	}
	for _, tt := range tests {
		ts, err := ParseTimespan(tt.input)
		if err != nil {
			t.Errorf("ParseTimespan(%q): unexpected error: %v", tt.input, err)
			continue
		}
		if ts.Value != tt.value || ts.Unit != tt.unit {
			t.Errorf("ParseTimespan(%q) = {%d, %c}, want {%d, %c}", tt.input, ts.Value, ts.Unit, tt.value, tt.unit)
		}
	}
}

func TestParseTimespan_Invalid(t *testing.T) {
	for _, input := range []string{"", "h", "0d", "-1h", "3x", "abc", "3.5d"} {
		_, err := ParseTimespan(input)
		if err == nil {
			t.Errorf("ParseTimespan(%q): expected error", input)
		}
	}
}

func TestTimespan_AddTo_Hours(t *testing.T) {
	base := time.Date(2026, 3, 13, 8, 0, 0, 0, time.UTC)
	ts := Timespan{Value: 3, Unit: 'h'}
	result := ts.AddTo(base)
	expected := time.Date(2026, 3, 13, 11, 0, 0, 0, time.UTC)
	if !result.Equal(expected) {
		t.Errorf("AddTo hours: got %v, want %v", result, expected)
	}
}

func TestTimespan_AddTo_Days(t *testing.T) {
	base := time.Date(2026, 3, 13, 8, 0, 0, 0, time.UTC)
	ts := Timespan{Value: 3, Unit: 'd'}
	result := ts.AddTo(base)
	expected := time.Date(2026, 3, 16, 8, 0, 0, 0, time.UTC)
	if !result.Equal(expected) {
		t.Errorf("AddTo days: got %v, want %v", result, expected)
	}
}

func TestTimespan_AddTo_Weeks(t *testing.T) {
	base := time.Date(2026, 3, 13, 8, 0, 0, 0, time.UTC)
	ts := Timespan{Value: 1, Unit: 'w'}
	result := ts.AddTo(base)
	expected := time.Date(2026, 3, 20, 8, 0, 0, 0, time.UTC)
	if !result.Equal(expected) {
		t.Errorf("AddTo weeks: got %v, want %v", result, expected)
	}
}

func TestTodayAt(t *testing.T) {
	svc, err := NewService("UTC")
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	result := svc.TodayAt(9, 30)
	now := time.Now().UTC()
	if result.Year() != now.Year() || result.Month() != now.Month() || result.Day() != now.Day() {
		t.Errorf("TodayAt: wrong date: got %v", result)
	}
	if result.Hour() != 9 || result.Minute() != 30 {
		t.Errorf("TodayAt: got %02d:%02d, want 09:30", result.Hour(), result.Minute())
	}
}

func TestNow(t *testing.T) {
	svc, err := NewService("UTC")
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	result := svc.Now()
	if result.Location().String() != "UTC" {
		t.Errorf("Now: expected UTC location, got %s", result.Location())
	}
}
