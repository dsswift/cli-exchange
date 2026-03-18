package graph

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

var calendarSelectFields = "id,name,isDefaultCalendar,owner,color"
var eventSelectFields = "id,subject,start,end,location,isAllDay,showAs,organizer,attendees,body,recurrence,webLink,onlineMeetingUrl"

func (c *GraphClient) ListCalendars() ([]Calendar, error) {
	params := url.Values{}
	params.Set("$select", calendarSelectFields)

	data, err := c.do("GET", "/me/calendars", params, nil)
	if err != nil {
		return nil, err
	}

	var resp listResponse[Calendar]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing calendars: %w", err)
	}
	return resp.Value, nil
}

type ListEventsOptions struct {
	CalendarID string
	StartDate  *time.Time
	EndDate    *time.Time
	Limit      int
}

func (c *GraphClient) ListEvents(opts ListEventsOptions) ([]Event, error) {
	path := "/me/calendar/events"
	if opts.CalendarID != "" {
		path = fmt.Sprintf("/me/calendars/%s/events", opts.CalendarID)
	}

	params := url.Values{}
	params.Set("$select", eventSelectFields)
	params.Set("$orderby", "start/dateTime asc")

	limit := opts.Limit
	if limit <= 0 {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}
	params.Set("$top", fmt.Sprintf("%d", limit))

	var filters []string
	if opts.StartDate != nil {
		filters = append(filters, fmt.Sprintf("start/dateTime ge '%s'", opts.StartDate.Format(time.RFC3339)))
	}
	if opts.EndDate != nil {
		filters = append(filters, fmt.Sprintf("end/dateTime le '%s'", opts.EndDate.Format(time.RFC3339)))
	}

	if len(filters) > 0 {
		filter := filters[0]
		for _, f := range filters[1:] {
			filter += " and " + f
		}
		params.Set("$filter", filter)
	}

	data, err := c.do("GET", path, params, nil)
	if err != nil {
		return nil, err
	}

	var resp listResponse[Event]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing events: %w", err)
	}
	return resp.Value, nil
}

func (c *GraphClient) GetEvent(id string) (*Event, error) {
	params := url.Values{}
	params.Set("$select", eventSelectFields)

	data, err := c.do("GET", fmt.Sprintf("/me/events/%s", id), params, nil)
	if err != nil {
		return nil, err
	}

	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("parsing event: %w", err)
	}
	return &event, nil
}

type GetScheduleOptions struct {
	Emails          []string
	StartTime       time.Time
	EndTime         time.Time
	Timezone        string
	IntervalMinutes int
}

func (c *GraphClient) GetSchedule(opts GetScheduleOptions) ([]ScheduleInformation, error) {
	interval := opts.IntervalMinutes
	if interval <= 0 {
		interval = 30
	}

	body := map[string]any{
		"schedules":                opts.Emails,
		"startTime":               map[string]string{"dateTime": opts.StartTime.Format("2006-01-02T15:04:05"), "timeZone": opts.Timezone},
		"endTime":                 map[string]string{"dateTime": opts.EndTime.Format("2006-01-02T15:04:05"), "timeZone": opts.Timezone},
		"availabilityViewInterval": fmt.Sprintf("%d", interval),
	}

	data, err := c.do("POST", "/me/calendar/getSchedule", nil, body)
	if err != nil {
		return nil, err
	}

	var resp listResponse[ScheduleInformation]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing schedule: %w", err)
	}
	return resp.Value, nil
}
