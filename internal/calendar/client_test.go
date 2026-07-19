package calendar

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	gcal "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func newTestClient(t *testing.T, mux *http.ServeMux) *client {
	t.Helper()
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	svc, err := gcal.NewService(context.Background(),
		option.WithoutAuthentication(),
		option.WithEndpoint(ts.URL+"/"),
	)
	if err != nil {
		t.Fatalf("gcal.NewService: %v", err)
	}
	return &client{svc: svc}
}

func Test_Client(t *testing.T) {
	t.Run("ListCalendars", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/users/me/calendarList", func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(gcal.CalendarList{
				Items: []*gcal.CalendarListEntry{
					{Id: "cal1", Summary: "習慣"},
				},
			})
		})

		entries, err := newTestClient(t, mux).ListCalendars()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) != 1 {
			t.Fatalf("got %d entries, want 1", len(entries))
		}
		if entries[0].ID != "cal1" || entries[0].Name != "習慣" {
			t.Errorf("got %+v, want {ID:cal1 Name:習慣}", entries[0])
		}
	})

	t.Run("AddEvent", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/calendars/cal1/events", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			json.NewEncoder(w).Encode(gcal.Event{Id: "ev1"})
		})

		err := newTestClient(t, mux).AddEvent("cal1", "筋トレ", 30*time.Minute)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("GetEvents", func(t *testing.T) {
		tests := []struct {
			name       string
			items      []*gcal.Event
			wantCount  int
			wantTitle  string
		}{
			{
				name: "returns converted events",
				items: []*gcal.Event{
					{Summary: "筋トレ", Start: &gcal.EventDateTime{DateTime: "2026-07-20T10:00:00+09:00"}},
				},
				wantCount: 1,
				wantTitle: "筋トレ",
			},
			{
				name:      "empty list",
				items:     []*gcal.Event{},
				wantCount: 0,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				mux := http.NewServeMux()
				mux.HandleFunc("/calendars/cal1/events", func(w http.ResponseWriter, r *http.Request) {
					json.NewEncoder(w).Encode(gcal.Events{Items: tt.items})
				})

				events, err := newTestClient(t, mux).GetEvents("cal1", time.Now().AddDate(0, 0, -7), time.Now())
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(events) != tt.wantCount {
					t.Fatalf("got %d events, want %d", len(events), tt.wantCount)
				}
				if tt.wantCount > 0 && events[0].Title != tt.wantTitle {
					t.Errorf("got title %q, want %q", events[0].Title, tt.wantTitle)
				}
			})
		}
	})
}
