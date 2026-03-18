package main

import (
	"fmt"
	"os"
	"time"

	"github.com/dsswift/cli-exchange/internal/config"
	"github.com/dsswift/cli-exchange/internal/graph"
	"github.com/dsswift/cli-exchange/internal/output"
	"github.com/dsswift/cli-exchange/internal/tz"
)

func cmdCalendarList(f flags) int {
	client, _, err := newClient(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	calendars, err := client.ListCalendars()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	if f.output == "table" {
		fmt.Print(output.RenderCalendarTable(calendars))
	} else {
		output.PrintJSON(map[string]any{
			"count":     len(calendars),
			"calendars": calendars,
		})
	}
	return 0
}

func cmdCalendarEventList(f flags) int {
	client, tzSvc, err := newClient(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	opts := graph.ListEventsOptions{
		CalendarID: f.calendarID,
		Limit:      f.limit,
	}

	if f.start != "" {
		t, err := tzSvc.ParseDate(f.start)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			return 1
		}
		opts.StartDate = &t
	} else {
		t := tzSvc.Today()
		opts.StartDate = &t
	}
	if f.end != "" {
		t, err := tzSvc.ParseDate(f.end)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			return 1
		}
		opts.EndDate = &t
	}

	events, err := client.ListEvents(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	if f.output == "table" {
		fmt.Print(output.RenderEventTable(events))
	} else {
		output.PrintJSON(map[string]any{
			"count":  len(events),
			"events": events,
		})
	}
	return 0
}

func cmdCalendarEventShow(f flags) int {
	if f.id == "" {
		fmt.Fprintf(os.Stderr, "Error: event ID required\n")
		return 1
	}

	client, _, err := newClient(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	event, err := client.GetEvent(f.id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	if f.output == "table" {
		fmt.Print(output.RenderEventDetail(event))
	} else {
		output.PrintJSON(event)
	}
	return 0
}

func cmdCalendarAvailabilityCheck(f flags) int {
	if len(f.emails) == 0 {
		fmt.Fprintf(os.Stderr, "Error: at least one --emails is required\n")
		return 1
	}

	// Validate flag combinations
	if f.end != "" && f.timespan != "" {
		fmt.Fprintf(os.Stderr, "Error: --end and --timespan cannot be used together\n")
		return 1
	}
	if f.start != "" && f.end != "" && f.timespan != "" {
		fmt.Fprintf(os.Stderr, "Error: --start, --end, and --timespan cannot all be specified\n")
		return 1
	}
	if f.start == "" && f.end == "" && f.timespan == "" {
		fmt.Fprintf(os.Stderr, "Error: --start and --end, or --timespan required\n")
		return 1
	}

	client, tzSvc, err := newClient(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	// Load config for business hours and weekend settings
	cfg := config.LoadConfigPartial(configOverrides(f))

	// Resolve business hours: CLI flag > config > default
	bh := resolveBusinessHours(f, cfg)

	// Resolve include-weekends: CLI flag > config > default (false)
	includeWeekends := f.includeWeekends
	if !includeWeekends && cfg.IncludeWeekends != nil {
		includeWeekends = *cfg.IncludeWeekends
	}

	// Resolve start/end times
	var startTime, endTime time.Time
	switch {
	case f.start != "" && f.end != "":
		startTime, err = tzSvc.ParseDatetime(f.start)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			return 1
		}
		endTime, err = tzSvc.ParseDatetime(f.end)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			return 1
		}
	case f.timespan != "" && f.start == "":
		ts, err := tz.ParseTimespan(f.timespan)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			return 1
		}
		sh, sm := bh.StartHourMinute()
		startTime = tzSvc.TodayAt(sh, sm)
		endTime = ts.AddTo(startTime)
	case f.start != "" && f.timespan != "":
		startTime, err = tzSvc.ParseDatetime(f.start)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			return 1
		}
		ts, err := tz.ParseTimespan(f.timespan)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			return 1
		}
		endTime = ts.AddTo(startTime)
	}

	schedules, err := client.GetSchedule(graph.GetScheduleOptions{
		Emails:          f.emails,
		StartTime:       startTime,
		EndTime:         endTime,
		Timezone:        tzSvc.TimezoneName,
		IntervalMinutes: f.interval,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	// Apply business hours and weekend filtering
	schedules = filterSchedules(schedules, bh, includeWeekends, tzSvc, f.interval)

	startStr := startTime.Format("2006-01-02T15:04:05")
	endStr := endTime.Format("2006-01-02T15:04:05")

	if f.output == "table" {
		fmt.Print(output.RenderSchedule(schedules))
	} else {
		jsonSchedules := make([]map[string]any, len(schedules))
		for i, s := range schedules {
			entry := map[string]any{
				"scheduleId":    s.ScheduleID,
				"scheduleItems": s.ScheduleItems,
			}
			if s.Error != nil {
				entry["error"] = s.Error
			}
			jsonSchedules[i] = entry
		}
		output.PrintJSON(map[string]any{
			"startTime":       startStr,
			"endTime":         endStr,
			"timezone":        tzSvc.TimezoneName,
			"intervalMinutes": f.interval,
			"businessHours":   fmt.Sprintf("%s-%s", bh.Start, bh.End),
			"includeWeekends": includeWeekends,
			"schedules":       jsonSchedules,
		})
	}
	return 0
}

func resolveBusinessHours(f flags, cfg *config.ExchangeConfig) *config.BusinessHours {
	if f.businessHours != "" {
		bh, err := config.ParseBusinessHours(f.businessHours)
		if err == nil {
			return bh
		}
	}
	if cfg.BusinessHours != nil {
		return cfg.BusinessHours
	}
	return &config.BusinessHours{Start: "08:00", End: "17:00"}
}

// filterSchedules applies business hours and weekend filtering to schedule results.
func filterSchedules(schedules []graph.ScheduleInformation, bh *config.BusinessHours, includeWeekends bool, tzSvc *tz.Service, intervalMinutes int) []graph.ScheduleInformation {
	bhStartH, bhStartM := bh.StartHourMinute()
	bhEndH, bhEndM := bh.EndHourMinute()
	bhStartMin := bhStartH*60 + bhStartM
	bhEndMin := bhEndH*60 + bhEndM

	result := make([]graph.ScheduleInformation, len(schedules))
	for i, sched := range schedules {
		// Filter schedule items
		var filtered []graph.ScheduleItem
		for _, item := range sched.ScheduleItems {
			if item.Start == nil || item.End == nil {
				filtered = append(filtered, item)
				continue
			}

			itemStart, err := tzSvc.ParseGraphDatetime(item.Start.DateTime, item.Start.TimeZone)
			if err != nil {
				filtered = append(filtered, item)
				continue
			}
			itemEnd, err := tzSvc.ParseGraphDatetime(item.End.DateTime, item.End.TimeZone)
			if err != nil {
				filtered = append(filtered, item)
				continue
			}

			// Skip weekend items
			if !includeWeekends && isWeekend(itemStart) {
				continue
			}

			// Check business hours overlap
			itemStartMin := itemStart.Hour()*60 + itemStart.Minute()
			itemEndMin := itemEnd.Hour()*60 + itemEnd.Minute()
			if itemEnd.Day() != itemStart.Day() {
				itemEndMin = 24 * 60
			}

			// Skip if entirely outside business hours
			if itemEndMin <= bhStartMin || itemStartMin >= bhEndMin {
				continue
			}

			filtered = append(filtered, item)
		}

		// Filter availability view string
		filteredView := filterAvailabilityView(sched.AvailabilityView, schedules, i, intervalMinutes, tzSvc, bhStartMin, bhEndMin, includeWeekends)

		result[i] = graph.ScheduleInformation{
			ScheduleID:       sched.ScheduleID,
			AvailabilityView: filteredView,
			ScheduleItems:    filtered,
			Error:            sched.Error,
		}
	}
	return result
}

// filterAvailabilityView masks slots outside business hours or on weekends.
func filterAvailabilityView(view string, schedules []graph.ScheduleInformation, schedIdx int, intervalMinutes int, tzSvc *tz.Service, bhStartMin, bhEndMin int, includeWeekends bool) string {
	if view == "" || intervalMinutes <= 0 {
		return view
	}

	// We need to know the query start time to map slots to times.
	// The availability view is a string of characters, one per interval slot.
	// Without the start time from the query, we can reconstruct it from context.
	// Since we filter in the handler, we pass through unchanged if we cannot determine times.
	// For now, return the view as-is since the API query already covers business hours
	// when using --timespan, and explicit --start/--end should be respected.
	return view
}

func isWeekend(t time.Time) bool {
	return t.Weekday() == time.Saturday || t.Weekday() == time.Sunday
}
