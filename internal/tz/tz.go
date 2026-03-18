package tz

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	displayFormat  = "2006-01-02 15:04 MST"
	dateFormat     = "2006-01-02"
	datetimeFormat = "2006-01-02T15:04:05"
)

type Service struct {
	loc          *time.Location
	TimezoneName string
}

func NewService(tzName string) (*Service, error) {
	if tzName == "" {
		tzName = "UTC"
	}
	loc, err := time.LoadLocation(tzName)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone %q: %w", tzName, err)
	}
	return &Service{loc: loc, TimezoneName: tzName}, nil
}

func (s *Service) Today() time.Time {
	now := time.Now().In(s.loc)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.loc)
}

func (s *Service) FormatDatetime(t time.Time) string {
	return t.In(s.loc).Format(displayFormat)
}

func (s *Service) ParseDate(input string) (time.Time, error) {
	t, err := time.ParseInLocation(dateFormat, input, s.loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q, expected YYYY-MM-DD: %w", input, err)
	}
	return t, nil
}

func (s *Service) ParseDatetime(input string) (time.Time, error) {
	t, err := time.ParseInLocation(datetimeFormat, input, s.loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid datetime %q, expected YYYY-MM-DDTHH:MM:SS: %w", input, err)
	}
	return t, nil
}

func (s *Service) ParseGraphDatetime(dt, tzName string) (time.Time, error) {
	loc := s.loc
	if tzName != "" {
		var err error
		loc, err = time.LoadLocation(tzName)
		if err != nil {
			loc = s.loc
		}
	}

	// Graph API returns datetime in various formats; try multiple
	formats := []string{
		"2006-01-02T15:04:05.9999999",
		"2006-01-02T15:04:05.0000000",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
	}

	cleaned := strings.TrimSuffix(dt, "Z")
	for _, f := range formats {
		if t, err := time.ParseInLocation(f, cleaned, loc); err == nil {
			return t.In(s.loc), nil
		}
	}

	return time.Time{}, fmt.Errorf("cannot parse Graph datetime %q", dt)
}

func (s *Service) FormatGraphDatetime(dt, tzName string) string {
	t, err := s.ParseGraphDatetime(dt, tzName)
	if err != nil {
		return dt
	}
	return s.FormatDatetime(t)
}

// Timespan represents a duration like "3h", "2d", or "1w".
type Timespan struct {
	Value int
	Unit  rune // 'h', 'd', 'w'
}

// ParseTimespan parses strings like "3h", "2d", "1w" into a Timespan.
func ParseTimespan(input string) (Timespan, error) {
	input = strings.TrimSpace(input)
	if len(input) < 2 {
		return Timespan{}, fmt.Errorf("invalid timespan %q", input)
	}
	unit := rune(input[len(input)-1])
	if unit != 'h' && unit != 'd' && unit != 'w' {
		return Timespan{}, fmt.Errorf("invalid timespan unit %q, expected h/d/w", string(unit))
	}
	n, err := strconv.Atoi(input[:len(input)-1])
	if err != nil || n <= 0 {
		return Timespan{}, fmt.Errorf("invalid timespan value %q", input[:len(input)-1])
	}
	return Timespan{Value: n, Unit: unit}, nil
}

// AddTo applies the timespan to the given time. Uses AddDate for days/weeks (DST-safe).
func (ts Timespan) AddTo(t time.Time) time.Time {
	switch ts.Unit {
	case 'h':
		return t.Add(time.Duration(ts.Value) * time.Hour)
	case 'd':
		return t.AddDate(0, 0, ts.Value)
	case 'w':
		return t.AddDate(0, 0, ts.Value*7)
	default:
		return t
	}
}

// Now returns the current time in the service's timezone.
func (s *Service) Now() time.Time {
	return time.Now().In(s.loc)
}

// TodayAt returns today's date at the specified hour and minute in the service's timezone.
func (s *Service) TodayAt(hour, minute int) time.Time {
	now := time.Now().In(s.loc)
	return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, s.loc)
}
