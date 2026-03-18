package output

import (
	"fmt"
	"strings"

	"github.com/dsswift/cli-exchange/internal/graph"
)

func RenderCalendarTable(calendars []graph.Calendar) string {
	headers := []string{"Name", "Default", "Owner"}
	rows := make([][]string, len(calendars))
	for i, c := range calendars {
		owner := ""
		if c.Owner != nil {
			owner = c.Owner.Address
		}
		def := "no"
		if c.IsDefaultCalendar {
			def = "yes"
		}
		rows[i] = []string{c.Name, def, owner}
	}
	return BuildTable(headers, rows)
}

func RenderEventTable(events []graph.Event) string {
	headers := []string{"Subject", "Start", "End", "Location"}
	rows := make([][]string, len(events))
	for i, e := range events {
		start := ""
		if e.Start != nil {
			start = e.Start.DateTime
		}
		end := ""
		if e.End != nil {
			end = e.End.DateTime
		}
		loc := ""
		if e.Location != nil {
			loc = e.Location.DisplayName
		}
		rows[i] = []string{
			truncate(e.Subject, 40),
			start,
			end,
			truncate(loc, 30),
		}
	}
	return BuildTable(headers, rows)
}

func RenderEventDetail(e *graph.Event) string {
	var lines []string

	addField := func(label, value string) {
		if value != "" {
			lines = append(lines, fmt.Sprintf("%-12s %s", label+":", value))
		}
	}

	addField("ID", e.ID)
	addField("Subject", e.Subject)

	if e.Start != nil {
		addField("Start", fmt.Sprintf("%s (%s)", e.Start.DateTime, e.Start.TimeZone))
	}
	if e.End != nil {
		addField("End", fmt.Sprintf("%s (%s)", e.End.DateTime, e.End.TimeZone))
	}

	if e.Location != nil && e.Location.DisplayName != "" {
		addField("Location", e.Location.DisplayName)
	}

	addField("All Day", fmt.Sprintf("%t", e.IsAllDay))
	addField("Show As", e.ShowAs)

	if e.Organizer != nil {
		addField("Organizer", fmt.Sprintf("%s <%s>", e.Organizer.EmailAddress.Name, e.Organizer.EmailAddress.Address))
	}

	if len(e.Attendees) > 0 {
		for j, a := range e.Attendees {
			label := ""
			if j == 0 {
				label = "Attendees"
			}
			response := a.Status.Response
			lines = append(lines, fmt.Sprintf("%-12s %s <%s> [%s, %s]", label+":", a.EmailAddress.Name, a.EmailAddress.Address, a.Type, response))
		}
	}

	if e.OnlineMeetingURL != "" {
		addField("Meeting URL", e.OnlineMeetingURL)
	}

	if e.WebLink != "" {
		addField("Web Link", e.WebLink)
	}

	result := strings.Join(lines, "\n") + "\n"

	if e.Body != nil && e.Body.Content != "" {
		result += "\n" + e.Body.Content
	}

	return result
}

func RenderSchedule(schedules []graph.ScheduleInformation) string {
	var b strings.Builder
	for _, s := range schedules {
		fmt.Fprintf(&b, "Schedule: %s\n", s.ScheduleID)

		if s.Error != nil {
			fmt.Fprintf(&b, "Error:    %s\n", s.Error.Message)
		}

		if len(s.ScheduleItems) > 0 {
			headers := []string{"Status", "Start", "End", "Subject"}
			rows := make([][]string, len(s.ScheduleItems))
			for i, item := range s.ScheduleItems {
				subject := item.Subject
				if item.IsPrivate {
					subject = "[Private]"
				}
				start := ""
				if item.Start != nil {
					start = item.Start.DateTime
				}
				end := ""
				if item.End != nil {
					end = item.End.DateTime
				}
				rows[i] = []string{item.Status, start, end, subject}
			}
			b.WriteString(BuildTable(headers, rows))
		}
		b.WriteString("\n")
	}
	return b.String()
}
