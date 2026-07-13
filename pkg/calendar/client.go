package calendar

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2"
	gcal "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)


// Client abstracts the Google Calendar operations needed by habit-tracker.
type Client interface {
	ListCalendars() ([]string, error)
	ListEvents(calName string, since time.Time) ([]Occurrence, error)
	AddEvent(calName, summary string, start, end time.Time) error
}

// NewClient は認証済みの TokenSource から Google Calendar クライアントを生成する。
func NewClient(ctx context.Context, ts oauth2.TokenSource) (Client, error) {
	svc, err := gcal.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, err
	}
	return &googleClient{svc}, nil
}

type googleClient struct {
	svc *gcal.Service
}

func (c *googleClient) ListCalendars() ([]string, error) {
	var names []string
	err := c.svc.CalendarList.List().Pages(nil, func(page *gcal.CalendarList) error {
		for _, item := range page.Items {
			names = append(names, item.Summary)
		}
		return nil
	})
	return names, err
}

func (c *googleClient) ListEvents(calName string, since time.Time) ([]Occurrence, error) {
	calID, err := c.findCalendarID(calName)
	if err != nil {
		return nil, err
	}

	var items []*gcal.Event
	err = c.svc.Events.List(calID).
		SingleEvents(true).
		TimeMin(since.Format(time.RFC3339)).
		MaxResults(2500).
		Pages(nil, func(page *gcal.Events) error {
			items = append(items, page.Items...)
			return nil
		})
	if err != nil {
		return nil, err
	}
	return ToOccurrences(items), nil
}

func (c *googleClient) AddEvent(calName, summary string, start, end time.Time) error {
	calID, err := c.findCalendarID(calName)
	if err != nil {
		return err
	}
	_, err = c.svc.Events.Insert(calID, &gcal.Event{
		Summary: summary,
		Start:   &gcal.EventDateTime{DateTime: start.Format(time.RFC3339)},
		End:     &gcal.EventDateTime{DateTime: end.Format(time.RFC3339)},
	}).Do()
	if err != nil {
		return fmt.Errorf("イベント作成に失敗: %w", err)
	}
	return nil
}

func (c *googleClient) findCalendarID(name string) (string, error) {
	var id string
	err := c.svc.CalendarList.List().Pages(nil, func(page *gcal.CalendarList) error {
		for _, item := range page.Items {
			if item.Summary == name || item.SummaryOverride == name {
				id = item.Id
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if id == "" {
		return "", fmt.Errorf("カレンダー %q が見つかりません (`habit-tracker list` で確認できます)", name)
	}
	return id, nil
}
