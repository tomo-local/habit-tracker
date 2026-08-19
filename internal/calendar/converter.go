package calendar

import (
	"fmt"
	"log"
	"time"

	gcal "google.golang.org/api/calendar/v3"
)

func toCalendarEntries(items []*gcal.CalendarListEntry) []*CalendarEntry {
	entries := make([]*CalendarEntry, 0, len(items))
	for _, item := range items {
		entries = append(entries, &CalendarEntry{
			ID:   item.Id,
			Name: item.Summary,
		})
	}
	return entries
}

func toEvents(items []*gcal.Event) []*Event {
	events := make([]*Event, 0, len(items))
	for _, item := range items {
		e, err := toEvent(item)
		if err != nil {
			log.Printf("calendar: skipping malformed event %q: %v", item.Id, err)
			continue
		}
		if e != nil {
			events = append(events, e)
		}
	}
	return events
}

// toEvent converts a single Google Calendar event.
// A cancelled event (deleted instance of a recurring event) legitimately lacks
// most fields, so it is skipped silently (nil, nil). Anything else that fails
// to parse is treated as malformed data and reported via error, rather than
// being swallowed the same way as an intentionally-absent event.
func toEvent(item *gcal.Event) (*Event, error) {
	if item.Status == "cancelled" {
		return nil, nil
	}
	if item.Start == nil {
		return nil, fmt.Errorf("missing start time")
	}

	start, err := parseEventDateTime(item.Start)
	if err != nil {
		return nil, fmt.Errorf("parse start time: %w", err)
	}

	var end *time.Time
	if item.End != nil {
		t, err := parseEventDateTime(item.End)
		if err != nil {
			return nil, fmt.Errorf("parse end time: %w", err)
		}
		end = &t
	}

	return &Event{
		Title: item.Summary,
		Start: start,
		End:   end,
	}, nil
}

func parseEventDateTime(dt *gcal.EventDateTime) (time.Time, error) {
	switch {
	case dt.DateTime != "":
		return time.Parse(time.RFC3339, dt.DateTime)
	case dt.Date != "":
		return time.Parse("2006-01-02", dt.Date)
	default:
		return time.Time{}, fmt.Errorf("neither dateTime nor date is set")
	}
}
