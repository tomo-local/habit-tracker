package calendar

import (
	"context"
	"time"

	"golang.org/x/oauth2"
	gcal "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Calendar interface {
	ListCalendars() ([]*CalendarEntry, error)
	AddEvent(calendarID, title string, duration time.Duration) error
	GetEvents(calendarID string, start, end time.Time) ([]*Event, error)
}

func New(ctx context.Context, ts oauth2.TokenSource) (Calendar, error) {
	svc, err := gcal.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, err
	}
	return &client{svc: svc}, nil
}
