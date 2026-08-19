package calendar

import (
	"time"

	gcal "google.golang.org/api/calendar/v3"
)

type Client interface {
	ListCalendars() ([]*CalendarEntry, error)
	CreateCalendar(summary string) (*CalendarEntry, error)
	AddEvent(calendarID, title, description string, duration time.Duration) error
	GetEvents(calendarID string, start, end time.Time) ([]*Event, error)
}

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

func (c *client) CreateCalendar(summary string) (*CalendarEntry, error) {
	created, err := c.svc.Calendars.Insert(&gcal.Calendar{Summary: summary}).Do()
	if err != nil {
		return nil, err
	}
	return &CalendarEntry{ID: created.Id, Name: created.Summary}, nil
}

func (c *client) AddEvent(calendarID, title, description string, duration time.Duration) error {
	end := time.Now()
	start := end.Add(-duration)
	event := &gcal.Event{
		Summary:     title,
		Description: description,
		Start:       &gcal.EventDateTime{DateTime: start.Format(time.RFC3339)},
		End:         &gcal.EventDateTime{DateTime: end.Format(time.RFC3339)},
	}
	_, err := c.svc.Events.Insert(calendarID, event).Do()
	return err
}

func (c *client) GetEvents(calendarID string, start, end time.Time) ([]*Event, error) {
	var allItems []*gcal.Event
	req := c.svc.Events.List(calendarID).
		TimeMin(start.Format(time.RFC3339)).
		TimeMax(end.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime")
	for {
		resp, err := req.Do()
		if err != nil {
			return nil, err
		}
		allItems = append(allItems, resp.Items...)
		if resp.NextPageToken == "" {
			break
		}
		req = req.PageToken(resp.NextPageToken)
	}
	return toEvents(allItems), nil
}
