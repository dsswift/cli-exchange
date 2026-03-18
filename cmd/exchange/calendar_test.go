package main

import (
	"testing"

	"github.com/dsswift/cli-exchange/internal/config"
	"github.com/dsswift/cli-exchange/internal/graph"
	"github.com/dsswift/cli-exchange/internal/tz"
)

func TestFilterSchedules_RemovesWeekendItems(t *testing.T) {
	tzSvc, _ := tz.NewService("UTC")
	bh := &config.BusinessHours{Start: "08:00", End: "17:00"}

	schedules := []graph.ScheduleInformation{
		{
			ScheduleID: "user@example.com",
			ScheduleItems: []graph.ScheduleItem{
				{
					Status: "busy",
					Start:  &graph.DateTimeTimeZone{DateTime: "2026-03-14T09:00:00", TimeZone: "UTC"}, // Saturday
					End:    &graph.DateTimeTimeZone{DateTime: "2026-03-14T10:00:00", TimeZone: "UTC"},
				},
				{
					Status: "busy",
					Start:  &graph.DateTimeTimeZone{DateTime: "2026-03-16T09:00:00", TimeZone: "UTC"}, // Monday
					End:    &graph.DateTimeTimeZone{DateTime: "2026-03-16T10:00:00", TimeZone: "UTC"},
				},
			},
		},
	}

	result := filterSchedules(schedules, bh, false, tzSvc, 30)
	if len(result[0].ScheduleItems) != 1 {
		t.Fatalf("expected 1 item after weekend filter, got %d", len(result[0].ScheduleItems))
	}
	if result[0].ScheduleItems[0].Start.DateTime != "2026-03-16T09:00:00" {
		t.Errorf("expected Monday item to remain, got %s", result[0].ScheduleItems[0].Start.DateTime)
	}
}

func TestFilterSchedules_KeepsWeekendsWhenIncluded(t *testing.T) {
	tzSvc, _ := tz.NewService("UTC")
	bh := &config.BusinessHours{Start: "08:00", End: "17:00"}

	schedules := []graph.ScheduleInformation{
		{
			ScheduleID: "user@example.com",
			ScheduleItems: []graph.ScheduleItem{
				{
					Status: "busy",
					Start:  &graph.DateTimeTimeZone{DateTime: "2026-03-14T09:00:00", TimeZone: "UTC"},
					End:    &graph.DateTimeTimeZone{DateTime: "2026-03-14T10:00:00", TimeZone: "UTC"},
				},
			},
		},
	}

	result := filterSchedules(schedules, bh, true, tzSvc, 30)
	if len(result[0].ScheduleItems) != 1 {
		t.Fatalf("expected 1 item when weekends included, got %d", len(result[0].ScheduleItems))
	}
}

func TestFilterSchedules_RemovesOutsideBusinessHours(t *testing.T) {
	tzSvc, _ := tz.NewService("UTC")
	bh := &config.BusinessHours{Start: "09:00", End: "17:00"}

	schedules := []graph.ScheduleInformation{
		{
			ScheduleID: "user@example.com",
			ScheduleItems: []graph.ScheduleItem{
				{
					Status: "busy",
					Start:  &graph.DateTimeTimeZone{DateTime: "2026-03-16T07:00:00", TimeZone: "UTC"}, // Before business hours
					End:    &graph.DateTimeTimeZone{DateTime: "2026-03-16T08:00:00", TimeZone: "UTC"},
				},
				{
					Status: "busy",
					Start:  &graph.DateTimeTimeZone{DateTime: "2026-03-16T10:00:00", TimeZone: "UTC"}, // Within business hours
					End:    &graph.DateTimeTimeZone{DateTime: "2026-03-16T11:00:00", TimeZone: "UTC"},
				},
				{
					Status: "busy",
					Start:  &graph.DateTimeTimeZone{DateTime: "2026-03-16T18:00:00", TimeZone: "UTC"}, // After business hours
					End:    &graph.DateTimeTimeZone{DateTime: "2026-03-16T19:00:00", TimeZone: "UTC"},
				},
			},
		},
	}

	result := filterSchedules(schedules, bh, true, tzSvc, 30)
	if len(result[0].ScheduleItems) != 1 {
		t.Fatalf("expected 1 item within business hours, got %d", len(result[0].ScheduleItems))
	}
	if result[0].ScheduleItems[0].Start.DateTime != "2026-03-16T10:00:00" {
		t.Errorf("expected 10:00 item, got %s", result[0].ScheduleItems[0].Start.DateTime)
	}
}

func TestFilterSchedules_KeepsOverlappingItems(t *testing.T) {
	tzSvc, _ := tz.NewService("UTC")
	bh := &config.BusinessHours{Start: "09:00", End: "17:00"}

	schedules := []graph.ScheduleInformation{
		{
			ScheduleID: "user@example.com",
			ScheduleItems: []graph.ScheduleItem{
				{
					Status: "busy",
					Start:  &graph.DateTimeTimeZone{DateTime: "2026-03-16T08:00:00", TimeZone: "UTC"}, // Starts before, ends during
					End:    &graph.DateTimeTimeZone{DateTime: "2026-03-16T10:00:00", TimeZone: "UTC"},
				},
				{
					Status: "busy",
					Start:  &graph.DateTimeTimeZone{DateTime: "2026-03-16T16:00:00", TimeZone: "UTC"}, // Starts during, ends after
					End:    &graph.DateTimeTimeZone{DateTime: "2026-03-16T18:00:00", TimeZone: "UTC"},
				},
			},
		},
	}

	result := filterSchedules(schedules, bh, true, tzSvc, 30)
	if len(result[0].ScheduleItems) != 2 {
		t.Fatalf("expected 2 overlapping items to be kept, got %d", len(result[0].ScheduleItems))
	}
}

func TestResolveBusinessHours_Default(t *testing.T) {
	f := flags{}
	cfg := &config.ExchangeConfig{}
	bh := resolveBusinessHours(f, cfg)
	if bh.Start != "08:00" || bh.End != "17:00" {
		t.Errorf("expected default 08:00-17:00, got %s-%s", bh.Start, bh.End)
	}
}

func TestResolveBusinessHours_FromConfig(t *testing.T) {
	f := flags{}
	cfg := &config.ExchangeConfig{
		BusinessHours: &config.BusinessHours{Start: "07:00", End: "18:00"},
	}
	bh := resolveBusinessHours(f, cfg)
	if bh.Start != "07:00" || bh.End != "18:00" {
		t.Errorf("expected config 07:00-18:00, got %s-%s", bh.Start, bh.End)
	}
}

func TestResolveBusinessHours_FlagOverridesConfig(t *testing.T) {
	f := flags{businessHours: "06:00-20:00"}
	cfg := &config.ExchangeConfig{
		BusinessHours: &config.BusinessHours{Start: "07:00", End: "18:00"},
	}
	bh := resolveBusinessHours(f, cfg)
	if bh.Start != "06:00" || bh.End != "20:00" {
		t.Errorf("expected flag 06:00-20:00, got %s-%s", bh.Start, bh.End)
	}
}

func TestIsWeekend(t *testing.T) {
	tests := []struct {
		date     string
		weekend  bool
	}{
		{"2026-03-13", false}, // Friday
		{"2026-03-14", true},  // Saturday
		{"2026-03-15", true},  // Sunday
		{"2026-03-16", false}, // Monday
	}
	for _, tt := range tests {
		tm, _ := tz.NewService("UTC")
		parsed, _ := tm.ParseDate(tt.date)
		if isWeekend(parsed) != tt.weekend {
			t.Errorf("isWeekend(%s) = %v, want %v", tt.date, !tt.weekend, tt.weekend)
		}
	}
}
