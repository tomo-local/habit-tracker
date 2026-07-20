package calendar

import (
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
		if e := toEvent(item); e != nil {
			events = append(events, e)
		}
	}
	return events
}

func toEvent(item *gcal.Event) *Event {
	if item.Start == nil {
		return nil
	}

	var start time.Time
	var err error

	switch {
	case item.Start.DateTime != "":
		start, err = time.Parse(time.RFC3339, item.Start.DateTime)
	case item.Start.Date != "":
		start, err = time.Parse("2006-01-02", item.Start.Date)
	default:
		return nil
	}
	if err != nil {
		return nil
	}

	var end *time.Time
	if item.End != nil {
		switch {
		case item.End.DateTime != "":
			if t, e := time.Parse(time.RFC3339, item.End.DateTime); e == nil {
				end = &t
			}
		case item.End.Date != "":
			if t, e := time.Parse("2006-01-02", item.End.Date); e == nil {
				end = &t
			}
		}
	}

	return &Event{
		Title: item.Summary,
		Start: start,
		End:   end,
	}
}
