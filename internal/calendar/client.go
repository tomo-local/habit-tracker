package calendar

import (
	"time"

	gcal "google.golang.org/api/calendar/v3"
)

type client struct {
	svc *gcal.Service
}

func (c *client) ListCalendars() ([]*CalendarEntry, error) {
	list, err := c.svc.CalendarList.List().Do()
	if err != nil {
		return nil, err
	}
	return toCalendarEntries(list.Items), nil
}

func (c *client) AddEvent(calendarID, title string, duration time.Duration) error {
	end := time.Now()
	start := end.Add(-duration)
	event := &gcal.Event{
		Summary: title,
		Start:   &gcal.EventDateTime{DateTime: start.Format(time.RFC3339)},
		End:     &gcal.EventDateTime{DateTime: end.Format(time.RFC3339)},
	}
	_, err := c.svc.Events.Insert(calendarID, event).Do()
	return err
}

func (c *client) GetEvents(calendarID string, start, end time.Time) ([]*Event, error) {
	resp, err := c.svc.Events.List(calendarID).
		TimeMin(start.Format(time.RFC3339)).
		TimeMax(end.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		Do()
	if err != nil {
		return nil, err
	}
	return toEvents(resp.Items), nil
}
